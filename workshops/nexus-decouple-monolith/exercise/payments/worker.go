// The Payments team's Worker. Whatever NewWorker registers runs in the
// Payments process: a bug in it at 3 AM takes payments down.
//
// Do TODO 5 last, after TODO 4 has moved the Workflow onto a Nexus client.
package payments

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
	// TODO 5a start

	// Delete this import once TODO 5b is done. It compiles the Compliance
	// team's code into the Payments binary.
	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/compliance"
	// TODO 5a end
)

// NewWorker returns the Payments team's Worker. c must be connected to the
// Payments Namespace.
func NewWorker(c client.Client) worker.Worker {
	w := worker.New(c, shared.PaymentsTaskQueue, worker.Options{})
	w.RegisterWorkflow(PaymentProcessingWorkflow)
	w.RegisterWorkflow(ReviewCallerWorkflow)
	w.RegisterActivity(ValidatePayment)
	w.RegisterActivity(ExecutePayment)
	// TODO 5b start
	// Once the Workflow calls Compliance over Nexus, this Worker has no reason
	// to run the Compliance team's code. Delete the registration below, then
	// the import (TODO 5a).
	//
	// That deletion is the decoupling. Everything before it was wiring; this
	// is the line where the two teams stop sharing a process.
	w.RegisterActivity(compliance.CheckCompliance)
	// TODO 5b end
	return w
}
