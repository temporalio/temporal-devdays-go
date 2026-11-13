package sagacompensation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// part: 17-saga-04-when-to-register.md "heading: A compensation Activity" 2
// workflow.go
var chargeOptions = workflow.ActivityOptions{
	StartToCloseTimeout:    10 * time.Second,
	ScheduleToCloseTimeout: 2 * time.Minute,
}

var refundOptions = workflow.ActivityOptions{
	StartToCloseTimeout: 10 * time.Second,
}

// end part

// ChargeThenRefundWorkflow is not on the slide: the smallest Workflow that
// charges and then undoes it, so the test can watch the refund outlast an
// outage.
func ChargeThenRefundWorkflow(ctx workflow.Context, order Order) error {
	chargeCtx := workflow.WithActivityOptions(ctx, chargeOptions)
	if err := workflow.ExecuteActivity(chargeCtx, ChargePayment, order).Get(ctx, nil); err != nil {
		return err
	}
	refundCtx := workflow.WithActivityOptions(ctx, refundOptions)
	return workflow.ExecuteActivity(refundCtx, RefundPayment, order).Get(ctx, nil)
}
