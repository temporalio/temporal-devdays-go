// Package startexecute backs 08-start-vs-execute.md. The slides show only the
// client side; the Workflows they start are here, off the slides.
package startexecute

import "go.temporal.io/sdk/workflow"

type Order struct {
	OrderID string
}

type OrderResult struct {
	OrderID string
	Status  string
}

// OrderWorkflow finishes on its own: the shape that suits waiting at once.
func OrderWorkflow(ctx workflow.Context, order Order) (OrderResult, error) {
	return OrderResult{OrderID: order.OrderID, Status: "submitted"}, nil
}

const (
	ApproveSignal  = "approve"
	GetStatusQuery = "getStatus"
)

// ApprovalWorkflow waits for a person: the shape that needs the run kept.
func ApprovalWorkflow(ctx workflow.Context, order Order) (OrderResult, error) {
	status := "awaiting approval"
	if err := workflow.SetQueryHandler(ctx, GetStatusQuery, func() (string, error) {
		return status, nil
	}); err != nil {
		return OrderResult{}, err
	}
	var approver string
	workflow.GetSignalChannel(ctx, ApproveSignal).Receive(ctx, &approver)
	status = "approved by " + approver
	return OrderResult{OrderID: order.OrderID, Status: status}, nil
}
