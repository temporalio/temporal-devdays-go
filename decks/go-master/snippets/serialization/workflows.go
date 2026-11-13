// Package serialization backs 19-serialization.md.
package serialization

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

type Result struct {
	PatientID     string `json:"patientId"`
	ApprovedCents int64  `json:"approvedCents"`
}

func ProcessClaim(ctx context.Context, claim Claim) (Result, error) {
	return Result{PatientID: claim.PatientID, ApprovedCents: claim.BilledAmountCents}, nil
}

var opts = workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second}

// slide: 19-serialization.md "heading: Data serialization"
type Claim struct {
	PatientID         string `json:"patientId"`
	DrugNDC           string `json:"drugNdc"`
	BilledAmountCents int64  `json:"billedAmountCents"`
}

func ClaimWorkflow(ctx workflow.Context, claim Claim) (Result, error) {
	ctx = workflow.WithActivityOptions(ctx, opts)
	// claim was JSON on the wire a moment ago
	var result Result
	err := workflow.ExecuteActivity(ctx, ProcessClaim, claim).Get(ctx, &result)

	// and this crosses the wire on the way back out
	return result, err
}

// end slide

// Visit is the argument of the JSON-traps slide.
type Visit struct {
	At    time.Time
	Claim map[string]any
	copay int
}

// Echo returns its argument, so the value crosses the wire twice.
func Echo(ctx workflow.Context, v Visit) (Visit, error) {
	return v, nil
}

// Inspect reports what the Workflow itself received: the Go type and value of
// the amount inside Claim.
func Inspect(ctx workflow.Context, v Visit) (string, error) {
	return fmt.Sprintf("%T %v", v.Claim["amount"], v.Claim["amount"]), nil
}
