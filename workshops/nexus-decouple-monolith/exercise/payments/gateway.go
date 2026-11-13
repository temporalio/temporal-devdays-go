// [GIVEN] Simulated payment gateway. In production this would call Stripe,
// PayPal or SWIFT.
//
// The 10% failure rate is deliberate: it shows Temporal retrying an Activity
// for free. Because of it, never assert on a clean first attempt. Assert on
// the final Workflow state.
package payments

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

// validate reports whether a payment is well formed, and why not.
func validate(request shared.PaymentRequest) (bool, string) {
	if request.Amount <= 0 {
		return false, "invalid amount"
	}
	if request.SenderAccount == "" || request.ReceiverAccount == "" {
		return false, "missing account info"
	}
	return true, ""
}

// execute sends the payment to the banking network and returns its
// confirmation number.
func execute(ctx context.Context, request shared.PaymentRequest) (string, error) {
	select {
	case <-time.After(time.Duration(500+rand.IntN(500)) * time.Millisecond):
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// Simulate occasional gateway failures. Temporal retries automatically.
	if rand.Float64() < 0.1 {
		return "", fmt.Errorf("payment gateway timeout for %s, connection to banking network failed", request.TransactionID)
	}

	// Derived from the transaction ID, not the clock, so a retry of this
	// Activity returns the SAME confirmation number instead of minting a
	// second one. Idempotency is what makes an at-least-once retry safe.
	return "CONF-" + request.TransactionID, nil
}
