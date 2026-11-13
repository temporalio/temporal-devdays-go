package startexecute

import (
	"bytes"
	"errors"
	stdlog "log"
	"testing"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

// The slides name their Task Queue, so the Worker polls that one. Each test
// package has its own dev server, so nothing else shares it.
func startWorker(t *testing.T) {
	testenv.WorkerOn(t, "order-task-queue", func(r worker.Registry) {
		r.RegisterWorkflow(OrderWorkflow)
		r.RegisterWorkflow(ApprovalWorkflow)
	})
}

func capture(t *testing.T) *bytes.Buffer {
	var buf bytes.Buffer
	old := log
	log = stdlog.New(&buf, "", 0)
	t.Cleanup(func() { log = old })
	return &buf
}

func TestExecuteThenGetWaitsForTheResult(t *testing.T) {
	startWorker(t)
	out := capture(t)
	if err := ExecuteOrder(testenv.Ctx(t), testenv.Client(), Order{OrderID: testenv.ID("o")}); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "Result: submitted\n" {
		t.Fatalf("printed %q", got)
	}
}

func TestExecuteWorkflowReturnsWhileTheWorkflowIsStillRunning(t *testing.T) {
	startWorker(t)
	c := testenv.Client()
	run, err := c.ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("approval"), TaskQueue: "order-task-queue",
	}, ApprovalWorkflow, Order{OrderID: "o-3"})
	if err != nil {
		t.Fatal(err)
	}
	desc, err := c.DescribeWorkflowExecution(testenv.Ctx(t), run.GetID(), run.GetRunID())
	if err != nil {
		t.Fatal(err)
	}
	if s := desc.WorkflowExecutionInfo.Status; s != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
		t.Fatalf("status %v", s)
	}
	_ = c.TerminateWorkflow(testenv.Ctx(t), run.GetID(), "", "test over")
}

func TestTheRunCanQuerySignalAndWaitForTheResult(t *testing.T) {
	startWorker(t)
	out := capture(t)
	if err := StartApproval(testenv.Ctx(t), testenv.Client(), Order{OrderID: testenv.ID("o")}); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "Status: awaiting approval\nResult: approved by alice\n" {
		t.Fatalf("printed %q", got)
	}
}

func TestARunCanBeRebuiltFromTheWorkflowIDAlone(t *testing.T) {
	startWorker(t)
	c := testenv.Client()
	id := testenv.ID("approval")
	if _, err := c.ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: id, TaskQueue: "order-task-queue",
	}, ApprovalWorkflow, Order{OrderID: "o-4"}); err != nil {
		t.Fatal(err)
	}

	run := c.GetWorkflow(testenv.Ctx(t), id, "")
	value, err := c.QueryWorkflow(testenv.Ctx(t), id, "", GetStatusQuery)
	if err != nil {
		t.Fatal(err)
	}
	var status string
	if err := value.Get(&status); err != nil || status != "awaiting approval" {
		t.Fatalf("status %q, err %v", status, err)
	}
	if err := c.SignalWorkflow(testenv.Ctx(t), id, "", ApproveSignal, "bob"); err != nil {
		t.Fatal(err)
	}
	var result OrderResult
	if err := run.Get(testenv.Ctx(t), &result); err != nil || result.Status != "approved by bob" {
		t.Fatalf("result %+v, err %v", result, err)
	}
}

// A second start with a running Workflow's ID does not start a second
// Workflow. By default ExecuteWorkflow hands back the run already going;
// WorkflowExecutionErrorWhenAlreadyStarted turns that into an error.
func TestASecondStartWithARunningWorkflowsIDStartsNothingNew(t *testing.T) {
	startWorker(t)
	c := testenv.Client()
	opts := client.StartWorkflowOptions{ID: testenv.ID("approval"), TaskQueue: "order-task-queue"}
	first, err := c.ExecuteWorkflow(testenv.Ctx(t), opts, ApprovalWorkflow, Order{OrderID: "o-5"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := c.ExecuteWorkflow(testenv.Ctx(t), opts, ApprovalWorkflow, Order{OrderID: "o-5"})
	if err != nil {
		t.Fatal(err)
	}
	if second.GetRunID() != first.GetRunID() {
		t.Fatalf("a second run started: %s and %s", first.GetRunID(), second.GetRunID())
	}

	opts.WorkflowExecutionErrorWhenAlreadyStarted = true
	_, err = c.ExecuteWorkflow(testenv.Ctx(t), opts, ApprovalWorkflow, Order{OrderID: "o-5"})
	var already *serviceerror.WorkflowExecutionAlreadyStarted
	if !errors.As(err, &already) {
		t.Fatalf("want WorkflowExecutionAlreadyStarted, got %v", err)
	}
	_ = c.TerminateWorkflow(testenv.Ctx(t), first.GetID(), "", "test over")
}
