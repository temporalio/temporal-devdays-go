---
layout: eyebrow-hero
eyebrow: IMPLEMENTATION
---

# The order fulfillment Workflow

Register a compensation before or after each step, and run all compensations in
**reverse order** on failure.

<!--
Source: Temporal docs, Saga pattern page (Implementation).

Twenty seconds. The docs page shows a compact version of this Workflow. The
lab's version is the same shape with two additions the docs' own best practices
ask for: an error check around each compensation, and a log line when one
fails. The Go SDK has no Saga helper, so the compensation stack is ordinary
code: a slice of functions, a `defer`, and a loop that walks the slice
backwards. The slide's version also lets anything but an Activity failure
straight through; the code slide says why.
-->

---
layout: code-walkthrough
heading: Implementation
notes:
  - "The Workflow: an order in, a result out"
  - "A list of undos, kept in Workflow state. Replay rebuilds it after a crash"
  - "Everything that can fail goes inside the <code>try</code>"
  - "Register the undo <strong>before</strong> the step, on top of the list. Nothing runs yet"
  - "Step 1 runs as an Activity. Its undo is already on record"
  - "Same pair for the charge, so a charge that lands and then fails is still refunded"
  - 'Last step, no undo needed. No courier fails with <code class="ref">CourierUnavailable</code>'
  - "Happy path: return the tracking number. The undo list is simply dropped"
  - "Compensate <strong>Activity failures</strong> only. A bug in Workflow code retries instead of refunding"
  - "Log it. The Workflow logger is replay-aware"
  - "Walk the list from the top: newest first, so refund, then release"
  - "Each undo is itself an Activity, with its own timeouts and retries"
  - "Each undo has its own <code>catch</code>. One failure is logged, the rest still run"
  - "Rethrow the original failure. <strong>Compensating is not succeeding</strong>"
---

::code::

```txt {all|1|2|4|5|6|8-9|11|12|14|15|17|19|18-21|23}
workflow fulfillOrder(order):
  undos = []

  try:
    undos.push(releaseInventory)
    reserveInventory(order)

    undos.push(refundPayment)
    chargePayment(order)

    trackingNumber = dispatchCourier(order)
    return trackingNumber

  catch activityFailure:
    log("fulfillment failed, unwinding")

    for each undo in undos, newest first:
      try:
        undo(order)
      catch error:
        log("compensation failed"), keep going

    rethrow activityFailure
```

<!--
Pseudocode of workshops/saga-order-fulfillment/solution/workflow.go, plus one
guard the Workshop does not have yet (compensate Activity failures only). The
next slide is the same Workflow in real Go; this one is the shape, without the
syntax.

Five minutes, one click per note. Above the function, out of frame: one
`workflow.ActivityOptions` value gives every Activity a 10 second
Start-To-Close Timeout and two attempts, and
`type Compensation func(ctx workflow.Context) error` is the whole abstraction.
The SDK does not know which of the five Activities is an undo.

The walk is in execution order: forward through the `try`, then the `catch` and
the unwind. The builds below name the Go each pseudocode line becomes on the
next slide: `undos` is the `compensations` slice, `push` is `append`, each step
is `workflow.ExecuteActivity(...).Get(...)` with an `if err != nil` return, and
`try`/`catch activityFailure` is a deferred function that checks the named
`err` result with `errors.As(err, &activityErr)`.

The bubbles are the short version. The full explanation for each build:

- **Build 1 - Line 1: a Workflow definition.** In Go a Workflow is an exported function whose first parameter is a `workflow.Context`. The Worker registers it, and a client starts it by name on the fulfillment-tasks Task Queue. It takes an `Order` and returns an `OrderResult` and an `error`; Temporal serializes both values as JSON.

- **Build 2 - Line 2: the compensation slice.** A plain local slice, kept in Workflow state. If the Worker crashes, replay rebuilds it, so the Workflow still knows what it owes. `Compensation` is `func(ctx workflow.Context) error`: a closure over `order` that schedules one compensation Activity. A small helper, `undo(ReleaseInventoryIfReserved, order)`, builds one.

- **Build 3 - Line 4: the failure boundary.** In Go there is no `try`: every step returns an error, and the Workflow returns it at once. The `defer` registered at the top of the function is the boundary: it runs when the function returns, and sees the error through the named `err` result.

- **Build 4 - Line 5: register before the step.** `append` puts the undo on top of the slice before reserving. The closure is stored, not called, so nothing runs yet.

- **Build 5 - Line 6: step 1 runs.** `workflow.ExecuteActivity` schedules `ReserveInventory` on the Task Queue, and `.Get` blocks the Workflow until its result. The reservation lands in the inventory system, outside the Workflow. Its undo is already on record.

- **Build 6 - Lines 8-9: the same pair for the charge.** Register `RefundPaymentIfCharged`, then charge. Registering first means a charge that lands and then reports failure is still refunded.

- **Build 7 - Line 11: step 3 has no compensation.** `DispatchCourier` is last, so no later step can fail and leave it standing. When there is no courier, it returns a non-retryable `ApplicationError` of type `CourierUnavailable`, and the Workflow sees a `*temporal.ActivityError` that wraps it.

- **Build 8 - Line 12: the happy path.** If dispatch succeeds, the Workflow returns the tracking number with a nil error. The deferred function sees no error and returns: the slice is simply dropped, nothing is undone.

