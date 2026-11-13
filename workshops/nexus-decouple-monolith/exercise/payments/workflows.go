// The payment Workflow. Three steps:
//
//	Step 1: ValidatePayment    Payments team
//	Step 2: compliance check   Compliance team  <- this one crosses a team boundary
//	Step 3: ExecutePayment     Payments team
//
// In the monolith, step 2 is an Activity, so the Compliance team's code runs
// on this Worker. TODO 4 moves it behind a Nexus Endpoint.
package payments

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
	// TODO 4a start

	// Delete this import. This is the coupling: the Workflow below calls the
	// Compliance team's Activity, and Activities run on THIS Worker, so the
	// Compliance team's code executes inside the Payments process. TODO 4b
	// replaces the one place that uses it.
	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/compliance"
	// TODO 4a end
)

// PaymentProcessingWorkflow validates, clears and executes one payment.
func PaymentProcessingWorkflow(ctx workflow.Context, request shared.PaymentRequest) (shared.PaymentResult, error) {
	logger := workflow.GetLogger(ctx)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
		},
	})

	// Step 1: validate (Payments team).
	var valid bool
	if err := workflow.ExecuteActivity(ctx, ValidatePayment, request).Get(ctx, &valid); err != nil {
		return shared.PaymentResult{}, err
	}
	if !valid {
		return shared.PaymentResult{
			Success:       false,
			TransactionID: request.TransactionID,
			Status:        shared.StatusRejected,
			Error:         "Payment validation failed",
		}, nil
	}
	logger.Info("Step 1 passed: validation OK", "transactionId", request.TransactionID)

	// Step 2: compliance check, across the team boundary.
	compReq := shared.ComplianceRequest{
		TransactionID:   request.TransactionID,
		Amount:          request.Amount,
		SenderCountry:   request.SenderCountry,
		ReceiverCountry: request.ReceiverCountry,
		Description:     request.Description,
	}
	logger.Info("Step 2: calling compliance check", "transactionId", request.TransactionID)

	var result shared.ComplianceResult
	// TODO 4b start
	// Replace this Activity call with a Nexus call:
	//
	//	complianceClient := workflow.NewNexusClient(shared.ComplianceEndpoint, shared.ComplianceServiceName)
	//	err := complianceClient.ExecuteOperation(ctx, shared.CheckComplianceOperation, compReq,
	//		workflow.NexusOperationOptions{ScheduleToCloseTimeout: 10 * time.Minute},
	//	).Get(ctx, &result)
	//
	// Ten minutes covers the whole call including retries, which is what lets
	// it outlive the Compliance Worker going away. You prove that in
	// Challenge 5.
	//
	// Note what the Workflow names: an Endpoint, a Service and an Operation.
	// No Namespace, no Task Queue, no address. The Registry resolves those,
	// so this Workflow does not change when Compliance moves.
	err := workflow.ExecuteActivity(ctx, compliance.CheckCompliance, compReq).Get(ctx, &result)
	// TODO 4b end
	if err != nil {
		return shared.PaymentResult{}, err
	}
	logger.Info("Compliance result", "riskLevel", result.RiskLevel, "approved", result.Approved)

	// A declined payment is a business outcome, not a failure. The Workflow
	// completes successfully and reports the decision.
	if !result.Approved {
		return shared.PaymentResult{
			Success:       false,
			TransactionID: request.TransactionID,
			Status:        shared.StatusDeclinedCompliance,
			RiskLevel:     result.RiskLevel,
			Explanation:   result.Explanation,
		}, nil
	}

	// Step 3: execute, only if compliance approved.
	logger.Info("Step 3: executing payment", "transactionId", request.TransactionID)
	var confirmation string
	if err := workflow.ExecuteActivity(ctx, ExecutePayment, request).Get(ctx, &confirmation); err != nil {
		return shared.PaymentResult{}, err
	}

	return shared.PaymentResult{
		Success:            true,
		TransactionID:      request.TransactionID,
		Status:             shared.StatusCompleted,
		RiskLevel:          result.RiskLevel,
		Explanation:        result.Explanation,
		ConfirmationNumber: confirmation,
	}, nil
}

// ReviewCallerWorkflow sends a compliance review decision through Nexus.
//
// Routing the review through the Endpoint respects the team boundary: neither
// team needs to know the other's Workflow IDs or Update names.
func ReviewCallerWorkflow(ctx workflow.Context, request shared.ReviewRequest) (shared.ComplianceResult, error) {
	var result shared.ComplianceResult
	// TODO 4c start
	// Same client as 4b, but call shared.SubmitReviewOperation with the
	// request, a ScheduleToCloseTimeout of 10 * time.Second, and read the
	// answer into result.
	//
	// Seconds, not minutes: submitReview is a SYNC Operation, so the handler
	// has to answer inside the Nexus handler deadline.
	err := temporal.NewNonRetryableApplicationError(
		"TODO 4c: send the review through the Nexus Endpoint (payments/workflows.go)", "NotImplemented", nil)
	// TODO 4c end
	return result, err
}
