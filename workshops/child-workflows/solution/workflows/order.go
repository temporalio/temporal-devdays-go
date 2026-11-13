// ABOUTME: Challenge 1, the parent: runs one ShipmentWorkflow child and waits for it.
package workflows

import (
	"go.temporal.io/sdk/workflow"
)

// OrderWorkflow ships an order through one ShipmentWorkflow child and returns
// the child's tracking number.
func OrderWorkflow(ctx workflow.Context, order Order) (string, error) {
	// The child's options ride on the Context: its own Workflow ID here. It runs
	// on the parent's Task Queue, because none is set.
	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: "shipment-" + order.OrderID,
	})

	// ExecuteChildWorkflow starts the child and returns a Future at once. Get
	// blocks until the child returns. The child gets its own Event History.
	var tracking string
	err := workflow.ExecuteChildWorkflow(ctx, ShipmentWorkflow, order.OrderID).Get(ctx, &tracking)
	return tracking, err
}
