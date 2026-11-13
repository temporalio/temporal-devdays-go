// ABOUTME: Names and shapes shared by the Workflow, the Activity, the Worker, the
// starter and the test.
package subscription

import "time"

// TaskQueue is the Task Queue this Workshop's Worker polls and its starter targets.
const TaskQueue = "subscription-tasks"

// Cycle is one billing cycle. Compressed so the history grows fast enough to watch.
const Cycle = time.Second

// SubscriptionInput is the input of every run in a Continue-As-New chain.
// Whatever a run has done that is not in the input it hands to the next run is
// gone.
type SubscriptionInput struct {
	// SubscriptionID is the subscription being billed.
	SubscriptionID string
	// CyclesCompleted is how many billing cycles earlier runs already charged.
	CyclesCompleted int
	// MaxCycles stops the subscription after this many billing cycles; 0 means
	// bill forever.
	MaxCycles int
}

// SubscriptionResult is what the last run of the chain returns.
type SubscriptionResult struct {
	SubscriptionID string
	TotalCycles    int
}
