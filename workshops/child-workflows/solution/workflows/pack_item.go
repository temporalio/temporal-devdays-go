// ABOUTME: Challenges 3 and 4, the child: packs one item into a box.
package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// PackItemWorkflow packs one item into a box.
func PackItemWorkflow(ctx workflow.Context, item string) (string, error) {
	if err := workflow.Sleep(ctx, 3*time.Second); err != nil {
		return "", err
	}
	return "box:" + item, nil
}
