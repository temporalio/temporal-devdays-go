// ABOUTME: The tests behind each Challenge's Check (ADR 0003), one group per
// Challenge, run against a real dev server started from the temporal CLI. Each
// failure message is for the learner: it names the file, the function and the TODO.
package saga

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

var testClient client.Client

func TestMain(m *testing.M) {
	path, err := exec.LookPath("temporal")
	if err != nil {
		fmt.Println("The tests need the temporal CLI on PATH to start a dev server: https://docs.temporal.io/cli")
		os.Exit(1)
	}
	quiet := tlog.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr,
		&slog.HandlerOptions{Level: slog.LevelError + 4})))
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{
		ExistingPath:  path,
		LogLevel:      "error",
		ClientOptions: &client.Options{Logger: quiet},
	})
	if err != nil {
		fmt.Println("Could not start the Temporal dev server:", err)
		os.Exit(1)
	}
	testClient = server.Client()
	code := m.Run()
	_ = server.Stop()
	os.Exit(code)
}

// ---- Challenge 1: unwind (TODO 1) -------------------------------------------

func TestChallenge1_UnwindsNewestFirst(t *testing.T) {
	tq := startWorker(t, nil)
	order := newOrder(t)
	order.FailDispatch = true

	_, err := runOrder(t, tq, order)
	got := ledgerFor(order.OrderID)

	switch {
	case slices.Equal(got, []string{"reserve", "charge"}):
		t.Fatal("workflow.go, OrderFulfillmentWorkflow, TODO 1: an order failed at the courier and nothing was undone. The ledger stops at reserve, charge. Add the defer that runs the compensations when err is not nil.")
	case slices.Equal(got, []string{"reserve", "charge", "release", "refund"}):
		t.Fatal("workflow.go, OrderFulfillmentWorkflow, TODO 1: the undos ran oldest first, so the stock went back before the card was refunded. Walk compensations backwards, from the last index to 0.")
	case !slices.Equal(got, []string{"reserve", "charge", "refund", "release"}):
		t.Fatalf("workflow.go, OrderFulfillmentWorkflow, TODO 1: after a failed dispatch the ledger should read reserve, charge, refund, release. It reads: %s.", strings.Join(got, ", "))
	}
	if err == nil {
		t.Fatal("workflow.go, OrderFulfillmentWorkflow, TODO 1: the order failed at the courier, but the Workflow completed. Leave err as it is in the defer: compensating is not succeeding.")
	}
	if !hasErrorType(err, "CourierUnavailable") {
		t.Fatalf("workflow.go, OrderFulfillmentWorkflow, TODO 1: the Workflow should fail with the courier's error once the undos have run, but it failed with: %v", err)
	}
}

func TestChallenge1_CleanOrderRunsNoUndo(t *testing.T) {
	tq := startWorker(t, nil)
	order := newOrder(t)

	result, err := runOrder(t, tq, order)
	if err != nil {
		t.Fatalf("workflow.go, OrderFulfillmentWorkflow, TODO 1: a clean order failed: %v", err)
	}
	if got := ledgerFor(order.OrderID); !slices.Equal(got, []string{"reserve", "charge", "dispatch"}) {
		t.Fatalf("workflow.go, OrderFulfillmentWorkflow, TODO 1: a clean order should only reserve, charge and dispatch, but the ledger reads: %s. Return from the defer when err is nil.", strings.Join(got, ", "))
	}
	if result.TrackingNumber != "TRK-"+order.OrderID {
		t.Fatalf("workflow.go, OrderFulfillmentWorkflow: a clean order returned tracking number %q, want %q.", result.TrackingNumber, "TRK-"+order.OrderID)
	}
}

func TestChallenge1_OneFailingUndoDoesNotStrandTheRest(t *testing.T) {
	failingRefund := func(ctx context.Context, order Order) error {
		return temporal.NewNonRetryableApplicationError("refund service down", "RefundFailed", nil)
	}
	tq := startWorker(t, map[string]any{"RefundPaymentIfCharged": failingRefund})
	order := newOrder(t)
	order.FailDispatch = true

	_, _ = runOrder(t, tq, order)
	if got := ledgerFor(order.OrderID); !slices.Contains(got, "release") {
		t.Fatalf("workflow.go, OrderFulfillmentWorkflow, TODO 1: when the refund failed, the stock was never released (ledger: %s). Log a failed compensation and keep unwinding; do not return from the loop.", strings.Join(got, ", "))
	}
}

