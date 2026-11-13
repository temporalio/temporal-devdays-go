package designpatterns

import (
	"os"
	"time"

	"go.temporal.io/sdk/workflow"
)

// part: 04-design-patterns.md "is worth testing in isolation" 2
// workflow.go
func EligibilityWorkflow(ctx workflow.Context, req RFIRequest) (bool, error) {
	ctx = workflow.WithActivityOptions(ctx, opts)

	var isEligible bool
	err := workflow.ExecuteActivity(ctx, ValidateEligibility, req).Get(ctx, &isEligible)
	return isEligible, err
}

// end part

// part: 04-design-patterns.md "can fail on its own" 2
// workflow.go
func ReviewWorkflow(ctx workflow.Context, req RFIRequest, reviewer string) error {
	ctx = workflow.WithActivityOptions(ctx, opts)

	return workflow.ExecuteActivity(ctx, NotifyReviewer, req, reviewer).Get(ctx, nil)
}

// end part

// part: 04-design-patterns.md "takes significant time" 2
// workflow.go
func AnnualReportWorkflow(ctx workflow.Context, req RFIRequest) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Hour,
	})

	return workflow.ExecuteActivity(ctx, GenerateAnnualReport, req).Get(ctx, nil)
}

// end part

// slide: 05-workflows.md "Keep it in the Workflow if it… coordinates Activities"
func AdjudicationWorkflow(ctx workflow.Context, applicant Applicant) (Decision, error) {
	ctx = workflow.WithActivityOptions(ctx, opts)

	var credit CreditCheck
	if err := workflow.ExecuteActivity(ctx, RunCreditCheck, applicant).Get(ctx, &credit); err != nil {
		return "", err
	}
	var fraud FraudCheck
	if err := workflow.ExecuteActivity(ctx, CheckFraud, applicant, credit).Get(ctx, &fraud); err != nil {
		return "", err
	}
	var decision Decision
	err := workflow.ExecuteActivity(ctx, MakeDecision, applicant, credit, fraud).Get(ctx, &decision)
	return decision, err
}

// end slide

func ClaimWorkflow(ctx workflow.Context, claim Claim) (Approval, error) {
	// slide: 05-workflows.md "Keep it in the Workflow if it… must survive restarts"
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})

	var approval Approval
	err := workflow.ExecuteActivity(ctx, GetApproval, claim).Get(ctx, &approval)
	// end slide
	return approval, err
}

// CreateUserWorkflow runs the CreateUser Activity with retries, to show why it
// must be split.
func CreateUserWorkflow(ctx workflow.Context, data UserData) (User, error) {
	ctx = workflow.WithActivityOptions(ctx, opts)
	var user User
	err := workflow.ExecuteActivity(ctx, CreateUser, data).Get(ctx, &user)
	return user, err
}

// PricingWorkflow is the mistake on the "Make it an Activity" slide: it reads
// the Worker's environment in Workflow code and decides on it. It runs, and
// its history replays only while the environment stays the same.
func PricingWorkflow(ctx workflow.Context) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, opts)
	pricing := StandardPricing
	if os.Getenv("PRICING_TIER") == "premium" { // ❌ in a Workflow
		pricing = PremiumPricing
	}
	var tier string
	err := workflow.ExecuteActivity(ctx, pricing).Get(ctx, &tier)
	return tier, err
}
