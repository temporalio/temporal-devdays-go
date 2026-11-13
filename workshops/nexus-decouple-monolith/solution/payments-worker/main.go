// [GIVEN] Runs the Payments team's Worker: go run ./payments-worker
//
// What it registers is in payments/worker.go.
package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/payments"
	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

func main() {
	c, err := client.Dial(client.Options{HostPort: hostPort(), Namespace: shared.PaymentsNamespace, Logger: logger()})
	if err != nil {
		log.Fatalln("Unable to connect to Temporal:", err)
	}
	defer c.Close()

	w := payments.NewWorker(c)
	if err := w.Start(); err != nil {
		log.Fatalln("Unable to start the Payments Worker:", err)
	}
	defer w.Stop()

	fmt.Println()
	fmt.Println("  Payments Worker started on:", shared.PaymentsTaskQueue)
	fmt.Println("  Namespace:", shared.PaymentsNamespace)
	fmt.Println("  Press Ctrl+C to stop it.")
	fmt.Println()

	<-worker.InterruptCh()
}

// logger writes the SDK's INFO lines and above to the terminal, leaving out its
// DEBUG chatter.
func logger() sdklog.Logger {
	return sdklog.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

func hostPort() string {
	if address := os.Getenv("TEMPORAL_ADDRESS"); address != "" {
		return address
	}
	return client.DefaultHostPort
}
