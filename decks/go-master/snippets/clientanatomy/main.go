// The Client from 07b-client.md. Compiled and vetted, never run by the tests:
// it would connect to localhost:7233. main_test.go runs the same steps against
// the test's dev server.
package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"

	transfers "github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/workflows"
)

// slide: 07b-client.md "Go Client"
func main() {
	c, err := client.Dial(client.Options{HostPort: "localhost:7233"})
	if err != nil {
		log.Fatalln("unable to create client", err)
	}
	defer c.Close()

	transfer := transfers.Transfer{
		TransferID:  "tx-1001",
		FromAccount: "A-1",
		ToAccount:   "B-2",
		Amount:      250,
	}
	run, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        "transfer-" + transfer.TransferID,
		TaskQueue: "transfers",
	}, transfers.TransferWorkflow, transfer)
	if err != nil {
		log.Fatalln("unable to start the transfer", err)
	}
	var receipt transfers.Receipt
	if err := run.Get(context.Background(), &receipt); err != nil {
		log.Fatalln("transfer failed", err)
	}
	log.Println(receipt.Status)
}

// end slide
