// The Worker from 07-workers.md. Compiled and vetted, never run by the tests:
// it would connect to localhost:7233 and poll until interrupted. main_test.go
// builds the same Worker against the test's dev server instead.
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	orders "github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/activities"
)

// slide: 07-workers.md "Go Workers"
func main() {
	c, err := client.Dial(client.Options{HostPort: "localhost:7233"})
	if err != nil {
		log.Fatalln("unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "order-task-queue", worker.Options{})
	w.RegisterWorkflow(orders.OrderWorkflow)
	w.RegisterActivity(orders.SubmitOrder)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("unable to start worker", err)
	}
}

// end slide
