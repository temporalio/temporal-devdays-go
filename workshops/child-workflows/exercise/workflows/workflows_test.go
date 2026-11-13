// ABOUTME: One group of tests per Challenge, named TestChallenge<N>_..., so a Check
// runs exactly the Challenges done so far. Every failure message is what a
// learner sees on a failed Check: it names the file, the function and the TODO.
//
// Challenges 1 to 3 are run and observe, so their tests prove what the shipped
// Exercise already does. Challenge 4 is the one TODO.
package workflows_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/workshops/child-workflows/workflows"
)

var items = []string{"mug", "book", "lamp"}

// c is a Client of the dev server TestMain starts for this package.
var c client.Client

// TestMain starts a real Temporal dev server and one Worker for every test, so
// each Workflow a test runs writes a real Event History.
func TestMain(m *testing.M) { os.Exit(runTests(m)) }

func runTests(m *testing.M) int {
	quiet := tlog.NewStructuredLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
	options := testsuite.DevServerOptions{LogLevel: "error", ClientOptions: &client.Options{Logger: quiet}}
	// The Sandbox Image has the temporal CLI on the PATH, so a Check never
	// downloads a server. Elsewhere the SDK downloads one.
	if path, err := exec.LookPath("temporal"); err == nil {
		options.ExistingPath = path
	}
	server, err := testsuite.StartDevServer(context.Background(), options)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not start the Temporal dev server for the tests:", err)
		return 1
	}
	defer func() { _ = server.Stop() }()
	c = server.Client()

	// A panic in Workflow code fails the Workflow instead of retrying its task,
	// so a half-finished TODO fails the test rather than hanging it.
	w := worker.New(c, workflows.TaskQueue, worker.Options{WorkflowPanicPolicy: worker.FailWorkflow})
	w.RegisterWorkflow(workflows.OrderWorkflow)
	w.RegisterWorkflow(workflows.ShipmentWorkflow)
	w.RegisterWorkflow(workflows.ChainWorkflow)
	w.RegisterWorkflow(workflows.PackWorkflow)
	w.RegisterWorkflow(workflows.PackItemWorkflow)
	if err := w.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Could not start the Worker for the tests:", err)
		return 1
	}
	defer w.Stop()
	return m.Run()
}

func shortID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func newOrder(items ...string) workflows.Order {
	return workflows.Order{OrderID: "T-" + shortID(), Items: items}
}

// execute runs one Workflow to the end and stores its result in valuePtr.
func execute(t *testing.T, id string, workflow, arg, valuePtr any) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                       id,
		TaskQueue:                workflows.TaskQueue,
		WorkflowExecutionTimeout: 30 * time.Second,
	}, workflow, arg)
	if err != nil {
		t.Fatalf("Could not start Workflow %s: %v", id, err)
	}
	return run.Get(ctx, valuePtr)
}

// history reads one Workflow's Event History. Empty if it does not exist.
func history(t *testing.T, workflowID string) []*historypb.HistoryEvent {
	t.Helper()
	iter := c.GetWorkflowHistory(context.Background(), workflowID, "", false,
		enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	var events []*historypb.HistoryEvent
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return events
		}
		events = append(events, event)
	}
	return events
}

// parentOf is the Workflow ID of the parent that started workflowID, or "".
func parentOf(t *testing.T, workflowID string) string {
	t.Helper()
	events := history(t, workflowID)
	if len(events) == 0 {
		return ""
	}
	return events[0].GetWorkflowExecutionStartedEventAttributes().GetParentWorkflowExecution().GetWorkflowId()
}

// indexesOf is the position of every event of one type in an Event History.
func indexesOf(events []*historypb.HistoryEvent, eventType enumspb.EventType) []int {
	var at []int
	for i, e := range events {
		if e.GetEventType() == eventType {
			at = append(at, i)
		}
	}
	return at
}

func TestChallenge1_OrderWaitsForItsShipmentChild(t *testing.T) {
	order := newOrder("mug")
	id := "test-order-" + order.OrderID
	var tracking string
	if err := execute(t, id, workflows.OrderWorkflow, order, &tracking); err != nil {
		t.Fatalf("OrderWorkflow in workflows/order.go failed: %v", err)
	}
	if want := "TRACK-" + order.OrderID; tracking != want {
		t.Fatalf("OrderWorkflow in workflows/order.go returned %q. It must return what its ShipmentWorkflow child returns, %q", tracking, want)
	}
	if parent := parentOf(t, "shipment-"+order.OrderID); parent != id {
		t.Fatalf("OrderWorkflow in workflows/order.go must start shipment-<orderId> as its child")
	}
}

