package continueasnew

import (
	"errors"
	"slices"
	"testing"
	"time"

	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// The slide sleeps 30 days between charges, so these run in the SDK's
// in-memory test environment, which skips time instead of waiting it out. It
// also lets a test say what GetContinueAsNewSuggested returns, which a real
// dev server only does once the history is thousands of Events long.

func env() *testsuite.TestWorkflowEnvironment {
	var s testsuite.WorkflowTestSuite
	e := s.NewTestWorkflowEnvironment()
	e.RegisterWorkflow(SubscriptionWorkflow)
	e.RegisterActivity(ChargeMonthly)
	return e
}

func TestBillsMonthlyUntilCancelledAndCancellationReachesTheSleep(t *testing.T) {
	mu.Lock()
	charges = nil
	mu.Unlock()
	e := env()
	// Charges land on day 0, 30 and 60; cancel during the third month.
	e.RegisterDelayedCallback(e.CancelWorkflow, 75*24*time.Hour)
	e.ExecuteWorkflow(SubscriptionWorkflow, Subscription{CustomerID: "cust-1"})

	if !e.IsWorkflowCompleted() {
		t.Fatal("the subscription is still running")
	}
	var canceled *temporal.CanceledError
	if err := e.GetWorkflowError(); !errors.As(err, &canceled) {
		t.Fatalf("want a cancellation, got %v", err)
	}
	if got := Charges(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("charges %v", got)
	}
}

func TestContinuesAsNewWhenSuggestedCarryingOnlyTheState(t *testing.T) {
	mu.Lock()
	charges = nil
	mu.Unlock()
	e := env()
	e.SetContinueAsNewSuggested(true)
	e.ExecuteWorkflow(SubscriptionWorkflow, Subscription{CustomerID: "cust-2", MonthsBilled: 11})

	var can *workflow.ContinueAsNewError
	if err := e.GetWorkflowError(); !errors.As(err, &can) {
		t.Fatalf("want Continue-As-New, got %v", err)
	}
	if can.WorkflowType.Name != "SubscriptionWorkflow" {
		t.Fatalf("continued as %q", can.WorkflowType.Name)
	}
	// One charge in this run, then the new run starts with the month counter
	// carried forward in its input.
	if got := Charges(); !slices.Equal(got, []int{12}) {
		t.Fatalf("charges %v", got)
	}
	var next Subscription
	if err := converter.GetDefaultDataConverter().FromPayloads(can.Input, &next); err != nil {
		t.Fatal(err)
	}
	if next != (Subscription{CustomerID: "cust-2", MonthsBilled: 12}) {
		t.Fatalf("next run's input %+v", next)
	}
}
