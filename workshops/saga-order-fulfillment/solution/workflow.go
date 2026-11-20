// ABOUTME: Order fulfillment as a Saga. Three steps forward, two undos backward.
// There is no built-in Saga helper: the compensation stack is a slice of funcs,
// and a defer runs it newest-first on the way out.
package saga

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// compensation is one undo, ready to run. You hold these yourself and run them
// on the way out.
type compensation func(ctx workflow.Context) error

// undoActivityOptions are for the undos, not the forward steps. A forward step
// that keeps failing should give up quickly so the order can unwind. An undo
// that gives up strands the stock or the money, so it keeps retrying, with
// backoff, for up to five minutes.
var undoActivityOptions = workflow.ActivityOptions{
	StartToCloseTimeout:    10 * time.Second,
	ScheduleToCloseTimeout: 5 * time.Minute,
	RetryPolicy: &temporal.RetryPolicy{
		InitialInterval:    500 * time.Millisecond,
		BackoffCoefficient: 2,
		MaximumInterval:    10 * time.Second,
	},
}

// OrderFulfillmentWorkflow reserves the stock, charges the card and books the
// courier. When any step fails, it undoes the steps that ran, newest first, and
// still fails: compensating is not succeeding.
func OrderFulfillmentWorkflow(ctx workflow.Context, order Order) (result OrderResult, err error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: 100 * time.Millisecond,
			MaximumAttempts: 2,
		},
	})

	// Each undo is appended, so the newest is last.
	var compensations []compensation

	// TODO 1 start
	defer func() {
		if err == nil {
			return
		}
		logger := workflow.GetLogger(ctx)
		logger.Error("fulfillment failed, unwinding", "orderId", order.OrderID, "error", err)

		// A cancelled Workflow cancels ctx, and with it every Activity started on
		// it. The undos run on a disconnected context that cancellation never
		// reaches, with the undos' own retry policy.
		undoCtx, _ := workflow.NewDisconnectedContext(ctx)
		undoCtx = workflow.WithActivityOptions(undoCtx, undoActivityOptions)

		// Newest first: walk the slice backwards.
		for i := len(compensations) - 1; i >= 0; i-- {
			// One failing undo must not strand the rest. Log it and keep unwinding.
			// The cost: the Workflow still fails with the original error, so a
			// failed undo shows only here and as a failed Activity in the Event
			// History. In production this is where you raise something a human sees.
			if cerr := compensations[i](undoCtx); cerr != nil {
				logger.Error("compensation failed", "orderId", order.OrderID, "error", cerr)
			}
		}
		// err is left as it is, so the Workflow still fails with the original
		// failure once the unwind is done.
	}()
	// TODO 1 end

	// TODO 2a start
	// Each undo is registered BEFORE the step it undoes. A step can land its
	// effect and still report failure, and then nothing knows how to reverse it.
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, ReleaseInventoryIfReserved, order).Get(ctx, nil)
	})
	if err = workflow.ExecuteActivity(ctx, ReserveInventory, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	// TODO 2a end

	// TODO 2b start
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, RefundPaymentIfCharged, order).Get(ctx, nil)
	})
	if err = workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	// TODO 2b end

	var trackingNumber string
	if err = workflow.ExecuteActivity(ctx, DispatchCourier, order).Get(ctx, &trackingNumber); err != nil {
		return OrderResult{}, err
	}

	return OrderResult{OrderID: order.OrderID, TrackingNumber: trackingNumber}, nil
}
