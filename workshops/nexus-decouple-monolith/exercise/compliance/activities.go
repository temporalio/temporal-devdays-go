// [GIVEN] Compliance team Activity.
//
// In the monolith, PaymentProcessingWorkflow calls this Activity directly,
// which means Compliance's code runs inside the Payments Worker process. After
// decoupling, the Activity stays here but runs on the Compliance Worker,
// inside the Workflow that the Nexus handler starts.
package compliance

import (
	"context"

	"go.temporal.io/sdk/activity"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

// CheckCompliance scores a transaction's regulatory risk.
func CheckCompliance(ctx context.Context, request shared.ComplianceRequest) (shared.ComplianceResult, error) {
	activity.GetLogger(ctx).Info("[ComplianceChecker] Evaluating",
		"transactionId", request.TransactionID, "amount", request.Amount,
		"route", request.SenderCountry+" -> "+request.ReceiverCountry)
	return Check(request), nil
}
