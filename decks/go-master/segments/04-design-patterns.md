---
layout: code-walkthrough
heading: Make it an Activity if it…
notes:
  - Talks to external systems — APIs, databases, message queues
  - Does I/O — file system, network calls
  - Uses non-deterministic operations. Both compile in a Workflow, and both break its replay
---

::code::

```go {all|1-2|4-6|8-10}
// Talks to external systems
resp, err := http.Get("https://<api-endpoint>/projects/" + req.ProjectID)

// Does I/O
rows, err := db.QueryContext(ctx,
	"SELECT * FROM reviewer_history WHERE project_id = $1", req.ProjectID)

// Uses non-deterministic operations
id := uuid.NewString()            // ❌ in a Workflow
tier := os.Getenv("PRICING_TIER") // ❌ in a Workflow
```

<!--
Source: PPT slide 30. The segment as a whole is PPT slides 29-37, merged with
the expanded edition's slides 40-54 (`{{BLOCK:activity-patterns}}` and
`{{BLOCK:workflow-patterns}}`). Slide notes cite 29-37 where the slide follows
the Foundations deck, and 40-54 where it is one of the expanded slides.

The most important distinction in Temporal, and the one people get wrong for
the first month. Workflows are the coordinators: they make decisions, handle
timing and manage state, and they must be deterministic because they replay
from Event History. Activities are the workers: they talk to databases, call
APIs, send emails. They can fail, be retried, and use any Go module you like.

The mental model to repeat until it sticks: if it has side effects or talks to
the outside world, it is an Activity. If it is decision logic or coordination,
it stays in the Workflow.

This slide:

- **Build 1 -** anything across a network boundary. It can be slow, it can be
  down, and it needs its own retry policy.
- **Build 2 -** the same argument for disk and database.
- **Build 3 -** the determinism rule, made concrete: a value that can differ
  between the first run and the replay.

On build 3, say what actually happens rather than "it is banned": Workflow code
is ordinary Go running in your Worker, so both lines compile and run. Nothing
stops them. The trouble comes on replay: a fresh UUID or a Worker's environment
variable can differ the next time the Workflow's history is replayed, on
another Worker or after a deploy. A different value is silent until it changes
a decision; then the replay schedules a different command from the one in
Event History and the Workflow Task fails with a non-determinism error.
`snippets/designpatterns/designpatterns_test.go` proves exactly that with
`os.Getenv`: the history replays while the variable is the same, and fails
the moment it is not.

The same list in Go includes `time.Now()`, `time.Sleep`, `math/rand`,
goroutines, native channels and `select`, and ranging over a map. The SDK gives
a deterministic replacement for each, which is the determinism slide in the
Workflows section, and the `workflowcheck` analyzer finds the ones you missed.
For a UUID inside a Workflow, generate it in `workflow.SideEffect`, or in an
Activity. For configuration, pass it in as a Workflow argument or read it in an
Activity.

The original also listed two more reasons, and they get a slide each next:
complex validation you want to test on its own, and work that can fail or run
long by itself.
-->

---
layout: code-stack
heading: Make it an Activity if it… is worth testing in isolation
proseMax: 30
---

::code::

```go
// activities.go
var rfiID = regexp.MustCompile(`^RFI-\d{4}$`)

func ValidateEligibility(ctx context.Context, req RFIRequest) (bool, error) {
	rfiIDValid := rfiID.MatchString(req.RFIID)
	submittedRecently := req.SubmittedAt.After(time.Now().AddDate(0, 0, -30))
	withinCostThreshold := req.EstimatedCost < 50_000

	return rfiIDValid && submittedRecently && withinCostThreshold, nil
}

// workflow.go
func EligibilityWorkflow(ctx workflow.Context, req RFIRequest) (bool, error) {
	ctx = workflow.WithActivityOptions(ctx, opts)

	var isEligible bool
	err := workflow.ExecuteActivity(ctx, ValidateEligibility, req).Get(ctx, &isEligible)
	return isEligible, err
}
```

::default::

Complex validation logic. Fast, in-memory, no external dependency — and a clean
boundary to unit test on its own.

<!--
Source: PPT slide 42.
Code: snippets/designpatterns/activities.go and workflows.go, run by
snippets/designpatterns/designpatterns_test.go.

Note the tension with the "Keep it in the Workflow" slides later in this
segment, and name it rather than hoping nobody notices: these three checks are
pure, in-memory, and deterministic (the Activity may call `time.Now()`; a
Workflow would use `workflow.Now(ctx)`), so by the "keep it in the Workflow"
rule they could live in the Workflow.

The reason to pull them out anyway is testability. You get a named unit with a
contract you can test with a plain `go test`, no Workflow test environment. The
test file does exactly that, alongside the run through a real Worker.

`opts` is the Activity options from the anatomy slide: every Activity call
needs a Start-To-Close or Schedule-To-Close timeout, set on the context.
-->

---
layout: code-stack
heading: Make it an Activity if it… can fail on its own
proseMax: 30
---

::code::

```go
// activities.go
func NotifyReviewer(ctx context.Context, req RFIRequest, reviewer string) error {
	return emailService.Send(ctx, reviewer, fmt.Sprintf("RFI %s requires your review", req.RFIID))
}

// workflow.go
func ReviewWorkflow(ctx workflow.Context, req RFIRequest, reviewer string) error {
	ctx = workflow.WithActivityOptions(ctx, opts)

	return workflow.ExecuteActivity(ctx, NotifyReviewer, req, reviewer).Get(ctx, nil)
}
```

::default::

Sending mail fails for reasons that have nothing to do with you. Wrapped in an
Activity, Temporal retries it until it succeeds or the policy runs out.

<!--
Source: PPT slide 43.

