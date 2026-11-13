// The Checks' tests. Each Challenge's tests are named TestChallenge<N>_... and
// run against a real Temporal dev server that this package starts once, with
// the same two Namespaces and the same Nexus Endpoint as the lab.
//
// Run them all:            go test ./tests/
// Run one Challenge's:     go test ./tests/ -run 'TestChallenge3_'
package tests

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	nexuspb "go.temporal.io/api/nexus/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
)

// The lab's names, written out rather than read from shared/types.go, so a
// test notices if a constant there stops matching the lab's Endpoint.
const (
	paymentsNamespace   = "payments-namespace"
	complianceNamespace = "compliance-namespace"
	complianceQueue     = "compliance-risk"
	complianceEndpoint  = "compliance-endpoint"
	// The tests' own caller Workflows run here, in the Payments Namespace.
	testQueue = "challenge-tests"
)

var (
	once      sync.Once
	server    *testsuite.DevServer
	serverErr error
)

func TestMain(m *testing.M) {
	code := m.Run()
	if server != nil {
		_ = server.Stop()
	}
	os.Exit(code)
}

// quiet keeps Worker and client chatter out of the test output, so a failing
// test's own message is the first thing a learner reads.
var quiet = sdklog.NewStructuredLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))

// clients starts the dev server on first use and returns one client per
// Namespace.
func clients(t *testing.T) (payments, compliance client.Client) {
	t.Helper()
	once.Do(func() { serverErr = startServer() })
	if serverErr != nil {
		t.Fatalf("The test could not start its Temporal dev server: %v", serverErr)
	}
	payments, err := client.NewClientFromExisting(server.Client(), client.Options{Namespace: paymentsNamespace, Logger: quiet})
	if err != nil {
		t.Fatalf("Could not connect to the test dev server: %v", err)
	}
	compliance, err = client.NewClientFromExisting(server.Client(), client.Options{Namespace: complianceNamespace, Logger: quiet})
	if err != nil {
		t.Fatalf("Could not connect to the test dev server: %v", err)
	}
	t.Cleanup(payments.Close)
	t.Cleanup(compliance.Close)
	return payments, compliance
}

func startServer() error {
	path, err := exec.LookPath("temporal")
	if err != nil {
		return errors.New("the temporal CLI is not on PATH")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	server, err = testsuite.StartDevServer(ctx, testsuite.DevServerOptions{
		ExistingPath:  path,
		LogLevel:      "error",
		ClientOptions: &client.Options{Namespace: paymentsNamespace, Logger: quiet},
		ExtraArgs:     []string{"--namespace", complianceNamespace},
	})
	if err != nil {
		return err
	}
	_, err = server.Client().OperatorService().CreateNexusEndpoint(ctx, &operatorservice.CreateNexusEndpointRequest{
		Spec: &nexuspb.EndpointSpec{
			Name: complianceEndpoint,
			Target: &nexuspb.EndpointTarget{Variant: &nexuspb.EndpointTarget_Worker_{
				Worker: &nexuspb.EndpointTarget_Worker{Namespace: complianceNamespace, TaskQueue: complianceQueue},
			}},
		},
	})
	if err != nil {
		return fmt.Errorf("creating the Nexus Endpoint: %w", err)
	}
	return nil
}

// history returns the Event History of a Workflow's latest run.
func history(t *testing.T, c client.Client, workflowID string) []*historypb.HistoryEvent {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var events []*historypb.HistoryEvent
	iter := c.GetWorkflowHistory(ctx, workflowID, "", false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			t.Fatalf("Could not read the Event History of %s: %v", workflowID, err)
		}
		events = append(events, event)
	}
	return events
}

// waitForEvent waits until a Workflow's history holds an event of the given
// type, and reports whether it did before the timeout.
func waitForEvent(t *testing.T, c client.Client, workflowID string, eventType enumspb.EventType, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := c.DescribeWorkflowExecution(context.Background(), workflowID, ""); err == nil {
			for _, event := range history(t, c, workflowID) {
				if event.GetEventType() == eventType {
					return true
				}
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}

// await waits for a Workflow's result. It fails the test, naming where, if the
// Workflow fails, gets stuck, or is still running at the deadline.
func await(t *testing.T, ctx context.Context, c client.Client, run client.WorkflowRun, valuePtr any, where string) {
	t.Helper()
	for {
		short, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := run.Get(short, valuePtr)
		cancel()
		if err == nil {
			return
		}
		desc, derr := c.DescribeWorkflowExecution(context.Background(), run.GetID(), "")
		running := derr == nil &&
			desc.GetWorkflowExecutionInfo().GetStatus() == enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING
		if !running {
			if err = run.Get(context.Background(), valuePtr); err == nil {
				return
			}
			t.Fatalf("%s: %s failed: %s", where, run.GetID(), cause(err))
		}
		if reason := stuckReason(t, c, run.GetID()); reason != "" {
			t.Fatalf("%s: %s is stuck. %s", where, run.GetID(), reason)
		}
		if ctx.Err() != nil {
			t.Fatalf("%s: %s did not finish in time. Open it in the Temporal UI and read its Event History.", where, run.GetID())
		}
	}
}

// stuckReason says why a running Workflow cannot make progress, or returns ""
// if it can.
func stuckReason(t *testing.T, c client.Client, workflowID string) string {
	t.Helper()
	desc, err := c.DescribeWorkflowExecution(context.Background(), workflowID, "")
	if err != nil {
		return ""
	}
	if desc.GetPendingWorkflowTask().GetAttempt() > 2 {
		failure := ""
		for _, event := range history(t, c, workflowID) {
			if attrs := event.GetWorkflowTaskFailedEventAttributes(); attrs != nil {
				failure = firstLine(attrs.GetFailure().GetMessage())
			}
		}
		reason := "Its Workflow Task keeps failing: " + strings.TrimSuffix(failure, ".") + "."
		if strings.Contains(failure, "endpoint") {
			reason += " The lab's Nexus Endpoint is named \"" + complianceEndpoint + "\": check ComplianceEndpoint in shared/types.go."
		}
		return reason
	}
	for _, op := range desc.GetPendingNexusOperations() {
		if op.GetAttempt() > 2 && op.GetLastAttemptFailure() != nil {
			return "Its Nexus Operation " + op.GetOperation() + " keeps failing: " + firstLine(op.GetLastAttemptFailure().GetMessage())
		}
	}
	return ""
}

// cause is the innermost error's message: what actually went wrong, without
// the Workflow and Operation wrappers around it.
func cause(err error) string {
	for {
		next := errors.Unwrap(err)
		if next == nil {
			return firstLine(err.Error())
		}
		err = next
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
