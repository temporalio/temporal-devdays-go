// ABOUTME: The one Activity a subscription runs: charge one billing cycle.
package subscription

import (
	"context"
	"fmt"
)

// ChargeSubscription charges one billing cycle and returns a receipt id.
func ChargeSubscription(ctx context.Context, subscriptionID string, cycle int) (string, error) {
	receipt := fmt.Sprintf("rcpt-%s-%d", subscriptionID, cycle)
	fmt.Printf("[activity] charged %s cycle %d -> %s\n", subscriptionID, cycle, receipt)
	return receipt, nil
}
