// Package designpatterns backs 04-design-patterns.md and the "Keep it in the
// Workflow" slides of 05-workflows.md.
package designpatterns

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type RFIRequest struct {
	RFIID         string
	ProjectID     string
	SubmittedAt   time.Time
	EstimatedCost int
}

type Applicant struct {
	Name   string
	Income int
}

type CreditCheck struct{ Score int }

type FraudCheck struct{ Flagged bool }

type Decision string

type Claim struct {
	ClaimID string
	Amount  int
}

type Approval struct {
	ClaimID  string
	Approved bool
}

type UserData struct{ Email string }

type User struct {
	ID    int
	Email string
}

// opts are the Activity options the slides' Workflows set before they call an
// Activity. Short retries keep the tests fast.
var opts = workflow.ActivityOptions{
	StartToCloseTimeout: time.Minute,
	RetryPolicy:         &temporal.RetryPolicy{InitialInterval: 200 * time.Millisecond},
}
