package updates

import (
	"context"
	"sync"
)

// The Activity the Update handler in 11-updates.md waits on. It records every
// amount it debits so the tests can see which Updates reached the handler, and
// can be held shut so a test controls when it finishes.
var (
	mu      sync.Mutex
	debited []int
	gate    = closed()
)

func closed() chan struct{} {
	c := make(chan struct{})
	close(c)
	return c
}

// Hold makes every DebitAccount call wait until Release.
func Hold() {
	mu.Lock()
	defer mu.Unlock()
	gate = make(chan struct{})
}

// Release lets held DebitAccount calls finish.
func Release() {
	mu.Lock()
	defer mu.Unlock()
	select {
	case <-gate:
	default:
		close(gate)
	}
}

// Debited returns the amounts debited so far, and Reset clears them.
func Debited() []int {
	mu.Lock()
	defer mu.Unlock()
	return append([]int(nil), debited...)
}

func Reset() {
	mu.Lock()
	defer mu.Unlock()
	debited = nil
}

func DebitAccount(ctx context.Context, amount int) error {
	mu.Lock()
	g := gate
	mu.Unlock()
	select {
	case <-g:
	case <-ctx.Done():
		return ctx.Err()
	}
	mu.Lock()
	defer mu.Unlock()
	debited = append(debited, amount)
	return nil
}
