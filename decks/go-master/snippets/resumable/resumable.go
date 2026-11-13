// Package resumable is the Workflow the Resumable Activity console and demo in
// 11-updates.md act out: a transfer to an unknown account parks instead of
// failing, and resumes when a Signal brings a corrected account.
package resumable

import (
	"errors"
	"regexp"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	RetryWithCorrectionSignal = "retryWithCorrection"
	StatusQuery               = "getStatus"
)

var validAccount = regexp.MustCompile(`^\d{8,12}$`)

func TransferWorkflow(ctx workflow.Context, transfer Transfer) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
	})

	status := "PENDING"
	account := transfer.ToAccount
	fix := ""

	if err := workflow.SetQueryHandler(ctx, StatusQuery, func() (string, error) {
		return status, nil
	}); err != nil {
		return "", err
	}
	corrections := workflow.GetSignalChannel(ctx, RetryWithCorrectionSignal)
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var acct string
			corrections.Receive(ctx, &acct)
			// Validate in the handler: a malformed correction is ignored.
			if validAccount.MatchString(acct) {
				fix = acct
			}
		}
	})

	for {
		status = "TRANSFERRING"
		attempt := transfer
		attempt.ToAccount = account
		var receipt string
		err := workflow.ExecuteActivity(ctx, ExecuteTransfer, attempt).Get(ctx, &receipt)
		if err == nil {
			status = "COMPLETED"
			return receipt, nil
		}
		// Park only on an Activity failure. A cancellation, or a bug in the
		// Workflow itself, is not bad data.
		var activityErr *temporal.ActivityError
		if !errors.As(err, &activityErr) || temporal.IsCanceledError(err) {
			return "", err
		}
		status = "AWAITING_CORRECTION"
		if err := workflow.Await(ctx, func() bool { return fix != "" }); err != nil {
			return "", err
		}
		account, fix = fix, ""
	}
}
