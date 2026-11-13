package resumable

import (
	"slices"
	"testing"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func status(t *testing.T, id string) string {
	t.Helper()
	v, err := testenv.Client().QueryWorkflow(testenv.Ctx(t), id, "", StatusQuery)
	if err != nil {
		t.Fatal(err)
	}
	var s string
	if err := v.Get(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestParksOnABadAccountIgnoresAMalformedFixAndResumes(t *testing.T) {
	Reset()
	tq := testenv.Worker(t, func(r worker.Registry) {
		r.RegisterWorkflow(TransferWorkflow)
		r.RegisterActivity(ExecuteTransfer)
	})
	c := testenv.Client()
	run, err := c.ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: testenv.ID("transfer"), TaskQueue: tq},
		TransferWorkflow, Transfer{FromAccount: "10000001", ToAccount: "99999999", Amount: 500})
	if err != nil {
		t.Fatal(err)
	}
	id := run.GetID()

	deadline := time.Now().Add(10 * time.Second)
	for status(t, id) != "AWAITING_CORRECTION" {
		if time.Now().After(deadline) {
			t.Fatal("the transfer never parked")
		}
		time.Sleep(100 * time.Millisecond)
	}
	// Non-retryable: one attempt, not the Retry Policy's three.
	if got := Attempts(); !slices.Equal(got, []string{"99999999"}) {
		t.Fatalf("attempts %v", got)
	}

	// The handler drops a correction in the wrong format; the Workflow stays parked.
	if err := c.SignalWorkflow(testenv.Ctx(t), id, "", RetryWithCorrectionSignal, "not-an-account"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	if s := status(t, id); s != "AWAITING_CORRECTION" {
		t.Fatalf("status %q after a malformed fix", s)
	}
	if got := Attempts(); !slices.Equal(got, []string{"99999999"}) {
		t.Fatalf("attempts %v", got)
	}

	if err := c.SignalWorkflow(testenv.Ctx(t), id, "", RetryWithCorrectionSignal, "12345678"); err != nil {
		t.Fatal(err)
	}
	var receipt string
	if err := run.Get(testenv.Ctx(t), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt != "Transfer of 500 to 12345678 completed" {
		t.Fatalf("receipt %q", receipt)
	}
	if got := Attempts(); !slices.Equal(got, []string{"99999999", "12345678"}) {
		t.Fatalf("attempts %v", got)
	}
	// A Query on the finished Workflow reports where it ended, as the slide shows.
	if s := status(t, id); s != "COMPLETED" {
		t.Fatalf("final status %q", s)
	}
	// Same Workflow Execution throughout: no restart, no new run.
	desc, err := c.DescribeWorkflowExecution(testenv.Ctx(t), id, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := desc.WorkflowExecutionInfo.Execution.RunId; got != run.GetRunID() {
		t.Fatalf("run ID changed from %s to %s", run.GetRunID(), got)
	}
}
