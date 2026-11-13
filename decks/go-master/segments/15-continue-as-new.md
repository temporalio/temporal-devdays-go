---
layout: section
---

<ZiggyHero scene="forever">

# Continue-As-New

Workflows that run forever, with a history that doesn't.

</ZiggyHero>

<!--
Source: PPT slides 113-120.
-->

---
layout: default
---

# Workflows that run forever

- Some processes genuinely never end: subscription billing, IoT polling,
  perpetual monitoring
- Temporal stores every Event in the Workflow's Event History
- An infinite Workflow creates infinite history
- Eventually: memory pressure, slow replay, and hard limits

<!--
Source: PPT slide 114.

A problem that only surfaces in production, which makes it easy to ignore in
development. A Workflow that runs fine for a week starts degrading after a
month.

The numbers: the Server warns at 10,240 Events or 10 MB of history, and
terminates the Workflow Execution once history passes 51,200 Events or 50 MB.
-->

---
layout: default
---

# What is Continue-As-New?

- Closes the current Workflow Execution and immediately starts a fresh one
- Same Workflow ID, same Workflow Type, **fresh history**
- The state you need to carry forward is passed as input to the new run
- From the outside, it looks like one continuous Workflow

<!--
Source: PPT slide 115.

The analogy that works: a software release. The application keeps running, but
the old version is swapped for a new one and users do not see the seam. Same
here. Observers see one Workflow ID that has been running for months; underneath
it has restarted dozens of times with a clean history each time.

Each run gets a new Run ID under the same Workflow ID. A Client that gets the
Workflow by ID follows the chain: `c.GetWorkflow(ctx, id, "").Get(ctx, &out)`
waits for the last run.
-->

---
layout: default
---

# When to use it

- Workflows that run indefinitely, or for very long periods
- Workflows that process unbounded queues or event streams
- Periodic Workflows that loop on a Timer
- Any Workflow approaching Temporal's history limits
- A clean point to adopt new code: the new run has no old history to replay

<!--
Source: PPT slide 116.

The Server stores the complete Event History for the lifetime of a Workflow
Execution. It logs a warning after 10,240 Events, and keeps warning as Events
are added. It terminates the Execution when history exceeds 51,200 Events,
contains more than 2,000 Updates, or contains more than 10,000 Signals.

The last bullet deserves a sentence. Replay always runs the Worker's current
code against the recorded history, which is why changing Workflow code can
cause non-determinism errors in long-running Workflows. A new run starts with
empty history, so there is nothing old to replay against the new code. (With
Worker Versioning, a Pinned run's continuation inherits its version unless you
ask for `AUTO_UPGRADE`.)
-->

---
layout: code-walkthrough
heading: "Continue-As-New in practice: a monthly subscription"
notes:
  - Bill every month for as long as the subscription lives. No end date
  - Charge the card, then record the month in <code>MonthsBilled</code>, the state carried forward
  - A durable 30-day Timer. The Workflow costs nothing while it waits
  - Ask the SDK. <code>GetContinueAsNewSuggested</code> turns true as this run's history nears its limits
  - Start a fresh run with an empty history and <strong>only</strong> the state it needs. <code>NewContinueAsNewError</code>
---

::code::

```go {all|6|7-10|11-13|15|16}
func SubscriptionWorkflow(ctx workflow.Context, sub Subscription) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	})

	for {
		if err := workflow.ExecuteActivity(ctx, ChargeMonthly, sub).Get(ctx, nil); err != nil {
			return err
		}
		sub.MonthsBilled++
		if err := workflow.Sleep(ctx, 30*24*time.Hour); err != nil {
			return err
		}

		if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
			return workflow.NewContinueAsNewError(ctx, SubscriptionWorkflow, sub)
		}
	}
}
```

<!--
Source: PPT slide 118. The original was a status monitor that continued as new
on every loop and built a `newContinueAsNewStub`. This is the shape you would
actually ship: a subscription that bills every month for years, and stays in one
run until the Server says its history is getting big.
Code: snippets/continueasnew/continueasnew.go, run by
snippets/continueasnew/continueasnew_test.go in the SDK's in-memory test
environment, which skips the 30-day Timers and can force the suggestion on.

- **Build 1 - the loop.** No end date. Without Continue-As-New, a subscription
  that runs for ten years builds one enormous history.
- **Build 2 - the charge.** An ordinary Activity with its own retries. The month
  counter lives on `sub`, the input, which is why it survives the jump to a new
  run.
- **Build 3 - the Timer.** Two Events in history and no Worker slot for 30 days;
  the Worker can evict the Workflow from its cache while it waits.
  `30*24*time.Hour` keeps the slide simple; a real biller would compute the
  next calendar date from `workflow.Now(ctx)`, the Workflow's deterministic
  clock, never `time.Now()`.
- **Build 4 - the check.** `workflow.GetInfo(ctx).GetContinueAsNewSuggested()`
  is the Server's hint, set well before the warning thresholds from two slides
  back: by default around 4K Events or 4 MB of history, or 90% of the Update
  limit. Call `GetInfo` each time rather than caching it; the value changes as
  the run goes on. Checking it every month, at a safe point between charges,
  beats picking an arbitrary "every N months". (`GetCurrentHistoryLength()`
  and `GetCurrentHistorySize()` are there too if you want your own threshold.)
- **Build 5 - continue.** `NewContinueAsNewError` builds an error; returning it
  from the Workflow function ends the run, and the new run starts at the top of
  `SubscriptionWorkflow` with `sub` as its input: same Workflow ID, fresh
  history. The arguments are checked against the Workflow's parameters when the
  error is built. It only works when it is returned from the Workflow function
  itself, so pass it straight up.

How does it stop? Cancel the Workflow (`c.CancelWorkflow`, or
`temporal workflow cancel`). The cancellation reaches the 30-day sleep:
`workflow.Sleep` returns a `*temporal.CanceledError` and the subscription ends
as Cancelled. If you want a final "subscription closed" Activity, check with
`temporal.IsCanceledError(err)` and run it on a context from
`workflow.NewDisconnectedContext(ctx)`, which the cancellation does not reach.

If the Workflow has Signal or Update handlers (a "change plan" Signal, say),
wait for them before continuing, or Continue-As-New cuts them off:
`workflow.Await(ctx, func() bool { return workflow.AllHandlersFinished(ctx) })`
just before returning the Continue-As-New error. Signals still unread on a
channel must be drained first, too: a `ReceiveAsync` loop.
-->

---
layout: default
---

# What state to carry forward

- Only carry what you actually need. This is a fresh start, not a full backup
- State should be serializable and reasonably small
- **Good candidates:** counters, last processed ID, configuration, accumulated results
- **Bad candidates:** full Event History, large blobs. Use Claim Check instead

<!--
Source: PPT slide 119.

The design decision to think through. The input to the new run is your
checkpoint. Keep it lean: if your state struct grows without bound, you have
a different problem. Large state belongs in blob storage, with a reference
passed as the checkpoint.

It goes through the Data Converter like any Workflow argument, so by default it
is JSON: only exported struct fields survive, an unexported field arrives as
its zero value, and an interface-typed field comes back as a `map[string]any`.
-->

---
layout: default
---

# What you got

- Truly perpetual Workflows without history bloat
- Clean, predictable memory use and replay performance
- Explicit state management at each continuation boundary
- Works with Child Workflows and Versioning

<!--
Source: PPT slide 120.

Close on the connections. Continue-As-New pairs naturally with Child Workflows
for long-lived parent orchestrators, and the state you carry forward needs the
same versioning discipline as your Workflow code. These patterns compose.
-->
