// ABOUTME: Starts one parent Workflow and prints what came back.
//
//	go run ./starter child             one parent, one child
//	go run ./starter chain --fail      a chain 5 levels deep that fails at the bottom
//	go run ./starter chain --depth 3   a shorter chain that succeeds
//	go run ./starter pack              fan out one child per item, then fan in
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"

	"github.com/temporalio/temporal-devdays-go/workshops/child-workflows/workflows"
)

const usage = `Usage: go run ./starter <scenario> [flags]

  child               one parent, one child
  chain --fail        a chain 5 levels deep that fails at the bottom
  chain --depth 3     a shorter chain that succeeds
  pack                fan out one child per item, then fan in`

func newID(prefix string) string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return prefix + "-" + strings.ToUpper(hex.EncodeToString(b))
}

// resultOf waits for a Workflow's result, saying so if no Worker picks it up.
// Without a Worker the Workflow just sits on the Task Queue, so the starter
// would otherwise wait forever with nothing on screen.
func resultOf(ctx context.Context, run client.WorkflowRun, valuePtr any) error {
	nudge := time.AfterFunc(5*time.Second, func() {
		fmt.Printf("Still waiting on %s. Is the Worker running? In the Worker tab, run:\n  go run ./worker\n", run.GetID())
	})
	defer nudge.Stop()
	return run.Get(ctx, valuePtr)
}

func runChild(ctx context.Context, c client.Client) error {
	order := workflows.Order{OrderID: newID("ORD"), Items: []string{"mug"}}
	workflowID := "order-" + order.OrderID
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: workflowID, TaskQueue: workflows.TaskQueue},
		workflows.OrderWorkflow, order)
	if err != nil {
		return err
	}
	var tracking string
	if err := resultOf(ctx, run, &tracking); err != nil {
		return err
	}
	fmt.Printf("Parent  %s  completed\n", workflowID)
	fmt.Printf("Child   shipment-%s  returned %s\n", order.OrderID, tracking)
	return nil
}

func runChain(ctx context.Context, c client.Client, depth int, failAtBottom bool) error {
	root := newID("chain")
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: root, TaskQueue: workflows.TaskQueue},
		workflows.ChainWorkflow, workflows.ChainRequest{Depth: depth, FailAtBottom: failAtBottom})
	if err != nil {
		return err
	}
	var levels int
	err = resultOf(ctx, run, &levels)
	if err == nil {
		fmt.Printf("Chain %s completed: %d levels of children below the top\n", root, levels)
		return nil
	}
	var failed *temporal.WorkflowExecutionError
	if !errors.As(err, &failed) {
		return err
	}
	// Unwrap one level at a time. Every parent wraps its child's failure in a
	// ChildWorkflowExecutionError, so the real error is at the very bottom.
	fmt.Printf("Chain %s failed. The failure, one wrapper per level:\n\n", root)
	cause := errors.Unwrap(failed)
	for {
		child, ok := cause.(*temporal.ChildWorkflowExecutionError)
		if !ok {
			break
		}
		fmt.Printf("  ChildWorkflowExecutionError  from %s\n", child.WorkflowID())
		cause = child.Unwrap()
	}
	var app *temporal.ApplicationError
	if errors.As(cause, &app) {
		fmt.Printf("  ApplicationError: %s: %s\n", app.Type(), app.Message())
	} else {
		fmt.Printf("  %v\n", cause)
	}
	return nil
}

func runPack(ctx context.Context, c client.Client) error {
	order := workflows.Order{OrderID: newID("ORD"), Items: []string{"mug", "book", "lamp", "plant", "clock"}}
	workflowID := "pack-" + order.OrderID
	started := time.Now()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: workflowID, TaskQueue: workflows.TaskQueue},
		workflows.PackWorkflow, order)
	if err != nil {
		return err
	}
	var boxes []string
	if err := resultOf(ctx, run, &boxes); err != nil {
		return err
	}
	took := time.Since(started).Seconds()
	fmt.Printf("Packed %d of %d items in %.1fs: [%s]\n", len(boxes), len(order.Items), took, strings.Join(boxes, ", "))
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	scenario := os.Args[1]
	flags := flag.NewFlagSet(scenario, flag.ExitOnError)
	flags.Usage = func() { fmt.Fprintln(os.Stderr, usage) }
	depth := flags.Int("depth", 5, "how many levels of children to start below the top")
	fail := flags.Bool("fail", false, "the bottom of the chain fails")
	_ = flags.Parse(os.Args[2:])

	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}
	c, err := client.Dial(client.Options{
		HostPort: address,
		// Warnings and errors only, so the terminal shows this Workshop's own lines.
		Logger: tlog.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr,
			&slog.HandlerOptions{Level: slog.LevelWarn}))),
	})
	if err != nil {
		log.Fatalln("Unable to connect to Temporal:", err)
	}
	defer c.Close()

	ctx := context.Background()
	switch scenario {
	case "child":
		err = runChild(ctx, c)
	case "chain":
		err = runChain(ctx, c, *depth, *fail)
	case "pack":
		err = runPack(ctx, c)
	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	if err != nil {
		log.Fatalln(err)
	}
}
