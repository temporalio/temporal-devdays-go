// The Compliance team's Nexus handlers. Payments depends on the contract in
// shared/service.go. This file is the code that answers those calls, and only
// the Compliance team owns it.
package compliance

import (
	"context"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

// workflowIDFor gives the same ID both times: one Operation creates the
// Workflow, the other looks it up.
func workflowIDFor(transactionID string) string {
	return "compliance-" + transactionID
}

// CheckComplianceOperation is the SLOW call, and the worked example for
// TODO 2. A risky payment may need a person to approve it, so an answer can be
// minutes or hours away. This Operation must not sit and wait for it.
//
// NewWorkflowRunOperation starts a Workflow and returns at once, handing back
// a token. Temporal holds that token and delivers the result whenever the
// Workflow finishes.
//
// A sync Operation would NOT work here: it is cut off at the 10-second handler
// deadline, and when Temporal retried the failed call it would start a SECOND
// compliance check for the same payment.
var CheckComplianceOperation = temporalnexus.NewWorkflowRunOperation(
	shared.CheckComplianceOperation.Name(),
	ComplianceWorkflow,
	func(ctx context.Context, input shared.ComplianceRequest, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
		return client.StartWorkflowOptions{
			// Business-meaningful and deduped: a retried Operation attaches to
			// the same compliance check rather than starting a second one.
			ID: workflowIDFor(input.TransactionID),
			// TaskQueue defaults to the queue this Operation is handled on.
		}, nil
	},
)

// SubmitReviewOperation is the FAST call. By now the compliance check is
// running and parked waiting for a decision. All this does is find that
// Workflow and hand it the yes or no, a few milliseconds, so a sync handler is
// the right shape.
var SubmitReviewOperation = nexus.NewSyncOperation(
	shared.SubmitReviewOperation.Name(),
	func(ctx context.Context, input shared.ReviewRequest, options nexus.StartOperationOptions) (shared.ComplianceResult, error) {
		// TODO 2 start
		c := temporalnexus.GetClient(ctx)
		handle, err := c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
			WorkflowID:   workflowIDFor(input.TransactionID),
			UpdateName:   ReviewUpdate,
			Args:         []any{input.Approved, input.Explanation},
			WaitForStage: client.WorkflowUpdateStageCompleted,
		})
		if err != nil {
			return shared.ComplianceResult{}, err
		}
		var result shared.ComplianceResult
		err = handle.Get(ctx, &result)
		return result, err
		// TODO 2 end
	},
)

// NewNexusService binds this team's handlers to the Service name in the
// contract, one handler per Operation.
func NewNexusService() (*nexus.Service, error) {
	service := nexus.NewService(shared.ComplianceServiceName)
	if err := service.Register(CheckComplianceOperation, SubmitReviewOperation); err != nil {
		return nil, err
	}
	return service, nil
}
