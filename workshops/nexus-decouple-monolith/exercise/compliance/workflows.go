// [GIVEN] Compliance Workflow. Pre-built for this exercise.
//
// Runs an automated compliance check and, for MEDIUM-risk transactions, waits
// for a human reviewer to approve or deny it through a Workflow Update.
//
//	LOW risk  -> auto-approved, returns after the sleep
//	HIGH risk -> auto-denied, returns after the sleep
//	MEDIUM    -> pauses, waits for the review Update
//
// The 10-second sleep is intentional. It gives you a window to stop the
// Compliance Worker with Ctrl+C and restart it, so you can watch the Nexus
// Operation survive the outage. The timer runs on the server, so nothing is
// lost when the Worker goes away.
package compliance

import (
	"errors"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

// ReviewUpdate is the Update the Compliance team's sync Nexus handler sends
// when a human decides. Its arguments are (approved bool, explanation string).
const ReviewUpdate = "review"

// ComplianceWorkflow checks one transaction and, if it is MEDIUM risk, waits
// for a reviewer.
func ComplianceWorkflow(ctx workflow.Context, request shared.ComplianceRequest) (shared.ComplianceResult, error) {
	var auto, review *shared.ComplianceResult

	err := workflow.SetUpdateHandlerWithOptions(ctx, ReviewUpdate,
		func(ctx workflow.Context, approved bool, explanation string) (shared.ComplianceResult, error) {
			result := shared.ComplianceResult{
				TransactionID: request.TransactionID,
				Approved:      approved,
				RiskLevel:     shared.RiskMedium,
				Explanation:   explanation,
			}
			review = &result
			return result, nil
		},
		workflow.UpdateHandlerOptions{
			Validator: func(ctx workflow.Context, approved bool, explanation string) error {
				if auto == nil || auto.RiskLevel != shared.RiskMedium {
					return errors.New("workflow is not awaiting review")
				}
				if review != nil {
					return errors.New("review already submitted")
				}
				return nil
			},
		})
	if err != nil {
		return shared.ComplianceResult{}, err
	}

	// Step 1: automated compliance check.
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})
	var result shared.ComplianceResult
	if err := workflow.ExecuteActivity(ctx, CheckCompliance, request).Get(ctx, &result); err != nil {
		return shared.ComplianceResult{}, err
	}
	auto = &result

	// Durable delay. Stop the Compliance Worker mid-sleep, restart it, and the
	// Workflow resumes on its own.
	if err := workflow.Sleep(ctx, 10*time.Second); err != nil {
		return shared.ComplianceResult{}, err
	}

	// Step 2: LOW or HIGH risk returns now.
	if result.RiskLevel != shared.RiskMedium {
		return result, nil
	}

	// Step 3: MEDIUM risk waits for human review.
	//
	// Bounded on purpose. An unbounded wait would leave this Workflow Running
	// forever when nobody reviews. Eight minutes is deliberately shorter than
	// the caller's 10-minute ScheduleToCloseTimeout, so the handler reports a
	// decision rather than letting the caller time out first.
	reviewed, err := workflow.AwaitWithTimeout(ctx, 8*time.Minute, func() bool { return review != nil })
	if err != nil {
		return shared.ComplianceResult{}, err
	}
	if !reviewed {
		return shared.ComplianceResult{
			TransactionID: request.TransactionID,
			Approved:      false,
			RiskLevel:     shared.RiskMedium,
			Explanation:   "No reviewer responded within 8 minutes. Not approved.",
		}, nil
	}
	return *review, nil
}
