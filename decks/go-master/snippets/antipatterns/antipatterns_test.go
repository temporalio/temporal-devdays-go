package antipatterns

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

// The business-ID and idempotency tests run on a local dev server. The polling
// and three-day-hold slides wait for minutes and days, so those run in the
// SDK's in-memory test environment, which skips time.
func TestMain(m *testing.M) { testenv.Main(m) }

func register(r worker.Registry) {
	r.RegisterWorkflow(TransferWorkflow)
	r.RegisterWorkflow(DebitWorkflow)
	r.RegisterActivity(PostDebit)
}

func terminate(runs ...client.WorkflowRun) {
	for _, r := range runs {
		_ = testenv.Client().TerminateWorkflow(context.Background(), r.GetID(), "", "test over")
	}
}

func TestARandomWorkflowIDLetsADoubleTapStartTwoTransfers(t *testing.T) {
	testenv.WorkerOn(t, "payments", register)
	ctx, c := testenv.Ctx(t), testenv.Client()
	transfer := Transfer{TransferID: testenv.ID("t")}
	a, err := StartWithRandomID(ctx, c, transfer)
	if err != nil {
		t.Fatal(err)
	}
	b, err := StartWithRandomID(ctx, c, transfer)
	if err != nil {
		t.Fatal(err)
	}
	defer terminate(a, b)
	if a.GetID() == b.GetID() || a.GetRunID() == b.GetRunID() {
		t.Fatal("two starts with random IDs should be two transfers")
	}
}

func TestTheBusinessIDGivesTheSecondStartTheRunningTransfer(t *testing.T) {
	testenv.WorkerOn(t, "payments", register)
	ctx, c := testenv.Ctx(t), testenv.Client()
	transfer := Transfer{TransferID: testenv.ID("t")}
	first, err := StartWithTransferID(ctx, c, transfer)
	if err != nil {
		t.Fatal(err)
	}
	defer terminate(first)
	second, err := StartWithTransferID(ctx, c, transfer)
	if err != nil {
		t.Fatal(err)
	}
	if second.GetRunID() != first.GetRunID() {
		t.Fatalf("the second start began run %s beside %s", second.GetRunID(), first.GetRunID())
	}

	// Asked to, ExecuteWorkflow reports the collision as an error instead.
	_, err = c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID: "transfer-" + transfer.TransferID, TaskQueue: "payments",
		WorkflowExecutionErrorWhenAlreadyStarted: true,
	}, TransferWorkflow, transfer)
	var already *serviceerror.WorkflowExecutionAlreadyStarted
	if !errors.As(err, &already) {
		t.Fatalf("want WorkflowExecutionAlreadyStarted, got %v", err)
	}
}

func TestTheRunIDAndActivityIDKeyIsTheSameOnTheRetry(t *testing.T) {
	ledger.reset()
	ledger.failOnceAfterCommit()
	tq := testenv.Worker(t, register)
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("debit"), TaskQueue: tq,
	}, DebitWorkflow, Debit{AccountID: "acct-1", Amount: 80})
	if err != nil {
		t.Fatal(err)
	}
	if err := run.Get(testenv.Ctx(t), nil); err != nil {
		t.Fatal(err)
	}
	ledger.Lock()
	defer ledger.Unlock()
	if len(ledger.requests) != 2 || ledger.requests[0] != ledger.requests[1] {
		t.Fatalf("want two requests with one key, got %v", ledger.requests)
	}
	if len(ledger.postings) != 1 {
		t.Fatalf("want one posting, got %d", len(ledger.postings))
	}
}

// timeSkipping runs wf in the in-memory test environment and returns the
// durations of the Timers it started.
func timeSkipping(t *testing.T, wf any, arg any) []time.Duration {
	t.Helper()
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(CheckSettlement)
	env.RegisterActivity(AwaitSettlement)
	env.RegisterActivity(ReleaseFunds)
	var timers []time.Duration
	env.SetOnTimerScheduledListener(func(_ string, d time.Duration) { timers = append(timers, d) })
	env.ExecuteWorkflow(wf, arg)
	if !env.IsWorkflowCompleted() {
		t.Fatal("the Workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatal(err)
	}
	return timers
}

func activityIDs(cs []check) []string {
	var ids []string
	for _, c := range cs {
		if !slices.Contains(ids, c.activityID) {
			ids = append(ids, c.activityID)
		}
	}
	return ids
}

func TestPollingInALoopAddsAnActivityAndATimerOnEveryPass(t *testing.T) {
	settleOnCheck(4)
	timers := timeSkipping(t, LoopPollingWorkflow, "wire-1")
	cs := checks()
	if len(cs) != 4 || len(activityIDs(cs)) != 4 {
		t.Fatalf("want four checks as four Activities, got %+v", cs)
	}
	if len(timers) != 3 {
		t.Fatalf("want three Timers, got %v", timers)
	}
}

func TestPollingWithTheRetryPolicyKeepsOneActivity(t *testing.T) {
	settleOnCheck(4)
	timers := timeSkipping(t, RetryPollingWorkflow, "wire-2")
	cs := checks()
	if len(cs) != 4 || len(activityIDs(cs)) != 1 {
		t.Fatalf("want four checks as attempts of one Activity, got %+v", cs)
	}
	if cs[3].attempt != 4 {
		t.Fatalf("want the last check to be attempt 4, got %d", cs[3].attempt)
	}
	if len(timers) != 0 {
		t.Fatalf("want no Timers, got %v", timers)
	}
}

func TestAThreeDayHoldIsOneDurableTimer(t *testing.T) {
	released.Lock()
	released.ids = nil
	released.Unlock()
	timers := timeSkipping(t, DepositHoldWorkflow, Deposit{DepositID: "dep-1"})
	if !slices.Equal(timers, []time.Duration{72 * time.Hour}) {
		t.Fatalf("want one three-day Timer, got %v", timers)
	}
	released.Lock()
	defer released.Unlock()
	if !slices.Equal(released.ids, []string{"dep-1"}) {
		t.Fatalf("released %v", released.ids)
	}
}
