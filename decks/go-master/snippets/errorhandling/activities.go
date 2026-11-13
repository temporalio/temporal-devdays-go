// Package errorhandling backs 24-error-handling.md: raising errors from an
// Activity, handling them in a Workflow, and failing a Workflow on purpose.
package errorhandling

import (
	"context"
	"sync/atomic"

	"go.temporal.io/sdk/temporal"
)

type Order struct {
	OrderID string
}

type paymentGateway struct{}

func (paymentGateway) Charge(ctx context.Context, order Order) (string, error) {
	return "charged-" + order.OrderID, nil
}

var gateway paymentGateway

// Every card is empty here, so ChargePayment always takes the error path.
func cardHasFunds(ctx context.Context, order Order) bool { return false }

// slide: 24-error-handling.md "Raising errors from an Activity"
func ChargePayment(ctx context.Context, order Order) (string, error) {
	if !cardHasFunds(ctx, order) {
		return "", temporal.NewApplicationErrorWithOptions(
			"payment failed: insufficient funds",
			"PaymentError",
			temporal.ApplicationErrorOptions{NonRetryable: true},
		)
	}
	return gateway.Charge(ctx, order)
}

// end slide

// Attempts counts ChargePayment attempts, so a test can prove a
// non-retryable error ran once.
var Attempts atomic.Int32

// CountedChargePayment wraps the slide's Activity to count its attempts
// without changing the slide. The test registers it under ChargePayment's name.
func CountedChargePayment(ctx context.Context, order Order) (string, error) {
	Attempts.Add(1)
	return ChargePayment(ctx, order)
}

// GatewayDown stands in for ChargePayment in the test where nothing in the
// Workflow handles the error.
func GatewayDown(ctx context.Context, order Order) (string, error) {
	return "", temporal.NewNonRetryableApplicationError("boom", "GatewayDown", nil)
}
