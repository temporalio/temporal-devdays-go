package signals

import (
	"testing"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func register(r worker.Registry) {
	r.RegisterWorkflow(WireTransferWorkflow)
	r.RegisterWorkflow(AsyncWireTransferWorkflow)
}

func start(t *testing.T, tq string, wf any, wireID string) client.WorkflowRun {
	t.Helper()
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID(wireID), TaskQueue: tq,
	}, wf, WireTransfer{WireID: wireID})
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

func signal(t *testing.T, run client.WorkflowRun, officer string) {
	t.Helper()
	if err := testenv.Client().SignalWorkflow(testenv.Ctx(t), run.GetID(), "", ApproveWireSignal, officer); err != nil {
		t.Fatal(err)
	}
}

func TestASignalUnblocksTheWaitingWorkflow(t *testing.T) {
	tq := testenv.Worker(t, register)
	run := start(t, tq, WireTransferWorkflow, "wire-1")
	signal(t, run, "alice")
	if got := result(t, run); got != "wire-1 approved by alice" {
		t.Fatalf("got %q", got)
	}
}

func TestTheSignalAndItsArgumentAreRecordedInEventHistory(t *testing.T) {
	tq := testenv.Worker(t, register)
	run := start(t, tq, WireTransferWorkflow, "wire-2")
	signal(t, run, "alice")
	result(t, run)
	iter := testenv.Client().GetWorkflowHistory(testenv.Ctx(t), run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		if a := e.GetWorkflowExecutionSignaledEventAttributes(); a != nil {
			if a.SignalName != "approveWire" {
				t.Fatalf("signal name %q", a.SignalName)
			}
			if got := string(a.Input.Payloads[0].Data); got != `"alice"` {
				t.Fatalf("signal input %s", got)
			}
			return
		}
	}
	t.Fatal("no WorkflowExecutionSignaled event")
}

func TestASignalIsAcceptedWhileNoWorkerRuns(t *testing.T) {
	tq := testenv.ID("signals")
	run := start(t, tq, WireTransferWorkflow, "wire-3")
	// No Worker polls this Task Queue yet; the Server records the Signal anyway.
	signal(t, run, "carol")
	testenv.WorkerOn(t, tq, register)
	if got := result(t, run); got != "wire-3 approved by carol" {
		t.Fatalf("got %q", got)
	}
}

func TestASignalCanBeHandledInACoroutine(t *testing.T) {
	tq := testenv.Worker(t, register)
	run := start(t, tq, AsyncWireTransferWorkflow, "wire-4")
	signal(t, run, "dave")
	if got := result(t, run); got != "wire-4 approved by dave" {
		t.Fatalf("got %q", got)
	}
}
