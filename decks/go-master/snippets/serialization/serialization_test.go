package serialization

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func register(r worker.Registry) {
	r.RegisterWorkflow(ClaimWorkflow)
	r.RegisterWorkflow(Echo)
	r.RegisterWorkflow(Inspect)
	r.RegisterActivity(ProcessClaim)
}

func start(tq string) client.StartWorkflowOptions {
	return client.StartWorkflowOptions{ID: testenv.ID("serialization"), TaskQueue: tq}
}

func TestDefaultConverterRoundTripsAStructThroughAWorkflowAndAnActivity(t *testing.T) {
	tq := testenv.Worker(t, register)
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), start(tq), ClaimWorkflow,
		Claim{PatientID: "p-1", DrugNDC: "0002-1433-80", BilledAmountCents: 12_500})
	if err != nil {
		t.Fatal(err)
	}
	var got Result
	if err := run.Get(testenv.Ctx(t), &got); err != nil {
		t.Fatal(err)
	}
	if got != (Result{PatientID: "p-1", ApprovedCents: 12_500}) {
		t.Fatalf("got %+v", got)
	}
}

func TestJSONKeepsTimesButFloatsAnyNumbersAndDropsUnexportedFields(t *testing.T) {
	tq := testenv.Worker(t, register)
	received, err := JSONTraps(testenv.Ctx(t), testenv.Client(), start(tq))
	if err != nil {
		t.Fatal(err)
	}
	if !received.At.Equal(time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)) {
		t.Errorf("At = %v", received.At)
	}
	if f, ok := received.Claim["amount"].(float64); !ok || f != 9_007_199_254_740_992 {
		t.Errorf("amount = %T %v; want float64 9007199254740992", received.Claim["amount"], received.Claim["amount"])
	}
	if received.copay != 0 {
		t.Errorf("copay = %d; want 0", received.copay)
	}
}

func TestTheWorkflowItselfReceivesAFloat64(t *testing.T) {
	tq := testenv.Worker(t, register)
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), start(tq), Inspect,
		Visit{Claim: map[string]any{"amount": int64(9_007_199_254_740_993)}})
	if err != nil {
		t.Fatal(err)
	}
	var got string
	if err := run.Get(testenv.Ctx(t), &got); err != nil {
		t.Fatal(err)
	}
	if got != "float64 9.007199254740992e+15" {
		t.Fatalf("got %q", got)
	}
}

func TestAValueJSONCannotEncodeFailsBeforeAnythingIsSent(t *testing.T) {
	tq := testenv.ID("never-polled")
	_, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), start(tq), "Echo", map[string]any{"done": make(chan int)})
	if err == nil || !strings.Contains(err.Error(), "unable to encode") {
		t.Fatalf("got %v; want an encode error", err)
	}
}

func TestACustomConverterKeepsEveryDigitInTheWorkflowAndBack(t *testing.T) {
	tq := testenv.ID("serialization")
	c, w := Connect(testenv.HostPort(), tq)
	defer c.Close()
	register(w)
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	v := Visit{Claim: map[string]any{"amount": int64(9_007_199_254_740_993)}}
	run, err := c.ExecuteWorkflow(testenv.Ctx(t), start(tq), Inspect, v)
	if err != nil {
		t.Fatal(err)
	}
	var inside string
	if err := run.Get(testenv.Ctx(t), &inside); err != nil {
		t.Fatal(err)
	}
	if inside != "json.Number 9007199254740993" {
		t.Fatalf("the Workflow received %q", inside)
	}

	run, err = c.ExecuteWorkflow(testenv.Ctx(t), start(tq), Echo, v)
	if err != nil {
		t.Fatal(err)
	}
	var back Visit
	if err := run.Get(testenv.Ctx(t), &back); err != nil {
		t.Fatal(err)
	}
	if n, ok := back.Claim["amount"].(json.Number); !ok || n.String() != "9007199254740993" {
		t.Fatalf("the Client received %T %v", back.Claim["amount"], back.Claim["amount"])
	}
}

func TestAClientWithoutTheConverterCannotReadTheResult(t *testing.T) {
	tq := testenv.ID("serialization")
	c, w := Connect(testenv.HostPort(), tq)
	defer c.Close()
	register(w)
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	run, err := c.ExecuteWorkflow(testenv.Ctx(t), start(tq), Inspect, Visit{})
	if err != nil {
		t.Fatal(err)
	}
	if err := run.Get(testenv.Ctx(t), nil); err != nil {
		t.Fatal(err)
	}
	// testenv.Client() has the default converter only.
	var got string
	err = testenv.Client().GetWorkflow(testenv.Ctx(t), run.GetID(), run.GetRunID()).Get(testenv.Ctx(t), &got)
	if err == nil || !strings.Contains(err.Error(), "encoding") {
		t.Fatalf("got %v, %q; want an unknown-encoding error", err, got)
	}
	t.Log(err)
}
