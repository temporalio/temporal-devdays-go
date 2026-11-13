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
	// Nothing of Compliance's is compiled into this Worker.
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
	// Payments' Activities only. Compliance is remote, behind Nexus.
	// TODO 5b end
	return w
}
