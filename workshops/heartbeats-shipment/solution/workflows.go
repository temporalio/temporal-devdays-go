// ABOUTME: One shipment, packed by one long-running Activity.
package shipment

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ShipmentWorkflow packs one shipment and returns what was packed.
func ShipmentWorkflow(ctx workflow.Context, input ShipmentInput) (ShipmentResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// A ceiling for the whole shipment, which takes about 12 seconds.
		StartToCloseTimeout: 20 * time.Second,
		// TODO 1a start: give up on an attempt when no Heartbeat arrives for 2 seconds
		HeartbeatTimeout: 2 * time.Second,
		// TODO 1a end
		RetryPolicy: &temporal.RetryPolicy{InitialInterval: time.Second, MaximumAttempts: 10},
	})

	var result ShipmentResult
	err := workflow.ExecuteActivity(ctx, PackShipment, input).Get(ctx, &result)
	return result, err
}
