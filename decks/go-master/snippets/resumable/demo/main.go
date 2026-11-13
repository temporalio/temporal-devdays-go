// The Worker for the Resumable Activity demo in 11-updates.md.
// Start a dev server (`temporal server start-dev`), then from the snippets
// folder: `go run ./resumable/demo`. The commands to drive it are in that demo
// slide's speaker notes.
package main

import (
	"log"
	"strings"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/resumable"
)

func main() {
	c, err := client.Dial(client.Options{HostPort: "localhost:7233"})
	if err != nil {
		log.Fatalln("connecting to Temporal:", err)
	}
	defer c.Close()

	w := worker.New(c, "resumable-demo", worker.Options{})
	w.RegisterWorkflow(resumable.TransferWorkflow)
	w.RegisterActivity(resumable.ExecuteTransfer)

	log.Println("Worker polling resumable-demo. Accounts that exist:", strings.Join(resumable.KnownAccounts, ", "))
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln(err)
	}
}
