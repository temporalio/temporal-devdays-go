// ABOUTME: Hosts every parent and every child Workflow.
//
//	go run ./worker
package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/workshops/child-workflows/workflows"
)

func main() {
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}
	c, err := client.Dial(client.Options{
		HostPort: address,
		// Warnings and errors only, so the terminal shows this Workshop's own lines.
		Logger: tlog.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr,
			&slog.HandlerOptions{Level: slog.LevelWarn}))),
	})
	if err != nil {
		log.Fatalln("Unable to connect to Temporal:", err)
	}
	defer c.Close()

	// A parent and its children can share one Worker and one Task Queue: a
	// child runs on its parent's Task Queue unless told otherwise.
	w := worker.New(c, workflows.TaskQueue, worker.Options{})
	w.RegisterWorkflow(workflows.OrderWorkflow)
	w.RegisterWorkflow(workflows.ShipmentWorkflow)
	w.RegisterWorkflow(workflows.ChainWorkflow)
	w.RegisterWorkflow(workflows.PackWorkflow)
	w.RegisterWorkflow(workflows.PackItemWorkflow)

	fmt.Printf("Worker polling %s. Ctrl+C to stop.\n", workflows.TaskQueue)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Worker stopped:", err)
	}
}
