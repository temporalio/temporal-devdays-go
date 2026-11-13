package timeouts

import (
	"strings"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func register(r worker.Registry) {
	for _, wf := range []any{OrderWorkflow, MaxAttemptsWorkflow, NonRetryableWorkflow, NoTimeoutWorkflow, BusyWorkflow} {
		r.RegisterWorkflow(wf)
	}
	for _, a := range []any{ChargePayment, ChargeFlakyPayment, ChargeDeclinedCard} {
		r.RegisterActivity(a)
	}
}

var order = Order{OrderID: "o-1"}

func execute(t *testing.T, tq string, wf any, args ...any) client.WorkflowRun {
	t.Helper()
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("timeouts"), TaskQueue: tq,
	}, wf, args...)
	if err != nil {
		t.Fatal(err)
	}
	return run
}

func result(t *testing.T, run client.WorkflowRun) string {
	t.Helper()
	var out string
	if err := run.Get(testenv.Ctx(t), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func history(t *testing.T, run client.WorkflowRun) []*historypb.HistoryEvent {
	t.Helper()
	var events []*historypb.HistoryEvent
	iter := testenv.Client().GetWorkflowHistory(testenv.Ctx(t), run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		events = append(events, e)
	}
	return events
}

func count(events []*historypb.HistoryEvent, typ enumspb.EventType) int {
	n := 0
	for _, e := range events {
		if e.EventType == typ {
			n++
		}
	}
	return n
}

// waitForTaskFailure polls history until a WorkflowTaskFailed appears, and
// returns its failure message.
func waitForTaskFailure(t *testing.T, run client.WorkflowRun) string {
	t.Helper()
	for range 50 {
		for _, e := range history(t, run) {
			if a := e.GetWorkflowTaskFailedEventAttributes(); a != nil {
				return a.GetFailure().GetMessage()
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("no WorkflowTaskFailed event")
	return ""
}

func status(t *testing.T, run client.WorkflowRun) enumspb.WorkflowExecutionStatus {
	t.Helper()
	d, err := testenv.Client().DescribeWorkflowExecution(testenv.Ctx(t), run.GetID(), run.GetRunID())
	if err != nil {
		t.Fatal(err)
	}
	return d.WorkflowExecutionInfo.Status
}

func terminate(t *testing.T, run client.WorkflowRun) {
	t.Helper()
	if err := testenv.Client().TerminateWorkflow(testenv.Ctx(t), run.GetID(), "", "test done"); err != nil {
		t.Fatal(err)
	}
}

func TestTheClientSetsTheWorkflowExecutionTimeout(t *testing.T) {
	tq := testenv.Worker(t, register)
	run, err := StartOrder(testenv.Ctx(t), testenv.Client(), order, tq, testenv.ID("order"))
	if err != nil {
		t.Fatal(err)
	}
	if got := result(t, run); got != "charged-o-1" {
		t.Fatalf("got %q", got)
	}
	d, err := testenv.Client().DescribeWorkflowExecution(testenv.Ctx(t), run.GetID(), run.GetRunID())
	if err != nil {
		t.Fatal(err)
	}
	if got := d.ExecutionConfig.WorkflowExecutionTimeout.AsDuration(); got != 30*24*time.Hour {
		t.Fatalf("Workflow Execution Timeout %v", got)
	}
}

func TestTheWorkflowSetsTheActivityStartToCloseTimeout(t *testing.T) {
	tq := testenv.Worker(t, register)
	run := execute(t, tq, OrderWorkflow, order)
	if got := result(t, run); got != "charged-o-1" {
		t.Fatalf("got %q", got)
	}
	for _, e := range history(t, run) {
		if a := e.GetActivityTaskScheduledEventAttributes(); a != nil {
			if got := a.StartToCloseTimeout.AsDuration(); got != 10*time.Second {
				t.Fatalf("Start-To-Close %v", got)
			}
			return
		}
	}
	t.Fatal("no ActivityTaskScheduled event")
}

// Maximum Attempts: failed attempts are not Events; only the last one is, with
// its attempt number.
func TestMaximumAttemptsLeavesOnlyTheLastAttemptInHistory(t *testing.T) {
	tq := testenv.Worker(t, register)
	run := execute(t, tq, MaxAttemptsWorkflow, order)
	// An errors.New error reaches the Workflow as an ApplicationError with no Type.
	if got := result(t, run); got != "MaximumAttemptsReached:" {
		t.Fatalf("got %q", got)
	}
	events := history(t, run)
	if n := count(events, enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED); n != 1 {
		t.Fatalf("%d ActivityTaskScheduled", n)
	}
	if n := count(events, enumspb.EVENT_TYPE_ACTIVITY_TASK_STARTED); n != 1 {
		t.Fatalf("%d ActivityTaskStarted", n)
	}
	if n := count(events, enumspb.EVENT_TYPE_ACTIVITY_TASK_FAILED); n != 1 {
		t.Fatalf("%d ActivityTaskFailed", n)
	}
	for _, e := range events {
		if a := e.GetActivityTaskStartedEventAttributes(); a != nil && a.Attempt != 3 {
			t.Fatalf("ActivityTaskStarted attempt %d, want 3", a.Attempt)
		}
	}
}

func TestNonRetryableErrorTypesStopsAfterOneAttemptAndKeepsTheType(t *testing.T) {
	tq := testenv.Worker(t, register)
	run := execute(t, tq, NonRetryableWorkflow, order)
	if got := result(t, run); got != "NonRetryableFailure:CardDeclined" {
		t.Fatalf("got %q", got)
	}
	for _, e := range history(t, run) {
		if a := e.GetActivityTaskStartedEventAttributes(); a != nil && a.Attempt != 1 {
			t.Fatalf("ActivityTaskStarted attempt %d, want 1", a.Attempt)
		}
	}
}

func TestAnActivityWithNoStartOrScheduleToCloseFailsTheWorkflowTask(t *testing.T) {
	tq := testenv.Worker(t, register)
	run := execute(t, tq, NoTimeoutWorkflow, order)
	msg := waitForTaskFailure(t, run)
	t.Log(msg)
	if !strings.Contains(msg, "StartToClose") {
		t.Fatalf("Workflow Task failure %q does not name the missing timeout", msg)
	}
	if s := status(t, run); s != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
		t.Fatalf("status %v, want RUNNING", s)
	}
	terminate(t, run)
}

func TestTheDeadlockDetectorFailsAWorkflowTaskThatNeverYields(t *testing.T) {
	tq := testenv.Worker(t, register)
	run := execute(t, tq, BusyWorkflow)
	msg := waitForTaskFailure(t, run)
	t.Log(msg)
	if !strings.Contains(msg, "deadlock") {
		t.Fatalf("Workflow Task failure %q is not the deadlock detector", msg)
	}
	if s := status(t, run); s != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
		t.Fatalf("status %v, want RUNNING", s)
	}
	terminate(t, run)
}
