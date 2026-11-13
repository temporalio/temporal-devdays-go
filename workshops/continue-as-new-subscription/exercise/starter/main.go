// ABOUTME: Bills one subscription for 30 cycles and waits for it to finish.
//
//	go run ./starter           subscription-demo
//	go run ./starter <tag>     subscription-<tag>
//
// Starting again with the same ID replaces the old subscription.
package main

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"log/slog"
	"os"
	"strings"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"

	"github.com/temporalio/temporal-devdays-go/workshops/continue-as-new-subscription/subscription"
)

const targetCycles = 30

func main() {
	tag := "demo"
	if len(os.Args) > 1 && strings.TrimSpace(os.Args[1]) != "" {
		tag = strings.TrimSpace(os.Args[1])
	}
	workflowID := "subscription-" + tag

	c, err := client.Dial(client.Options{
		HostPort: os.Getenv("TEMPORAL_ADDRESS"), // empty means localhost:7233
		Logger:   log.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))),
	})
	if err != nil {
		stdlog.Fatalln("Unable to connect to Temporal:", err)
	}
	defer c.Close()

	ctx := context.Background()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                subscription.TaskQueue,
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_TERMINATE_EXISTING,
	}, subscription.SubscriptionWorkflow, subscription.SubscriptionInput{
		SubscriptionID:  workflowID,
		CyclesCompleted: 0,
		MaxCycles:       targetCycles,
	})
	if err != nil {
		stdlog.Fatalln("Unable to start the subscription:", err)
	}
	fmt.Printf("Billing %s for %d cycles. Waiting for it to finish...\n", workflowID, targetCycles)

	// Follows the chain: waits for the last run, however many rollovers it takes.
	var result subscription.SubscriptionResult
	err = run.Get(ctx, &result)
	var terminated *temporal.TerminatedError
	switch {
	case errors.As(err, &terminated):
		fmt.Printf("\n%s was terminated by the Temporal Service: %s\n", workflowID, terminationReason(ctx, c, workflowID))
	case err != nil:
		stdlog.Fatalln("The subscription failed:", err)
	default:
		fmt.Printf("\nDone. Billed %d cycles.\n", result.TotalCycles)
	}
}

// terminationReason reads the reason off the last Event of the latest run.
func terminationReason(ctx context.Context, c client.Client, workflowID string) string {
	iter := c.GetWorkflowHistory(ctx, workflowID, "", false, enumspb.HISTORY_EVENT_FILTER_TYPE_CLOSE_EVENT)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			break
		}
		if attrs := event.GetWorkflowExecutionTerminatedEventAttributes(); attrs != nil {
			return attrs.GetReason()
		}
	}
	return "terminated"
}
