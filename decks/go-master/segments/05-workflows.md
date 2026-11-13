---
layout: section
---

<ZiggyHero scene="conduct">

# Go Workflows

The orchestration. Plain Go code that Temporal can replay.

</ZiggyHero>

<!--
Source: PPT slides 38-50.

One slide, stepped through ten times. Resist reading all ten annotations at
once: the point of the build is that the room looks at one line at a time.

The original split this into an interface slide and an implementation slide.
Go has no separate interface to declare. One exported function is both the
definition and the implementation, so the two walkthroughs collapse into one.
The determinism rules follow it, because they are a discipline the code keeps,
and the room should see them right after its first Workflow.
-->

---
layout: code-walkthrough
heading: Go Workflow anatomy
notes:
  - "Every Workflow takes a <code>workflow.Context</code> first, and uses the <code class='ref'>workflow</code> package for anything that waits"
  - "A nil <code>*Activities</code>: the Workflow takes its methods for their <strong>names and signatures</strong>, and never calls them"
  - "Options travel on a context. <code>WithActivityOptions</code> derives one per kind of call"
  - "Options per context, set by the caller: 5 seconds for a fraud check, 30 seconds and five attempts to move money"
  - An exported function, registered on the Worker. Its name is the Workflow <strong>Type</strong>
  - "All arguments and return values must be <strong>serializable</strong>, and the last result is an <code>error</code>"
  - "<code>ExecuteActivity</code> <strong>schedules</strong> the Activity and <code>Get</code> waits for it. Branching on the result is a business decision, so it lives here"
  - The money leaves the source account. Once it completes, it is in Event History for good
  - "If the deposit fails for good, put the money back. A compensation, in a plain <code>err != nil</code> branch"
  - The receipt is the Workflow's result, recorded in history too
---

::code::

```go {all|1|2|3,6|4,7-8|1|1|11-17|19-21|22-27|28}
func TransferWorkflow(ctx workflow.Context, t Transfer) (Receipt, error) {
	var a *Activities
	fraudCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	})
	bankCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 5},
	})

	var flagged bool
	if err := workflow.ExecuteActivity(fraudCtx, a.CheckFraud, t).Get(ctx, &flagged); err != nil {
		return Receipt{}, err
	}
	if flagged {
		return Receipt{Status: "rejected", Reason: "flagged by the fraud check"}, nil
	}

	if err := workflow.ExecuteActivity(bankCtx, a.Withdraw, t).Get(ctx, nil); err != nil {
		return Receipt{}, err
	}
	if err := workflow.ExecuteActivity(bankCtx, a.Deposit, t).Get(ctx, nil); err != nil {
		if rerr := workflow.ExecuteActivity(bankCtx, a.Refund, t).Get(ctx, nil); rerr != nil {
			return Receipt{}, rerr
		}
		return Receipt{Status: "refunded", Reason: err.Error()}, nil
	}
	return Receipt{Status: "completed"}, nil
}
```

<!--
Source: PPT slides 39-50, one build per slide in the original. The original
was a one-line order Workflow; this is a money transfer, so the same anatomy
carries a business decision and a compensation.
Code: snippets/workflows/transfer.go, run by
snippets/workflows/workflows_test.go.

- **Build 1 - the context.** The SDK is split by concern: `workflow`,
  `activity`, `client`, `worker`. A Workflow function takes a
  `workflow.Context`, never a `context.Context`, and everything that waits goes
  through the `workflow` package: Activities, Timers, Signals, coroutines. The
  split is the guardrail. Workflow code is ordinary Go in your Worker process,
  so determinism is a discipline you keep, and the next slide is the list of
  rules.
- **Build 2 - the nil pointer.** The Activities are methods on a struct, so a
  real Worker can give them a database pool or an HTTP client. The Workflow
  only needs the methods' names and signatures, so it declares a nil
  `*Activities` and passes `a.Withdraw` as a method value. Nothing is ever
  called on it. Passing the function rather than the string `"Withdraw"` gets
  you compile-time checks on the name; the arguments are checked when the
  Activity is scheduled.
- **Build 3 - the contexts.** Activity options do not go on the call; they go
  on a context derived with `workflow.WithActivityOptions`. Two derived
  contexts is normal: each carries its own options, and you pass the right one
  to each call.
- **Build 4 - the options.** `StartToCloseTimeout` is how long one attempt may
  run. At least one of it or `ScheduleToCloseTimeout` is required. A fraud
  check that takes more than 5 seconds is broken; a bank call gets longer, and
  five attempts before the Workflow gives up on it. Note where the options
  live: on the caller. The Workflow decides how long it will wait; the
  Activity has no opinion.
