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
	// Leave this out and the Worker still starts. Payments' calls just go
	// unanswered. That is the failure this guards against.
	service, err := NewNexusService()
	if err != nil {
		return nil, err
	}
	w.RegisterNexusService(service)
	// TODO 3 end

	return w, nil
}
