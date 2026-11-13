package designpatterns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func register(r worker.Registry) {
	for _, wf := range []any{EligibilityWorkflow, ReviewWorkflow, AnnualReportWorkflow,
		AdjudicationWorkflow, ClaimWorkflow, CreateUserWorkflow, PricingWorkflow} {
		r.RegisterWorkflow(wf)
	}
	for _, a := range []any{ValidateEligibility, NotifyReviewer, GenerateAnnualReport, RunCreditCheck,
		CheckFraud, MakeDecision, GetApproval, CreateUser, StandardPricing, PremiumPricing} {
		r.RegisterActivity(a)
	}
}

func reset() {
	emailService.failuresLeft, emailService.sent = 0, nil
	db.rows = nil
}

func execute(t *testing.T, tq string, wf any, out any, args ...any) client.WorkflowRun {
	t.Helper()
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{
		ID: testenv.ID("patterns"), TaskQueue: tq,
	}, wf, args...)
	if err != nil {
		t.Fatal(err)
	}
	if err := run.Get(testenv.Ctx(t), out); err != nil {
		t.Fatal(err)
	}
	return run
}

func history(t *testing.T, run client.WorkflowRun) *historypb.History {
	t.Helper()
	iter := testenv.Client().GetWorkflowHistory(context.Background(), run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	h := &historypb.History{}
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		h.Events = append(h.Events, e)
	}
	return h
}

var rfi = RFIRequest{RFIID: "RFI-0042", ProjectID: "p-7", SubmittedAt: time.Now(), EstimatedCost: 12_000}

// Make it an Activity: reading the environment in a Workflow runs, and the
// history it writes no longer replays once the environment differs.
func TestEnvironmentReadInAWorkflowBreaksReplay(t *testing.T) {
	reset()
	t.Setenv("PRICING_TIER", "premium")
	tq := testenv.Worker(t, register)
	var tier string
	run := execute(t, tq, PricingWorkflow, &tier)
	if tier != "premium" {
		t.Fatalf("tier %q", tier)
	}
	h := history(t, run)

	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(PricingWorkflow)
	if err := replayer.ReplayWorkflowHistory(testenv.Quiet, h); err != nil {
		t.Fatalf("same environment should replay: %v", err)
	}
	t.Setenv("PRICING_TIER", "standard")
	err := replayer.ReplayWorkflowHistory(testenv.Quiet, h)
	if err == nil || !strings.Contains(err.Error(), "nondeterministic") {
		t.Fatalf("a different environment should fail replay with a non-determinism error, got %v", err)
	}
}

func TestValidateEligibilityChecksAllThreeRules(t *testing.T) {
	reset()
	tq := testenv.Worker(t, register)
	var ok bool
	execute(t, tq, EligibilityWorkflow, &ok, rfi)
	if !ok {
		t.Fatal("a fresh, cheap, well-formed RFI should be eligible")
	}
	stale := rfi
	stale.SubmittedAt = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	execute(t, tq, EligibilityWorkflow, &ok, stale)
	if ok {
		t.Fatal("a stale RFI should not be eligible")
	}
}

// It is a plain function, so it also tests without any Temporal at all.
func TestValidateEligibilityTestsOnItsOwn(t *testing.T) {
	bad := rfi
	bad.RFIID = "RFI-42"
	if ok, _ := ValidateEligibility(context.Background(), bad); ok {
		t.Fatal("a malformed RFI ID should not be eligible")
	}
}

func TestNotifyReviewerAFailedSendIsRetriedAndTheWorkflowNeverSeesIt(t *testing.T) {
	reset()
	emailService.failuresLeft = 2
	tq := testenv.Worker(t, register)
	execute(t, tq, ReviewWorkflow, nil, rfi, "rita@example.com")
	if want := []string{"rita@example.com: RFI RFI-0042 requires your review"}; !slices.Equal(emailService.sent, want) {
		t.Fatalf("sent %v", emailService.sent)
	}
}

func TestGenerateAnnualReportRunsUnderTheCallersTwoHourTimeout(t *testing.T) {
	reset()
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Write(make([]byte, 8))
	}))
	defer srv.Close()
	old := RFIAPI
	RFIAPI = srv.URL
	defer func() { RFIAPI = old }()

	tq := testenv.Worker(t, register)
	run := execute(t, tq, AnnualReportWorkflow, nil, rfi)
	for _, e := range history(t, run).Events {
		if a := e.GetActivityTaskScheduledEventAttributes(); a != nil {
			if got := a.StartToCloseTimeout.AsDuration(); got != 2*time.Hour {
				t.Fatalf("StartToCloseTimeout %v", got)
			}
		}
	}
	if !slices.Equal(paths, []string{"/rfi/p-7/annual"}) {
		t.Fatalf("requests %v", paths)
	}
}

func TestAdjudicationWorkflowChainsThreeActivities(t *testing.T) {
	reset()
	tq := testenv.Worker(t, register)
	var d Decision
	run := execute(t, tq, AdjudicationWorkflow, &d, Applicant{Name: "Ada", Income: 90_000})
	if d != "approved" {
		t.Fatalf("decision %q", d)
	}
	completed := 0
	for _, e := range history(t, run).Events {
		if e.EventType == enumspb.EVENT_TYPE_ACTIVITY_TASK_COMPLETED {
			completed++
		}
	}
	if completed != 3 {
		t.Fatalf("%d Activities completed, want 3", completed)
	}
	execute(t, tq, AdjudicationWorkflow, &d, Applicant{Name: "Bob", Income: 20_000})
	if d != "declined" {
		t.Fatalf("decision %q", d)
	}
}

func TestMustSurviveRestartsTheActivityResultIsWrittenToHistory(t *testing.T) {
	reset()
	tq := testenv.Worker(t, register)
	var a Approval
	run := execute(t, tq, ClaimWorkflow, &a, Claim{ClaimID: "c-1", Amount: 500})
	if a != (Approval{ClaimID: "c-1", Approved: true}) {
		t.Fatalf("approval %+v", a)
	}
	for _, e := range history(t, run).Events {
		if e.EventType == enumspb.EVENT_TYPE_ACTIVITY_TASK_COMPLETED {
			return
		}
	}
	t.Fatal("no ActivityTaskCompleted event in history")
}

func TestWhenToSplitAFailedEmailRetriesTheWholeActivityAndAddsTheUserTwice(t *testing.T) {
	reset()
	emailService.failuresLeft = 1
	tq := testenv.Worker(t, register)
	var u User
	execute(t, tq, CreateUserWorkflow, &u, UserData{Email: "ada@example.com"})
	if u != (User{ID: 2, Email: "ada@example.com"}) {
		t.Fatalf("user %+v", u)
	}
	if len(db.rows) != 2 {
		t.Fatalf("%d rows, want the duplicate 2", len(db.rows))
	}
}
