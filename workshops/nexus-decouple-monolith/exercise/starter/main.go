// [GIVEN] Starts three payment Workflows: go run ./starter
//
// All three start before any result is collected. That matters once Nexus is
// in play: TXN-B is MEDIUM risk, so its handler Workflow pauses for human
// review and would otherwise hold TXN-C up behind it. Starting them together
// lets TXN-A and TXN-C finish while TXN-B sits visibly pending, which is the
// point of the human-in-the-loop path.
//
// The starter does not know or care whether compliance runs locally or via
// Nexus.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/payments"
	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

// How long to wait before calling a Workflow unresolved. TXN-A and TXN-C
// finish in roughly 12 seconds (the handler Workflow sleeps for 10 of those).
const collectDeadline = 25 * time.Second

var transactions = []shared.PaymentRequest{
	{
		TransactionID: "TXN-A", Amount: 250, Currency: "USD",
		SenderCountry: "US", ReceiverCountry: "US",
		Description: "Routine supplier payment", SenderAccount: "ACC-001", ReceiverAccount: "ACC-002",
	},
	{
		TransactionID: "TXN-B", Amount: 12_000, Currency: "USD",
		SenderCountry: "US", ReceiverCountry: "UK",
		Description: "International consulting fee", SenderAccount: "ACC-003", ReceiverAccount: "ACC-004",
	},
	{
		TransactionID: "TXN-C", Amount: 75_000, Currency: "USD",
		SenderCountry: "US", ReceiverCountry: "US",
		Description: "Large capital transfer", SenderAccount: "ACC-005", ReceiverAccount: "ACC-006",
	},
}

type outcome struct {
	result shared.PaymentResult
	err    error
	done   bool
}

func main() {
	fmt.Println()
	fmt.Println("  PAYMENT STARTER: running 3 transactions through Temporal")
	fmt.Println()

	c, err := client.Dial(client.Options{HostPort: hostPort(), Namespace: shared.PaymentsNamespace})
	if err != nil {
		log.Fatalln("Unable to connect to Temporal:", err)
	}
	defer c.Close()

	// Start all three without waiting for results, so a pause in one cannot
	// hold up the others.
	runs := make([]client.WorkflowRun, len(transactions))
	for i, txn := range transactions {
		run, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
			ID:        "payment-" + txn.TransactionID,
			TaskQueue: shared.PaymentsTaskQueue,
		}, payments.PaymentProcessingWorkflow, txn)
		if err != nil {
			log.Fatalln("Unable to start payment-"+txn.TransactionID+":", err)
		}
		runs[i] = run
		fmt.Printf("  Started: payment-%s\n", txn.TransactionID)
		fmt.Printf("    Amount: $%.2f | Route: %s -> %s\n", txn.Amount, txn.SenderCountry, txn.ReceiverCountry)
	}

	fmt.Println()
	fmt.Println("  All 3 started. Collecting results...")
	fmt.Println()

	// Wait on all three against one shared deadline. Waiting one at a time
	// would let TXN-B's review pause burn the whole budget, so TXN-C would be
	// reported as unresolved even though it finished seconds earlier.
	ctx, cancel := context.WithTimeout(context.Background(), collectDeadline)
	defer cancel()
	outcomes := make([]outcome, len(runs))
	var wg sync.WaitGroup
	for i, run := range runs {
		wg.Go(func() {
			var result shared.PaymentResult
			err := run.Get(ctx, &result)
			if ctx.Err() != nil {
				return // still running at the deadline
			}
			outcomes[i] = outcome{result: result, err: err, done: true}
		})
	}
	wg.Wait()

	unresolved := 0
	for i, txn := range transactions {
		o := outcomes[i]
		fmt.Println("  ----------------------------------------------------")
		fmt.Println("  " + txn.TransactionID)
		switch {
		case !o.done:
			unresolved++
			fmt.Println("  Result: STILL RUNNING")
			fmt.Println("  Reason: either it is parked for human review (MEDIUM risk), or the")
			fmt.Println("          Compliance handler is not answering right now.")
			fmt.Println("  Action: check the Temporal UI. If it is parked for review, run")
			fmt.Println("          'go run ./review-starter'.")
		case o.err != nil:
			fmt.Println("  Result: FAILED")
			fmt.Println("  Error: ", o.err)
		default:
			r := o.result
			fmt.Println("  Result:", r.Status)
			fmt.Println("  Risk:  ", orNA(string(r.RiskLevel)))
			fmt.Println("  Reason:", orNA(r.Explanation))
			if r.ConfirmationNumber != "" {
				fmt.Println("  Conf#: ", r.ConfirmationNumber)
			}
			if r.Error != "" {
				fmt.Println("  Error: ", r.Error)
			}
		}
		fmt.Println()
	}

	fmt.Println("  ----------------------------------------------------")
	if unresolved > 0 {
		fmt.Printf("  %d transaction(s) have not resolved yet.\n", unresolved)
		fmt.Println("  Parked for review? Run 'go run ./review-starter'.")
		fmt.Println("  Compliance Worker down? Start it and they resume on their own.")
	} else {
		fmt.Println("  All 3 transactions resolved!")
	}
	fmt.Println("  Check the Temporal UI: http://localhost:8233")
}

func orNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func hostPort() string {
	if address := os.Getenv("TEMPORAL_ADDRESS"); address != "" {
		return address
	}
	return client.DefaultHostPort
}
