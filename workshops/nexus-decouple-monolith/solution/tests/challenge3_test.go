package tests

import (
	"context"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/compliance"
	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

// Challenge 3 is the Compliance side only, so these tests bring their own
// caller: two small Workflows in the Payments Namespace that call each
// Operation through the Endpoint, the way Payments will in Challenge 4.

func testCheckCaller(ctx workflow.Context, request shared.ComplianceRequest) (shared.ComplianceResult, error) {
	var result shared.ComplianceResult
	err := workflow.NewNexusClient(complianceEndpoint, shared.ComplianceServiceName).
		ExecuteOperation(ctx, shared.CheckComplianceOperation, request,
			workflow.NexusOperationOptions{ScheduleToCloseTimeout: time.Minute}).
		Get(ctx, &result)
	return result, err
}

func testReviewCaller(ctx workflow.Context, request shared.ReviewRequest) (shared.ComplianceResult, error) {
	var result shared.ComplianceResult
	err := workflow.NewNexusClient(complianceEndpoint, shared.ComplianceServiceName).
		ExecuteOperation(ctx, shared.SubmitReviewOperation, request,
			workflow.NexusOperationOptions{ScheduleToCloseTimeout: 10 * time.Second}).
		Get(ctx, &result)
	return result, err
}

// startComplianceWorker runs compliance.NewWorker for the length of a test.
func startComplianceWorker(t *testing.T, c client.Client) {
	t.Helper()
	w, err := compliance.NewWorker(c)
	if err != nil {
		t.Fatalf("compliance/worker.go, NewWorker: it returned an error: %v", err)
	}
	if err := w.Start(); err != nil {
		t.Fatalf("compliance/worker.go, NewWorker: the Worker it returns did not start: %v", err)
	}
	t.Cleanup(w.Stop)
}

// startTestCallers runs the two caller Workflows above for the length of a test.
func startTestCallers(t *testing.T, c client.Client) {
	t.Helper()
	w := worker.New(c, testQueue, worker.Options{})
	w.RegisterWorkflow(testCheckCaller)
	w.RegisterWorkflow(testReviewCaller)
	if err := w.Start(); err != nil {
		t.Fatalf("Could not start the test's caller Worker: %v", err)
	}
	t.Cleanup(w.Stop)
}

// start begins a Workflow on the tests' own Task Queue.
func start(t *testing.T, ctx context.Context, c client.Client, id string, workflowFn any, arg any) client.WorkflowRun {
	t.Helper()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: id, TaskQueue: testQueue}, workflowFn, arg)
	if err != nil {
		t.Fatalf("Could not start the test's Workflow %s: %v", id, err)
	}
	return run
}

func TestChallenge3_ComplianceWorkerAnswersNexus(t *testing.T) {
	paymentsClient, complianceClient := clients(t)
	startComplianceWorker(t, complianceClient)

	// A Worker polls for Nexus tasks only when a Nexus Service is registered.
	deadline := time.Now().Add(10 * time.Second)
	for {
		resp, err := complianceClient.DescribeTaskQueue(context.Background(), complianceQueue, enumspb.TASK_QUEUE_TYPE_NEXUS)
		if err == nil && len(resp.GetPollers()) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("compliance/worker.go, NewWorker, TODO 3: the Compliance Worker starts, but it is not polling for Nexus tasks, so every call from Payments goes unanswered. Register the Service from NewNexusService() with w.RegisterNexusService(...).")
		}
		time.Sleep(250 * time.Millisecond)
	}

	// And the async Operation, given, answers through the Endpoint.
	startTestCallers(t, paymentsClient)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	request := shared.ComplianceRequest{TransactionID: "TXN-T3-LOW", Amount: 250, SenderCountry: "US", ReceiverCountry: "US", Description: "test"}
	run := start(t, ctx, paymentsClient, "test-check-"+request.TransactionID, testCheckCaller, request)
	var result shared.ComplianceResult
	await(t, ctx, paymentsClient, run, &result, "compliance/handler.go, calling checkCompliance through "+complianceEndpoint)
	if !result.Approved || result.RiskLevel != shared.RiskLow {
		t.Fatalf("compliance/handler.go: checkCompliance returned %+v for a routine $250 payment, want an approved LOW result.", result)
	}
}

func TestChallenge3_SubmitReviewUpdatesTheRunningCheck(t *testing.T) {
	paymentsClient, complianceClient := clients(t)
	startComplianceWorker(t, complianceClient)
	startTestCallers(t, paymentsClient)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// A MEDIUM-risk check, which parks waiting for a reviewer.
	check := shared.ComplianceRequest{TransactionID: "TXN-T3-MEDIUM", Amount: 12_000, SenderCountry: "US", ReceiverCountry: "UK", Description: "test"}
	checkRun := start(t, ctx, paymentsClient, "test-check-"+check.TransactionID, testCheckCaller, check)
	// It accepts a review once its automated check has run: the 10-second
	// timer starts right after.
	if !waitForEvent(t, complianceClient, "compliance-"+check.TransactionID, enumspb.EVENT_TYPE_TIMER_STARTED, 20*time.Second) {
		t.Fatalf("compliance/handler.go: checkCompliance did not start the Workflow compliance-%s within 20 seconds.", check.TransactionID)
	}

	review := shared.ReviewRequest{TransactionID: check.TransactionID, Approved: true, Explanation: "Approved by the Challenge 3 test"}
	reviewRun := start(t, ctx, paymentsClient, "test-review-"+check.TransactionID, testReviewCaller, review)
	var reviewed shared.ComplianceResult
	await(t, ctx, paymentsClient, reviewRun, &reviewed, "compliance/handler.go, SubmitReviewOperation, TODO 2: calling submitReview")
	if !reviewed.Approved || reviewed.Explanation != review.Explanation {
		t.Fatalf("compliance/handler.go, SubmitReviewOperation, TODO 2: submitReview returned %+v. It must return the result of the %q Update, which echoes the reviewer's decision and explanation.", reviewed, compliance.ReviewUpdate)
	}

	var final shared.ComplianceResult
	await(t, ctx, paymentsClient, checkRun, &final, "compliance/handler.go, the parked compliance check")
	if !final.Approved || final.Explanation != review.Explanation {
		t.Fatalf("compliance/handler.go, SubmitReviewOperation, TODO 2: the review never reached the Workflow compliance-%s; it finished with %+v. Send the Update to workflowIDFor(input.TransactionID).", check.TransactionID, final)
	}
}
