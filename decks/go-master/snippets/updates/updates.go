// Package updates backs the Updates walkthrough and the Signals, Queries and
// Updates comparison in 11-updates.md.
package updates

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// debitAccount runs the DebitAccount Activity. It is off the slide, the way a
// helper that sets Activity Options usually is.
func debitAccount(ctx workflow.Context, amount int) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	return workflow.ExecuteActivity(ctx, DebitAccount, amount).Get(ctx, nil)
}

// slide: 11-updates.md "Go Updates"
const WithdrawUpdate = "withdraw"

func AccountWorkflow(ctx workflow.Context, opening int) error {
	balance := opening

	err := workflow.SetUpdateHandlerWithOptions(ctx, WithdrawUpdate,
		func(ctx workflow.Context, amount int) (int, error) {
			if err := debitAccount(ctx, amount); err != nil {
				return 0, err
			}
			balance -= amount
			return balance, nil
		},
		workflow.UpdateHandlerOptions{
			Validator: func(ctx workflow.Context, amount int) error {
				if amount > balance {
					return fmt.Errorf("insufficient funds: balance is %d", balance)
				}
				return nil
			},
		},
	)
	if err != nil {
		return err
	}

	return workflow.Await(ctx, func() bool { return balance == 0 })
}

// end slide

// LockedAccountWorkflow is not on a slide: it is the fix the speaker notes
// describe for two concurrent withdraw(80) calls against a balance of 100 that
// both pass the Validator. A workflow.Mutex lets one handler run at a time,
// and the handler re-checks once it holds the lock. Returning an error from
// the handler fails only that Update, not the Workflow.
func LockedAccountWorkflow(ctx workflow.Context, opening int) error {
	balance := opening
	mu := workflow.NewMutex(ctx)

	err := workflow.SetUpdateHandlerWithOptions(ctx, WithdrawUpdate,
		func(ctx workflow.Context, amount int) (int, error) {
			if err := mu.Lock(ctx); err != nil {
				return 0, err
			}
			defer mu.Unlock()
			if amount > balance {
				return 0, temporal.NewNonRetryableApplicationError(
					fmt.Sprintf("insufficient funds: balance is %d", balance), "InsufficientFunds", nil)
			}
			if err := debitAccount(ctx, amount); err != nil {
				return 0, err
			}
			balance -= amount
			return balance, nil
		},
		workflow.UpdateHandlerOptions{
			Validator: func(ctx workflow.Context, amount int) error {
				if amount > balance {
					return fmt.Errorf("insufficient funds: balance is %d", balance)
				}
				return nil
			},
		},
	)
	if err != nil {
		return err
	}

	return workflow.Await(ctx, func() bool { return balance == 0 })
}
