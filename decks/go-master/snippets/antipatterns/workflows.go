package antipatterns

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Transfer struct {
	TransferID string
}

type Deposit struct {
	DepositID string
}

var checkOptions = workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second}

var releaseOptions = workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second}

// TransferWorkflow is what the business-ID slide starts. It stays running long
// enough for the second start to meet it.
func TransferWorkflow(ctx workflow.Context, transfer Transfer) error {
	return workflow.Sleep(ctx, 24*time.Hour)
}

// DebitWorkflow runs the idempotency slide's PostDebit, retried after the
// ledger committed.
func DebitWorkflow(ctx workflow.Context, debit Debit) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{InitialInterval: time.Second},
	})
	var key string
	err := workflow.ExecuteActivity(ctx, PostDebit, debit).Get(ctx, &key)
	return key, err
}

func LoopPollingWorkflow(ctx workflow.Context, wireID string) error {
	ctx = workflow.WithActivityOptions(ctx, checkOptions)
	// part: 17b-patterns-and-anti-patterns.md "Polling an external system" 1
	// ❌ Every pass adds an Activity and a Timer to history
	var status string
	err := workflow.ExecuteActivity(ctx, CheckSettlement, wireID).Get(ctx, &status)
	for err == nil && status != "SETTLED" {
		if err = workflow.Sleep(ctx, time.Minute); err == nil {
			err = workflow.ExecuteActivity(ctx, CheckSettlement, wireID).Get(ctx, &status)
		}
	}
	// end part
	return err
}

func RetryPollingWorkflow(ctx workflow.Context, wireID string) error {
	// part: 17b-patterns-and-anti-patterns.md "Polling an external system" 2
	// ✅ One Activity. The Retry Policy does the waiting
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Minute,
			BackoffCoefficient: 1,
		},
	})
	return workflow.ExecuteActivity(ctx, AwaitSettlement, wireID).Get(ctx, nil)
	// end part
}

// part: 17b-patterns-and-anti-patterns.md "Sleeping in an Activity" 2
// ✅ A durable Timer in the Workflow, then a quick Activity
func DepositHoldWorkflow(ctx workflow.Context, deposit Deposit) error {
	if err := workflow.Sleep(ctx, 72*time.Hour); err != nil {
		return err
	}
	ctx = workflow.WithActivityOptions(ctx, releaseOptions)
	return workflow.ExecuteActivity(ctx, ReleaseFunds, deposit.DepositID).Get(ctx, nil)
}

// end part
