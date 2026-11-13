// ABOUTME: A subscription that bills every cycle until it reaches its last one.
// Every charge and every timer appends to one Event History, and that history has
// a hard ceiling. Continue-As-New closes this run and starts a fresh one: same
// Workflow ID, new Run ID, new short history.
package subscription

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// SubscriptionWorkflow bills one cycle at a time until MaxCycles have been billed.
func SubscriptionWorkflow(ctx workflow.Context, input SubscriptionInput) (SubscriptionResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	cyclesCompleted := input.CyclesCompleted

	for input.MaxCycles == 0 || cyclesCompleted < input.MaxCycles {
		err := workflow.ExecuteActivity(ctx, ChargeSubscription, input.SubscriptionID, cyclesCompleted+1).Get(ctx, nil)
		if err != nil {
			return SubscriptionResult{}, err
		}
		cyclesCompleted++
		if err := workflow.Sleep(ctx, Cycle); err != nil {
			return SubscriptionResult{}, err
		}

		// TODO 1 start: when the Service suggests it, continue as new with what this run has done
		if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
			return SubscriptionResult{}, workflow.NewContinueAsNewError(ctx, SubscriptionWorkflow, SubscriptionInput{
				SubscriptionID:  input.SubscriptionID,
				CyclesCompleted: cyclesCompleted,
				MaxCycles:       input.MaxCycles,
			})
		}
		// TODO 1 end
	}
	return SubscriptionResult{SubscriptionID: input.SubscriptionID, TotalCycles: cyclesCompleted}, nil
}
