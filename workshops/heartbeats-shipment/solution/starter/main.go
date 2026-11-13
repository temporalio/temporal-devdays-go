// ABOUTME: Packs one shipment and waits for the result.
//
//	go run ./starter          shipment-demo
//	go run ./starter <tag>    shipment-<tag>
//
// Starting again with the same ID replaces the old shipment, so a run left
// retrying never blocks the next experiment.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"

	shipment "github.com/temporalio/temporal-devdays-go/workshops/heartbeats-shipment"
)

func main() {
	tag := "demo"
	if len(os.Args) > 1 && strings.TrimSpace(os.Args[1]) != "" {
		tag = strings.TrimSpace(os.Args[1])
	}
	workflowID := "shipment-" + tag

	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}
	c, err := client.Dial(client.Options{
		HostPort: address,
		Logger:   tlog.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))),
	})
	if err != nil {
		log.Fatalln("Unable to connect to Temporal:", err)
	}
	defer c.Close()

	fmt.Printf("Packing %s. %d items.\n", workflowID, shipment.BatchSize)
	fmt.Println("Watch the Worker terminal count them.")

	ctx := context.Background()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                shipment.TaskQueue,
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_TERMINATE_EXISTING,
	}, shipment.ShipmentWorkflow, shipment.ShipmentInput{ShipmentID: workflowID, ItemCount: shipment.BatchSize})
	if err != nil {
		log.Fatalln("Unable to start the Workflow:", err)
	}

	var result shipment.ShipmentResult
	if err := run.Get(ctx, &result); err != nil {
		if temporal.IsCanceledError(err) {
			fmt.Printf("\n%s was cancelled. It ended Canceled, not Completed.\n", workflowID)
			return
		}
		log.Fatalln("The shipment failed:", err)
	}
	fmt.Printf("\nPacked %d items, finished on attempt %d.\n", result.Packed, result.FinishedOnAttempt)
	if result.FinishedOnAttempt > 1 {
		fmt.Println("An earlier attempt was interrupted. This one finished the job.")
	}
}
