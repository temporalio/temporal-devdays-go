// ABOUTME: The shipment Worker: hosts the Workflow and the packing Activity.
//
//	go run ./worker
//
// Ctrl+C is a crash, on purpose. Nothing here catches the signal, so the
// process exits on the spot without telling Temporal, the way a lost node does,
// and the running attempt is left Started.
package main

import (
	"log"
	"log/slog"
	"os"

	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"

	shipment "github.com/temporalio/temporal-devdays-go/workshops/heartbeats-shipment"
)

func main() {
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}
	c, err := client.Dial(client.Options{
		HostPort: address,
		// Warnings and errors only, so the terminal shows this Workshop's own lines.
		Logger: tlog.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))),
	})
	if err != nil {
		log.Fatalln("Unable to connect to Temporal:", err)
	}
	defer c.Close()

	w := worker.New(c, shipment.TaskQueue, worker.Options{})
	w.RegisterWorkflow(shipment.ShipmentWorkflow)
	w.RegisterActivity(shipment.PackShipment)

	log.Printf("Shipment Worker polling %s. Ctrl+C to crash it.", shipment.TaskQueue)
	// No interrupt channel: worker.InterruptCh() would turn Ctrl+C into a
	// graceful Shutdown that fails the running attempt so it retries at once,
	// and the lab would show nothing. Run returns only on a fatal error.
	if err := w.Run(nil); err != nil {
		log.Fatalln("Worker stopped:", err)
	}
}
