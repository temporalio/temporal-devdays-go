package errorhandling

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type OrderResult struct {
	OrderID string
}

var paymentOptions = workflow.ActivityOptions{
	StartToCloseTimeout: 30 * time.Second,
}

func requestNewCard(ctx workflow.Context, order Order) (string, error) {
	return "new card requested", nil
}

// slide: 24-error-handling.md "Handling an ActivityError"
func PaymentWorkflow(ctx workflow.Context, order Order) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, paymentOptions)
	var receipt string
	err := workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, &receipt)
	var activityErr *temporal.ActivityError
	if errors.As(err, &activityErr) {
		var appErr *temporal.ApplicationError
		switch cause := activityErr.Unwrap(); {
		case temporal.IsTimeoutError(cause):
			// the Activity timed out
		case temporal.IsCanceledError(cause):
			// the Activity was canceled
		case errors.As(cause, &appErr) && appErr.Type() == "PaymentError":
			return requestNewCard(ctx, order)
		}
	}
	return receipt, err
}

// end slide

// slide: 24-error-handling.md "Failing a Workflow with"
func CriticalWorkflow(ctx workflow.Context, order Order) (OrderResult, error) {
	// Returning an error fails the Workflow Execution
	return OrderResult{}, temporal.NewApplicationError(
		"unrecoverable state",
		"CriticalError",
	)
}

// end slide

// PlainErrorWorkflow backs the speaker note: any returned error fails the
// Workflow Execution, not only an ApplicationError.
func PlainErrorWorkflow(ctx workflow.Context, order Order) (OrderResult, error) {
	return OrderResult{}, fmt.Errorf("invalid order %s", order.OrderID)
}

// BuggyWorkflow backs the speaker note: a panic fails the Workflow Task, not
// the Workflow, and Temporal retries it until the code is fixed.
func BuggyWorkflow(ctx workflow.Context, order Order) (OrderResult, error) {
	var lines []string
	_ = lines[len(order.OrderID)] // index out of range: a bug
	return OrderResult{}, nil
}

// ParentWorkflow backs the speaker note: a failed Child Workflow arrives as a
// ChildWorkflowExecutionError, with the child's error as its cause.
func ParentWorkflow(ctx workflow.Context, order Order) (string, error) {
	err := workflow.ExecuteChildWorkflow(ctx, CriticalWorkflow, order).Get(ctx, nil)
	var childErr *temporal.ChildWorkflowExecutionError
	var appErr *temporal.ApplicationError
	if errors.As(err, &childErr) && errors.As(err, &appErr) {
		return "child failed: " + appErr.Type(), nil
	}
	if err != nil {
		return "", err
	}
	return "child succeeded", nil
}
