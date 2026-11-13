package heartbeat

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

// A real dev server, not the in-memory test environment: keeping the last
// Heartbeat's details when an attempt fails is the Service's job, and it is
// the whole point of the Heartbeat slide.
func TestMain(m *testing.M) { testenv.Main(m) }

func register(r worker.Registry) {
	r.RegisterWorkflow(PackWorkflow)
	r.RegisterWorkflow(ProcessWorkflow)
	r.RegisterActivity(PackShipment)
	r.RegisterActivity(ProcessItems)
}

func TestARetryResumesFromTheLastHeartbeatsDetails(t *testing.T) {
	mu.Lock()
	packed = nil
	mu.Unlock()
	tq := testenv.Worker(t, register)
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: testenv.ID("pack"), TaskQueue: tq},
		PackWorkflow, ShipmentInput{ShipmentID: "s-1", ItemCount: 5})
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := run.Get(testenv.Ctx(t), &n); err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("packed %d", n)
	}
	// Attempt 1 dies on item 2. Attempt 2 starts at item 2, not item 0.
	mu.Lock()
	defer mu.Unlock()
	if !slices.Equal(packed, []int{0, 1, 2, 2, 3, 4}) {
		t.Fatalf("packed %v", packed)
	}
}

func TestCancellationArrivesAtAHeartbeatAndTheActivityCleansUp(t *testing.T) {
	mu.Lock()
	locksReleased = nil
	mu.Unlock()
	tq := testenv.Worker(t, register)
	items := make([]string, 100)
	for i := range items {
		items[i] = fmt.Sprintf("item-%d", i)
	}
	c := testenv.Client()
	run, err := c.ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: testenv.ID("process"), TaskQueue: tq},
		ProcessWorkflow, ProcessInput{Items: items})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	if err := c.CancelWorkflow(testenv.Ctx(t), run.GetID(), ""); err != nil {
		t.Fatal(err)
	}
	err = run.Get(testenv.Ctx(t), nil)
	var canceled *temporal.CanceledError
	if !errors.As(err, &canceled) {
		t.Fatalf("want a cancellation, got %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if !slices.Equal(locksReleased, []bool{true}) {
		t.Fatalf("locks released %v", locksReleased)
	}
}
