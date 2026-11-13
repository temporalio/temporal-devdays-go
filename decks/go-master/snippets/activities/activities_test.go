package activities

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

var order = Order{OrderID: "o-1", SKU: "sku-1", Quantity: 2}

// ordersAPI serves /submit, answering with the given statuses in turn (then
// 200), and records the Orders it was posted.
func ordersAPI(t *testing.T, statuses ...int) *[]Order {
	t.Helper()
	var mu sync.Mutex
	var posted []Order
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		var o Order
		_ = json.NewDecoder(r.Body).Decode(&o)
		posted = append(posted, o)
		status := http.StatusOK
		if len(statuses) > 0 {
			status, statuses = statuses[0], statuses[1:]
		}
		w.WriteHeader(status)
		if status == http.StatusOK {
			_ = json.NewEncoder(w).Encode(OrderResult{OrderID: o.OrderID, Status: "submitted"})
		}
	}))
	t.Cleanup(srv.Close)
	old := OrdersAPI
	OrdersAPI = srv.URL
	t.Cleanup(func() { OrdersAPI = old })
	return &posted
}

func register(r worker.Registry) {
	r.RegisterWorkflow(OrderWorkflow)
	r.RegisterActivity(SubmitOrder)
}

func start(t *testing.T, wf any) client.WorkflowRun {
	t.Helper()
	tq := testenv.Worker(t, register)
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("order"), TaskQueue: tq,
	}, wf, order)
	if err != nil {
		t.Fatal(err)
	}
	return run
}

func result(t *testing.T, run client.WorkflowRun) OrderResult {
	t.Helper()
	var r OrderResult
	if err := run.Get(testenv.Ctx(t), &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestTheWorkflowSchedulesTheActivityAndReturnsItsResult(t *testing.T) {
	posted := ordersAPI(t)
	got := result(t, start(t, OrderWorkflow))
	if got != (OrderResult{OrderID: "o-1", Status: "submitted"}) {
		t.Fatalf("result %+v", got)
	}
	// The Order crossed into the Activity as JSON and was posted as-is.
	if len(*posted) != 1 || (*posted)[0] != order {
		t.Fatalf("posted %+v", *posted)
	}
}

func TestTheActivityTypeIsTheFunctionName(t *testing.T) {
	ordersAPI(t)
	run := start(t, OrderWorkflow)
	result(t, run)
	iter := testenv.Client().GetWorkflowHistory(context.Background(), run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		if a := e.GetActivityTaskScheduledEventAttributes(); a != nil {
			if a.ActivityType.Name != "SubmitOrder" {
				t.Fatalf("Activity Type %q", a.ActivityType.Name)
			}
			return
		}
	}
	t.Fatal("no ActivityTaskScheduled event")
}

func TestAnActivityThatReturnsAnErrorIsRetriedByThePlatform(t *testing.T) {
	posted := ordersAPI(t, http.StatusServiceUnavailable, http.StatusOK)
	run := start(t, OrderWorkflow)
	if got := result(t, run); got.Status != "submitted" {
		t.Fatalf("result %+v", got)
	}
	iter := testenv.Client().GetWorkflowHistory(context.Background(), run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	var attempt int32
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		if a := e.GetActivityTaskStartedEventAttributes(); a != nil {
			attempt = a.Attempt
		}
	}
	// One ActivityTaskStarted event, recording the attempt that succeeded: the
	// failed attempt was retried inside the same scheduled Activity.
	if attempt != 2 {
		t.Fatalf("succeeded on attempt %d, want 2", attempt)
	}
	if len(*posted) != 2 {
		t.Fatalf("%d calls to the API, want 2", len(*posted))
	}
}
