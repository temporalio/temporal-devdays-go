---
layout: section
---

# Anti-patterns

Three ways to write a Saga that quietly loses money.

<!--
Six minutes, and the heart of the session. Everything before this was setup.

Frame it as a confession, not a lecture: all three of these are what a careful
engineer writes on the first attempt. None of them fail loudly. All three pass a
happy-path test.
-->

---
layout: code-stack
heading: "Anti-pattern 1: register after the step"
proseMax: 36
---

::code::

```go {1-3|4-6|8-13|all}
	if err = workflow.ExecuteActivity(ctx, ReserveInventory, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, ReleaseInventoryIfReserved, order).Get(ctx, nil)
	})

	if err = workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, RefundPaymentIfCharged, order).Get(ctx, nil)
	})
```

::default::

Reads better. Feels safer — why register an undo for something that has not happened yet?
Because when the step returns an error, the `return` fires first. There is a **window where
the effect exists and nothing knows how to reverse it.**

<!--
Source: workshops/saga-order-fulfillment/exercise/workflow.go, lines 48-53 and
58-63. This is literally the code they will open in the editor. TODO 2a and 2b
ask them to move each `append` above its step.

- **Build 1 -** the step, and Go's usual early return on error.
- **Build 2 -** the registration, which only runs if the step said it worked.
- **Build 3 -** the same shape for the charge. This is the one that costs money.

Ask the room which of the two orderings they would have written. Most people say
this one, and they say it with confidence, because in a single-process mental
model the step either happened or it did not. In Go it also looks idiomatic:
check the error, then carry on.

The window is not a race condition and it is not about concurrency. It is about
a step that lands its effect and then reports failure. Next slide is the proof,
so do not explain the mechanism here — just plant the question: what if
`ChargePayment` returns an error *after* the money moved?

Cost of getting it wrong, in each direction:
- Register before, effect never landed: the undo runs against nothing. Harmless,
  provided the undo is idempotent — which is anti-pattern 2, and why these two
  are a pair rather than two independent rules.
- Register after, effect landed: nothing ever reverses it. Silent, permanent,
  and discovered by a customer.

One is a no-op. The other is a stranded charge. That asymmetry is the entire
argument, and it is the sentence to repeat if you only get one.
-->

---
layout: code-stack
heading: The step that lands, then fails
proseMax: 34
---

::code::

```go {1-4|6-10|all}
func ChargePayment(ctx context.Context, order Order) (string, error) {
	// The effect lands first, exactly as a real gateway would capture the money
	// before returning. Then the call reports failure.
	record("charge:" + order.OrderID)

	if order.ChargeFailsAfterEffect {
		return "", temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("payment gateway timed out after capturing %d cents", order.AmountCents),
			"PaymentGatewayTimeout", nil)
	}
	return fmt.Sprintf("charged %d cents", order.AmountCents), nil
}
```

::default::

`ChargeFailsAfterEffect` — the scenario switch that makes this concrete. The money moved.
The Workflow was told it did not. **An error is not proof that nothing happened.**

<!--
Source: workshops/saga-order-fulfillment/solution/activities.go, lines 24-35,
verbatim, and the `ChargeFailsAfterEffect` doc comment in order.go.

This slide exists to kill one specific and very common assumption: that a
returned error means the call had no effect. Every engineer in this room has
written code that assumes it — `if err != nil { return }` is the most-typed
line in Go.

The real-world version is not exotic. A gateway captures the funds, then the
connection drops before the 200 comes back. A write commits, then the ack is
lost. A function does its work, then times out. Your client sees a failure. The
world saw a success.

Note the failure is non-retryable. That matters: a retry would double-charge,
because `ChargePayment` itself is not idempotent. In production you would give
this an idempotency key and let it retry safely — and that is worth a sentence,
because it is the other half of the same lesson. The lab keeps it non-retryable
so the compensation path is unambiguous and fast.

Now close the loop from the previous slide out loud: register after the step,
run this scenario, and the refund was never appended when the defer fires. The
charge is stranded. Register before, and the refund is already on the stack
before the gateway ever sees the request.
-->

---
layout: code-stack
heading: "Anti-pattern 2: an undo that is not idempotent"
proseMax: 32
---

::code::

```go {2-5|6|all}
func ReleaseInventoryIfReserved(ctx context.Context, order Order) error {
	if !isReserved(order.OrderID) {
		fmt.Printf("[compensate] nothing reserved for %s, no-op\n", order.OrderID)
		return nil
	}
	record("release:" + order.OrderID)
	return nil
}
```

::default::

The `If` in the name is a promise the body has to keep. Same shape in
`RefundPaymentIfCharged`, guarded by `isCharged`.

<!--
Source: workshops/saga-order-fulfillment/solution/activities.go, lines 51-60,
verbatim with the TODO markers trimmed. TODO 3a and 3b in the exercise are
exactly this guard, missing.

- **Build 1 -** the guard. Nothing reserved, nothing to release, and the
  Activity still completes successfully.
- **Build 2 -** the effect, reached only when there is something to undo.

Two reasons an undo runs when there is nothing to undo, and both are ordinary:
1. **The effect never landed.** You registered the compensation before the step,
   so if the step fails cleanly the undo is on the stack anyway. That is by
   design — it is the price of anti-pattern 1's fix.
2. **Temporal retried it.** Compensations are Activities. They get the same
   retry policy as everything else. A refund whose response is lost gets called
   again, and a second refund sends money the wrong way.

So the naming convention is not decoration. `ReleaseInventoryIfReserved`, not
`ReleaseInventory`. The name states the precondition, and a reviewer who sees
`IfReserved` with no `if` in the body has caught the bug from the diff.

`isReserved` stands in for the check a real system makes — reading the
reservation's status, or the gateway's own idempotency key. The lab keeps it in
memory so the effect is visible in the Worker terminal; the shape is the same.

If someone asks why not make the *forward* steps idempotent instead: you should,
and that is a different fix for a different problem. Idempotent forward steps
let you retry. Idempotent undos let you compensate. You need both.
-->

---
layout: checklist
eyebrow: "ANTI-PATTERN 3"
title: The unwind that gives up
subtitle: "Five lines people get wrong in the defer, and what each one costs."
items:
  - "return on the first failing undo — every undo after it is abandoned"
  - "err = nil after compensating — the Workflow completes green over an order that never shipped"
  - "A forward loop over the slice — the stock goes back before the card is refunded"
  - "Undos on ctx, not a disconnected context — a cancelled order never unwinds"
  - "No log on a failed compensation — money is stranded and nothing knows"
---

<!--
Source: the defer in
workshops/saga-order-fulfillment/solution/workflow.go, read as a list of what
happens when each of its parts is wrong. TODO 1 in the exercise asks for all
five, and the Challenge 1 Check has a test for each of the first four.

Walk them one at a time and name the failure, not the rule. The rule is
forgettable; "green tick over a failed order" is not.

Row two is the one to spend time on with this audience. A team running
production Workflows has dashboards built on Workflow status. Clearing the error
does not just hide the failure from the caller — it poisons every metric and
alert downstream of it. Compensating is not succeeding.

Row three is the LIFO trap in its most realistic form. Nobody writes a forward
loop to be wrong on purpose; they write `for _, undo := range compensations`
because that is the idiomatic way to read a slice, and `append` because that is
the idiomatic way to grow one. Two idiomatic choices compose into the wrong
order.

Row four is the one a happy-path test never finds, because nothing is
cancelled in a happy path. It only shows up the day someone cancels an order in
the Temporal UI.

Row five is the honest limit of the pattern, already raised on the unwind slide.
Repeat it here in one breath and move on to the results.
-->