- **Build 5 - the function.** There is no annotation. Registering the function
  on a Worker (`w.RegisterWorkflow(TransferWorkflow)`) is the whole
  registration. The Workflow Type is the function name, `TransferWorkflow`, and
  a Client in another codebase can start it by that string. Rename the function
  and you rename the Type, which matters once Workflows of the old name are
  running; `RegisterWorkflowWithOptions` can pin the name.
- **Build 6 - the signature.** Arguments and return values cross a process
  boundary as JSON by default. Structs with exported fields, primitives, slices
  and maps work out of the box; unexported fields are dropped, and a
  `time.Time` round-trips as a string. A single struct argument is the
  convention: you can add a field later without breaking running Workflows.
  The `error` is how the Workflow fails.
- **Build 7 - the call, and the decision.** It looks like a function call and
  it is not one. `ExecuteActivity` writes a command to Temporal, which schedules
  the Activity on a Worker that may be a different process or machine, and
  returns a Future; `Get` blocks the Workflow until the result is in. What to
  do with a flagged transfer is policy, so it is an `if` here, visible in the
  code and in the history, not buried in the Activity.
- **Build 8 - the withdrawal.** The result is recorded in Event History. If the
  Worker crashes after this line, the replay reads it back rather than
  withdrawing twice.
- **Build 9 - the compensation.** After five failed attempts, or one
  non-retryable failure such as a closed account, `Get` returns a
  `*temporal.ActivityError` and the Workflow refunds the source account. That
  is the Saga pattern in its smallest form; the Compensations section comes back
  to it.
- **Build 10 - the result.** The receipt goes back to whoever is waiting on the
  Workflow: the Client, on the Client slide.

If you only have time for three builds, take 2, 7 and 9.
-->

---
layout: code-walkthrough
heading: Go determinism rules
notes:
  - "<code>workflow.Now</code>, not time.Now(): the time of the Workflow Task, the same on every replay"
  - "<code>workflow.NewChannel</code> and <code>workflow.Go</code>, never a native channel or goroutine. The SDK runs coroutines one at a time, in a fixed order"
  - "<code>workflow.NewSelector</code> in place of select, and <code>NewTimer</code> in place of time.After: a durable Timer"
  - "Randomness and UUIDs go in <code>SideEffect</code>: run once, recorded, read back on replay"
  - "Go randomizes map order. <code>DeterministicKeys</code> sorts the keys before you range"
---

::code::

```go {all|2|4-7|10-13|16-18|24}
func PackingWorkflow(ctx workflow.Context, items map[string]int) (Packing, error) {
	started := workflow.Now(ctx)

	packed := workflow.NewChannel(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		packed.Send(ctx, "box-1")
	})

	box := "none"
	sel := workflow.NewSelector(ctx)
	sel.AddReceive(packed, func(c workflow.ReceiveChannel, _ bool) { c.Receive(ctx, &box) })
	sel.AddFuture(workflow.NewTimer(ctx, time.Hour), func(workflow.Future) {})
	sel.Select(ctx)

	var label string
	err := workflow.SideEffect(ctx, func(workflow.Context) any {
		return fmt.Sprintf("PKG-%06d", rand.IntN(1_000_000))
	}).Get(&label)
	if err != nil {
		return Packing{}, err
	}

	var lines []string
	for _, sku := range workflow.DeterministicKeys(items) {
		lines = append(lines, fmt.Sprintf("%s x%d", sku, items[sku]))
	}
	return Packing{Started: started, Box: box, Label: label, Lines: lines}, nil
}
```

<!--
New to the Go Deck. Code: snippets/workflows/determinism.go, run by
snippets/workflows/workflows_test.go, which also replays its Event History
three times and checks the SideEffect is one recorded marker.

Why this slide exists: Workflow code is ordinary Go running in your Worker, and
a replay re-runs it against the Event History. It must make the same decisions
in the same order every time. Each line here is a standard-library habit with
its deterministic replacement.

- **Build 1 - the clock.** `time.Now()` differs on every replay.
  `workflow.Now(ctx)` is the time the current Workflow Task started, recorded in
  history, so it replays identically. The same goes for `time.Sleep`: use
  `workflow.Sleep(ctx, d)`, a durable Timer.
