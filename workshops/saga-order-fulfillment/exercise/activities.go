// ABOUTME: The five Activities an order touches. Three forward steps, two undos.
// Each undo is named "IfX" because it must tolerate running when X never happened.
package saga

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/temporal"
)

// ReserveInventory holds the stock for the order.
func ReserveInventory(ctx context.Context, order Order) (string, error) {
	if order.FailReserve {
		// Fails BEFORE recording, so nothing lands. The undo still runs.
		return "", temporal.NewNonRetryableApplicationError(
			"inventory service unavailable", "InventoryUnavailable", nil)
	}
	record("reserve:" + order.OrderID)
	return fmt.Sprintf("reserved %d x %s", order.Quantity, order.SKU), nil
}

// ChargePayment takes the money.
func ChargePayment(ctx context.Context, order Order) (string, error) {
	// The effect lands first, exactly as a real gateway would capture the money
	// before returning. Then the call reports failure.
	record("charge:" + order.OrderID)

	if order.ChargeFailsAfterEffect {
		return "", temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("payment gateway timed out after capturing %d cents", order.AmountCents),
			"PaymentGatewayTimeout", nil)
	}
	return fmt.Sprintf("charged %d cents", order.AmountCents), nil
}

// DispatchCourier books the pickup and returns its tracking number.
func DispatchCourier(ctx context.Context, order Order) (string, error) {
	if order.FailDispatch {
		return "", temporal.NewNonRetryableApplicationError(
			"no courier available for this route", "CourierUnavailable", nil)
	}
	record("dispatch:" + order.OrderID)
	return "TRK-" + order.OrderID, nil
}

// ReleaseInventoryIfReserved is the undo for ReserveInventory. Named IfReserved
// because it has to tolerate being called when nothing is reserved: the
// reservation may never have landed, and Temporal retries a compensation like
// any other Activity, so this can run more than once.
func ReleaseInventoryIfReserved(ctx context.Context, order Order) error {
	// TODO 3a start
	// Guard this with isReserved(order.OrderID). When nothing is reserved, print
	// "[compensate] nothing reserved for <order ID>, no-op" and return nil.
	// TODO 3a end
	record("release:" + order.OrderID)
	return nil
}

// RefundPaymentIfCharged is the undo for ChargePayment. Same idempotency
// requirement, same reasons.
func RefundPaymentIfCharged(ctx context.Context, order Order) error {
	// TODO 3b start
	// Same guard, with isCharged(order.OrderID). Refunding a charge that never
	// landed, or refunding it twice, sends money the wrong way.
	// TODO 3b end
	record("refund:" + order.OrderID)
	return nil
}
