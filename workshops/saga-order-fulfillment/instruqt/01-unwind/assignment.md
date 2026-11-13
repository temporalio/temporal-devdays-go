---
slug: unwind
type: challenge
title: 1. Unwinding an order that failed halfway
teaser: An order reserves stock and charges a card, then fails. Watch nobody clean
  up, then write the unwind.
notes:
- type: text
  contents: |-
    # Two steps went through. The third failed. Nothing rolls back.

    Placing an order takes three steps, one after another: hold the stock,
    charge the card, book the courier.

    When the courier step fails, the first two have already happened. The stock
    is held and the money is gone. No database transaction spans three separate
    systems, so there is nothing to roll back.

    Somebody has to undo them, in the right order, with code that works even
    when it cannot tell what actually happened.

    **Author:** [Nikolay Advolodkin](https://www.linkedin.com/in/nikolayadvolodkin/), Staff Developer Advocate
tabs:
- title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/saga-order-fulfillment
  port: 8080
- title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/saga-order-fulfillment/exercise
- title: Client
  type: terminal
  hostname: workshop
  workdir: /root/workshop/saga-order-fulfillment/exercise
- title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: intermediate
timelimit: 900
---

# The setup

Every order takes three steps, each one a separate system:

1. Reserve the stock. Inventory service.
2. Charge the card. Payment gateway.
3. Book the courier. Logistics.

No transaction spans those three. When step 3 fails, steps 1 and 2 have already
happened.

The [button label="Exercise" background="#444CE7"](tab-0) tab opens two trees.
You work in `exercise/`. Every file path below starts there. `solution/` is the
same tree, finished, for when you are stuck. The editor saves as you type.

The blue buttons in this text are clickable. Each one jumps to the tab it
names.

# Watch it fail

Start the Worker in the [button label="Worker" background="#444CE7"](tab-1) tab:

```bash,run
go run ./worker
```

It prints `Fulfillment Worker polling fulfillment-tasks` and then sits there.
Leave it running. Every `[ledger]` line in this lab appears in that tab.

Now place an order that fails at the courier, from the
[button label="Client" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter -fail-dispatch
```

Among the Worker's log lines are two ledger entries, and then nothing:

```bash,nocopy
[ledger] reserve:ORD-XXXXXX
[ledger] charge:ORD-XXXXXX
```

Stock held, card charged, courier failed, nothing undone. The customer paid for
an order that will never ship.

The ledger is something this Workshop made up. The Event History is real, and
it is what you get in production. Open the Workflow in the
[button label="Temporal UI" background="#444CE7"](tab-3) tab:

| Event | Activity |
|---|---|
| Activity Task Completed | `ReserveInventory` |
| Activity Task Completed | `ChargePayment` |
| Activity Task Failed | `DispatchCourier` |
| Workflow Execution Failed | |

What matters is what is missing. There is no `ReleaseInventoryIfReserved` and no
`RefundPaymentIfCharged` anywhere in the history. After every run in this lab,
ask the same question. Which Activities are in the history, and which should be
there and are not?

# Unwind (TODO 1)

**Observe.** Open `exercise/workflow.go` in the
[button label="Exercise" background="#444CE7"](tab-0) tab. Press `Ctrl+P`
(`Cmd+P` on a Mac), type the file name, then `Ctrl+G` to jump to a line.

<!-- snippet: exercise/workflow.go TODO 1 before=3 -->
Lines 29-43 of `OrderFulfillmentWorkflow` hold the compensations, and the empty
space where they should run:

```go
	// Each undo is appended, so the newest is last.
	var compensations []compensation

	// TODO 1 start
	// Nothing unwinds yet. Defer a func that, when err is not nil:
	//
	//   - makes undoCtx, _ := workflow.NewDisconnectedContext(ctx). A cancelled
	//     Workflow cancels ctx, and an undo started on it would never run.
	//   - walks compensations BACKWARDS, from len(compensations)-1 down to 0, and
	//     calls each one with undoCtx. The newest undo goes first.
	//   - logs a failed undo with workflow.GetLogger(ctx) and keeps going, so one
	//     failing undo does not strand the rest.
	//
	// Leave err as it is: the Workflow must still fail once the unwind is done.
	// TODO 1 end
```

**Learn.** There is no built-in Saga helper. You hold the compensations
yourself: `compensations` is a slice of funcs, and each step below appends its
undo to it. Nothing runs them yet.

Three things decide how they run:

- **When.** A `defer` runs on every way out of the function. The Workflow
  returns a named `err`, so the deferred func can see whether the order failed.
- **Which way.** Later steps are built on earlier ones. Undo the newest first,
  or you release stock while a charge still stands against it.
- **On which context.** When a Workflow is cancelled, its `ctx` is cancelled,
  and an Activity started on a cancelled context never runs. A context from
  `workflow.NewDisconnectedContext` is not cancelled with it, so the undos still
  run.

**Do.** Between the TODO 1 markers, add a `defer` that, when `err` is not nil:

- [ ] makes a disconnected context, `undoCtx`
- [ ] calls each compensation with `undoCtx`, from the last one to the first
- [ ] logs a compensation that fails, and keeps going

Leave `err` alone. Compensating is not succeeding, and the caller has to learn
the order did not go through.

<details>
<summary>Hint</summary>

A backwards walk over a slice:

```go
for i := len(items) - 1; i >= 0; i-- {
	use(items[i])
}
```

</details>

<details>
<summary>Show solution</summary>

```go
	defer func() {
		if err == nil {
			return
		}
		logger := workflow.GetLogger(ctx)
		logger.Error("fulfillment failed, unwinding", "orderId", order.OrderID, "error", err)

		// A cancelled Workflow cancels ctx, and with it every Activity started on
		// it. The undos run on a disconnected context that cancellation never reaches.
		undoCtx, _ := workflow.NewDisconnectedContext(ctx)

		// Newest first: walk the slice backwards.
		for i := len(compensations) - 1; i >= 0; i-- {
			// One failing undo must not strand the rest. Log it and keep unwinding;
			// in production this is where you would raise something a human sees.
			if cerr := compensations[i](undoCtx); cerr != nil {
				logger.Error("compensation failed", "orderId", order.OrderID, "error", cerr)
			}
		}
		// err is left as it is, so the Workflow still fails with the original
		// failure once the unwind is done.
	}()
```

</details>

# Run it

Go builds your code into the Worker when it starts, so a running Worker still
has the old code. In the [button label="Worker" background="#444CE7"](tab-1)
tab, press `Ctrl+C`, then start it again:

```bash,run
go run ./worker
```

In the [button label="Client" background="#444CE7"](tab-2) tab, run the same
order again:

```bash,run
go run ./starter -fail-dispatch
```

Four ledger lines now:

```bash,nocopy
[ledger] reserve:ORD-XXXXXX
[ledger] charge:ORD-XXXXXX
[ledger] refund:ORD-XXXXXX
[ledger] release:ORD-XXXXXX
```

The refund runs before the stock is released, newest undo first.

Open the new Workflow in the
[button label="Temporal UI" background="#444CE7"](tab-3) tab. The history now
carries `RefundPaymentIfCharged` and then `ReleaseInventoryIfReserved` after the
failure, so the order they ran in is recorded, not just logged. The Workflow
still ends **Failed**, which is correct.

# Checkpoint

Click **Check**. It runs the tests for this Challenge against their own dev
server: the undos run newest first, a clean order runs none, one failing undo
does not stop the rest, and a cancelled order still unwinds. If one fails, the
message names the file, the function and the TODO.

---

**Please share your feedback so we can make better content for you.** The
**Feedback** tab takes a few seconds, and it is how we find out which parts of
this landed.