func TestChallenge1_UndosRunWhenCancelled(t *testing.T) {
	dispatching := make(chan struct{})
	var once sync.Once
	slowDispatch := func(ctx context.Context, order Order) (string, error) {
		once.Do(func() { close(dispatching) })
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(30 * time.Second):
			return "", errors.New("dispatch was never cancelled")
		}
	}
	tq := startWorker(t, map[string]any{"DispatchCourier": slowDispatch})
	order := newOrder(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	run, err := testClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID: "fulfillment-" + order.OrderID, TaskQueue: tq,
	}, OrderFulfillmentWorkflow, order)
	if err != nil {
		t.Fatalf("could not start the Workflow: %v", err)
	}
	select {
	case <-dispatching:
	case <-ctx.Done():
		t.Fatal("workflow.go, OrderFulfillmentWorkflow: the order never reached DispatchCourier.")
	}
	if err := testClient.CancelWorkflow(ctx, run.GetID(), ""); err != nil {
		t.Fatalf("could not cancel the Workflow: %v", err)
	}
	_ = run.Get(ctx, nil)

	if got := ledgerFor(order.OrderID); !slices.Equal(got, []string{"reserve", "charge", "refund", "release"}) {
		t.Fatalf("workflow.go, OrderFulfillmentWorkflow, TODO 1: the order was cancelled after the charge landed, and the ledger reads: %s. Run the undos on the context from workflow.NewDisconnectedContext(ctx), which the cancellation never reaches.", strings.Join(got, ", "))
	}
}

// ---- Challenge 2: register before you run (TODO 2a, TODO 2b) ----------------

func TestChallenge2_ReleaseIsRegisteredBeforeReserve(t *testing.T) {
	tq := startWorker(t, nil)
	order := newOrder(t)
	order.FailReserve = true

	_, _ = runOrder(t, tq, order)
	if !slices.Contains(scheduledActivities(t, order), "ReleaseInventoryIfReserved") {
		t.Fatal("workflow.go, OrderFulfillmentWorkflow, TODO 2a: ReserveInventory failed and ReleaseInventoryIfReserved was never called. Register the release above the ReserveInventory call, so the undo is on the stack before the step runs.")
	}
}

func TestChallenge2_RefundIsRegisteredBeforeCharge(t *testing.T) {
	tq := startWorker(t, nil)
	order := newOrder(t)
	order.ChargeFailsAfterEffect = true

	_, err := runOrder(t, tq, order)
	got := ledgerFor(order.OrderID)
	if slices.Equal(got, []string{"reserve", "charge", "release"}) {
		t.Fatal("workflow.go, OrderFulfillmentWorkflow, TODO 2b: the charge landed and then ChargePayment reported failure, so the refund was never registered and the money is still gone. Register the refund above the ChargePayment call.")
	}
	if !slices.Equal(got, []string{"reserve", "charge", "refund", "release"}) {
		t.Fatalf("workflow.go, OrderFulfillmentWorkflow, TODO 2b: after a charge that landed and then failed, the ledger should read reserve, charge, refund, release. It reads: %s.", strings.Join(got, ", "))
	}
	if !hasErrorType(err, "PaymentGatewayTimeout") {
		t.Fatalf("workflow.go, OrderFulfillmentWorkflow: the Workflow should fail with the gateway's error once the undos have run, but it ended with: %v", err)
	}
}

// ---- Challenge 3: make the undo safe to repeat (TODO 3a, TODO 3b) -----------

func TestChallenge3_ReleaseWithNothingReservedIsANoOp(t *testing.T) {
	tq := startWorker(t, nil)
	order := newOrder(t)
	order.FailReserve = true

	_, _ = runOrder(t, tq, order)
	if got := ledgerFor(order.OrderID); slices.Contains(got, "release") {
		t.Fatal("activities.go, ReleaseInventoryIfReserved, TODO 3a: ReserveInventory failed before reserving anything, and the undo released stock anyway. Return early when isReserved(order.OrderID) is false.")
	}
}

