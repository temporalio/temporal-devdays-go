package tuning

import (
	"fmt"
	"strings"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func execute(t *testing.T, tq string) string {
	t.Helper()
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("order"), TaskQueue: tq,
	}, OrderWorkflow, "A-1")
	if err != nil {
		t.Fatal(err)
	}
	var out string
	if err := run.Get(testenv.Ctx(t), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestAWorkerWithAResourceBasedTunerRunsWorkflowsAndActivities(t *testing.T) {
	w, err := NewTunedWorker(testenv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()
	if got := execute(t, "order-task-queue"); got != "submitted A-1" {
		t.Fatalf("got %q", got)
	}
}

func TestATunerCannotBeCombinedWithFixedSlotCounts(t *testing.T) {
	tuner, err := worker.NewFixedSizeTuner(worker.FixedSizeTunerOptions{NumWorkflowSlots: 40})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		msg := fmt.Sprint(recover())
		if !strings.Contains(msg, "cannot set MaxConcurrentWorkflowTaskExecutionSize") || !strings.Contains(msg, "with Tuner") {
			t.Fatalf("worker.New did not refuse a Tuner with fixed slot counts; recovered %q", msg)
		}
	}()
	worker.New(testenv.Client(), testenv.ID("tuning"), worker.Options{
		Tuner:                                  tuner,
		MaxConcurrentWorkflowTaskExecutionSize: 40,
	})
}

func TestTheOptionsOnTheSlidesBuildWorkingWorkers(t *testing.T) {
	register := func(r worker.Registry) {
		r.RegisterWorkflow(OrderWorkflow)
		r.RegisterActivity(SubmitOrder)
	}
	tq := testenv.Worker(t, register, FixedSlots)
	if got := execute(t, tq); got != "submitted A-1" {
		t.Fatalf("fixed slots: got %q", got)
	}

	tuner, err := Composite()
	if err != nil {
		t.Fatal(err)
	}
	tq = testenv.Worker(t, register, worker.Options{Tuner: tuner})
	if got := execute(t, tq); got != "submitted A-1" {
		t.Fatalf("composite tuner: got %q", got)
	}
}

func TestWorkflowCodeThatNeverYieldsFailsItsWorkflowTaskAndTheWorkflowStaysOpen(t *testing.T) {
	tq := testenv.Worker(t, func(r worker.Registry) { r.RegisterWorkflow(BusyWorkflow) })
	c := testenv.Client()
	ctx := testenv.Ctx(t)
	// A few seconds of spinning: well past the one-second deadlock detector.
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: testenv.ID("busy"), TaskQueue: tq}, BusyWorkflow, int64(4_000_000_000))
	if err != nil {
		t.Fatal(err)
	}
	defer c.TerminateWorkflow(ctx, run.GetID(), "", "test done")

	var failure string
	for deadline := time.Now().Add(20 * time.Second); failure == "" && time.Now().Before(deadline); time.Sleep(500 * time.Millisecond) {
		iter := c.GetWorkflowHistory(ctx, run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		for iter.HasNext() {
			e, err := iter.Next()
			if err != nil {
				t.Fatal(err)
			}
			if a := e.GetWorkflowTaskFailedEventAttributes(); a != nil {
				failure = a.GetFailure().GetMessage()
				break
			}
		}
	}
	if !strings.Contains(failure, "Potential deadlock detected") || !strings.Contains(failure, "didn't yield for over 1s") {
		t.Fatalf("Workflow Task failure %q, want the deadlock detector's", failure)
	}
	desc, err := c.DescribeWorkflowExecution(ctx, run.GetID(), run.GetRunID())
	if err != nil {
		t.Fatal(err)
	}
	if s := desc.GetWorkflowExecutionInfo().GetStatus(); s != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
		t.Fatalf("status %v, want RUNNING", s)
	}
}
