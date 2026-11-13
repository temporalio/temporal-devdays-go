// ABOUTME: Challenge 2: starts itself as a child, one level shallower each time.
package workflows

import (
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ChainWorkflow returns how many levels of children ran below this one.
func ChainWorkflow(ctx workflow.Context, request ChainRequest) (int, error) {
	if request.Depth == 0 {
		if request.FailAtBottom {
			return 0, temporal.NewApplicationError("the bottom of the chain failed", "BottomOfChain")
		}
		return 0, nil
	}
	// Each child's ID is its parent's ID plus one segment, so the ID alone tells
	// you how deep you are: chain-ABC123/4/3/2/1/0.
	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s/%d", workflow.GetInfo(ctx).WorkflowExecution.ID, request.Depth-1),
	})
	var below int
	next := ChainRequest{Depth: request.Depth - 1, FailAtBottom: request.FailAtBottom}
	if err := workflow.ExecuteChildWorkflow(ctx, ChainWorkflow, next).Get(ctx, &below); err != nil {
		return 0, err
	}
	return below + 1, nil
}
