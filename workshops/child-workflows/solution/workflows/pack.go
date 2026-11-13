// ABOUTME: Challenges 3 and 4, the parent: fans out one PackItemWorkflow per item, then fans in.
package workflows

import (
	"fmt"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
)

// PackWorkflow returns one box per item, in item order.
func PackWorkflow(ctx workflow.Context, order Order) ([]string, error) {
	// Fan out: start one child per item, without waiting for any of them to finish.
	// ExecuteChildWorkflow returns a Future straight away; keep one per item.
	boxes := make([]workflow.ChildWorkflowFuture, len(order.Items))
	for i, item := range order.Items {
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("pack-%s-%d", order.OrderID, i),
			// The default, written out: when this parent closes, every child
			// still running is terminated.
			ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_TERMINATE,
		})
		boxes[i] = workflow.ExecuteChildWorkflow(childCtx, PackItemWorkflow, item)
	}
	// Wait for every child to start, not to finish.
	for _, box := range boxes {
		if err := box.GetChildWorkflowExecution().Get(ctx, nil); err != nil {
			return nil, err
		}
	}

	// TODO 1 start
	packed := make([]string, len(boxes))
	for i, box := range boxes {
		if err := box.Get(ctx, &packed[i]); err != nil {
			return nil, err
		}
	}
	return packed, nil
	// TODO 1 end
}
