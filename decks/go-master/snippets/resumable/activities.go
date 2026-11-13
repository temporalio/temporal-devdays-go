package resumable

import (
	"context"
	"fmt"
	"sync"

	"go.temporal.io/sdk/temporal"
)

// Transfer keeps lower-case JSON names so the demo's
// `temporal workflow start --input` line reads naturally.
type Transfer struct {
	FromAccount string `json:"fromAccount"`
	ToAccount   string `json:"toAccount"`
	Amount      int    `json:"amount"`
}

// KnownAccounts are the accounts the bank has. Any other account is a
// permanent failure: retrying the same input can never succeed, so it fails at
// once instead of burning the Retry Policy's attempts.
var KnownAccounts = []string{"10000001", "12345678"}

var (
	mu       sync.Mutex
	attempts []string
)

// Attempts returns the account of every ExecuteTransfer call so far; Reset
// clears them.
func Attempts() []string {
	mu.Lock()
	defer mu.Unlock()
	return append([]string(nil), attempts...)
}

func Reset() {
	mu.Lock()
	defer mu.Unlock()
	attempts = nil
}

func ExecuteTransfer(ctx context.Context, transfer Transfer) (string, error) {
	mu.Lock()
	attempts = append(attempts, transfer.ToAccount)
	mu.Unlock()
	for _, known := range KnownAccounts {
		if transfer.ToAccount == known {
			return fmt.Sprintf("Transfer of %d to %s completed", transfer.Amount, transfer.ToAccount), nil
		}
	}
	return "", temporal.NewNonRetryableApplicationError(
		fmt.Sprintf("Account %s not found", transfer.ToAccount), "AccountNotFoundError", nil)
}
