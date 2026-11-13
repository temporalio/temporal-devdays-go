package queries

import (
	"testing"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func historyLength(t *testing.T, id, runID string) int {
	t.Helper()
	iter := testenv.Client().GetWorkflowHistory(testenv.Ctx(t), id, runID, false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	n := 0
	for iter.HasNext() {
		if _, err := iter.Next(); err != nil {
			t.Fatal(err)
		}
		n++
	}
	return n
}

func TestAQueryReadsStateAndAddsNothingToEventHistory(t *testing.T) {
	tq := testenv.Worker(t, func(r worker.Registry) { r.RegisterWorkflow(LoanWorkflow) })
	c := testenv.Client()
	run, err := c.ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: testenv.ID("loan"), TaskQueue: tq}, LoanWorkflow, Loan{LoanID: "loan-1"})
	if err != nil {
		t.Fatal(err)
	}
	var res LoanResult
	if err := run.Get(testenv.Ctx(t), &res); err != nil {
		t.Fatal(err)
	}
	if res != (LoanResult{LoanID: "loan-1", Status: "approved"}) {
		t.Fatalf("result %+v", res)
	}
	before := historyLength(t, run.GetID(), run.GetRunID())

	// Queries still answer after the Workflow has completed.
	for range 5 {
		v, err := c.QueryWorkflow(testenv.Ctx(t), run.GetID(), run.GetRunID(), LoanStatusQuery)
		if err != nil {
			t.Fatal(err)
		}
		var status string
		if err := v.Get(&status); err != nil {
			t.Fatal(err)
		}
		if status != "approved" {
			t.Fatalf("status %q", status)
		}
	}

	if after := historyLength(t, run.GetID(), run.GetRunID()); after != before {
		t.Fatalf("history grew from %d to %d events", before, after)
	}
}
