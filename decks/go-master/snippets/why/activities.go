// Package why backs 01-why-temporal.md "The Temporal version": the two steps
// of a transfer, with no retry loop, lock or checkpoint. The bank is an
// in-memory fake; tests make it fail a set number of times.
package why

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Transfer struct {
	TransferID  string
	FromAccount string
	ToAccount   string
	Amount      int
}

// InsufficientFundsError's type name is the error type the Workflow's Retry
// Policy lists as non-retryable: the SDK reports a Go error by its type name.
type InsufficientFundsError struct {
	Account string
	Balance int
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("%s has %d", e.Account, e.Balance)
}

// FakeBank stands in for the bank's API client.
type FakeBank struct {
	mu                  sync.Mutex
	Balances            map[string]int
	DepositFailuresLeft int
	Withdrawals         int
	Deposits            int
}

func (b *FakeBank) Withdraw(_ context.Context, account string, amount int) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Withdrawals++
	if b.Balances[account] < amount {
		return &InsufficientFundsError{Account: account, Balance: b.Balances[account]}
	}
	b.Balances[account] -= amount
	return nil
}

func (b *FakeBank) Deposit(_ context.Context, account string, amount int) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Deposits++
	if b.DepositFailuresLeft > 0 {
		b.DepositFailuresLeft--
		return errors.New("bank unavailable")
	}
	b.Balances[account] += amount
	return nil
}

var bank = &FakeBank{Balances: map[string]int{}}

// part: 01-why-temporal.md "The Temporal version" 1
// activities.go
func Withdraw(ctx context.Context, t Transfer) error {
	return bank.Withdraw(ctx, t.FromAccount, t.Amount)
}

func Deposit(ctx context.Context, t Transfer) error {
	return bank.Deposit(ctx, t.ToAccount, t.Amount)
}

// end part
