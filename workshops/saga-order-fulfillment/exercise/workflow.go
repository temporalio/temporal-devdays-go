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
	// Nothing unwinds yet. Defer a func that, when err is not nil:
	//
	//   - makes undoCtx, _ := workflow.NewDisconnectedContext(ctx). A cancelled
	//     Workflow cancels ctx, and an undo started on it would never run.
	//   - walks compensations BACKWARDS, from len(compensations)-1 down to 0, and
	//     calls each one with undoCtx. The newest undo goes first.
	//   - logs a failed undo with workflow.GetLogger(ctx) and keeps going, so one
	//     failing undo does not strand the rest.
	//
	// Leave err as it is: the Workflow must still fail once the unwind is done.
	// TODO 1 end

	// TODO 2a start
	// Move the append ABOVE the ReserveInventory call. A step can land its
	// effect and still report failure, and then nothing knows how to reverse it.
	if err = workflow.ExecuteActivity(ctx, ReserveInventory, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, ReleaseInventoryIfReserved, order).Get(ctx, nil)
	})
	// TODO 2a end

	// TODO 2b start
	// Same move, above the ChargePayment call. This is the one that costs real money.
	if err = workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, RefundPaymentIfCharged, order).Get(ctx, nil)
	})
	// TODO 2b end

	var trackingNumber string
	if err = workflow.ExecuteActivity(ctx, DispatchCourier, order).Get(ctx, &trackingNumber); err != nil {
		return OrderResult{}, err
	}

	return OrderResult{OrderID: order.OrderID, TrackingNumber: trackingNumber}, nil
}
