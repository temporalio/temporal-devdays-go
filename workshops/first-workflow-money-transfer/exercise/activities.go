// ABOUTME: The two steps of a transfer. Each one talks to the bank, so each is an
// Activity: code that can fail, and that Temporal retries for you.
// The bank calls are simulated: each prints a line and takes half a second.
package moneytransfer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
)

// Withdraw takes the money out of the source account.
func Withdraw(ctx context.Context, details PaymentDetails) (string, error) {
	attempt := activity.GetInfo(ctx).Attempt
	fmt.Printf("[activity] withdraw attempt %d: $%d from %s\n", attempt, details.Amount, details.SourceAccount)
	callTheBank()
	return "W-" + details.ReferenceID, nil
}

// Deposit puts the money into the target account.
func Deposit(ctx context.Context, details PaymentDetails) (string, error) {
	attempt := activity.GetInfo(ctx).Attempt
	fmt.Printf("[activity] deposit attempt %d: $%d to %s\n", attempt, details.Amount, details.TargetAccount)

	// TODO 1 start: write the audit note for a large deposit without crashing
	if details.Amount >= LargeDeposit {
		note := make([]string, 0, 2)
		note[0] = fmt.Sprintf("Large deposit of $%d", details.Amount)
		note[1] = "from " + details.SourceAccount
		fmt.Println("[activity] audit:", strings.Join(note, " "))
	}
	// TODO 1 end

	callTheBank()
	return "D-" + details.ReferenceID, nil
}

func callTheBank() {
	time.Sleep(500 * time.Millisecond)
}