func TestChallenge3_UndosAreSafeToRepeat(t *testing.T) {
	tq := startWorker(t, nil)
	order := newOrder(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	run, err := testClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID: "undo-twice-" + order.OrderID, TaskQueue: tq,
	}, undoTwiceWorkflow, order)
	if err != nil {
		t.Fatalf("could not start the Workflow: %v", err)
	}
	if err := run.Get(ctx, nil); err != nil {
		t.Fatalf("a Workflow that runs each undo twice failed: %v", err)
	}

	got := ledgerFor(order.OrderID)
	if n := countOf(got, "release"); n != 1 {
		t.Fatalf("activities.go, ReleaseInventoryIfReserved, TODO 3a: ReleaseInventoryIfReserved ran twice for one reservation and released %d times. Return early when isReserved(order.OrderID) is false.", n)
	}
	if n := countOf(got, "refund"); n != 1 {
		t.Fatalf("activities.go, RefundPaymentIfCharged, TODO 3b: RefundPaymentIfCharged ran twice for one charge and refunded %d times. Temporal retries a compensation like any other Activity. Return early when isCharged(order.OrderID) is false.", n)
	}
}

// undoTwiceWorkflow reserves and charges, then runs each undo twice, the way a
// retried compensation would.
func undoTwiceWorkflow(ctx workflow.Context, order Order) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second})
	for _, step := range []any{ReserveInventory, ChargePayment,
		RefundPaymentIfCharged, RefundPaymentIfCharged,
		ReleaseInventoryIfReserved, ReleaseInventoryIfReserved} {
		if err := workflow.ExecuteActivity(ctx, step, order).Get(ctx, nil); err != nil {
			return err
		}
	}
	return nil
}

// ---- helpers ------------------------------------------------------------------

// startWorker runs a Worker on a Task Queue of its own, with the real Workflow
// and Activities, and any Activity swapped out by name.
func startWorker(t *testing.T, overrides map[string]any) string {
	t.Helper()
	tq := fmt.Sprintf("saga-test-%s-%d", t.Name(), time.Now().UnixNano())
	w := worker.New(testClient, tq, worker.Options{})
	w.RegisterWorkflow(OrderFulfillmentWorkflow)
	w.RegisterWorkflow(undoTwiceWorkflow)
	activities := map[string]any{
		"ReserveInventory":           ReserveInventory,
		"ChargePayment":              ChargePayment,
		"DispatchCourier":            DispatchCourier,
		"ReleaseInventoryIfReserved": ReleaseInventoryIfReserved,
		"RefundPaymentIfCharged":     RefundPaymentIfCharged,
	}
	for name, fn := range overrides {
		activities[name] = fn
	}
	for name, fn := range activities {
		w.RegisterActivityWithOptions(fn, activity.RegisterOptions{Name: name})
	}
	if err := w.Start(); err != nil {
		t.Fatalf("could not start a test Worker: %v", err)
	}
	t.Cleanup(w.Stop)
	return tq
}

func newOrder(t *testing.T) Order {
	return Order{
		OrderID:     fmt.Sprintf("ORD-%s-%d", strings.ReplaceAll(t.Name(), "/", "-"), time.Now().UnixNano()),
		SKU:         "widget",
		Quantity:    1,
		AmountCents: 4999,
	}
}

func runOrder(t *testing.T, tq string, order Order) (OrderResult, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	run, err := testClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID: "fulfillment-" + order.OrderID, TaskQueue: tq,
	}, OrderFulfillmentWorkflow, order)
	if err != nil {
		t.Fatalf("could not start the Workflow: %v", err)
	}
	var result OrderResult
	err = run.Get(ctx, &result)
	if ctx.Err() != nil {
		t.Fatal("workflow.go, OrderFulfillmentWorkflow: the order did not finish in 30 seconds. Check the last change you made.")
	}
	return result, err
}

// scheduledActivities reads the order's Event History and returns the Activity
// types it scheduled, in order.
func scheduledActivities(t *testing.T, order Order) []string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	iter := testClient.GetWorkflowHistory(ctx, "fulfillment-"+order.OrderID, "", false,
		enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	var names []string
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			t.Fatalf("could not read the Event History: %v", err)
		}
		if attrs := event.GetActivityTaskScheduledEventAttributes(); attrs != nil {
			names = append(names, attrs.GetActivityType().GetName())
		}
	}
	return names
}

func hasErrorType(err error, errType string) bool {
	var appErr *temporal.ApplicationError
	return errors.As(err, &appErr) && appErr.Type() == errType
}

func countOf(entries []string, action string) int {
	n := 0
	for _, e := range entries {
		if e == action {
			n++
		}
	}
	return n
}
