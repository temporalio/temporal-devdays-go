package saga

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func register(r worker.Registry) {
	r.RegisterWorkflow(OrderFulfillmentWorkflow)
	r.RegisterWorkflow(BuggyFulfillmentWorkflow)
	r.RegisterActivity(ReserveInventory)
	r.RegisterActivity(ChargePayment)
	r.RegisterActivity(DispatchCourier)
	r.RegisterActivity(ReleaseInventoryIfReserved)
	r.RegisterActivity(RefundPaymentIfCharged)
}

func newOrder(switches func(*Order)) Order {
	o := Order{OrderID: testenv.ID("order"), SKU: "widget", Quantity: 1, AmountCents: 4999}
	if switches != nil {
		switches(&o)
	}
	return o
}

func fulfil(t *testing.T, o Order, reg func(worker.Registry)) (OrderResult, error) {
	t.Helper()
	tq := testenv.Worker(t, reg)
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("fulfil"), TaskQueue: tq,
	}, OrderFulfillmentWorkflow, o)
	if err != nil {
		t.Fatal(err)
	}
	var res OrderResult
	err = run.Get(testenv.Ctx(t), &res)
	return res, err
}

// failureType checks the Workflow failed with the Activity failure it
// returned, and gives the type of that Activity's own ApplicationError.
func failureType(t *testing.T, err error) string {
	t.Helper()
	var wfErr *temporal.WorkflowExecutionError
	if !errors.As(err, &wfErr) {
		t.Fatalf("the Workflow should have failed, got %v", err)
	}
	var actErr *temporal.ActivityError
	if !errors.As(err, &actErr) {
		t.Fatalf("the Workflow should fail with the Activity's failure, got %v", err)
	}
	var appErr *temporal.ApplicationError
	if !errors.As(actErr, &appErr) {
		t.Fatalf("no ApplicationError under %v", actErr)
	}
	return appErr.Type()
}

func expectLedger(t *testing.T, orderID string, want ...string) {
	t.Helper()
	if got := ForOrder(orderID); !slices.Equal(got, want) {
		t.Fatalf("ledger %v, want %v", got, want)
	}
}

func TestTheHappyPathRunsNoUndo(t *testing.T) {
	o := newOrder(nil)
	res, err := fulfil(t, o, register)
	if err != nil {
		t.Fatal(err)
	}
	if res != (OrderResult{OrderID: o.OrderID, TrackingNumber: "TRK-" + o.OrderID}) {
		t.Fatalf("result %+v", res)
	}
	expectLedger(t, o.OrderID, "reserve", "charge", "dispatch")
}

func TestAFailedDispatchUnwindsNewestFirstAndReturnsTheOriginalFailure(t *testing.T) {
	o := newOrder(func(o *Order) { o.FailDispatch = true })
	_, err := fulfil(t, o, register)
	if got := failureType(t, err); got != "CourierUnavailable" {
		t.Fatalf("failure type %q", got)
	}
	expectLedger(t, o.OrderID, "reserve", "charge", "refund", "release")
}

func TestAChargeThatLandsThenFailsIsStillRefunded(t *testing.T) {
	o := newOrder(func(o *Order) { o.ChargeFailsAfterEffect = true })
	_, err := fulfil(t, o, register)
	if got := failureType(t, err); got != "PaymentGatewayTimeout" {
		t.Fatalf("failure type %q", got)
	}
	expectLedger(t, o.OrderID, "reserve", "charge", "refund", "release")
}

func TestAnUndoWhoseStepNeverLandedIsANoOp(t *testing.T) {
	o := newOrder(func(o *Order) { o.FailReserve = true })
	_, err := fulfil(t, o, register)
	if got := failureType(t, err); got != "InventoryUnavailable" {
		t.Fatalf("failure type %q", got)
	}
	// ReleaseInventoryIfReserved ran (it was registered first) and found nothing.
	expectLedger(t, o.OrderID)
}

func TestOneFailingCompensationDoesNotStrandTheRest(t *testing.T) {
	o := newOrder(func(o *Order) { o.FailDispatch = true })
	broken := func(r worker.Registry) {
		r.RegisterWorkflow(OrderFulfillmentWorkflow)
		r.RegisterActivity(ReserveInventory)
		r.RegisterActivity(ChargePayment)
		r.RegisterActivity(DispatchCourier)
		r.RegisterActivity(ReleaseInventoryIfReserved)
		r.RegisterActivityWithOptions(func(ctx context.Context, order Order) error {
			return temporal.NewNonRetryableApplicationError("refund service down", "RefundUnavailable", nil)
		}, activity.RegisterOptions{Name: "RefundPaymentIfCharged"})
	}
	_, err := fulfil(t, o, broken)
	// The original failure, not the compensation's, is what the caller sees.
	if got := failureType(t, err); got != "CourierUnavailable" {
		t.Fatalf("failure type %q", got)
	}
	expectLedger(t, o.OrderID, "reserve", "charge", "release")
}

func TestABugInWorkflowCodeRetriesTheTaskAndRefundsNothing(t *testing.T) {
	o := newOrder(nil)
	tq := testenv.Worker(t, register)
	c := testenv.Client()
	run, err := c.ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("buggy"), TaskQueue: tq,
	}, BuggyFulfillmentWorkflow, o)
	if err != nil {
		t.Fatal(err)
	}
	defer c.TerminateWorkflow(context.Background(), run.GetID(), "", "test over")
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if taskFailed(t, run) {
			desc, err := c.DescribeWorkflowExecution(testenv.Ctx(t), run.GetID(), "")
			if err != nil {
				t.Fatal(err)
			}
			if s := desc.WorkflowExecutionInfo.Status; s != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
				t.Fatalf("a panic should leave the Workflow running, got %v", s)
			}
			expectLedger(t, o.OrderID, "reserve")
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("no WorkflowTaskFailed event")
}

func taskFailed(t *testing.T, run client.WorkflowRun) bool {
	iter := testenv.Client().GetWorkflowHistory(testenv.Ctx(t), run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		if e.GetWorkflowTaskFailedEventAttributes() != nil {
			return true
		}
	}
	return false
}

func TestAnUndoRegisteredAfterItsStepUsesTheForwardResult(t *testing.T) {
	ResetJournal()
	tq := testenv.Worker(t, func(r worker.Registry) {
		r.RegisterWorkflow(TitleTransferWorkflow)
		r.RegisterActivity(CheckFraud)
		r.RegisterActivity(ProcessPayment)
		r.RegisterActivity(RefundPayment)
		r.RegisterActivity(TransferTitle)
		r.RegisterActivity(RevertTitleTransfer)
	})
	o := TitleOrder{OrderID: testenv.ID("title"), FailAt: "transferTitle"}
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("title"), TaskQueue: tq,
	}, TitleTransferWorkflow, o)
	if err != nil {
		t.Fatal(err)
	}
	failureType(t, run.Get(testenv.Ctx(t), nil))
	// The title transfer failed before its undo was registered, so only the
	// refund runs, with the confirmation the payment returned.
	want := []string{"checkFraud", "processPayment", "transferTitle", "refundPayment:conf-" + o.OrderID}
	if got := Journal(); !slices.Equal(got, want) {
		t.Fatalf("journal %v, want %v", got, want)
	}
}
