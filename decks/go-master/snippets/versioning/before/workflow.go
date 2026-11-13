// Package before is the Workflow as first deployed: A, then B.
package before

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/versioning"
)

var (
	ActivityA = versioning.ActivityA
	ActivityB = versioning.ActivityB
)

// slide: 18-versioning.md "Before: the original Workflow"
func MyWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	if err := workflow.ExecuteActivity(ctx, ActivityA).Get(ctx, nil); err != nil {
		return err
	}
	return workflow.ExecuteActivity(ctx, ActivityB).Get(ctx, nil)
}

// end slide
