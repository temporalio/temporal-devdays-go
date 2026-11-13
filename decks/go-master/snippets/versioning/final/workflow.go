// Package final is step 3 of 3, and also the unsafe edit: A, then C, with no
// GetVersion at all. A non-determinism error against any history that ran B.
package final

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/versioning"
)

var (
	ActivityA = versioning.ActivityA
	ActivityC = versioning.ActivityC
)

func MyWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	if err := workflow.ExecuteActivity(ctx, ActivityA).Get(ctx, nil); err != nil {
		return err
	}
	return workflow.ExecuteActivity(ctx, ActivityC).Get(ctx, nil)
}
