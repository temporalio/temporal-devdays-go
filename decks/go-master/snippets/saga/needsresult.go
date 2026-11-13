package saga

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// A Saga whose undo needs what its forward step returned: the refund needs the
// payment's confirmation, so it can only be registered after the payment. No
// slide shows this file; it backs the "register after" side of the comparison
// in 17-saga-04-when-to-register.md.

type TitleOrder struct {
	OrderID string
	FailAt  string
}

// journal is every step and undo that ran, in order.
var journal struct {
	sync.Mutex
	entries []string
}

func note(entry string) {
	journal.Lock()
	defer journal.Unlock()
	journal.entries = append(journal.entries, entry)
}

// Journal returns a copy of the journal, and ResetJournal clears it.
func Journal() []string {
	journal.Lock()
	defer journal.Unlock()
	return append([]string(nil), journal.entries...)
}

func ResetJournal() {
	journal.Lock()
	defer journal.Unlock()
	journal.entries = nil
}

func CheckFraud(ctx context.Context, order TitleOrder) error {
	note("checkFraud")
	return nil
}

func ProcessPayment(ctx context.Context, order TitleOrder) (string, error) {
	note("processPayment")
	return "conf-" + order.OrderID, nil
}

func RefundPayment(ctx context.Context, order TitleOrder, confirmation string) error {
	note("refundPayment:" + confirmation)
	return nil
}

func TransferTitle(ctx context.Context, order TitleOrder) error {
	note("transferTitle")
	if order.FailAt == "transferTitle" {
		return temporal.NewNonRetryableApplicationError("title transfer failed", "TitleTransferFailed", nil)
	}
	return nil
}

func RevertTitleTransfer(ctx context.Context, order TitleOrder) error {
	note("revertTitleTransfer")
	return nil
}

func TitleTransferWorkflow(ctx workflow.Context, order TitleOrder) (err error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})
	var compensations []Compensation
	defer func() {
		var activityErr *temporal.ActivityError
		if !errors.As(err, &activityErr) {
			return
		}
		ctx, _ := workflow.NewDisconnectedContext(ctx)
		for i := len(compensations) - 1; i >= 0; i-- {
			_ = compensations[i](ctx)
		}
	}()

	if err := workflow.ExecuteActivity(ctx, CheckFraud, order).Get(ctx, nil); err != nil {
		return err
	}

	// The undo needs the confirmation, so it is registered after the step.
	var confirmation string
	if err := workflow.ExecuteActivity(ctx, ProcessPayment, order).Get(ctx, &confirmation); err != nil {
		return err
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, RefundPayment, order, confirmation).Get(ctx, nil)
	})

	if err := workflow.ExecuteActivity(ctx, TransferTitle, order).Get(ctx, nil); err != nil {
		return err
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, RevertTitleTransfer, order).Get(ctx, nil)
	})
	return nil
}
