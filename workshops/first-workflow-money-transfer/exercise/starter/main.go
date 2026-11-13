// ABOUTME: Starts one transfer and waits for its result.
//
//	go run ./starter                  $100 from A-100 to B-200
//	go run ./starter -amount 2500     $2,500, a large deposit
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"

	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"

	moneytransfer "github.com/temporalio/temporal-devdays-go/workshops/first-workflow-money-transfer"
)

func main() {
	amount := flag.Int("amount", 100, "dollars to transfer")
	flag.Parse()

	details := moneytransfer.PaymentDetails{
		SourceAccount: "A-100",
		TargetAccount: "B-200",
		Amount:        *amount,
		ReferenceID:   fmt.Sprintf("%06x", rand.N(1<<24)),
	}
	workflowID := "transfer-" + details.ReferenceID

	c, err := client.Dial(client.Options{
		HostPort: address(),
		Logger:   sdklog.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))),
	})
	if err != nil {
		log.Fatalln("Can't reach the Temporal Service:", err)
	}
	defer c.Close()

	ctx := context.Background()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: moneytransfer.TaskQueue,
	}, moneytransfer.MoneyTransferWorkflow, details)
	if err != nil {
		log.Fatalln("Can't start the transfer:", err)
	}
	fmt.Printf("Started %s: $%d from %s to %s.\n", workflowID, details.Amount, details.SourceAccount, details.TargetAccount)
	fmt.Printf("Find it in the Temporal UI: Workflows › %s\n", workflowID)
	fmt.Println("Waiting for the result...")

	var result string
	if err := run.Get(ctx, &result); err != nil {
		log.Fatalln("The transfer failed:", err)
	}
	fmt.Println(result)
}

func address() string {
	if a := os.Getenv("TEMPORAL_ADDRESS"); a != "" {
		return a
	}
	return client.DefaultHostPort
}
