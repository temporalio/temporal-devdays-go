// ABOUTME: The fulfillment Worker. Hosts the Saga Workflow and all five Activities.
// Every [ledger] line in the lab is printed here.
//
//	go run ./worker
//
// Go builds your code into the Worker when it starts. After an edit, stop it
// with Ctrl+C and run it again.
package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"

	saga "github.com/temporalio/temporal-devdays-go/workshops/saga-order-fulfillment"
)

func main() {
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = client.DefaultHostPort
	}
	c, err := client.Dial(client.Options{
		HostPort: address,
		// Warnings and errors only, so the [ledger] lines stand out. Failed
		// Activities and the unwind are logged at ERROR, so they still show.
		Logger: tlog.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelWarn}))),
	})
	if err != nil {
		log.Fatalln("Unable to connect to Temporal:", err)
	}
	defer c.Close()

	w := worker.New(c, saga.TaskQueue, worker.Options{})
	w.RegisterWorkflow(saga.OrderFulfillmentWorkflow)
	w.RegisterActivity(saga.ReserveInventory)
	w.RegisterActivity(saga.ChargePayment)
	w.RegisterActivity(saga.DispatchCourier)
	w.RegisterActivity(saga.ReleaseInventoryIfReserved)
	w.RegisterActivity(saga.RefundPaymentIfCharged)

	fmt.Printf("Fulfillment Worker polling %s\n", saga.TaskQueue)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Worker stopped:", err)
	}
}
