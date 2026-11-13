package tests

import (
	"context"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/payments"
	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

// startPaymentsWorker runs payments.NewWorker for the length of a test, plus
// any test Workflow registered on it first.
func startPaymentsWorker(t *testing.T, c client.Client, extra ...func(workflow.Context) (string, error)) {
	t.Helper()
	w := payments.NewWorker(c)
	for _, wf := range extra {
		w.RegisterWorkflow(wf)
	}
	if err := w.Start(); err != nil {
		t.Fatalf("payments/worker.go, NewWorker: the Worker it returns did not start: %v", err)
	}
	t.Cleanup(w.Stop)
}

// startOnPayments begins a Workflow on the Payments Task Queue, as the
// starters do.
func startOnPayments(t *testing.T, ctx context.Context, c client.Client, id string, workflowFn any, args ...any) client.WorkflowRun {
	t.Helper()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: id, TaskQueue: shared.PaymentsTaskQueue}, workflowFn, args...)
	if err != nil {
		t.Fatalf("Could not start the Workflow %s: %v", id, err)
	}
	return run
}

func TestChallenge4_PaymentCallsComplianceOverNexus(t *testing.T) {
	paymentsClient, complianceClient := clients(t)
	startComplianceWorker(t, complianceClient)
	startPaymentsWorker(t, paymentsClient)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	request := shared.PaymentRequest{
		TransactionID: "TXN-T4-LOW", Amount: 250, Currency: "USD", SenderCountry: "US", ReceiverCountry: "US",
		Description: "test", SenderAccount: "ACC-001", ReceiverAccount: "ACC-002",
	}
	run := startOnPayments(t, ctx, paymentsClient, "payment-"+request.TransactionID, payments.PaymentProcessingWorkflow, request)
	var result shared.PaymentResult
	await(t, ctx, paymentsClient, run, &result, "payments/workflows.go, PaymentProcessingWorkflow")

	calledNexus := false
	for _, event := range history(t, paymentsClient, run.GetID()) {
		if attrs := event.GetActivityTaskScheduledEventAttributes(); attrs != nil && attrs.GetActivityType().GetName() == "CheckCompliance" {
			t.Fatalf("payments/workflows.go, PaymentProcessingWorkflow, TODO 4b: step 2 still runs the CheckCompliance Activity on the Payments Worker. Call the checkCompliance Operation through a Nexus client instead.")
		}
		if attrs := event.GetNexusOperationScheduledEventAttributes(); attrs != nil {
			calledNexus = true
			if attrs.GetOperation() != shared.CheckComplianceOperation.Name() {
				t.Fatalf("payments/workflows.go, PaymentProcessingWorkflow, TODO 4b: step 2 calls the Nexus Operation %q. It must call shared.CheckComplianceOperation.", attrs.GetOperation())
			}
		}
	}
	if !calledNexus {
		t.Fatalf("payments/workflows.go, PaymentProcessingWorkflow, TODO 4b: the Event History of %s has no Nexus Operation Scheduled event. Step 2 must call checkCompliance through a Nexus client.", run.GetID())
	}
	if result.Status != shared.StatusCompleted || result.RiskLevel != shared.RiskLow {
		t.Fatalf("payments/workflows.go, PaymentProcessingWorkflow, TODO 4b: a routine $250 payment finished %s with risk %q, want COMPLETED and LOW. Read the Nexus result into result.", result.Status, result.RiskLevel)
	}
}

func TestChallenge4_ReviewCallerSubmitsThroughNexus(t *testing.T) {
	paymentsClient, complianceClient := clients(t)
	startComplianceWorker(t, complianceClient)
	startPaymentsWorker(t, paymentsClient)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	request := shared.PaymentRequest{
		TransactionID: "TXN-T4-MEDIUM", Amount: 12_000, Currency: "USD", SenderCountry: "US", ReceiverCountry: "UK",
		Description: "test", SenderAccount: "ACC-003", ReceiverAccount: "ACC-004",
	}
	paymentRun := startOnPayments(t, ctx, paymentsClient, "payment-"+request.TransactionID, payments.PaymentProcessingWorkflow, request)
	if !waitForEvent(t, complianceClient, "compliance-"+request.TransactionID, enumspb.EVENT_TYPE_TIMER_STARTED, 20*time.Second) {
		t.Fatalf("payments/workflows.go, PaymentProcessingWorkflow, TODO 4b: a MEDIUM-risk payment did not start the Workflow compliance-%s within 20 seconds.", request.TransactionID)
	}

	review := shared.ReviewRequest{TransactionID: request.TransactionID, Approved: true, Explanation: "Approved by the Challenge 4 test"}
	reviewRun := startOnPayments(t, ctx, paymentsClient, "review-"+request.TransactionID, payments.ReviewCallerWorkflow, review)
	var reviewed shared.ComplianceResult
	await(t, ctx, paymentsClient, reviewRun, &reviewed, "payments/workflows.go, ReviewCallerWorkflow, TODO 4c")
	if !reviewed.Approved || reviewed.Explanation != review.Explanation {
		t.Fatalf("payments/workflows.go, ReviewCallerWorkflow, TODO 4c: it returned %+v. Return the result of the submitReview Operation.", reviewed)
	}
	calledNexus := false
	for _, event := range history(t, paymentsClient, reviewRun.GetID()) {
		if attrs := event.GetNexusOperationScheduledEventAttributes(); attrs != nil && attrs.GetOperation() == shared.SubmitReviewOperation.Name() {
			calledNexus = true
		}
	}
	if !calledNexus {
		t.Fatalf("payments/workflows.go, ReviewCallerWorkflow, TODO 4c: it must call the submitReview Operation through a Nexus client.")
	}

	var result shared.PaymentResult
	await(t, ctx, paymentsClient, paymentRun, &result, "payments/workflows.go, PaymentProcessingWorkflow")
	if result.Status != shared.StatusCompleted || result.Explanation != review.Explanation {
		t.Fatalf("payments/workflows.go, PaymentProcessingWorkflow: after the review, %s finished %s with %q, want COMPLETED with the reviewer's explanation.", paymentRun.GetID(), result.Status, result.Explanation)
	}
}

// activityProbe asks the Payments Worker to run the Compliance team's
// Activity, once, by name. It must not be able to.
func activityProbe(ctx workflow.Context) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})
	var result shared.ComplianceResult
	err := workflow.ExecuteActivity(ctx, "CheckCompliance", shared.ComplianceRequest{TransactionID: "TXN-PROBE", Amount: 1}).Get(ctx, &result)
	if err == nil {
		return "registered", nil
	}
	return "not registered", nil
}

func TestChallenge4_PaymentsWorkerRunsOnlyPaymentsCode(t *testing.T) {
	paymentsClient, _ := clients(t)
	startPaymentsWorker(t, paymentsClient, activityProbe)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	run := startOnPayments(t, ctx, paymentsClient, "test-activity-probe", activityProbe)
	var answer string
	await(t, ctx, paymentsClient, run, &answer, "payments/worker.go, NewWorker")
	if answer == "registered" {
		t.Fatalf("payments/worker.go, NewWorker, TODO 5b: the Payments Worker still registers compliance.CheckCompliance, so the Compliance team's code runs in the Payments process. Delete that registration, then the import (TODO 5a).")
	}
}
