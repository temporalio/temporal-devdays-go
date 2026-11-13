package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// slide: 05-workflows.md "Go Workflow anatomy"
func TransferWorkflow(ctx workflow.Context, t Transfer) (Receipt, error) {
	var a *Activities
	fraudCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	})
	bankCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 5},
	})

	var flagged bool
	if err := workflow.ExecuteActivity(fraudCtx, a.CheckFraud, t).Get(ctx, &flagged); err != nil {
		return Receipt{}, err
	}
	if flagged {
		return Receipt{Status: "rejected", Reason: "flagged by the fraud check"}, nil
	}

	if err := workflow.ExecuteActivity(bankCtx, a.Withdraw, t).Get(ctx, nil); err != nil {
		return Receipt{}, err
	}
	if err := workflow.ExecuteActivity(bankCtx, a.Deposit, t).Get(ctx, nil); err != nil {
		if rerr := workflow.ExecuteActivity(bankCtx, a.Refund, t).Get(ctx, nil); rerr != nil {
			return Receipt{}, rerr
		}
		return Receipt{Status: "refunded", Reason: err.Error()}, nil
	}
	return Receipt{Status: "completed"}, nil
}

// end slide
