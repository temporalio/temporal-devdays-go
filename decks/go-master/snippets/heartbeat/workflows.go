package heartbeat

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func PackWorkflow(ctx workflow.Context, input ShipmentInput) (int, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		HeartbeatTimeout:    10 * time.Second,
	})
	var n int
	err := workflow.ExecuteActivity(ctx, PackShipment, input).Get(ctx, &n)
	return n, err
}

func ProcessWorkflow(ctx workflow.Context, input ProcessInput) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    time.Second,
		WaitForCancellation: true,
	})
	return workflow.ExecuteActivity(ctx, ProcessItems, input).Get(ctx, nil)
}
