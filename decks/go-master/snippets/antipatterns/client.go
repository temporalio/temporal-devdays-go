package antipatterns

import (
	"context"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

// The two starts on the business-ID slide.

func StartWithRandomID(ctx context.Context, c client.Client, transfer Transfer) (client.WorkflowRun, error) {
	// part: 17b-patterns-and-anti-patterns.md "The business ID is the Workflow ID" 1
	// ❌ A random ID: a double-tapped "Send" starts two transfers
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        uuid.NewString(),
		TaskQueue: "payments",
	}, TransferWorkflow, transfer)
	// end part
	return run, err
}

func StartWithTransferID(ctx context.Context, c client.Client, transfer Transfer) (client.WorkflowRun, error) {
	// part: 17b-patterns-and-anti-patterns.md "The business ID is the Workflow ID" 2
	// ✅ The transfer's own ID: the second start gets the first one back
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        "transfer-" + transfer.TransferID,
		TaskQueue: "payments",
	}, TransferWorkflow, transfer)
	// end part
	return run, err
}
