// ABOUTME: The Track's Check (ADR 0003), one test per part of TODO 1, in the
// order a learner does them, so the first part still missing is the one
// reported. A failing test's message is exactly what a learner sees when Check
// fails, so every message names the file, the function and the TODO.
package shipment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

// devClient talks to the dev server every test in this package shares.
var devClient client.Client

// TestMain starts one real dev server for the package, from the `temporal` CLI
// on PATH when there is one (the Sandbox Image has it, so a Check never
// downloads anything).
func TestMain(m *testing.M) {
	quiet := tlog.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	opts := testsuite.DevServerOptions{
		LogLevel:      "error",
		ClientOptions: &client.Options{Logger: quiet},
	}
	if path, err := exec.LookPath("temporal"); err == nil {
		opts.ExistingPath = path
	}
	server, err := testsuite.StartDevServer(context.Background(), opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not start the Temporal dev server:", err)
		os.Exit(1)
	}
	devClient = server.Client()
	code := m.Run()
	_ = server.Stop()
	os.Exit(code)
}

// beat is one Heartbeat an Activity recorded, and the Attempt that recorded it.
type beat struct {
	attempt int32
	details any
}

// recorder sees every Heartbeat and every Activity error on one test's Worker,
// before the SDK throttles anything.
type recorder struct {
	interceptor.WorkerInterceptorBase
	mu      sync.Mutex
	beats   []beat
	lastErr error
}

func (r *recorder) InterceptActivity(_ context.Context, next interceptor.ActivityInboundInterceptor) interceptor.ActivityInboundInterceptor {
	in := &recordingInbound{root: r}
	in.Next = next
	return in
}

// fromAttempt returns the details of every Heartbeat sent by Attempt n or later.
func (r *recorder) fromAttempt(n int32) []any {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []any{}
	for _, b := range r.beats {
		if b.attempt >= n {
			out = append(out, b.details)
		}
	}
	return out
}

type recordingInbound struct {
	interceptor.ActivityInboundInterceptorBase
	root *recorder
}

func (i *recordingInbound) Init(outbound interceptor.ActivityOutboundInterceptor) error {
	out := &recordingOutbound{root: i.root}
	out.Next = outbound
	return i.Next.Init(out)
}

func (i *recordingInbound) ExecuteActivity(ctx context.Context, in *interceptor.ExecuteActivityInput) (any, error) {
	result, err := i.Next.ExecuteActivity(ctx, in)
	if err != nil && !errors.Is(err, activity.ErrResultPending) {
		i.root.mu.Lock()
		i.root.lastErr = err
		i.root.mu.Unlock()
	}
	return result, err
}

type recordingOutbound struct {
	interceptor.ActivityOutboundInterceptorBase
	root *recorder
}

func (o *recordingOutbound) RecordHeartbeat(ctx context.Context, details ...any) {
	var d any
	if len(details) > 0 {
		d = details[0]
	}
	o.root.mu.Lock()
	o.root.beats = append(o.root.beats, beat{attempt: activity.GetInfo(ctx).Attempt, details: d})
	o.root.mu.Unlock()
	o.Next.RecordHeartbeat(ctx, details...)
}

// ship runs the learner's ShipmentWorkflow on the dev server, on a Worker of
// its own whose PackShipment is pack, and waits up to timeout for the result.
func ship(t *testing.T, pack func(context.Context, ShipmentInput) (ShipmentResult, error), items int, timeout time.Duration) (ShipmentResult, *recorder, error) {
	t.Helper()
	rec := &recorder{}
	queue := TaskQueue + "-" + t.Name()
	w := worker.New(devClient, queue, worker.Options{Interceptors: []interceptor.WorkerInterceptor{rec}})
	w.RegisterWorkflow(ShipmentWorkflow)
	w.RegisterActivityWithOptions(pack, activity.RegisterOptions{Name: "PackShipment"})
	if err := w.Start(); err != nil {
		t.Fatalf("The test Worker did not start: %v. This is a Track bug, not your code: tell the instructor.", err)
	}
	defer w.Stop()

	ctx := context.Background()
	run, err := devClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                       fmt.Sprintf("test-%s-%d", t.Name(), time.Now().UnixNano()),
		TaskQueue:                queue,
		WorkflowExecutionTimeout: timeout,
	}, ShipmentWorkflow, ShipmentInput{ShipmentID: "test", ItemCount: items})
	if err != nil {
		t.Fatalf("Could not start ShipmentWorkflow: %v. This is a Track bug, not your code: tell the instructor.", err)
	}
	var result ShipmentResult
	err = run.Get(ctx, &result)
	return result, rec, err
}

