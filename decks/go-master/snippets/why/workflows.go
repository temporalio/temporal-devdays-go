package why

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// part: 01-why-temporal.md "The Temporal version" 2
// workflows.go
func TransferWorkflow(ctx workflow.Context, t Transfer) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			NonRetryableErrorTypes: []string{"InsufficientFundsError"},
		},
	})

	if err := workflow.ExecuteActivity(ctx, Withdraw, t).Get(ctx, nil); err != nil {
		return err
	}
	return workflow.ExecuteActivity(ctx, Deposit, t).Get(ctx, nil)
}

// end part
