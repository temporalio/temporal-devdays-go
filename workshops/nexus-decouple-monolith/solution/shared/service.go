// The shared contract between the Payments and Compliance teams. Like an
// OpenAPI spec, but durable. Payments calls the Operations it names;
// Compliance implements a handler for each one.
//
// Both teams import THIS package. It is the only thing they share.
package shared

import "github.com/nexus-rpc/sdk-go/nexus"

// ComplianceServiceName is the Service name on the wire. Caller and handler
// must agree on it. It is not the Endpoint name.
const ComplianceServiceName = "ComplianceNexusService"

// TODO 1 start
var (
	// CheckComplianceOperation is the slow path. Backed by a Workflow,
	// because a risky payment may need a human.
	CheckComplianceOperation = nexus.NewOperationReference[ComplianceRequest, ComplianceResult]("checkCompliance")

	// SubmitReviewOperation is the fast path. A human decision handed to an
	// already-running compliance check.
	SubmitReviewOperation = nexus.NewOperationReference[ReviewRequest, ComplianceResult]("submitReview")
)

// TODO 1 end
