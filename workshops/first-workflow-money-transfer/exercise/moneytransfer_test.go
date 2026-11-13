// ABOUTME: Proves the Solution, and the coding half of the Track's Check (ADR 0003).
// A failing test's message is exactly what a learner sees when Check fails, so
// every message names the file, the function and the TODO.
package moneytransfer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

// server is a real Temporal dev server, started once for every test here from
// the `temporal` CLI on the PATH, so a Check never downloads anything.
var server *testsuite.DevServer

func TestMain(m *testing.M) {
	quiet := sdklog.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	opts := testsuite.DevServerOptions{LogLevel: "error", ClientOptions: &client.Options{Logger: quiet}}
	if path, err := exec.LookPath("temporal"); err == nil {
		opts.ExistingPath = path
	}
	s, err := testsuite.StartDevServer(context.Background(), opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not start the Temporal dev server for the tests:", err)
		os.Exit(1)
	}
	server = s
	code := m.Run()
	_ = s.Stop()
	os.Exit(code)
}

func details(amount int) PaymentDetails {
	return PaymentDetails{
		SourceAccount: "A-100",
		TargetAccount: "B-200",
		Amount:        amount,
		ReferenceID:   fmt.Sprintf("%06x", rand.N(1<<24)),
	}
}

// transfer runs one whole transfer on a test Worker and returns its result, or
// the error that stopped it.
func transfer(t *testing.T, amount int) (string, error) {
	t.Helper()
	c := server.Client()
	w := worker.New(c, TaskQueue, worker.Options{})
	w.RegisterWorkflow(MoneyTransferWorkflow)
	w.RegisterActivity(Withdraw)
	w.RegisterActivity(Deposit)
	if err := w.Start(); err != nil {
		t.Fatalf("The test Worker did not start: %v", err)
	}
	defer w.Stop()

	ctx := context.Background()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("test-%x", rand.Uint64()),
		TaskQueue: TaskQueue,
		// A deposit that keeps failing retries forever; stop waiting after 10 seconds.
		WorkflowExecutionTimeout: 10 * time.Second,
	}, MoneyTransferWorkflow, details(amount))
	if err != nil {
		return "", err
	}
	var result string
	err = run.Get(ctx, &result)
	return result, err
}

func TestChallenge1_SmallTransferCompletes(t *testing.T) {
	result, err := transfer(t, 100)
	if err != nil || !strings.HasPrefix(result, "Transfer complete") {
		t.Fatalf("A $100 transfer did not complete (got %q, error: %v). You only run code in Challenge 1: "+
			"if you edited exercise/workflow.go or exercise/activities.go, undo it.", result, err)
	}
}

func TestChallenge2_LargeDepositSucceeds(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(Deposit)
	if _, err := env.ExecuteActivity(Deposit, details(2500)); err != nil {
		if cause := errors.Unwrap(err); cause != nil {
			err = cause
		}
		t.Fatalf("A $2,500 deposit still fails with %q. TODO 1 in exercise/activities.go, Deposit: "+
			"the audit note's slice has room for two lines but holds none, so note[0] is out of range. "+
			"Give it a length of 2.", err.Error())
	}
}

func TestSolution_LargeTransferCompletes(t *testing.T) {
	result, err := transfer(t, 2500)
	if err != nil || !strings.HasPrefix(result, "Transfer complete") {
		t.Fatalf("A $2,500 transfer did not complete (got %q, error: %v).", result, err)
	}
}
