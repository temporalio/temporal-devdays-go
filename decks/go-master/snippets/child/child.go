// Package child backs 14-child-workflows.md.
package child

import (
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
)

type Order struct {
	OrderID  string
	Shipment string
}

type Shipment struct {
	TrackingNumber string
}

func ShipmentWorkflow(ctx workflow.Context, shipment string) (Shipment, error) {
	return Shipment{TrackingNumber: "TRACK-" + shipment}, nil
}

// slide: 14-child-workflows.md "Spawning a child"
func OrderWorkflow(ctx workflow.Context, order Order) (string, error) {
	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:        "shipment-" + order.OrderID,
		ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
	})

	var shipment Shipment
	err := workflow.ExecuteChildWorkflow(ctx, ShipmentWorkflow, order.Shipment).Get(ctx, &shipment)
	if err != nil {
		return "", err
	}
	return shipment.TrackingNumber, nil
}

// end slide
