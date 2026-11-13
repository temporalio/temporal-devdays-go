package antipatterns

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

// slide: 17-saga-07-related.md "An idempotency key that survives retries"
func PostDebit(ctx context.Context, debit Debit) (string, error) {
	info := activity.GetInfo(ctx)
	idempotencyKey := info.WorkflowExecution.RunID + "-" + info.ActivityID

	return ledger.Post(ctx, debit, idempotencyKey)
}

// end slide

// settlement is the payment network the polling slide waits on. A wire
// settles on the Nth status check; every check is recorded with the Activity
// ID and Attempt that made it.
var settlement struct {
	sync.Mutex
	settlesOnCheck int
	checks         []check
}

type check struct {
	activityID string
	attempt    int32
}

func settleOnCheck(n int) {
	settlement.Lock()
	defer settlement.Unlock()
	settlement.settlesOnCheck, settlement.checks = n, nil
}

func checks() []check {
	settlement.Lock()
	defer settlement.Unlock()
	return append([]check(nil), settlement.checks...)
}

func CheckSettlement(ctx context.Context, wireID string) (string, error) {
	info := activity.GetInfo(ctx)
	settlement.Lock()
	defer settlement.Unlock()
	settlement.checks = append(settlement.checks, check{info.ActivityID, info.Attempt})
	if len(settlement.checks) >= settlement.settlesOnCheck {
		return "SETTLED", nil
	}
	return "PENDING", nil
}

// AwaitSettlement is the ✅ half of the polling slide: one check per attempt,
// and a retryable failure while the wire has not settled, so the Retry Policy
// does the waiting.
func AwaitSettlement(ctx context.Context, wireID string) error {
	status, err := CheckSettlement(ctx, wireID)
	if err != nil {
		return err
	}
	if status != "SETTLED" {
		return temporal.NewApplicationError(fmt.Sprintf("wire %s not settled yet", wireID), "NotSettled")
	}
	return nil
}

// released records the deposits the ✅ half of the sleeping slide releases.
var released struct {
	sync.Mutex
	ids []string
}

func ReleaseFunds(ctx context.Context, depositID string) error {
	released.Lock()
	defer released.Unlock()
	released.ids = append(released.ids, depositID)
	return nil
}

// part: 17b-patterns-and-anti-patterns.md "Sleeping in an Activity" 1
// ❌ An Activity that waits out a three-day hold
func ReleaseAfterHold(ctx context.Context, deposit Deposit) error {
	time.Sleep(72 * time.Hour)
	return bank.Release(ctx, deposit.DepositID)
}

// end part
