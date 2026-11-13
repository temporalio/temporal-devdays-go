// ABOUTME: The Worker: runs the money transfer Workflow and its Activities.
//
//	go run ./worker
//
// It does not reload when you save. A Worker runs the code it started with, so
// after an edit you stop it and start it again.
//
// Ctrl+C is a crash, on purpose: the process exits without telling Temporal,
// the way a server that loses power does.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"

	moneytransfer "github.com/temporalio/temporal-devdays-go/workshops/first-workflow-money-transfer"
)

func main() {
	// No Workflow cache, so the Server never saves the next step for this Worker
	// alone. After a crash, the restarted Worker picks it up at once.
	worker.SetStickyWorkflowCacheSize(0)

	// Warnings and errors only, so the terminal shows this Workshop's own lines.
	logs := oneLineFailures{slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})}
	c, err := client.Dial(client.Options{
		HostPort: address(),
		Logger:   sdklog.NewStructuredLogger(slog.New(logs)),
	})
	if err != nil {
		log.Fatalln("Can't reach the Temporal Service:", err)
	}
	defer c.Close()

	w := worker.New(c, moneytransfer.TaskQueue, worker.Options{})
	w.RegisterWorkflow(moneytransfer.MoneyTransferWorkflow)
	w.RegisterActivity(moneytransfer.Withdraw)
	w.RegisterActivity(moneytransfer.Deposit)

	if err := w.Start(); err != nil {
		log.Fatalln("Can't start the Worker:", err)
	}
	fmt.Printf("Worker polling the %s Task Queue. Ctrl+C crashes it.\n", moneytransfer.TaskQueue)

	// Run until the process is killed. There is no signal handler, so Ctrl+C
	// ends the process on the spot and the Worker never tells Temporal: a crash.
	select {}
}

func address() string {
	if a := os.Getenv("TEMPORAL_ADDRESS"); a != "" {
		return a
	}
	return client.DefaultHostPort
}

// oneLineFailures prints a failed Activity attempt as one line. The SDK's own
// report of it is a full error, with a stack trace for a panic, every 2 seconds.
type oneLineFailures struct{ slog.Handler }

func (h oneLineFailures) Handle(ctx context.Context, r slog.Record) error {
	if r.Message != "Activity error." && r.Message != "Activity panic." {
		return h.Handler.Handle(ctx, r)
	}
	cause := "unknown error"
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "Error" || a.Key == "PanicError" {
			cause = a.Value.String()
		}
		return true
	})
	fmt.Printf("[activity] failed: %s. Temporal will retry.\n", cause)
	return nil
}

func (h oneLineFailures) WithAttrs(attrs []slog.Attr) slog.Handler {
	return oneLineFailures{h.Handler.WithAttrs(attrs)}
}

func (h oneLineFailures) WithGroup(name string) slog.Handler {
	return oneLineFailures{h.Handler.WithGroup(name)}
}
