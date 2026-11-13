// ABOUTME: The subscription Worker: hosts the Workflow and its one Activity.
// The [activity] lines print here, one per billed cycle.
//
//	go run ./worker
package main

import (
	"fmt"
	stdlog "log"
	"log/slog"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/workshops/continue-as-new-subscription/subscription"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort: os.Getenv("TEMPORAL_ADDRESS"), // empty means localhost:7233
		// Warnings and errors only, so the terminal shows this Workshop's own lines.
		Logger: log.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))),
	})
	if err != nil {
		stdlog.Fatalln("Unable to connect to Temporal:", err)
	}
	defer c.Close()

	w := worker.New(c, subscription.TaskQueue, worker.Options{})
	w.RegisterWorkflow(subscription.SubscriptionWorkflow)
	w.RegisterActivity(subscription.ChargeSubscription)

	fmt.Printf("Subscription Worker polling %s. Ctrl+C to stop.\n", subscription.TaskQueue)
	if err := w.Run(worker.InterruptCh()); err != nil {
		stdlog.Fatalln("Worker stopped:", err)
	}
}
