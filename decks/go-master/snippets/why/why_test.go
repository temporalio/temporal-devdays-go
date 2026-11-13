// 01-why-temporal "The Temporal version": retries are configuration. The
// Activities have no retry loop, yet a bank that fails the deposit twice still
// completes the transfer, with one withdrawal. Insufficient funds is listed as
// non-retryable, so it fails at once.
package why

import (
	"errors"
	"testing"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

var transfer = Transfer{TransferID: "4471", FromAccount: "alice", ToAccount: "bob", Amount: 500}

func register(r worker.Registry) {
	r.RegisterWorkflow(TransferWorkflow)
	r.RegisterActivity(Withdraw)
	r.RegisterActivity(Deposit)
}

// The tests share the package-level fake bank, so they do not run in parallel.
func reset(alice int) {
	bank.mu.Lock()
	defer bank.mu.Unlock()
	bank.Balances = map[string]int{"alice": alice}
	bank.DepositFailuresLeft = 0
	bank.Withdrawals, bank.Deposits = 0, 0
}

func execute(t *testing.T) error {
	t.Helper()
	tq := testenv.Worker(t, register)
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("transfer"), TaskQueue: tq,
	}, TransferWorkflow, transfer)
	if err != nil {
		t.Fatal(err)
	}
	return run.Get(testenv.Ctx(t), nil)
}

func TestTheBankFailsTheDepositTwiceAndTheTransferCompletesWithOneWithdrawal(t *testing.T) {
	reset(1000)
	bank.DepositFailuresLeft = 2
	if err := execute(t); err != nil {
		t.Fatal(err)
	}
	if bank.Withdrawals != 1 || bank.Deposits != 3 {
		t.Fatalf("withdrawals=%d deposits=%d, want 1 and 3", bank.Withdrawals, bank.Deposits)
	}
	if bank.Balances["alice"] != 500 || bank.Balances["bob"] != 500 {
		t.Fatalf("balances %v", bank.Balances)
	}
}

func TestInsufficientFundsIsNotRetriedAndNothingIsDeposited(t *testing.T) {
	reset(100)
	err := execute(t)
	var wee *temporal.WorkflowExecutionError
	if !errors.As(err, &wee) {
		t.Fatalf("want a WorkflowExecutionError, got %v", err)
	}
	var appErr *temporal.ApplicationError
	if !errors.As(err, &appErr) || appErr.Type() != "InsufficientFundsError" {
		t.Fatalf("want an ApplicationError of type InsufficientFundsError, got %v", err)
	}
	if bank.Withdrawals != 1 || bank.Deposits != 0 {
		t.Fatalf("withdrawals=%d deposits=%d, want 1 and 0", bank.Withdrawals, bank.Deposits)
	}
}
