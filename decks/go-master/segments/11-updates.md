---
layout: section
---

<ZiggyHero scene="update">

# Go Updates

Synchronous write with result

</ZiggyHero>

<!--
Source: not in the original PowerPoint, which went from Queries straight to the
quiz. Added when the Deck was revised.

Updates complete the set: Signals write, Queries read, Updates do both.

If the room has a request/response use case (an API that must tell the caller
whether the change worked), this is the segment that answers "do I need a
database next to my Workflow for that?" No.
-->

---
layout: code-walkthrough
heading: Go Updates
notes:
  - "<code>WithdrawUpdate</code> names a message the caller <strong>waits on</strong>. The handler's signature says what it takes and returns"
  - Unlike a Signal it returns a value, and unlike a Query it can change state
  - It can block on Activities and Timers. The caller waits until it finishes
  - Change state, then return the result to the caller
  - "The <code>Validator</code> runs first, before anything is written to history"
  - Return an error to <strong>reject</strong> the Update. The caller gets an error; history is untouched
---

::code::

```go {all|1|7|8|11-12|15|16-18}
const WithdrawUpdate = "withdraw"

func AccountWorkflow(ctx workflow.Context, opening int) error {
	balance := opening

	err := workflow.SetUpdateHandlerWithOptions(ctx, WithdrawUpdate,
		func(ctx workflow.Context, amount int) (int, error) {
			if err := debitAccount(ctx, amount); err != nil {
				return 0, err
			}
			balance -= amount
			return balance, nil
		},
		workflow.UpdateHandlerOptions{
			Validator: func(ctx workflow.Context, amount int) error {
				if amount > balance {
					return fmt.Errorf("insufficient funds: balance is %d", balance)
				}
				return nil
			},
		},
	)
	if err != nil {
		return err
	}

	return workflow.Await(ctx, func() bool { return balance == 0 })
}
```

<!--
Source: not in the original PowerPoint; added when the Deck was revised.
Code: snippets/updates/updates.go, run by snippets/updates/updates_test.go.
`debitAccount` is a helper off the slide that sets Activity Options and runs
the `DebitAccount` Activity.

The scenario: an ATM or a payments API withdraws from an account and must tell
the customer the new balance, or refuse the withdrawal outright. A Signal
cannot answer and a Query cannot change the balance.

- **Build 1 - the name.** An exported constant, like Signals and Queries. The
  string is the Update name. What it takes and returns is the handler's own
  signature: an `int` amount in, an `int` balance and an `error` out.
- **Build 2 - the handler.** It returns a value. That is the point: a Signal
  returns nothing and a Query must not change state; an Update does both.
- **Build 3 - blocking work.** An Update handler can run Activities, Timers and
  `workflow.Await`. Each handler runs in its own Workflow coroutine. The
  caller's `UpdateWorkflow` with `WorkflowUpdateStageCompleted` waits until
  the handler returns, so the balance it gets back reflects a debit that
  actually posted.
- **Build 4 - change and answer.** The return value goes back to the caller and
  is recorded in history as `WorkflowExecutionUpdateCompleted`.
- **Build 5 - the Validator.** Passed in `workflow.UpdateHandlerOptions`. It
  runs before the Update is accepted, has the handler's parameters and returns
  only an `error`. It must not change state or block.
- **Build 6 - reject.** Returning an error (or panicking) in the Validator
  rejects the Update: the caller gets the error, and nothing is written to
  history. After acceptance it is different: an error the handler returns fails
  just that Update, recorded in history, and the Workflow carries on. A panic
  in the handler is not an Update failure: it fails the Workflow Task, which
  retries while the caller waits.

One gotcha worth saying, and in a bank it is the one that matters: the
Validator runs when the Update arrives, but the handler blocks on an Activity
before it subtracts. Only one Workflow coroutine runs at a time, yet handlers
still interleave at every blocking call: on a balance of 100, two concurrent
`withdraw(80)` calls can both pass the Validator, both debit, and leave the
account at -60. The fix: serialize the handlers with `workflow.NewMutex(ctx)`
and re-check the balance once you hold the lock
(snippets/updates/updates.go, `LockedAccountWorkflow`). A `sync.Mutex` is the
wrong tool here: it would block the Worker's goroutine, not the coroutine.
-->

---
layout: default
---

# Signals, Queries, and Updates

```go
handle, err := c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
	WorkflowID: workflowID, UpdateName: WithdrawUpdate, Args: []any{50},
	WaitForStage: client.WorkflowUpdateStageCompleted,
})
if err == nil {
	err = handle.Get(ctx, &balance)
}
```

<div class="three-way">

