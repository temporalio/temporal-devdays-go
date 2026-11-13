// Package tuning backs 26-worker-tuning.md: the resource-based tuner, every
// Worker option the tuning slides name, and the deadlock detector that stops
// Workflow code which never yields.
package tuning

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

func SubmitOrder(ctx context.Context, orderID string) (string, error) {
	return "submitted " + orderID, nil
}

func OrderWorkflow(ctx workflow.Context, orderID string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})
	var out string
	err := workflow.ExecuteActivity(ctx, SubmitOrder, orderID).Get(ctx, &out)
	return out, err
}

// BusyWorkflow is synchronous CPU work in Workflow code that never blocks on
// anything the SDK schedules, so it never yields back to the Worker.
// Each step depends on the last, so the compiler cannot fold the loop away.
func BusyWorkflow(ctx workflow.Context, iterations int64) (uint64, error) {
	var acc uint64 = 1
	for i := int64(0); i < iterations; i++ {
		acc = acc*6364136223846793005 + 1442695040888963407
	}
	return acc, nil
}
