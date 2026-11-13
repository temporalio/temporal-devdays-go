package main

import (
	"context"
	"testing"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
	transfers "github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/workflows"
)

func TestMain(m *testing.M) { testenv.Main(m) }

// The slide's steps against the test's dev server, with a Worker polling the
// slide's Task Queue.
func TestTheClientStartsTheTransferAndReadsTheReceipt(t *testing.T) {
	testenv.WorkerOn(t, "transfers", func(r worker.Registry) {
		r.RegisterWorkflow(transfers.TransferWorkflow)
		r.RegisterActivity(transfers.NewActivities())
	})
	c := testenv.Client()

	transfer := transfers.Transfer{TransferID: "tx-1001", FromAccount: "A-1", ToAccount: "B-2", Amount: 250}
	run, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        "transfer-" + transfer.TransferID,
		TaskQueue: "transfers",
	}, transfers.TransferWorkflow, transfer)
	if err != nil {
		t.Fatal(err)
	}
	if run.GetID() != "transfer-tx-1001" {
		t.Fatalf("Workflow ID %q", run.GetID())
	}
	var receipt transfers.Receipt
	if err := run.Get(testenv.Ctx(t), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Status != "completed" {
		t.Fatalf("receipt %+v", receipt)
	}
}

// Passing the function, not its name, lets the Client check the input before
// anything is sent: a wrong argument type fails in the Client.
func TestTheClientChecksTheInputAgainstTheWorkflowsParameters(t *testing.T) {
	_, err := testenv.Client().ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        testenv.ID("transfer"),
		TaskQueue: "transfers",
	}, transfers.TransferWorkflow, "tx-1001")
	if err == nil {
		t.Fatal("a string where a Transfer belongs should be refused")
	}
}
