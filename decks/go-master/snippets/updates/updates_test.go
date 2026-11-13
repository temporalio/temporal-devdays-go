package updates

import (
	"slices"
	"testing"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func register(r worker.Registry) {
	r.RegisterWorkflow(AccountWorkflow)
	r.RegisterWorkflow(LockedAccountWorkflow)
	r.RegisterActivity(DebitAccount)
}

// The tests share the package's Activity gate, so they run one at a time.
func setup(t *testing.T, wf any) (string, client.WorkflowRun) {
	t.Helper()
	Reset()
	Release()
	tq := testenv.Worker(t, register)
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: testenv.ID("account"), TaskQueue: tq}, wf, 100)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		Release()
		_ = testenv.Client().TerminateWorkflow(testenv.Ctx(t), run.GetID(), "", "test over")
	})
	return tq, run
}

func update(t *testing.T, id string, amount int, stage client.WorkflowUpdateStage) client.WorkflowUpdateHandle {
	t.Helper()
	h, err := testenv.Client().UpdateWorkflow(testenv.Ctx(t), client.UpdateWorkflowOptions{
		WorkflowID: id, UpdateName: WithdrawUpdate, Args: []any{amount}, WaitForStage: stage,
	})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func balance(t *testing.T, h client.WorkflowUpdateHandle) (int, error) {
	t.Helper()
	var b int
	err := h.Get(testenv.Ctx(t), &b)
	return b, err
}

func mustBalance(t *testing.T, h client.WorkflowUpdateHandle) int {
	t.Helper()
	b, err := balance(t, h)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestAnUpdateReturnsAValueAndTheValidatorRejectsFirst(t *testing.T) {
	_, run := setup(t, AccountWorkflow)
	id := run.GetID()
	c := testenv.Client()
	if b := mustBalance(t, update(t, id, 30, client.WorkflowUpdateStageCompleted)); b != 70 {
		t.Fatalf("balance %d", b)
	}
	// The Validator rejects: the caller gets an error.
	h, err := c.UpdateWorkflow(testenv.Ctx(t), client.UpdateWorkflowOptions{
		WorkflowID: id, UpdateName: WithdrawUpdate, Args: []any{500}, WaitForStage: client.WorkflowUpdateStageCompleted,
	})
	if err == nil {
		_, err = balance(t, h)
	}
	if err == nil {
		t.Fatal("an overdraft was accepted")
	}
	if b := mustBalance(t, update(t, id, 20, client.WorkflowUpdateStageCompleted)); b != 50 {
		t.Fatalf("balance %d", b)
	}
	b, err := WithdrawFifty(testenv.Ctx(t), c, id)
	if err != nil || b != 0 {
		t.Fatalf("WithdrawFifty: %d, %v", b, err)
	}
	if err := run.Get(testenv.Ctx(t), nil); err != nil {
		t.Fatal(err)
	}

	// The rejected Update never reached the handler, so never debited the
	// account, and left nothing in Event History.
	if got := Debited(); !slices.Equal(got, []int{30, 20, 50}) {
		t.Fatalf("debited %v", got)
	}
	accepted, completed := 0, 0
	iter := c.GetWorkflowHistory(testenv.Ctx(t), id, run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		if e.GetWorkflowExecutionUpdateAcceptedEventAttributes() != nil {
			accepted++
		}
		if e.GetWorkflowExecutionUpdateCompletedEventAttributes() != nil {
			completed++
		}
	}
	if accepted != 3 || completed != 3 {
		t.Fatalf("%d accepted, %d completed; want 3 and 3", accepted, completed)
	}
}

func TestWaitingForAcceptedReturnsBeforeTheResult(t *testing.T) {
	_, run := setup(t, AccountWorkflow)
	Hold()
	h := update(t, run.GetID(), 10, client.WorkflowUpdateStageAccepted)
	// Accepted, but the handler is still waiting on its Activity.
	if got := Debited(); len(got) != 0 {
		t.Fatalf("debited %v before release", got)
	}
	Release()
	if b := mustBalance(t, h); b != 90 {
		t.Fatalf("balance %d", b)
	}
}

func TestTwoConcurrentWithdrawalsCanBothPassTheValidator(t *testing.T) {
	_, run := setup(t, AccountWorkflow)
	Hold()
	a := update(t, run.GetID(), 80, client.WorkflowUpdateStageAccepted)
	b := update(t, run.GetID(), 80, client.WorkflowUpdateStageAccepted)
	Release()
	got := []int{mustBalance(t, a), mustBalance(t, b)}
	slices.Sort(got)
	if !slices.Equal(got, []int{-60, 20}) {
		t.Fatalf("balances %v; want the overdraft [-60 20]", got)
	}
	if d := Debited(); !slices.Equal(d, []int{80, 80}) {
		t.Fatalf("debited %v", d)
	}
}

func TestAMutexAndARecheckStopTheOverdraft(t *testing.T) {
	_, run := setup(t, LockedAccountWorkflow)
	Hold()
	a := update(t, run.GetID(), 80, client.WorkflowUpdateStageAccepted)
	b := update(t, run.GetID(), 80, client.WorkflowUpdateStageAccepted)
	Release()
	if got := mustBalance(t, a); got != 20 {
		t.Fatalf("first withdrawal left %d", got)
	}
	if _, err := balance(t, b); err == nil {
		t.Fatal("the second withdrawal succeeded")
	}
	if d := Debited(); !slices.Equal(d, []int{80}) {
		t.Fatalf("debited %v", d)
	}
	// Only the Update failed; the Workflow is still running.
	desc, err := testenv.Client().DescribeWorkflowExecution(testenv.Ctx(t), run.GetID(), "")
	if err != nil {
		t.Fatal(err)
	}
	if s := desc.WorkflowExecutionInfo.Status; s != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
		t.Fatalf("status %v", s)
	}
}
