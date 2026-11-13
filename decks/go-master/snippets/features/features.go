// Package features backs 22-more-features.md: Schedules, Serverless Workers,
// and Priority and Fairness.
package features

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Order struct {
	ID       string
	TenantID string
}

func ChargePayment(ctx context.Context, order Order) (string, error) {
	return fmt.Sprintf("charged %s for %s", order.ID, order.TenantID), nil
}

func BuildReport(ctx context.Context) (string, error) {
	return "report ready", nil
}

// ReportWorkflow is what the Schedules slide starts every morning. Not on a
// slide.
func ReportWorkflow(ctx workflow.Context) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})
	var report string
	err := workflow.ExecuteActivity(ctx, BuildReport).Get(ctx, &report)
	return report, err
}

// slide: 22-more-features.md "heading: Priority and Fairness in Go"
func OrderWorkflow(ctx workflow.Context, order Order) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		Priority: temporal.Priority{
			PriorityKey:    1,
			FairnessKey:    order.TenantID,
			FairnessWeight: 2,
		},
	})
	var receipt string
	err := workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, &receipt)
	return receipt, err
}

// end slide
