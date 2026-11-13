package features

import (
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func register(r worker.Registry) {
	r.RegisterWorkflow(ReportWorkflow)
	r.RegisterWorkflow(OrderWorkflow)
	r.RegisterActivity(BuildReport)
	r.RegisterActivity(ChargePayment)
}

func TestTheScheduleLivesOnTheServiceAndEachRunIsAWorkflow(t *testing.T) {
	ctx := testenv.Ctx(t)
	handle, err := CreateDailyReport(ctx, testenv.Client())
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Delete(ctx)

	desc, err := handle.Describe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if handle.GetID() != "daily-report" {
		t.Fatalf("schedule ID %q", handle.GetID())
	}
	if desc.Schedule.Policy.Overlap != enumspb.SCHEDULE_OVERLAP_POLICY_SKIP {
		t.Fatalf("overlap %v", desc.Schedule.Policy.Overlap)
	}
	if next := desc.Info.NextActionTimes[0].UTC(); next.Hour() != 9 || next.Minute() != 0 {
		t.Fatalf("next run at %v, want 09:00 UTC", next)
	}

	testenv.WorkerOn(t, "reports", register)
	if err := handle.Trigger(ctx, client.ScheduleTriggerOptions{}); err != nil {
		t.Fatal(err)
	}
	for range 80 {
		desc, err := handle.Describe(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if n := len(desc.Info.RecentActions); n > 0 {
			id := desc.Info.RecentActions[n-1].StartWorkflowResult.WorkflowID
			var report string
			if err := testenv.Client().GetWorkflow(ctx, id, "").Get(ctx, &report); err != nil {
				t.Fatal(err)
			}
			if report != "report ready" {
				t.Fatalf("report %q", report)
			}
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatal("the Schedule never started a Workflow")
}

func TestTheActivityIsScheduledWithThePriorityFromTheSlide(t *testing.T) {
	ctx := testenv.Ctx(t)
	tq := testenv.Worker(t, register)
	run, err := testenv.Client().ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID: testenv.ID("order"), TaskQueue: tq,
	}, OrderWorkflow, Order{ID: "A-1", TenantID: "tenant-a"})
	if err != nil {
		t.Fatal(err)
	}
	var receipt string
	if err := run.Get(ctx, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt != "charged A-1 for tenant-a" {
		t.Fatalf("receipt %q", receipt)
	}
	iter := testenv.Client().GetWorkflowHistory(ctx, run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		if a := e.GetActivityTaskScheduledEventAttributes(); a != nil {
			p := a.GetPriority()
			if p.GetPriorityKey() != 1 || p.GetFairnessKey() != "tenant-a" || p.GetFairnessWeight() != 2 {
				t.Fatalf("priority on ActivityTaskScheduled: %v", p)
			}
			return
		}
	}
	t.Fatal("no ActivityTaskScheduled event")
}
