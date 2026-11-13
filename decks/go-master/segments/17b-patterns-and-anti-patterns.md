---
layout: section
---

<ZiggyHero scene="dodont">

# Patterns and Anti-Patterns

The shapes that hold up in production, and the ones that page you at night.

</ZiggyHero>

<!--
Source: docs.temporal.io/design-patterns, plus the limits in
docs.temporal.io/evaluate/cloud/limits.

A bank runs through the whole section, the same one as the Signals, Queries and
Updates segments: transfers, wires, deposits, accounts. Every pattern here is a
fix for a specific way a payment goes wrong.

Pacing: three pattern slides. Resumable Activity and its demo come earlier,
right after Updates, since the pattern is built from a Signal and a Query.
-->

---
layout: code-stack
heading: The business ID is the Workflow ID
proseMax: 30
---

::code::

```go
// ❌ A random ID: a double-tapped "Send" starts two transfers
run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
	ID:        uuid.NewString(),
	TaskQueue: "payments",
}, TransferWorkflow, transfer)

// ✅ The transfer's own ID: the second start gets the first one back
run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
	ID:        "transfer-" + transfer.TransferID,
	TaskQueue: "payments",
}, TransferWorkflow, transfer)
```

::default::

Two running Workflows can never share an ID. Make the ID the thing that must not
happen twice, and the second start returns the transfer that is already running
instead of starting another.

<!--
Source: docs.temporal.io/design-patterns/entity-workflow#best-practices and
temporal.io/blog/idempotency-and-durable-execution#idempotent-workflows.
Code: snippets/antipatterns/client.go, run by
snippets/antipatterns/antipatterns_test.go, which starts each version twice.

The ❌ version is the default most people write first, because the samples use
a UUID. A mobile app retries on a flaky connection, a user double-taps, a
load balancer replays a request: each one is a second transfer.

The ✅ version uses the ID the transfer already has. The blog's advice is to
generate that key as close to the user as possible: the app creates the
transfer ID when the user opens the form, not when the backend receives it.

What the second start does: the Service refuses to start a second run with the
same ID, and `ExecuteWorkflow` hands back a `WorkflowRun` for the one already
running, with the same Run ID. The caller can `Get` its result as if it had
started it. To be told instead, set `WorkflowExecutionErrorWhenAlreadyStarted:
true`, and the second start returns a
`*serviceerror.WorkflowExecutionAlreadyStarted`. The test checks both.

Two policies to know by name, without reading them out:
- `WorkflowIDConflictPolicy` is about a *running* Workflow with the same ID. The
  default fails the start; `WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING` returns
  the running one even when you asked to be told.
- `WorkflowIDReusePolicy` is about a *closed* one. The default allows a new run
  with the same ID once the first has finished. For a transfer that must never
  run twice, set `WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE`; it holds for the
  Namespace's retention period.

Bonus: the transfer is now findable by its own ID in the Temporal UI and the
CLI.
-->

---
layout: code-stack
heading: Polling an external system
proseMax: 30
---

::code::

```go
// ❌ Every pass adds an Activity and a Timer to history
var status string
err := workflow.ExecuteActivity(ctx, CheckSettlement, wireID).Get(ctx, &status)
for err == nil && status != "SETTLED" {
	if err = workflow.Sleep(ctx, time.Minute); err == nil {
		err = workflow.ExecuteActivity(ctx, CheckSettlement, wireID).Get(ctx, &status)
	}
}

// ✅ One Activity. The Retry Policy does the waiting
ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
	StartToCloseTimeout: 10 * time.Second,
	RetryPolicy: &temporal.RetryPolicy{
		InitialInterval:    time.Minute,
		BackoffCoefficient: 1,
	},
})
return workflow.ExecuteActivity(ctx, AwaitSettlement, wireID).Get(ctx, nil)
```

::default::

`AwaitSettlement` checks once and fails if the wire hasn't settled. Retries are
not recorded in history, so it can poll for days and history stays one Activity.

<!--
Source: docs.temporal.io/design-patterns/polling#infrequent-polling-resource-efficient.
Code: snippets/antipatterns/workflows.go and activities.go, run by
snippets/antipatterns/antipatterns_test.go in the SDK's time-skipping test
environment: four checks are four Activities and three Timers in the ❌
version, and four Attempts of one Activity, with no Timers, in the ✅ one.

The scenario: a wire has gone out on a payment network that has no webhook, and
the Workflow must wait for it to settle.

`BackoffCoefficient: 1` is what turns retries into a fixed poll interval:
every wait is the Initial Interval. The Activity returns a retryable
`temporal.NewApplicationError(...)` when the wire is still pending.

The docs split polling three ways, and it is worth naming the other two:
- Every second or faster: loop *inside* one Activity, heartbeating each pass.
- Every minute or slower: this slide.
- Several steps per poll, or parameters that change between polls: a Child
  Workflow that loops with `workflow.Sleep` and calls Continue-As-New.

If the network can call you back, skip polling: complete the Activity
asynchronously from the webhook (return `activity.ErrResultPending`, then
`client.CompleteActivity` with the Task Token).
-->

---
layout: code-stack
heading: Sleeping in an Activity
proseMax: 30
---

::code::

```go
// ❌ An Activity that waits out a three-day hold
func ReleaseAfterHold(ctx context.Context, deposit Deposit) error {
	time.Sleep(72 * time.Hour)
	return bank.Release(ctx, deposit.DepositID)
}

// ✅ A durable Timer in the Workflow, then a quick Activity
func DepositHoldWorkflow(ctx workflow.Context, deposit Deposit) error {
	if err := workflow.Sleep(ctx, 72*time.Hour); err != nil {
		return err
	}
	ctx = workflow.WithActivityOptions(ctx, releaseOptions)
	return workflow.ExecuteActivity(ctx, ReleaseFunds, deposit.DepositID).Get(ctx, nil)
}
```

::default::

The ❌ Activity holds a Worker slot for three days and starts over if the Worker
restarts. `workflow.Sleep` is a Timer on the Server: no slot, no restart.

<!--
Source: docs.temporal.io/develop/go/timers.
Code: snippets/antipatterns/activities.go (`ReleaseAfterHold`) and
workflows.go (`DepositHoldWorkflow`), run by
snippets/antipatterns/antipatterns_test.go in the time-skipping test
environment: one Timer of exactly three days, then the release.

The scenario: a deposited cheque is held for three business days before the
funds are released.

Everything wrong with the ❌ version, in the order people discover it:
- It needs a Start-To-Close Timeout longer than three days.
- If the Worker restarts on day two, nothing notices until that timeout
  expires, and then the retry waits another three days from zero. A Heartbeat
  Timeout would notice sooner, but the retry still starts the wait over.
- It occupies an Activity slot the whole time, so a Worker can only hold as
  many deposits as `MaxConcurrentActivityExecutionSize` allows.
- `time.Sleep` does not even watch `ctx`, so cancelling the Activity does not
  wake it. A `select` on `ctx.Done()` and `time.After` would, but it would
  still hold the slot.

The ✅ version costs two Events (TimerStarted, TimerFired), survives any number
of Worker restarts, and holds a million deposits as easily as one. Inside
Workflow code, `workflow.Sleep` is the only sleep: `time.Sleep` there would
block the Workflow's goroutine outside Temporal's control and break replay.
-->
