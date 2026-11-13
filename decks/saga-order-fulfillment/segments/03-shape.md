---
layout: default
title: You hold the compensations
---

# You hold the compensations

There is no built-in helper. A Saga here is a slice of funcs, a `defer`, and a loop
that walks backwards — about twenty lines, and that is the whole of today's lab.

Writing it yourself is not a downgrade. Every line is visible, and when the unwind
does something surprising at 3am the answer is in your file. The hard part —
durability — comes from the Workflow those twenty lines run inside, not from the
lines themselves.

<!--
Say this flatly and move on; it takes twenty seconds.

The point to land is that the absence of a helper is not an absence of a feature.
The bookkeeping a helper would do is trivial: hold a slice, run it backwards. The
part that is genuinely hard is that the slice survives a process dying halfway
through the unwind, and that comes from Temporal, not from any library.

`defer` is the Go-shaped part. It already means "on every way out of this
function", which is exactly when a Saga unwinds, so the language does the
plumbing.

If someone asks whether they should write their own reusable helper: sure, and it
will be about twenty lines. Do it after they have seen the mechanism, not before.
-->

---
layout: code-stack
heading: One type, one slice
proseMax: 34
---

::code::

```go {1|3|4-10|12|all}
type compensation func(ctx workflow.Context) error

func OrderFulfillmentWorkflow(ctx workflow.Context, order Order) (result OrderResult, err error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: 100 * time.Millisecond,
			MaximumAttempts: 2,
		},
	})

	var compensations []compensation
```

::default::

Five Activities: three forward steps and two undos. **The undos are Activities like any
other** — they get the same timeout and the same retry policy, which is exactly why they
have to be idempotent.

<!--
Source: workshops/saga-order-fulfillment/solution/workflow.go, lines 15-30.
Verbatim, doc comments trimmed.

- **Build 1 -** `type compensation func(ctx workflow.Context) error` is the
  entire abstraction. A func that takes a Workflow context and returns an error.
  Nothing else. It takes the context as an argument so the unwind can choose
  which context the undo runs on — that choice comes back on the unwind slide.
- **Build 2 -** the signature. `err` is a **named** result, and that is
  load-bearing: a deferred func can read it, so the unwind can see whether the
  Workflow is on its way out with a failure.
- **Build 3 -** one set of Activity options for every Activity this Workflow
  runs, undos included. There is no separate mechanism for compensations; the
  SDK does not know which of the five is an undo. `MaximumAttempts: 2` is a
  teaching setting, not a production one — it keeps the lab's failure paths
  fast. Say so, or someone will copy it.
- **Build 4 -** a plain nil slice. `append` puts the newest undo last.

The consequence of build 3 is the one to plant now, because it becomes
anti-pattern number two in five minutes: Temporal will retry a compensation the
same way it retries a charge. A compensation that is not safe to run twice is a
bug waiting for a network blip.
-->

---
layout: code-stack
heading: Register before you act
proseMax: 32
---

::code::

```go {1|3-8|10-15|all}
	var compensations []compensation

	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, ReleaseInventoryIfReserved, order).Get(ctx, nil)
	})
	if err = workflow.ExecuteActivity(ctx, ReserveInventory, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}

	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, RefundPaymentIfCharged, order).Get(ctx, nil)
	})
	if err = workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
```

::default::

`append` puts each new undo at the **end**, so walking the slice from the last index down
runs them newest-first. Each one is appended **before** the step it undoes.

<!--
Source: workshops/saga-order-fulfillment/solution/workflow.go, lines 30-74.
Verbatim, with the defer between them (next slide), comments and TODO markers
trimmed.

- **Build 1 -** a plain local slice. It lives in Workflow state, so it is part
  of what Temporal reconstructs on replay. A Worker that dies mid-order comes
  back knowing what it owes. The defer that unwinds it sits right below this
  line; it is the next slide.
- **Build 2 and 3 -** read the pairs out loud in the order they appear:
  register, then act. Register, then act. Say it twice — it is the single most
  important line ordering in the session and it is deliberately the opposite of
  what feels natural.
- **All -** off the slide, `DispatchCourier` runs next and registers nothing.
  It is last, so nothing depends on it.

Note that every failure path is a plain `return OrderResult{}, err`. The steps
do not know a Saga exists. Because `err` is the named result, the deferred
unwind sees it on the way out, whichever `return` fired.

Note the closure captures `order`. That is why each undo needs only a context.
The closure's own `ctx` parameter shadows the Workflow's on purpose: the undo
runs on whatever context the unwind hands it.
-->

---
layout: code-stack
heading: The unwind
proseMax: 34
---

::code::

```go {1-4|5-6|8|10-15|16|all}
	defer func() {
		if err == nil {
			return
		}
		logger := workflow.GetLogger(ctx)
		logger.Error("fulfillment failed, unwinding", "orderId", order.OrderID, "error", err)

		undoCtx, _ := workflow.NewDisconnectedContext(ctx)

		for i := len(compensations) - 1; i >= 0; i-- {
			// One failing undo must not strand the rest.
			if cerr := compensations[i](undoCtx); cerr != nil {
				logger.Error("compensation failed", "orderId", order.OrderID, "error", cerr)
			}
		}
	}()
```

::default::

Undos run on a **disconnected** context, so a cancelled order still unwinds. Each failure
is logged and the loop **keeps going**. `err` is left alone: compensating is not succeeding.

<!--
Source: workshops/saga-order-fulfillment/solution/workflow.go, lines 33-54.
Verbatim, comments trimmed.

- **Build 1 -** the defer, and its guard. A clean order returns with `err` nil
  and nothing unwinds. Forget the guard and every successful order refunds
  itself.
- **Build 2 -** `workflow.GetLogger` is replay-aware, so the line does not spam
  on every replay.
- **Build 3 -** the disconnected context, the part people never think about
  until it bites. Cancel a Workflow and its `ctx` is cancelled; an Activity
  started on a cancelled context is never scheduled. Run the undos on `ctx` and
  a cancelled order — the most ordinary way an order stops — strands both the
  reservation and the charge. `workflow.NewDisconnectedContext` gives a context
  the cancellation never reaches. The lab's Check cancels an order mid-dispatch
  to prove it.
- **Build 4 -** the loop, backwards, and the per-undo error check, which is the
  part people cut. `return` on the first failure instead and the first failing
  undo abandons every undo after it — the refund fails, and the stock stays
  reserved forever. Two problems instead of one.
- **Build 5 -** what is not there. Nothing assigns `err`, so the Workflow still
  fails with the courier's error once the unwind is done. Set `err = nil` here
  because the compensation "handled" it, and the Workflow completes, the caller
  books a delivery, and the Temporal UI shows a green tick over a failed order.
  That is strictly worse than the bug you started with.

The honest gap, and say it before someone finds it: a compensation that fails
after its retries are exhausted is logged and stepped over. Money is now
stranded and only a log line knows. In production this is where you raise
something a human sees — a Signal to a remediation Workflow, a Search Attribute
you alert on, a row in a queue. The comment in the lab code says exactly this.
Nothing in a Saga makes the undo path infallible; it only makes it recorded.
-->
