// ABOUTME: The money transfer Workflow: withdraw, then deposit.
package moneytransfer

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// MoneyTransferWorkflow moves details.Amount dollars from the source account to
// the target account.
func MoneyTransferWorkflow(ctx workflow.Context, details PaymentDetails) (string, error) {
	// Retry a failed step every 2 seconds at most, with no limit on attempts:
	// a step that keeps failing waits for someone to fix it.
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumInterval: 2 * time.Second},
	})

	var withdrawID, depositID string
	if err := workflow.ExecuteActivity(ctx, Withdraw, details).Get(ctx, &withdrawID); err != nil {
		return "", err
	}
	if err := workflow.ExecuteActivity(ctx, Deposit, details).Get(ctx, &depositID); err != nil {
		return "", err
	}
	return "Transfer complete: " + withdrawID + ", " + depositID, nil
}
