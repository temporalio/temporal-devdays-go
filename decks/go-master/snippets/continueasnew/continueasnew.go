// Package continueasnew backs 15-continue-as-new.md.
package continueasnew

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// slide: 15-continue-as-new.md "Continue-As-New in practice"
func SubscriptionWorkflow(ctx workflow.Context, sub Subscription) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	})

	for {
		if err := workflow.ExecuteActivity(ctx, ChargeMonthly, sub).Get(ctx, nil); err != nil {
			return err
		}
		sub.MonthsBilled++
		if err := workflow.Sleep(ctx, 30*24*time.Hour); err != nil {
			return err
		}

		if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
			return workflow.NewContinueAsNewError(ctx, SubscriptionWorkflow, sub)
		}
	}
}

// end slide
