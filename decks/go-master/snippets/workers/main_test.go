package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	orders "github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/activities"
	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

// The slide's Worker, built against the test's dev server: it runs the
// Workflow it registered, and Run returns cleanly when the interrupt channel
// fires, as it does on Ctrl-C.
func TestTheWorkerRunsUntilInterrupted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var o orders.Order
		_ = json.NewDecoder(r.Body).Decode(&o)
		_ = json.NewEncoder(w).Encode(orders.OrderResult{OrderID: o.OrderID, Status: "submitted"})
	}))
	defer srv.Close()
	orders.OrdersAPI = srv.URL

	w := worker.New(testenv.Client(), "order-task-queue", worker.Options{})
	w.RegisterWorkflow(orders.OrderWorkflow)
	w.RegisterActivity(orders.SubmitOrder)

	interrupt := make(chan any)
	done := make(chan error, 1)
	go func() { done <- w.Run(interrupt) }()

	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("order"), TaskQueue: "order-task-queue",
	}, orders.OrderWorkflow, orders.Order{OrderID: "o-1", SKU: "sku-1", Quantity: 1})
	if err != nil {
		t.Fatal(err)
	}
	var result orders.OrderResult
	if err := run.Get(testenv.Ctx(t), &result); err != nil || result.Status != "submitted" {
		t.Fatalf("result %+v, err %v", result, err)
	}

	close(interrupt)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v after the interrupt", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("Run did not return after the interrupt")
	}
}

// A Workflow started on a Task Queue nobody polls is accepted, and waits.
func TestATaskQueueTypoStartsFineAndWaitsForever(t *testing.T) {
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("order"), TaskQueue: "order-taks-queue",
	}, orders.OrderWorkflow, orders.Order{OrderID: "o-2"})
	if err != nil {
		t.Fatalf("the start should be accepted: %v", err)
	}
	time.Sleep(2 * time.Second)
	desc, err := testenv.Client().DescribeWorkflowExecution(testenv.Ctx(t), run.GetID(), run.GetRunID())
	if err != nil {
		t.Fatal(err)
	}
	if s := desc.WorkflowExecutionInfo.Status; s != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
		t.Fatalf("status %v, want Running", s)
	}
	_ = testenv.Client().TerminateWorkflow(testenv.Ctx(t), run.GetID(), "", "test over")
}
