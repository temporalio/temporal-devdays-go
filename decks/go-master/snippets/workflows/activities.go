package workflows

import (
	"context"
	"fmt"
	"sync"

	"go.temporal.io/sdk/temporal"
)

// Activities are the bank calls the transfer Workflow on the anatomy slide
// makes. Not on a slide: an in-memory ledger stands in for the bank, so the
// tests can check balances and close an account.
type Activities struct {
	mu       sync.Mutex
	Balances map[string]int
	Closed   map[string]bool
}

func NewActivities() *Activities {
	return &Activities{Balances: map[string]int{}, Closed: map[string]bool{}}
}

func (a *Activities) move(account string, amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Balances[account] += amount
}

// CheckFraud reports true when the transfer looks fraudulent.
func (a *Activities) CheckFraud(ctx context.Context, t Transfer) (bool, error) {
	return t.Amount > 10_000, nil
}

func (a *Activities) Withdraw(ctx context.Context, t Transfer) error {
	a.move(t.FromAccount, -t.Amount)
	return nil
}

func (a *Activities) Deposit(ctx context.Context, t Transfer) error {
	a.mu.Lock()
	closed := a.Closed[t.ToAccount]
	a.mu.Unlock()
	if closed {
		return temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("account %s is closed", t.ToAccount), "AccountClosed", nil)
	}
	a.move(t.ToAccount, t.Amount)
	return nil
}

func (a *Activities) Refund(ctx context.Context, t Transfer) error {
	a.move(t.FromAccount, t.Amount)
	return nil
}
