package saga

import (
	"errors"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// As in the Saga Workshop: a 10 second Start-To-Close Timeout and two attempts.
var activityOptions = workflow.ActivityOptions{
	StartToCloseTimeout: 10 * time.Second,
	RetryPolicy: &temporal.RetryPolicy{
		InitialInterval: 100 * time.Millisecond,
		MaximumAttempts: 2,
	},
}

// Compensation is one undo, ready to run.
type Compensation func(ctx workflow.Context) error

// undo returns the Compensation that runs one compensation Activity for the
// order.
func undo(activity any, order Order) Compensation {
	return func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, activity, order).Get(ctx, nil)
	}
}

// slide: 17-saga-03-implementation.md "# Implementation: the unwind"
func OrderFulfillmentWorkflow(ctx workflow.Context, order Order) (_ OrderResult, err error) {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	var compensations []Compensation

	defer func() {
		var activityErr *temporal.ActivityError
		if !errors.As(err, &activityErr) {
			return
		}
		logger := workflow.GetLogger(ctx)
		logger.Error("fulfillment failed, unwinding", "order", order.OrderID, "error", err)
		ctx, _ := workflow.NewDisconnectedContext(ctx)
		for i := len(compensations) - 1; i >= 0; i-- {
			if cerr := compensations[i](ctx); cerr != nil {
				logger.Error("compensation failed", "order", order.OrderID, "error", cerr)
			}
		}
	}()

	compensations = append(compensations, undo(ReleaseInventoryIfReserved, order))
	if err := workflow.ExecuteActivity(ctx, ReserveInventory, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	compensations = append(compensations, undo(RefundPaymentIfCharged, order))
	if err := workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	var tracking string
	if err := workflow.ExecuteActivity(ctx, DispatchCourier, order).Get(ctx, &tracking); err != nil {
		return OrderResult{}, err
	}
	return OrderResult{OrderID: order.OrderID, TrackingNumber: tracking}, nil
}

// end slide

// BuggyFulfillmentWorkflow is not on a slide: it backs the speaker note that
// only Activity failures are compensated. A panic in Workflow code fails the
// Workflow Task, not the Workflow, and the deferred unwind sees no error, so
// nothing is refunded while the Task retries.
func BuggyFulfillmentWorkflow(ctx workflow.Context, order Order) (_ OrderResult, err error) {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	var compensations []Compensation
	defer func() {
		var activityErr *temporal.ActivityError
		if !errors.As(err, &activityErr) {
			return
		}
		ctx, _ := workflow.NewDisconnectedContext(ctx)
		for i := len(compensations) - 1; i >= 0; i-- {
			_ = compensations[i](ctx)
		}
	}()
	compensations = append(compensations, undo(ReleaseInventoryIfReserved, order))
	if err := workflow.ExecuteActivity(ctx, ReserveInventory, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	var items []string
	_ = items[order.Quantity] // the bug: index out of range
	return OrderResult{}, nil
}