- **Build 2 - concurrency.** A goroutine is scheduled by the Go runtime, so two
  of them race differently on every run. `workflow.Go` starts a coroutine the
  SDK schedules itself, one at a time, in a fixed order, and `workflow.Channel`
  is the channel that works with it. A `sync.Mutex` is unnecessary inside a
  Workflow for the same reason.
- **Build 3 - select.** Go's `select` picks at random among ready cases.
  `workflow.Selector` picks the first ready one in a deterministic order, and
  works with channels, Futures and Timers. `workflow.NewTimer` is the durable
  `time.After`: it survives the Worker restarting.
- **Build 4 - randomness.** `math/rand` and a UUID library give a new value on
  replay. `workflow.SideEffect` runs the function once, records the value in
  history as a marker, and returns the recorded value on every replay. For
  anything that does I/O, use an Activity instead.
- **Build 5 - maps.** Go randomizes map iteration order on purpose. If the order
  can change a decision (here, the order of the packing lines), range over
  `workflow.DeterministicKeys(m)`, which returns the keys sorted.

What breaks when you get it wrong: nothing, at first. The code compiles and the
first run works. The failure comes on replay, after a Worker restart or a cache
eviction, as a non-determinism error that fails the Workflow Task (not the
Workflow) until the code is fixed. The next slide is how to catch it before it
ships.
-->

---
layout: code-stack
heading: Catch it before it ships — workflowcheck
proseMax: 34
---

::code::

```bash
$ go install go.temporal.io/sdk/contrib/tools/workflowcheck@latest
$ workflowcheck ./...
workflows/ship.go:9:6: orders.ShipWorkflow is non-deterministic, reason: calls non-deterministic function time.Now
  time.Now is non-deterministic, reason: declared non-deterministic
workflows/ship.go:9:6: orders.ShipWorkflow is non-deterministic, reason: starts goroutine
workflows/ship.go:9:6: orders.ShipWorkflow is non-deterministic, reason: sends to channel
workflows/ship.go:9:6: orders.ShipWorkflow is non-deterministic, reason: receives from channel
workflows/ship.go:9:6: orders.ShipWorkflow is non-deterministic, reason: iterates over map
```

::default::

A static analyzer from the Go SDK. It finds every Workflow, follows its calls,
and names each one that reaches the clock, a goroutine, a native channel or a
map range. Run it in CI next to `go vet`.

<!--
New to the Go Deck.

The output above is real: it is what `workflowcheck` (v0.6.0) printed against
a Workflow that calls `time.Now()`, starts a goroutine, sends and waits on a
native channel and ranges over a map, with the absolute path dropped and the
file and package renamed to read well. The same
tool reports nothing for the determinism slide's `PackingWorkflow`, including
its `rand.IntN` inside `workflow.SideEffect`.

What to say about it:

- It is static: it sees calls through your own helpers and through third-party
  packages, which is where the surprises are. `-show-pos` adds the line of each
  offending call.
- It is conservative. When it flags something you know is safe, mark the line
  with a `//workflowcheck:ignore` comment, and say why in the same comment.
- It is not the only net. Replay tests (`worker.NewWorkflowReplayer`) run real
  histories against new code; the Versioning section comes back to them.
-->

---
layout: subsection
eyebrow: DESIGN PATTERNS
---

# Keep it in the Workflow if it…

<!--
Source: PPT slide 31, expanded into the expanded edition's slides 45-49,
`{{BLOCK:workflow-patterns}}`.

The Workflow half of the split, taught right after the anatomy slide. The
Activity half ("Make it an Activity if it…") comes right after the Activities
section.
Five criteria, one slide each, all short:

- Makes business logic decisions
- Coordinates between Activities
- Manages timing and state transitions
- Is a pure calculation
- Needs to survive process restarts

The last one is the one people underrate. A local variable in a Workflow
survives a Worker crash, because it is rebuilt by replay. A local variable in an
Activity does not.
-->

---
layout: code-stack
heading: Keep it in the Workflow if it… makes business decisions
proseMax: 26
---

::code::

```go
func AdjudicationWorkflow(ctx workflow.Context, claim Claim) (ApprovalStatus, error) {
	if claim.Amount > 10_000 || claim.Type == "Surgery" {
		return RequiresReview, nil
	}

	// continue to activities...
}
```

::default::

The decision itself is the orchestration. It is deterministic, it is cheap, and
it belongs where you can read it.

<!--
Source: PPT slide 32.

This is what "orchestration" means in practice: an `if`. There is nothing to
retry and nothing to time out, so wrapping it in an Activity buys nothing and
costs a round trip to the Server.

