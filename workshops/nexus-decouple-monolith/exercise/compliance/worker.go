// The Compliance team's Worker. This process is the only place their code runs.
package compliance

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

// NewWorker returns the Compliance team's Worker, polling the Task Queue the
// Nexus Endpoint targets. c must be connected to the Compliance Namespace.
func NewWorker(c client.Client) (worker.Worker, error) {
	w := worker.New(c, shared.ComplianceTaskQueue, worker.Options{})
	w.RegisterWorkflow(ComplianceWorkflow)
	w.RegisterActivity(CheckCompliance)

	// TODO 3 start
	// A Worker only handles work it has been told about. This one already
	// knows its Workflow and its Activity. What it does not know yet is the
	// Nexus Service, the thing that makes this team callable by Payments at
	// all.
	//
	// Build the Service with NewNexusService() (compliance/handler.go),
	// return the error if there is one, and register it on w with
	// w.RegisterNexusService(...).
	//
	// Leave it out and the Worker still starts. Payments' calls just go
	// unanswered. That is the failure you are guarding against.
	//
	// Stuck? "Ask AI" on https://docs.temporal.io:
	//   "what do I register on a Go Worker that handles Nexus Operations?"
	// TODO 3 end

	return w, nil
}
