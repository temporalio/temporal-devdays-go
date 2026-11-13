package saga

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

// The Saga Workshop's five Activities: three forward steps and two undos.

func ReserveInventory(ctx context.Context, order Order) (string, error) {
	if order.FailReserve {
		return "", temporal.NewNonRetryableApplicationError("inventory service unavailable", "InventoryUnavailable", nil)
	}
	record("reserve:" + order.OrderID)
	return fmt.Sprintf("reserved %d x %s", order.Quantity, order.SKU), nil
}

func ChargePayment(ctx context.Context, order Order) (string, error) {
	record("charge:" + order.OrderID)
	if order.ChargeFailsAfterEffect {
		return "", temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("payment gateway timed out after capturing %d cents", order.AmountCents),
			"PaymentGatewayTimeout", nil)
	}
	return fmt.Sprintf("charged %d cents", order.AmountCents), nil
}

func DispatchCourier(ctx context.Context, order Order) (string, error) {
	if order.FailDispatch {
		return "", temporal.NewNonRetryableApplicationError("no courier available for this route", "CourierUnavailable", nil)
	}
	record("dispatch:" + order.OrderID)
	return "TRK-" + order.OrderID, nil
}

func ReleaseInventoryIfReserved(ctx context.Context, order Order) error {
	if !isReserved(order.OrderID) {
		activity.GetLogger(ctx).Info("nothing reserved, no-op", "order", order.OrderID)
		return nil
	}
	record("release:" + order.OrderID)
	return nil
}

func RefundPaymentIfCharged(ctx context.Context, order Order) error {
	if !isCharged(order.OrderID) {
		activity.GetLogger(ctx).Info("nothing charged, no-op", "order", order.OrderID)
		return nil
	}
	record("refund:" + order.OrderID)
	return nil
}
