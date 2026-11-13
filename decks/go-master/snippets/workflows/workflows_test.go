package workflows

import (
	"context"
	"slices"
	"strings"
	"testing"

	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

var transfer = Transfer{TransferID: "tx-1", FromAccount: "acct-a", ToAccount: "acct-b", Amount: 250}

func runTransfer(t *testing.T, bank *Activities, tr Transfer) Receipt {
	t.Helper()
	tq := testenv.Worker(t, func(r worker.Registry) {
		r.RegisterWorkflow(TransferWorkflow)
		r.RegisterActivity(bank)
	})
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("transfer"), TaskQueue: tq,
	}, TransferWorkflow, tr)
	if err != nil {
		t.Fatal(err)
	}
	var receipt Receipt
	if err := run.Get(testenv.Ctx(t), &receipt); err != nil {
		t.Fatal(err)
	}
	return receipt
}

func TestTheTransferWorkflowMovesTheMoneyAndReturnsAReceipt(t *testing.T) {
	bank := NewActivities()
	if got := runTransfer(t, bank, transfer); got.Status != "completed" {
		t.Fatalf("receipt %+v", got)
	}
	if bank.Balances["acct-a"] != -250 || bank.Balances["acct-b"] != 250 {
		t.Fatalf("balances %v", bank.Balances)
	}
}

func TestAFlaggedTransferIsRejectedBeforeAnyMoneyMoves(t *testing.T) {
	bank := NewActivities()
	tr := transfer
	tr.Amount = 50_000
	if got := runTransfer(t, bank, tr); got.Status != "rejected" {
		t.Fatalf("receipt %+v", got)
	}
	if len(bank.Balances) != 0 {
		t.Fatalf("money moved: %v", bank.Balances)
	}
}

func TestADepositThatFailsForGoodIsCompensatedWithARefund(t *testing.T) {
	bank := NewActivities()
	bank.Closed["acct-b"] = true
	got := runTransfer(t, bank, transfer)
	if got.Status != "refunded" || !strings.Contains(got.Reason, "account acct-b is closed") {
		t.Fatalf("receipt %+v", got)
	}
	if bank.Balances["acct-a"] != 0 {
		t.Fatalf("source account not refunded: %v", bank.Balances)
	}
	if _, ok := bank.Balances["acct-b"]; ok {
		t.Fatalf("money reached the closed account: %v", bank.Balances)
	}
}

func TestTheWorkflowTypeIsTheFunctionName(t *testing.T) {
	tq := testenv.Worker(t, func(r worker.Registry) {
		r.RegisterWorkflow(TransferWorkflow)
		r.RegisterActivity(NewActivities())
	})
	// Started by name alone, as a Client in another codebase would.
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("transfer"), TaskQueue: tq,
	}, "TransferWorkflow", transfer)
	if err != nil {
		t.Fatal(err)
	}
	var receipt Receipt
	if err := run.Get(testenv.Ctx(t), &receipt); err != nil || receipt.Status != "completed" {
		t.Fatalf("receipt %+v, err %v", receipt, err)
	}
}

// The determinism slide: the Workflow runs, ranges over the map in sorted
// order, and its Event History replays without a non-determinism error.
func TestTheDeterministicWorkflowRunsAndReplays(t *testing.T) {
	tq := testenv.Worker(t, func(r worker.Registry) { r.RegisterWorkflow(PackingWorkflow) })
	items := map[string]int{"sku-c": 3, "sku-a": 1, "sku-b": 2}
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("packing"), TaskQueue: tq,
	}, PackingWorkflow, items)
	if err != nil {
		t.Fatal(err)
	}
	var p Packing
	if err := run.Get(testenv.Ctx(t), &p); err != nil {
		t.Fatal(err)
	}
	if want := []string{"sku-a x1", "sku-b x2", "sku-c x3"}; !slices.Equal(p.Lines, want) {
		t.Fatalf("lines %v, want %v", p.Lines, want)
	}
	if p.Box != "box-1" || !strings.HasPrefix(p.Label, "PKG-") || p.Started.IsZero() {
		t.Fatalf("packing %+v", p)
	}

	history := fetchHistory(t, run)
	// The random label is recorded once, as a marker, and read back on replay.
	markers := 0
	for _, e := range history.Events {
		if e.EventType == enumspb.EVENT_TYPE_MARKER_RECORDED {
			markers++
		}
	}
	if markers != 1 {
		t.Fatalf("%d MarkerRecorded events, want 1 for the SideEffect", markers)
	}
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(PackingWorkflow)
	for i := 0; i < 3; i++ {
		if err := replayer.ReplayWorkflowHistory(testenv.Quiet, history); err != nil {
			t.Fatalf("replay %d: %v", i+1, err)
		}
	}
}

func fetchHistory(t *testing.T, run client.WorkflowRun) *historypb.History {
	t.Helper()
	iter := testenv.Client().GetWorkflowHistory(context.Background(), run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	h := &historypb.History{}
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		h.Events = append(h.Events, e)
	}
	return h
}