// packOrFail runs the real PackShipment through the Workflow, and turns a
// failure into a message that names the TODOs.
func packOrFail(t *testing.T, pack func(context.Context, ShipmentInput) (ShipmentResult, error), items int, when string) (ShipmentResult, *recorder) {
	t.Helper()
	result, rec, err := ship(t, pack, items, 10*time.Second)
	if err != nil {
		rec.mu.Lock()
		lastErr := rec.lastErr
		rec.mu.Unlock()
		if errors.Is(lastErr, temporal.ErrNoData) {
			t.Fatalf("PackShipment returned %q %s: there was no Heartbeat to read yet. TODO 1c in activities.go, PackShipment: call activity.GetHeartbeatDetails only when activity.HasHeartbeatDetails(ctx) is true, and keep 0 otherwise.", lastErr.Error(), when)
		}
		cause := "it never finished"
		if lastErr != nil {
			cause = "it returned the error: " + lastErr.Error()
		}
		t.Fatalf("PackShipment did not pack %d items %s: %s. Check TODO 1b (activity.RecordHeartbeat(ctx, packed)) and TODO 1c (only read the details when activity.HasHeartbeatDetails(ctx)) in activities.go, PackShipment.", items, when, cause)
	}
	return result, rec
}

// --- TODO 1a: detection -----------------------------------------------------

func TestChallenge1_SilentAttemptIsRetriedQuickly(t *testing.T) {
	// Attempt 1 goes silent and never reports back, the way a crashed Worker
	// does. It sends no Heartbeat: a Heartbeat Timeout counts from the
	// attempt's start.
	workerDiesOnAttempt1 := func(ctx context.Context, input ShipmentInput) (ShipmentResult, error) {
		attempt := activity.GetInfo(ctx).Attempt
		if attempt == 1 {
			return ShipmentResult{}, activity.ErrResultPending
		}
		return ShipmentResult{ShipmentID: input.ShipmentID, Packed: input.ItemCount, FinishedOnAttempt: attempt}, nil
	}

	// Below the 20-second start-to-close timeout, well above a 2-second Heartbeat Timeout.
	result, _, err := ship(t, workerDiesOnAttempt1, 10, 8*time.Second)

	if err != nil || result.FinishedOnAttempt != 2 {
		t.Fatal("A packing attempt went silent and Temporal waited out the whole 20-second start-to-close timeout before retrying. TODO 1a in workflows.go, ShipmentWorkflow: add HeartbeatTimeout: 2 * time.Second to the ActivityOptions.")
	}
}

// --- TODO 1b: send the Heartbeat --------------------------------------------

func TestChallenge1_EveryItemIsHeartbeated(t *testing.T) {
	_, rec := packOrFail(t, PackShipment, 3, "on a first attempt")

	beats := rec.fromAttempt(1)
	if fmt.Sprint(beats) != "[1 2 3]" {
		got := "nothing"
		if len(beats) > 0 {
			got = fmt.Sprint(beats)
		}
		t.Fatalf("Packing 3 items Heartbeated %s. TODO 1b in activities.go, PackShipment: call activity.RecordHeartbeat(ctx, packed) after each item.", got)
	}
}

// --- TODO 1c: read it back --------------------------------------------------

func TestChallenge1_RetryResumesFromLastHeartbeat(t *testing.T) {
	// Attempt 1 Heartbeats 16 packed items and fails. Attempt 2 is the real
	// PackShipment, which should read 16 back and pack only items 17 and 18.
	crashAfter16 := func(ctx context.Context, input ShipmentInput) (ShipmentResult, error) {
		if activity.GetInfo(ctx).Attempt == 1 {
			activity.RecordHeartbeat(ctx, 16)
			return ShipmentResult{}, temporal.NewApplicationError("the Worker was lost after 16 items", "Crash")
		}
		return PackShipment(ctx, input)
	}

	result, rec := packOrFail(t, crashAfter16, 18, "on a retry")

	beats := rec.fromAttempt(2)
	startedAt := 0
	if len(beats) > 0 {
		if n, ok := beats[0].(int); ok {
			startedAt = n - 1
		}
	}
	if startedAt != 16 {
		t.Fatalf("The last attempt Heartbeated 16 packed items, and the retry started again from item %d. TODO 1c in activities.go, PackShipment: when activity.HasHeartbeatDetails(ctx), read the count with activity.GetHeartbeatDetails(ctx, &packed).", startedAt)
	}
	if result.Packed != 18 {
		t.Fatalf("Expected all 18 items packed and got %d. Check TODO 1c in activities.go, PackShipment.", result.Packed)
	}
}

func TestChallenge1_FirstAttemptStartsAtZero(t *testing.T) {
	result, rec := packOrFail(t, PackShipment, 3, "on a first attempt")

	if result.Packed != 3 || fmt.Sprint(rec.fromAttempt(1)) != "[1 2 3]" {
		t.Fatal("A first attempt has no Heartbeat to resume from and must start at item 0. TODO 1c in activities.go, PackShipment: read the details only when activity.HasHeartbeatDetails(ctx) is true, and keep 0 otherwise.")
	}
}