|                     | Signal              | Query          | Update                     |
| ------------------- | ------------------- | -------------- | -------------------------- |
| **Changes state**   | Yes                 | No             | Yes                        |
| **Returns a value** | No                  | Yes            | Yes                        |
| **Caller waits**    | Until it's recorded | For the answer | Until the handler finishes |
| **Can be rejected** | No                  | N/A            | Yes, by a validator        |
| **In history**      | Yes                 | No             | Yes, if accepted           |
| **Needs a Worker**  | No                  | Yes            | Yes                        |

</div>

<style>
/* table-compare splits two value columns; this table has three. */
.three-way th:not(:first-child),
.three-way td:not(:first-child) {
  width: 27%;
}
</style>

<!--
Source: not in the original PowerPoint; added when the Deck was revised.
Code: snippets/updates/client.go, run by snippets/updates/updates_test.go.

The Client code: `UpdateWorkflow` with `WorkflowUpdateStageCompleted` sends the
Update and waits for it to finish; `handle.Get` decodes the result, here the
balance after the withdrawal.
`WaitForStage: client.WorkflowUpdateStageAccepted` returns the handle as soon
as the Update is accepted, if the caller wants to do something else while the
handler runs; call `handle.Get` later. There is also Update-with-Start
(`c.UpdateWithStartWorkflow` with an operation from
`c.NewWithStartWorkflowOperation`), which starts the Workflow if it is not
running yet and sends the Update in one call: the standard shape for "open the
account if needed, then make the deposit".

Read the "Needs a Worker" row out loud. A Signal is accepted by the Server even
if every Worker is down; it waits in history. Queries and Updates need a Worker
to answer, so a caller blocked on an Update with no Workers just waits. Pass a
`ctx` from `context.WithTimeout` to bound it. Use Signals for fire-and-forget
(the wire approval), Updates when the caller needs to know (did the withdrawal
go through, and what is left).

Updates have their own limits: at most 10 in flight at once per Workflow, and
2,000 per run (rejected ones do not count). Continue-As-New resets the count.
-->

---
layout: comparison
heading: "Resumable Activity (Pause On Failure)"
leftTitle: What it is
leftFeatures:
  - "An Activity's retries run out, and the Workflow parks instead of failing"
  - "It waits, using nothing, for a human to Signal in corrected input"
  - "Then it runs the same Activity again, in the same Workflow"
rightTitle: When to use it
rightFeatures:
  - "The failure is bad input someone can fix: a mistyped account number"
  - "An invalid reference, or a missing record that will be created"
  - "Restarting from scratch is worse: the customer would re-enter everything"
---

<!--
Source: docs.temporal.io/design-patterns/resumable-activity, overview and
problem statement.

The framing that lands: retrying can't fix bad input, because the same wrong
account number fails the same way every time. The usual options are both bad:
fail the Workflow (the customer re-enters everything), or mark the error
non-retryable (same thing). What you want is for the Workflow to pause, using
nothing, until someone with authority supplies the right value, and then pick
up exactly where it stopped.

When *not* to use it: transient failures (timeouts, 503s) are the Retry
Policy's job, and failures nobody can correct should fail fast with a
non-retryable error. And it is not the Approval pattern: approval gates
progress on a decision; this recovers from a failure with corrected data.

Next slide: the room plays the on-call engineer for exactly this scenario.
-->

---
layout: default
clicks: 3
---

# Resumable Activity: you're on call

<div class="rt-slide">
<ResumableTransfer />
</div>

<style>
.rt-slide {
  height: 22rem;
}
</style>

<!--
Source: docs.temporal.io/design-patterns/resumable-activity, and its best
practices.
Code: snippets/resumable/resumable.go and activities.go, run by
snippets/resumable/resumable_test.go; the console is
components/ResumableTransfer.vue.

Hand the room the keyboard, figuratively: "you're the engineer on call; what
do you do first?" Each click (arrow key or clicker) performs the next action,
in order. The buttons jump to an action if someone calls one out of order;
Reset goes back to the start.

The pattern in one breath: when an Activity fails because its input is wrong,
retries cannot fix it and failing the Workflow makes the customer start over.
So the Workflow catches the failure, parks on `workflow.Await`, and waits for a
Signal carrying the corrected input. Then it runs the same Activity again.

What each button teaches:
- **Send the transfer.** `ExecuteTransfer` returns
  `temporal.NewNonRetryableApplicationError(..., "AccountNotFoundError", nil)`
  for an unknown account, so there is one attempt, not the Retry Policy's three
  (it is `MaximumAttempts: 3` for timeouts and 503s). The Workflow gets a
  `*temporal.ActivityError` back from `Get` and is Running, not Failed.
