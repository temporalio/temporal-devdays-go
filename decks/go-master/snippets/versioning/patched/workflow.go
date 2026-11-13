// Package patched is step 1 of 3: new runs take C, runs that started before
// the deploy keep B.
package patched

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/versioning"
)

var (
	ActivityA = versioning.ActivityA
	ActivityB = versioning.ActivityB
	ActivityC = versioning.ActivityC
)

// slide: 18-versioning.md "After: the versioned Workflow"
func MyWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	if err := workflow.ExecuteActivity(ctx, ActivityA).Get(ctx, nil); err != nil {
		return err
	}

	v := workflow.GetVersion(ctx, "use-activity-c", workflow.DefaultVersion, 1)
	if v == 1 {
		return workflow.ExecuteActivity(ctx, ActivityC).Get(ctx, nil)
	}
	return workflow.ExecuteActivity(ctx, ActivityB).Get(ctx, nil)
}

// end slide
