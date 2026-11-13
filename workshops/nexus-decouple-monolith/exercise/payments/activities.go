// [GIVEN] Payments team Activities. Thin wrappers around the gateway.
package payments

import (
	"context"

	"go.temporal.io/sdk/activity"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

// ValidatePayment checks the payment is well formed before anything moves.
func ValidatePayment(ctx context.Context, request shared.PaymentRequest) (bool, error) {
	ok, reason := validate(request)
	if !ok {
		activity.GetLogger(ctx).Info("[PaymentGateway] REJECTED", "transactionId", request.TransactionID, "reason", reason)
		return false, nil
	}
	activity.GetLogger(ctx).Info("[PaymentGateway] Validation passed", "transactionId", request.TransactionID)
	return true, nil
}

// ExecutePayment moves the money and returns the confirmation number.
func ExecutePayment(ctx context.Context, request shared.PaymentRequest) (string, error) {
	activity.GetLogger(ctx).Info("[PaymentGateway] Processing",
		"transactionId", request.TransactionID, "amount", request.Amount,
		"route", request.SenderCountry+" -> "+request.ReceiverCountry)
	confirmation, err := execute(ctx, request)
	if err != nil {
		return "", err
	}
	activity.GetLogger(ctx).Info("[PaymentGateway] Payment executed", "confirmationNumber", confirmation)
	return confirmation, nil
}