- **Build 9 - Line 14: compensate Activity failures only.** A `*temporal.ActivityError` is the business failure a Saga compensates. Anything else returns before a single undo is scheduled. A bug that panics in Workflow code, such as an index out of range, fails the Workflow Task, not the Workflow: the named `err` is still nil when the deferred function runs, so nothing is refunded, and Temporal retries the Task so you can fix and redeploy. The Workshop's code compensates on any error; without this check, a plain `errors.New` returned by mistake would schedule real refunds. One thing this slide does not handle: if someone cancels the Workflow, the error is a `*temporal.CanceledError` and it goes straight through. Compensating on cancellation means checking for it too; the disconnected context on the next slide is what lets undos run in a cancelled Workflow.

- **Build 10 - Line 15: log the failure.** `workflow.GetLogger(ctx)` is replay-aware: it does not repeat this line when the Workflow replays its history. `errors.As` can dig into the error for the Activity's own failure, here `CourierUnavailable`.

- **Build 11 - Line 17: newest first.** Each undo was appended to the end, so the loop walks the slice backwards, from `len(compensations) - 1` down to 0: `RefundPaymentIfCharged`, then `ReleaseInventoryIfReserved`. The undo order mirrors the forward order. A forward loop would release the stock before refunding the card.

- **Build 12 - Line 19: run one compensation.** `compensations[i](ctx)` calls the stored closure, which schedules the compensation Activity and waits for it. Compensations are Activities, so they get timeouts and retries too.

- **Build 13 - Lines 18-21: one failure must not strand the rest.** Each compensation's error is checked on its own. If one fails after its retries, it is logged and the loop moves on to the next. In production, this is where you alert a human.

- **Build 14 - Line 23: return the original failure.** The deferred function never touches `err`, so the original `*temporal.ActivityError` is what the Workflow returns. The Workflow Execution ends Failed and the caller learns the order did not go through. Compensating is not succeeding.
-->

---
layout: default
---

# Implementation: the unwind

<div class="uw">
<div class="uw-code">

```go {all|1-3|20|21|24|25|29|5-11|13-15|13-15|18}
func OrderFulfillmentWorkflow(ctx workflow.Context, order Order) (_ OrderResult, err error) {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	var compensations []Compensation

	defer func() {
		var activityErr *temporal.ActivityError
		if !errors.As(err, &activityErr) {
			return
		}
		logger := workflow.GetLogger(ctx)
		logger.Error("fulfillment failed, unwinding", "order", order.OrderID, "error", err)
		ctx, _ := workflow.NewDisconnectedContext(ctx)
		for i := len(compensations) - 1; i >= 0; i-- {
			if cerr := compensations[i](ctx); cerr != nil {
				logger.Error("compensation failed", "order", order.OrderID, "error", cerr)
			}
		}
	}()

	compensations = append(compensations, undo(ReleaseInventoryIfReserved, order))
	if err := workflow.ExecuteActivity(ctx, ReserveInventory, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	compensations = append(compensations, undo(RefundPaymentIfCharged, order))
	if err := workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	var tracking string
	if err := workflow.ExecuteActivity(ctx, DispatchCourier, order).Get(ctx, &tracking); err != nil {
		return OrderResult{}, err
	}
	return OrderResult{OrderID: order.OrderID, TrackingNumber: tracking}, nil
}
```

</div>
<SagaUnwind />
</div>

<style>
h1 { font-size: 1.8rem !important; margin-bottom: 0.5rem !important; }
.uw { display: grid; grid-template-columns: minmax(0, 1fr) 34%; gap: 1rem; height: 26.5rem; }
.uw-code { min-width: 0; overflow: hidden; }
.uw-code {
  --slidev-code-font-size: 0.52rem;
  --slidev-code-line-height: 0.74rem;
}
.uw-code :deep(pre.shiki) { margin: 0; padding: 0.5rem 0.6rem; }
.uw-code :deep(pre code), .uw-code :deep(pre .line) { font-size: 0.52rem !important; line-height: 0.74rem !important; }
</style>

<!--
Source: the same Workflow as the previous slide, in Go, with
components/SagaUnwind.vue.
Code: snippets/saga/workflow.go, run by snippets/saga/saga_test.go (happy
path, failed dispatch, a charge that fails after it lands, a reservation that
never landed, a compensation that fails, and a panic that refunds nothing).

Three minutes, ten clicks, one order whose courier fails. This is the Go the
pseudocode becomes; watch the right-hand side.

- **Read it in run order, not top to bottom.** The `defer` sits at the top,
  but it only registers the unwind. It runs when the function returns, after
  whichever step failed. The clicks follow execution: down through the steps,
  then back up into the deferred function.
- **Stack.** Each `append` puts an undo on top. Nothing on the stack runs
  until the courier fails. Then the unwind starts from the top: the refund,
  registered last, runs first. The code does not actually remove anything from
  the slice; the picture shows what is still owed.
- **`workflow.NewDisconnectedContext`.** The undos run on a context that is
  not cancelled with the Workflow. On a plain failure it changes nothing; it is
  there so the same unwind still schedules Activities if you extend the check
  to cancellation.
- **Event History.** This is what the Temporal UI shows for the same run, in
  order: the forward Activities and the failed `DispatchCourier` come first,
  the two compensation Activities come second, and `WorkflowExecutionFailed`
  closes it. Workflow Task Events are left out, as the UI's compact view does.

Open a real failed run in the Temporal UI after this slide, or in the lab, and
point at the same order of Events.
-->
