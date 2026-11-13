package startexecute

import (
	"context"
	stdlog "log"
	"os"

	"go.temporal.io/sdk/client"
)

// log is what the slides print with. The tests swap it for one that records.
var log = stdlog.New(os.Stdout, "", 0)

func ExecuteOrder(ctx context.Context, c client.Client, order Order) error {
	// slide: 08-start-vs-execute.md "Get() now: fire and wait"
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        "order-" + order.OrderID,
		TaskQueue: "order-task-queue",
	}, OrderWorkflow, order)
	if err != nil {
		return err
	}

	var result OrderResult
	if err := run.Get(ctx, &result); err != nil {
		return err
	}
	log.Printf("Result: %s", result.Status)
	// end slide
	return nil
}

func StartApproval(ctx context.Context, c client.Client, order Order) error {
	// slide: 08-start-vs-execute.md "Get() later: fire, and keep the run"
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        "approval-" + order.OrderID,
		TaskQueue: "order-task-queue",
	}, ApprovalWorkflow, order)
	if err != nil {
		return err
	}

	value, err := c.QueryWorkflow(ctx, run.GetID(), run.GetRunID(), GetStatusQuery)
	if err != nil {
		return err
	}
	var status string
	if err := value.Get(&status); err != nil {
		return err
	}
	log.Printf("Status: %s", status)

	err = c.SignalWorkflow(ctx, run.GetID(), run.GetRunID(), ApproveSignal, "alice")
	if err != nil {
		return err
	}

	var result OrderResult
	if err := run.Get(ctx, &result); err != nil {
		return err
	}
	log.Printf("Result: %s", result.Status)
	// end slide
	return nil
}