The practical argument, which lands better than the theoretical one: this
branch is the business rule. Put it in an Activity and it is invisible in the
Workflow history — you see "Activity completed" and not *why*. Keep it in the
Workflow and the decision is in the code and in the Event History.
-->

---
layout: code-stack
heading: Keep it in the Workflow if it… coordinates Activities
proseMax: 26
---

::code::

```go
func AdjudicationWorkflow(ctx workflow.Context, applicant Applicant) (Decision, error) {
	ctx = workflow.WithActivityOptions(ctx, opts)

	var credit CreditCheck
	if err := workflow.ExecuteActivity(ctx, RunCreditCheck, applicant).Get(ctx, &credit); err != nil {
		return "", err
	}
	var fraud FraudCheck
	if err := workflow.ExecuteActivity(ctx, CheckFraud, applicant, credit).Get(ctx, &fraud); err != nil {
		return "", err
	}
	var decision Decision
	err := workflow.ExecuteActivity(ctx, MakeDecision, applicant, credit, fraud).Get(ctx, &decision)
	return decision, err
}
```

::default::

Chaining Activities with results flowing between the steps. This is the
Workflow's actual job.

<!--
Source: PPT slide 46.
Code: snippets/designpatterns/workflows.go, run by
snippets/designpatterns/designpatterns_test.go, which also counts three
completed Activities in the history.

Point at how ordinary it looks. Three calls and three local variables, with
the usual `if err != nil` after each. No queues, no callbacks, no state
machine, no correlation IDs.

Then point at what is not written: every one of those three calls is
independently retried, independently timed out, and survives the process dying
between any two of them.
-->

---
layout: code-stack
heading: Keep it in the Workflow if it… manages timing
proseMax: 30
---

::code::

```go
// Sleep for 30 days — durable, costs nothing while waiting
err := workflow.Sleep(ctx, 30*24*time.Hour)

// Wait for a customer response — survives restarts
err = workflow.Await(ctx, func() bool { return response != nil })
```

::default::

Scheduling future actions and reacting to external events. Both are durable:
they survive Worker restarts and deploys.

<!--
Source: PPT slide 47.

The thirty-day sleep is the slide that changes how people think, so give it a
beat. No cron, no scheduler table, no polling job, no "wake up and check"
function. One line, and the process can be redeployed forty times in between.

The mechanism, if asked: Temporal records a Timer in history and the Workflow
is evicted from the Worker's cache. It is not a goroutine parked for a month.
This is also why it is `workflow.Sleep` and never `time.Sleep`.

`workflow.Await` is the primitive the Signals segment builds on: block until a
condition over Workflow state is true. Timing and events are the same mechanism
from the Workflow's point of view — both are "stop here until something
happens."
-->

---
layout: code-stack
heading: Keep it in the Workflow if it… is a pure calculation
proseMax: 30
---

::code::

```go
isValidEmail := companyEmail.MatchString(email)

riskScore := claim.Amount/1000 + len(claim.History)
```

::default::

Deterministic in, deterministic out. Same inputs, same answer, every replay.

<!--
Source: PPT slide 48.

These call nothing external, are completely idempotent, and produce the same
output on every replay. Round-tripping them through an Activity buys you a
Temporal Server call and two history events in exchange for nothing.

The honest caveat, because someone will push back with the previous
`ValidateEligibility` slide: the line between "pure calculation, keep it" and
"complex validation, extract it" is testability, not purity. Both are correct.
Extract when the logic is worth a test of its own.
-->

---
layout: code-stack
heading: Keep it in the Workflow if it… must survive restarts
proseMax: 38
---

::code::

```go
ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
	StartToCloseTimeout: 30 * time.Second,
})

var approval Approval
err := workflow.ExecuteActivity(ctx, GetApproval, claim).Get(ctx, &approval)
```

::default::

The Activity call is the **checkpoint**. Its result is written to history, so a
Worker restart replays up to this line and carries on from there.

<!--
Source: PPT slide 49.
Code: snippets/designpatterns/workflows.go, run by
snippets/designpatterns/designpatterns_test.go.

This is the slide that explains what "durable" actually means mechanically, so
do not rush it.

Everything before this line is replayed from history, not re-executed against
the outside world. Everything after it has not happened yet. The Activity
result is the boundary between the two.

The consequence worth stating: Workflow state persists across failures and
restarts without you writing a single line of persistence. No database, no
checkpoint table, no serialisation of your own.
-->
