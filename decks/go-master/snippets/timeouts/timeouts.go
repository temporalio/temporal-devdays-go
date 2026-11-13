// Package timeouts backs 23-timeouts-and-retries.md: who sets which timeout,
// and what Retry Policies leave in Event History.
package timeouts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Order struct {
	OrderID string
}

func ChargePayment(ctx context.Context, order Order) (string, error) {
	return "charged-" + order.OrderID, nil
}

// ChargeFlakyPayment is the Maximum Attempts diagram: a payment API that keeps
// returning 503.
func ChargeFlakyPayment(ctx context.Context, order Order) (string, error) {
	return "", errors.New("payment API returned 503")
}

// ChargeDeclinedCard is the Non-retryable diagram with the caller deciding:
// the Activity's error is retryable on its own.
func ChargeDeclinedCard(ctx context.Context, order Order) (string, error) {
	return "", temporal.NewApplicationError("card declined", "CardDeclined")
}

// StartOrder is the Workflow timeouts slide, wrapped so the test can call it.
func StartOrder(ctx context.Context, c client.Client, order Order, taskQueue, workflowID string) (client.WorkflowRun, error) {
	// slide: 23-timeouts-and-retries.md "# Workflow timeouts"
	we, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                taskQueue,
		WorkflowExecutionTimeout: 30 * 24 * time.Hour,
	}, OrderWorkflow, order)
	// end slide
	return we, err
}

func OrderWorkflow(ctx workflow.Context, order Order) (string, error) {
	// slide: 23-timeouts-and-retries.md "# Activity timeouts"
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	// end slide
	var receipt string
	err := workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, &receipt)
	return receipt, err
}

// failureOf reports what the Workflow saw when an Activity gave up: its retry
// state and the cause's type.
func failureOf(ctx workflow.Context, f workflow.Future) (string, error) {
	var receipt string
	err := f.Get(ctx, &receipt)
	var activityErr *temporal.ActivityError
	var appErr *temporal.ApplicationError
	if errors.As(err, &activityErr) && errors.As(err, &appErr) {
		return fmt.Sprintf("%s:%s", activityErr.RetryState(), appErr.Type()), nil
	}
	return receipt, err
}

// MaxAttemptsWorkflow is the Maximum Attempts diagram, with a short interval
// so the test is quick.
func MaxAttemptsWorkflow(ctx workflow.Context, order Order) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: 200 * time.Millisecond,
		},
	})
	return failureOf(ctx, workflow.ExecuteActivity(ctx, ChargeFlakyPayment, order))
}

// NonRetryableWorkflow is the Non-retryable diagram's second form: the caller
// lists the type.
func NonRetryableWorkflow(ctx workflow.Context, order Order) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			NonRetryableErrorTypes: []string{"CardDeclined"},
		},
	})
	return failureOf(ctx, workflow.ExecuteActivity(ctx, ChargeDeclinedCard, order))
}

// NoTimeoutWorkflow backs the speaker note: with neither Start-To-Close nor
// Schedule-To-Close, the Service refuses the command and the Workflow Task
// fails.
func NoTimeoutWorkflow(ctx workflow.Context, order Order) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{})
	var receipt string
	err := workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, &receipt)
	return receipt, err
}

// BusyWorkflow backs the speaker note on the Worker's deadlock detector:
// Workflow code that holds the Workflow Task without yielding past the
// Worker's DeadlockDetectionTimeout fails the Workflow Task.
func BusyWorkflow(ctx workflow.Context) error {
	time.Sleep(3 * time.Second) // never block a Workflow like this
	return nil
}
