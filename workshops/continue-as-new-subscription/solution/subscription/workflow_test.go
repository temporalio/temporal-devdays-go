// ABOUTME: The Track's Check (ADR 0003). A failing test's message is exactly what
// a learner sees when Check fails, so every message names the file, the function
// and the TODO.
//
// The Service suggests Continue-As-New only when a history nears its limits, so
// this starts a real dev server with the same two settings the Track's dev
// server uses, scaled down further so the test is quick: the Service suggests
// Continue-As-New at 20 Events and terminates a run past 100. A billing cycle
// adds about 11 Events, so a run that never rolls over is terminated around
// cycle 9, and a run that does rolls over every 2 cycles.
package subscription

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

const (
	testCycles = 12
	todo1      = "TODO 1 in subscription/workflow.go (SubscriptionWorkflow)"
)

var testClient client.Client

func TestMain(m *testing.M) {
	os.Exit(runWithDevServer(m))
}

func runWithDevServer(m *testing.M) int {
	ctx := context.Background()
	opts := testsuite.DevServerOptions{
		LogLevel: "error",
		ClientOptions: &client.Options{
			Logger: log.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))),
		},
		ExtraArgs: []string{
			"--dynamic-config-value", "limit.historyCount.suggestContinueAsNew=20",
			"--dynamic-config-value", "limit.historyCount.error=100",
		},
	}
	// The Sandbox Image has the temporal CLI on PATH, so a Check never downloads one.
	if path, err := exec.LookPath("temporal"); err == nil {
		opts.ExistingPath = path
	}
	server, err := testsuite.StartDevServer(ctx, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not start the Temporal dev server for the tests:", err)
		return 1
	}
	defer func() { _ = server.Stop() }()
	testClient = server.Client()
	return m.Run()
}

func TestChallenge1_SubscriptionRollsOverAndFinishes(t *testing.T) {
	ctx := context.Background()
	taskQueue := "test-" + uuid.NewString()
	subscriptionID := "test-" + uuid.NewString()[:8]

	w := worker.New(testClient, taskQueue, worker.Options{
		// Fail the Workflow on a panic instead of retrying its task, so a
		// half-finished TODO fails the test rather than hanging it.
		WorkflowPanicPolicy: worker.FailWorkflow,
	})
	w.RegisterWorkflow(SubscriptionWorkflow)
	w.RegisterActivity(ChargeSubscription)
	if err := w.Start(); err != nil {
		t.Fatalf("could not start a Worker for the test: %v", err)
	}
	defer w.Stop()

	run, err := testClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        subscriptionID,
		TaskQueue: taskQueue,
		// Generous: a Sandbox that has just booted can be slow. A run that never
		// rolls over is terminated by the Service long before this.
		WorkflowExecutionTimeout: 60 * time.Second,
	}, SubscriptionWorkflow, SubscriptionInput{SubscriptionID: subscriptionID, MaxCycles: testCycles})
	if err != nil {
		t.Fatalf("could not start SubscriptionWorkflow: %v", err)
	}
	// Read it now: Get follows the chain and moves the run's ID to the newest run.
	firstRunID := run.GetRunID()

	var result SubscriptionResult
	err = run.Get(ctx, &result)
	var terminated *temporal.TerminatedError
	var timedOut *temporal.TimeoutError
	switch {
	case errors.As(err, &terminated):
		t.Fatalf("The Temporal Service suggested Continue-As-New and the subscription never rolled over, so its "+
			"history hit the limit and the run was terminated. %s: when workflow.GetInfo(ctx).GetContinueAsNewSuggested() "+
			"is true, return workflow.NewContinueAsNewError(...).", todo1)
	case errors.As(err, &timedOut):
		t.Fatalf("The subscription rolled over but never reached %d cycles: each new run started counting from 0 again. "+
			"%s: pass cyclesCompleted in the SubscriptionInput you hand to workflow.NewContinueAsNewError(...).", testCycles, todo1)
	case err != nil:
		t.Fatalf("The subscription failed: %v. Check %s.", err, todo1)
	}

	if result.TotalCycles != testCycles {
		t.Fatalf("Asked for %d cycles and the subscription billed %d. Check %s.", testCycles, result.TotalCycles, todo1)
	}

	firstRun, err := testClient.DescribeWorkflowExecution(ctx, subscriptionID, firstRunID)
	if err != nil {
		t.Fatalf("could not describe the subscription's first run: %v", err)
	}
	if status := firstRun.GetWorkflowExecutionInfo().GetStatus(); status != enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW {
		t.Fatalf("The subscription finished without ever rolling over. %s: return workflow.NewContinueAsNewError(...) "+
			"when the Service suggests it.", todo1)
	}
}