func TestChallenge2_ChainCountsItsLevels(t *testing.T) {
	root := "test-chain-" + shortID()
	var levels int
	if err := execute(t, root, workflows.ChainWorkflow, workflows.ChainRequest{Depth: 3}, &levels); err != nil {
		t.Fatalf("ChainWorkflow in workflows/chain.go failed: %v", err)
	}
	if levels != 3 {
		t.Fatalf("ChainWorkflow in workflows/chain.go returned %d. It must return how many levels of children ran below it, 3", levels)
	}
	if parent := parentOf(t, root+"/2/1/0"); parent != root+"/2/1" {
		t.Fatalf("ChainWorkflow in workflows/chain.go must name each child <parent ID>/<depth>")
	}
}

func TestChallenge2_BottomFailureIsWrappedOncePerLevel(t *testing.T) {
	root := "test-chain-" + shortID()
	err := execute(t, root, workflows.ChainWorkflow, workflows.ChainRequest{Depth: 3, FailAtBottom: true}, nil)

	var failed *temporal.WorkflowExecutionError
	if !errors.As(err, &failed) {
		t.Fatalf("ChainWorkflow in workflows/chain.go must fail when the bottom of the chain fails")
	}
	wrappers := 0
	cause := errors.Unwrap(failed)
	for {
		child, ok := cause.(*temporal.ChildWorkflowExecutionError)
		if !ok {
			break
		}
		wrappers++
		cause = child.Unwrap()
	}
	if wrappers != 3 {
		t.Fatalf("A three-level chain must wrap the bottom failure in a ChildWorkflowExecutionError three times, not %d", wrappers)
	}
	var app *temporal.ApplicationError
	if !errors.As(cause, &app) || app.Type() != "BottomOfChain" {
		t.Fatalf("The bottom of the chain must fail with an ApplicationError of type BottomOfChain, not %v", cause)
	}
}

func TestChallenge3_PackStartsOneChildPerItem(t *testing.T) {
	order := newOrder(items...)
	id := "test-pack-" + order.OrderID
	var boxes []string
	if err := execute(t, id, workflows.PackWorkflow, order, &boxes); err != nil {
		t.Fatalf("PackWorkflow in workflows/pack.go failed: %v", err)
	}

	events := history(t, id)
	if n := len(indexesOf(events, enumspb.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_INITIATED)); n != len(items) {
		t.Fatalf("PackWorkflow in workflows/pack.go started %d children for %d items. It must start one PackItemWorkflow child per item", n, len(items))
	}
	if n := len(indexesOf(events, enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED)); n != len(items) {
		t.Fatalf("PackWorkflow in workflows/pack.go must wait for every child to start")
	}
}

func TestChallenge4_PackReturnsEveryBoxInItemOrder(t *testing.T) {
	order := newOrder(items...)
	var boxes []string
	if err := execute(t, "test-pack-"+order.OrderID, workflows.PackWorkflow, order, &boxes); err != nil {
		t.Fatalf("PackWorkflow in workflows/pack.go failed: %v", err)
	}
	if want := []string{"box:mug", "box:book", "box:lamp"}; !slices.Equal(boxes, want) {
		t.Fatalf("PackWorkflow in workflows/pack.go returned %q. Call Get on every Future in boxes, in item order, and return the boxes they hold (TODO 1)", boxes)
	}
}

func TestChallenge4_PackWaitsForEveryChildInParallel(t *testing.T) {
	order := newOrder(items...)
	id := "test-pack-" + order.OrderID
	var boxes []string
	if err := execute(t, id, workflows.PackWorkflow, order, &boxes); err != nil {
		t.Fatalf("PackWorkflow in workflows/pack.go failed: %v", err)
	}

	events := history(t, id)
	initiated := indexesOf(events, enumspb.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_INITIATED)
	completed := indexesOf(events, enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED)
	if len(completed) != len(items) {
		t.Fatalf("PackWorkflow in workflows/pack.go returned before its children finished, so the Parent Close Policy terminated them. Call Get on every Future in boxes (TODO 1)")
	}
	if initiated[len(initiated)-1] > completed[0] {
		t.Fatalf("PackWorkflow in workflows/pack.go must start every child before waiting for any of them. Leave the fan-out loop as it is and wait only in TODO 1")
	}
}
