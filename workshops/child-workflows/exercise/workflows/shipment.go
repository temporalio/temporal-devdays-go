// ABOUTME: Challenge 1, the child: ships one order and returns its tracking number.
package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// ShipmentWorkflow ships one order and returns its tracking number.
func ShipmentWorkflow(ctx workflow.Context, orderID string) (string, error) {
	// No Activities in this Workshop: every child only sleeps on a durable timer,
	// so the Temporal UI shows the parent-child structure and nothing else.
	if err := workflow.Sleep(ctx, 2*time.Second); err != nil {
		return "", err
	}
	return "TRACK-" + orderID, nil
}
