package why

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// RunWorker is the Worker from 03-temporal-concepts.md "One Worker, both kinds
// of code". It blocks until the process is interrupted, so the tests build the
// same Worker with testenv.Worker and the same register calls instead.
func RunWorker(c client.Client) error {
	// slide: 03-temporal-concepts.md "One Worker, both kinds of code"
	w := worker.New(c, "transfers", worker.Options{})
	w.RegisterWorkflow(TransferWorkflow)
	w.RegisterActivity(Withdraw)
	w.RegisterActivity(Deposit)

	err := w.Run(worker.InterruptCh())
	// end slide
	return err
}
