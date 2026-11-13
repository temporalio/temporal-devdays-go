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
// Declare the two Operations this Service offers. Each one is a name paired
// with an input type and an output type, and carries no implementation: only
// what both teams agree on.
//
//	CheckComplianceOperation   "checkCompliance"   ComplianceRequest -> ComplianceResult
//	SubmitReviewOperation      "submitReview"      ReviewRequest     -> ComplianceResult
//
// Write each one as nexus.NewOperationReference[Input, Output]("name").
//
// The two below are placeholders, with no name and no types, so the code
// compiles. Replace them. The Compliance handlers name their Operations after
// these, and the Payments Workflow calls them, so everything downstream
// builds on what you write here.
//
// Stuck? "Ask AI" on https://docs.temporal.io:
//
//	"how do I define a Nexus Service contract in Go?"
var (
	CheckComplianceOperation = nexus.NewOperationReference[any, any]("")
	SubmitReviewOperation    = nexus.NewOperationReference[any, any]("")
)

// TODO 1 end
