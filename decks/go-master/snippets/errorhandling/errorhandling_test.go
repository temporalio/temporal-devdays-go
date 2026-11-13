package errorhandling

import (
	"errors"
	"strings"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

var order = Order{OrderID: "o-4"}

func workflows(r worker.Registry) {
	for _, wf := range []any{PaymentWorkflow, CriticalWorkflow, PlainErrorWorkflow, BuggyWorkflow, ParentWorkflow} {
		r.RegisterWorkflow(wf)
	}
}

// withChargePayment registers fn under the name the Workflow calls.
func withChargePayment(fn any) func(worker.Registry) {
	return func(r worker.Registry) {
		workflows(r)
		r.RegisterActivityWithOptions(fn, activity.RegisterOptions{Name: "ChargePayment"})
	}
}

func execute(t *testing.T, tq string, wf any) client.WorkflowRun {
	t.Helper()
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("errors"), TaskQueue: tq,
	}, wf, order)
	if err != nil {
		t.Fatal(err)
	}
	return run
}

func TestANonRetryablePaymentErrorRunsOnceAndTheWorkflowTakesTheOtherPath(t *testing.T) {
	Attempts.Store(0)
	tq := testenv.Worker(t, withChargePayment(CountedChargePayment))
	var out string
	if err := execute(t, tq, PaymentWorkflow).Get(testenv.Ctx(t), &out); err != nil {
		t.Fatal(err)
	}
	if out != "new card requested" {
		t.Fatalf("got %q", out)
	}
	if n := Attempts.Load(); n != 1 {
		t.Fatalf("ChargePayment ran %d times, want 1", n)
	}
}

func TestAnApplicationErrorFailsTheWorkflowAndTheClientSeesItAsTheCause(t *testing.T) {
	tq := testenv.Worker(t, workflows)
	err := execute(t, tq, CriticalWorkflow).Get(testenv.Ctx(t), nil)
	var wfErr *temporal.WorkflowExecutionError
	if !errors.As(err, &wfErr) {
		t.Fatalf("got %T %v, want a WorkflowExecutionError", err, err)
	}
	var appErr *temporal.ApplicationError
	if !errors.As(err, &appErr) || appErr.Type() != "CriticalError" {
		t.Fatalf("cause %v, want an ApplicationError of type CriticalError", err)
	}
}

func TestAPlainReturnedErrorAlsoFailsTheWorkflow(t *testing.T) {
	tq := testenv.Worker(t, workflows)
	err := execute(t, tq, PlainErrorWorkflow).Get(testenv.Ctx(t), nil)
	var wfErr *temporal.WorkflowExecutionError
	if !errors.As(err, &wfErr) || !strings.Contains(err.Error(), "invalid order o-4") {
		t.Fatalf("got %v, want the Workflow failed with the returned error", err)
	}
}

func TestAPanicFailsTheWorkflowTaskAndTheWorkflowStaysOpen(t *testing.T) {
	tq := testenv.Worker(t, workflows)
	run := execute(t, tq, BuggyWorkflow)
	failed := false
	for i := 0; i < 50 && !failed; i++ {
		time.Sleep(200 * time.Millisecond)
		iter := testenv.Client().GetWorkflowHistory(testenv.Ctx(t), run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		for iter.HasNext() {
			e, err := iter.Next()
			if err != nil {
				t.Fatal(err)
			}
			failed = failed || e.EventType == enumspb.EVENT_TYPE_WORKFLOW_TASK_FAILED
		}
	}
	if !failed {
		t.Fatal("no WorkflowTaskFailed event")
	}
	d, err := testenv.Client().DescribeWorkflowExecution(testenv.Ctx(t), run.GetID(), run.GetRunID())
	if err != nil {
		t.Fatal(err)
	}
	if s := d.WorkflowExecutionInfo.Status; s != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
		t.Fatalf("status %v, want RUNNING", s)
	}
	_ = testenv.Client().TerminateWorkflow(testenv.Ctx(t), run.GetID(), "", "test done")
}

func TestFailWorkflowPanicPolicyMakesAPanicFailTheWorkflow(t *testing.T) {
	tq := testenv.Worker(t, workflows, worker.Options{WorkflowPanicPolicy: worker.FailWorkflow})
	err := execute(t, tq, BuggyWorkflow).Get(testenv.Ctx(t), nil)
	var wfErr *temporal.WorkflowExecutionError
	if !errors.As(err, &wfErr) || !strings.Contains(err.Error(), "index out of range") {
		t.Fatalf("got %v, want the Workflow failed by the panic", err)
	}
}

func TestAFailedChildArrivesAsChildWorkflowExecutionErrorWithItsCause(t *testing.T) {
	tq := testenv.Worker(t, workflows)
	var out string
	if err := execute(t, tq, ParentWorkflow).Get(testenv.Ctx(t), &out); err != nil {
		t.Fatal(err)
	}
	if out != "child failed: CriticalError" {
		t.Fatalf("got %q", out)
	}
}

func TestAnUnhandledActivityErrorFailsTheWorkflowWithTheFullCauseChain(t *testing.T) {
	tq := testenv.Worker(t, withChargePayment(GatewayDown))
	err := execute(t, tq, PaymentWorkflow).Get(testenv.Ctx(t), nil)
	var wfErr *temporal.WorkflowExecutionError
	var activityErr *temporal.ActivityError
	var appErr *temporal.ApplicationError
	if !errors.As(err, &wfErr) || !errors.As(err, &activityErr) || !errors.As(err, &appErr) {
		t.Fatalf("got %v, want WorkflowExecutionError > ActivityError > ApplicationError", err)
	}
	if appErr.Type() != "GatewayDown" {
		t.Fatalf("cause type %q", appErr.Type())
	}
}
