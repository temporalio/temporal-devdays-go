package sagacompensation

import (
	"testing"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

// The undo has no deadline and no attempt cap, so it outlasts an outage of the
// payment API, and its idempotency key keeps the retries to a single refund.
func TestTheRefundRetriesThroughAnOutageAndRefundsOnce(t *testing.T) {
	payments.Lock()
	payments.refundFailuresLeft = 3
	payments.Unlock()

	tq := testenv.Worker(t, func(r worker.Registry) {
		r.RegisterWorkflow(ChargeThenRefundWorkflow)
		r.RegisterActivity(ChargePayment)
		r.RegisterActivity(RefundPayment)
	})
	order := Order{OrderID: testenv.ID("order"), AmountCents: 500}
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("refund"), TaskQueue: tq,
	}, ChargeThenRefundWorkflow, order)
	if err != nil {
		t.Fatal(err)
	}
	// The default Retry Policy waits 1s, 2s, then 4s between attempts.
	if err := run.Get(testenv.Ctx(t), nil); err != nil {
		t.Fatal(err)
	}

	payments.Lock()
	defer payments.Unlock()
	if payments.refundCalls != 4 {
		t.Fatalf("refund called %d times, want 4", payments.refundCalls)
	}
	if len(payments.refunds) != 1 || payments.refunds["refund-"+order.OrderID] != order.OrderID {
		t.Fatalf("refunds %v, want one for %s", payments.refunds, order.OrderID)
	}
}

// The forward step has a deadline and the undo has none.
func TestOnlyTheForwardStepHasADeadline(t *testing.T) {
	if chargeOptions.ScheduleToCloseTimeout == 0 {
		t.Fatal("the charge should give up after its Schedule-To-Close Timeout")
	}
	if refundOptions.ScheduleToCloseTimeout != 0 || refundOptions.RetryPolicy != nil {
		t.Fatal("the refund should keep the default, unlimited Retry Policy")
	}
}