`emailService.Send` fails for any number of external reasons: SMTP timeout,
rate limiting, a network blip, someone's DNS.

The cost of getting this wrong: called directly from Workflow code, a transient
email failure is an error the Workflow has to handle itself, with no retry and
no record of the attempt. Wrapped in an Activity, it is a retry nobody hears
about. The test makes the first two sends fail and checks that the Workflow
completes with exactly one email sent.

Say the line again, because it is the whole segment: the Activity knows nothing
about retries. It returns an error, and Temporal handles the rest.
-->

---
layout: code-stack
heading: Make it an Activity if it… takes significant time
proseMax: 34
---

::code::

```go
// activities.go
func GenerateAnnualReport(ctx context.Context, req RFIRequest) error {
	// compiles all RFIs for the year into a report
	resp, err := http.Get(fmt.Sprintf("%s/rfi/%s/annual", RFIAPI, req.ProjectID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// workflow.go
func AnnualReportWorkflow(ctx workflow.Context, req RFIRequest) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Hour,
	})

	return workflow.ExecuteActivity(ctx, GenerateAnnualReport, req).Get(ctx, nil)
}
```

::default::

Two hours of work, and the timeout that allows it is set by the **caller**. The
Activity has no idea it has two hours.

<!--
Source: PPT slide 44.

Why this cannot stay in the Workflow: Workflow code must not do the I/O in the
first place, and a long computation there would hold the Workflow Task past its
10-second timeout, and the Worker's deadlock detector fails a Workflow
coroutine that does not yield for a second. As an Activity it gets its own
timeout, its own retries, and it can heartbeat.

As an Activity, the Worker can restart freely. Temporal reschedules the
Activity and retries it from the top, or, if it heartbeats its progress, from
the last checkpoint it recorded.

The timeout placement is the design point, and the same one you saw on the
Workflow anatomy slide: policy lives with the caller, in one place, not
scattered through the implementations. The test reads the
`ActivityTaskScheduled` Event and checks its Start-To-Close timeout is two
hours.

In production, build the request with `http.NewRequestWithContext(ctx, ...)`
so a cancelled Activity stops its call; the slide keeps `http.Get` short.

Heartbeating is the companion concept. For a two-hour Activity you want
`activity.RecordHeartbeat(ctx, ...)` calls inside the function body so Temporal
learns it has died in well under two hours. That gets its own segment —
forward-reference it, do not teach it here.
-->

---
layout: default
---

# Common mistakes

- Making an Activity just to log something → use `workflow.GetLogger(ctx)`
- Making an Activity for a simple calculation → keep it in the Workflow
- Putting a business decision in an Activity → it belongs in the Workflow
- One Activity doing two things → split it, so each part is safe to retry

<!--
Source: PPT slide 33 (the expanded edition's slides 50 and 53, which were
duplicates of each other), merged with the headlines from slides 35 and 36 so
the list is in one place. The next two slides give those two their code.

The first two come from treating "Activity" as "function." Every Activity call
is at least three Events in history (scheduled, started, completed) and a round
trip to the Server. For a log line, that is absurd.

Logging is the one worth a sentence: `workflow.GetLogger(ctx)` is replay-aware
and stays quiet while the Workflow is replaying, which is exactly what you want
and exactly what `log.Printf` or `fmt.Println` will not do.
-->

---
layout: code-stack
heading: A business decision hidden in an Activity
proseMax: 26
---

::code::

```go
// ❌ Don't do this — the rule is invisible in the Workflow and in history
var decision string
err := workflow.ExecuteActivity(ctx, EvaluateClaim, claim, coverage).Get(ctx, &decision)

// ✅ Do this — the rule is in the Workflow, where you can read it
switch {
case claim.Amount > coverage.MaxBenefit:
	decision = "partial_approval"
case claim.IsEmergency:
	decision = "fast_track"
default:
	decision = "standard_review"
}
```

::default::

The branch is the business logic. Decisions belong in the Workflow.

<!--
Source: PPT slide 35.

A common justification for the ❌ version is that `EvaluateClaim` might return
something different on replay. It will not: a completed Activity's result is
read from history on replay. The real cost is elsewhere: the decision is now
hidden behind a network call, retried as if it were I/O, and invisible when you
read the Workflow.

Ask the room which version they would rather read in six months during an
incident. That usually settles it faster than the determinism argument.

The second cost, beyond readability: in the ❌ version the Workflow history
records "EvaluateClaim completed" with an opaque result. In the ✅ version the
inputs are in history and the branch is in the code, so the decision is fully
reconstructable.
-->

---
layout: code-stack
heading: When to split an Activity
proseMax: 34
---

::code::

```go
// ❌ One Activity doing two things — not safe to retry
func CreateUser(ctx context.Context, data UserData) (User, error) {
	user, err := db.AddUser(ctx, data)
	if err != nil {
		return User{}, err
	}
	if err := emailService.SendWelcomeEmail(ctx, user.Email); err != nil {
		return User{}, err
	}

	return user, nil
}
```

::default::

Two different services in one Activity. If the email fails, the whole Activity
retries — and you create the user twice.

<!--
Source: PPT slide 36.
Code: snippets/designpatterns/activities.go. The test fails the first welcome
email and finds two rows in the fake database.

Think in failure modes rather than in tidiness. The question is never "are these
two things related?" It is "if the second one fails, is re-running the first one
safe?"

Here it is not. The insert succeeded, the email did not, and the retry inserts
again.

Two ways out, and it is worth naming both: split into two Activities so each
retries independently, or make the single Activity idempotent with a key the
database can deduplicate on. Splitting is the one to reach for first, because
the Workflow then records that the user was created and never re-runs it.
-->
