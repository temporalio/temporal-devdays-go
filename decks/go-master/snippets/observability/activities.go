// Package observability backs 25-observability.md: SDK metrics served to
// Prometheus, and logging from Workflows and Activities.
package observability

import (
	"context"

	"go.temporal.io/sdk/activity"
)

// slide: 25-observability.md "Logging from an Activity"
func ChargeCard(ctx context.Context, orderID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Charging card", "OrderID", orderID)
	receipt := "receipt-" + orderID
	logger.Info("Card charged", "Receipt", receipt)
	return receipt, nil
}

// end slide
