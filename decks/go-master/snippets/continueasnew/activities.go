package continueasnew

import (
	"context"
	"sync"
)

type Subscription struct {
	CustomerID   string
	MonthsBilled int
}

var (
	mu      sync.Mutex
	charges []int
)

// Charges returns every charge so far, by month number, across all runs.
func Charges() []int {
	mu.Lock()
	defer mu.Unlock()
	return append([]int(nil), charges...)
}

func ChargeMonthly(ctx context.Context, sub Subscription) error {
	mu.Lock()
	defer mu.Unlock()
	charges = append(charges, sub.MonthsBilled+1)
	return nil
}