- **Query getStatus.** The answer is `AWAITING_CORRECTION`, and the history
  does not grow. Point at that. In production, also set a Search Attribute with
  the status so operators can list every parked transfer, and alert when one
  parks rather than hoping someone opens the Temporal UI.
- **Signal "12345678".** The parked `workflow.Await` returns, the loop runs
  `ExecuteTransfer` again, and the Workflow completes. Same run ID throughout:
  the test checks it.

Two rules worth one sentence each: park only on a `*temporal.ActivityError`
(`errors.As`), and let cancellation through (`temporal.IsCanceledError`), so a
Workflow bug is not parked as if it were bad data; and if a Workflow may be
corrected many times, Continue-As-New before the history grows.

This is not the Approval pattern: approval gates progress on a decision; this
recovers from a failure with corrected data. The docs' version adds an approval
Signal after the transfer, which is the Signals segment's pattern.

The live demo on the next slide is this exact scenario against a real Worker.
-->

---
layout: comparison
heading: "Resumable Activity in practice"
leftTitle: Key advantages
leftFeatures:
  - "The Workflow stays alive throughout: no restart from scratch"
  - "No external database, queue or polling mechanism"
  - "Status, corrected input and any approval are durably persisted the entire time"
rightTitle: Best practices
rightFeatures:
  - "Allow a few automatic retries before parking, for transient failures"
  - "Expose the status through a Query"
  - "Validate the corrected input in the Signal handler"
  - "Continue-As-New if many correction cycles are expected"
  - "Notify the operator (email, Slack, a ticket) when it parks"
---

<!--
Source: docs.temporal.io/design-patterns/resumable-activity, Key advantages and
Best practices.

One minute, and every row maps to something the room just clicked through:

- **Stays alive.** The status pill said Running the whole time; one run ID from
  start to finish.
- **No database or queue.** The parked transfer, its status and the corrected
  account live in the Workflow. Nothing to provision, nothing to reconcile.
- **Durably persisted.** Kill every Worker while it is parked: the Signal is
  still accepted and the transfer resumes when a Worker comes back.

- **Retries before parking.** `MaximumAttempts: 3` handles a timeout or a 503;
  only the non-retryable `AccountNotFoundError` parks at once.
- **Query.** That was step 2. Pair it with a Search Attribute so operators can
  list every parked transfer, not just look one up.
- **Validate in the handler.** The coroutine that receives the Signal ignores
  anything that is not 8 to 12 digits, but a Signal cannot report that back:
  the sender sees success either way. If the operator needs an error back, make
  the correction an Update with a Validator.
- **Continue-As-New.** Each correction cycle adds Events; a transfer corrected
  once never needs it, a long-lived record corrected weekly does.
- **Notify.** The transition to AWAITING_CORRECTION is the place to send the
  alert, rather than hoping someone opens the Temporal UI.
-->

---
layout: section
---

<ZiggyHero scene="demo">

# Demo: Resumable Activity

A transfer to a mistyped account, parked, corrected, and sent.

</ZiggyHero>

<!--
The code is the previous slide's, in snippets/resumable/. The bank knows two
accounts: 10000001 and 12345678.

Terminal 1, the dev server (skip if one is running):

    temporal server start-dev

Terminal 2, the Worker, from decks/go-master/snippets:

    go run ./resumable/demo

Terminal 3, a transfer to an account that does not exist:

    temporal workflow start --type TransferWorkflow \
      --task-queue resumable-demo --workflow-id transfer-4471 \
      --input '{"fromAccount":"10000001","toAccount":"99999999","amount":500}'

Open the Temporal UI at localhost:8233 and find `transfer-4471`. Show the room
the ActivityTaskFailed Event with `AccountNotFoundError`, then point out the
Workflow is still Running, not Failed.

Ask it where it is:

    temporal workflow query --workflow-id transfer-4471 --name getStatus

It answers `"AWAITING_CORRECTION"`. Optional: send a malformed fix and query
again to show the handler ignored it:

    temporal workflow signal --workflow-id transfer-4471 \
      --name retryWithCorrection --input '"acct 1234"'

Now the real fix:

    temporal workflow signal --workflow-id transfer-4471 \
      --name retryWithCorrection --input '"12345678"'

    temporal workflow result --workflow-id transfer-4471

It completes with "Transfer of 500 to 12345678 completed". Back in the
Temporal UI, the whole story is one history: the failure, the Signal, the
second ExecuteTransfer, the result.

To run it again, use a new --workflow-id, or wait for this one to close: the
default reuse policy allows a new run once the first has finished.
-->
