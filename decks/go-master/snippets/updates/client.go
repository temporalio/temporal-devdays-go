package updates

import (
	"context"

	"go.temporal.io/sdk/client"
)

// WithdrawFifty is the Client code on the comparison slide in 11-updates.md.
func WithdrawFifty(ctx context.Context, c client.Client, workflowID string) (int, error) {
	var balance int
	// slide: 11-updates.md "Signals, Queries, and Updates"
	handle, err := c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
		WorkflowID: workflowID, UpdateName: WithdrawUpdate, Args: []any{50},
		WaitForStage: client.WorkflowUpdateStageCompleted,
	})
	if err == nil {
		err = handle.Get(ctx, &balance)
	}
	// end slide
	return balance, err
}
