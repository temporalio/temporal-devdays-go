package observability

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// slide: 25-observability.md "Logging from a Workflow"
func OrderWorkflow(ctx workflow.Context, orderID string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Order received", "OrderID", orderID)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})
	var receipt string
	if err := workflow.ExecuteActivity(ctx, ChargeCard, orderID).Get(ctx, &receipt); err != nil {
		return "", err
	}
	logger.Info("Order charged", "OrderID", orderID, "Receipt", receipt)
	return receipt, nil
}

// end slide
