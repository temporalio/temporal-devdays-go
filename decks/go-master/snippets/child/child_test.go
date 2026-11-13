package child

import (
	"testing"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func TestTheChildReturnsItsResultUnderItsOwnWorkflowID(t *testing.T) {
	tq := testenv.Worker(t, func(r worker.Registry) {
		r.RegisterWorkflow(OrderWorkflow)
		r.RegisterWorkflow(ShipmentWorkflow)
	})
	c := testenv.Client()
	orderID := testenv.ID("order")
	run, err := c.ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: testenv.ID("parent"), TaskQueue: tq},
		OrderWorkflow, Order{OrderID: orderID, Shipment: "box-7"})
	if err != nil {
		t.Fatal(err)
	}
	var tracking string
	if err := run.Get(testenv.Ctx(t), &tracking); err != nil {
		t.Fatal(err)
	}
	if tracking != "TRACK-box-7" {
		t.Fatalf("tracking %q", tracking)
	}

	// The child is its own Workflow Execution, findable by the ID the parent chose.
	desc, err := c.DescribeWorkflowExecution(testenv.Ctx(t), "shipment-"+orderID, "")
	if err != nil {
		t.Fatal(err)
	}
	info := desc.WorkflowExecutionInfo
	if info.Type.Name != "ShipmentWorkflow" {
		t.Fatalf("child type %q", info.Type.Name)
	}
	if info.Status != enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED {
		t.Fatalf("child status %v", info.Status)
	}
	if info.ParentExecution.GetWorkflowId() != run.GetID() {
		t.Fatalf("child's parent is %q", info.ParentExecution.GetWorkflowId())
	}
}
