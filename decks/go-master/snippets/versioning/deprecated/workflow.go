// Package deprecated is step 2 of 3: no run without the marker is left, so
// the old branch goes and the minimum supported version rises to 1.
package deprecated

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/versioning"
)

var (
	ActivityA = versioning.ActivityA
	ActivityC = versioning.ActivityC
)

var options = workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second}

// slide: 18-versioning.md "Retiring the old branch"
func MyWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, options)
	if err := workflow.ExecuteActivity(ctx, ActivityA).Get(ctx, nil); err != nil {
		return err
	}

	workflow.GetVersion(ctx, "use-activity-c", 1, 1)
	return workflow.ExecuteActivity(ctx, ActivityC).Get(ctx, nil)
}

// end slide
