---
slug: register-first
id: ing0dsw2w4la
type: challenge
title: 2. Register before you run
teaser: The gateway captures the money, then times out. The refund never runs.
notes:
- type: text
  contents: |-
    # The charge landed. The call said it failed.

    A payment gateway captures the money, then the connection drops before the
    reply comes back. The Workflow sees an error. The customer sees a charge.

    An error is not proof that nothing happened. If the undo is only registered
    once the step succeeds, this charge is never refunded.
tabs:
- id: iu9vbqph5viq
  title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/saga-order-fulfillment
  port: 8080
- id: tlmrdm8b6wbx
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/saga-order-fulfillment/exercise
- id: y0g4jbjuplvo
  title: Client
  type: terminal
  hostname: workshop
  workdir: /root/workshop/saga-order-fulfillment/exercise
- id: ucs8c8dcmtes
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: intermediate
timelimit: 600
enhanced_loading: null
---

# Goal

Make the order refund a charge that landed and then reported failure, by
registering each undo before the step it undoes.

**Why it matters:** this is the failure that costs money, and the unwind you
wrote in Challenge 1 cannot see it.

# Watch the money go

Each Challenge opens fresh terminals, so start the Worker again. In the
[button label="Worker" background="#444CE7"](tab-1) tab:

```bash,run
go run ./worker
```

In the [button label="Client" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter -charge-times-out
```

This is a payment gateway that captures the money and then times out. The charge
landed. The call reported failure anyway.

```bash,nocopy
[ledger] reserve:ORD-XXXXXX
[ledger] charge:ORD-XXXXXX
[ledger] release:ORD-XXXXXX
```

The stock came back. The refund never ran, so the money is still gone. Open the
Workflow in the [button label="Temporal UI" background="#444CE7"](tab-3) tab.
There is no `RefundPaymentIfCharged` event in the history at all, because the
Workflow never knew it owed one.

# Register before you run (TODO 2a, TODO 2b)

**Observe.** Open `exercise/workflow.go` in the
[button label="Exercise" background="#444CE7"](tab-0) tab.

<!-- snippet: exercise/workflow.go TODO 2a -->
Lines 45-54 reserve the stock, then register its undo:

```go
	// TODO 2a start
	// Move the append ABOVE the ReserveInventory call. A step can land its
	// effect and still report failure, and then nothing knows how to reverse it.
	if err = workflow.ExecuteActivity(ctx, ReserveInventory, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, ReleaseInventoryIfReserved, order).Get(ctx, nil)
	})
	// TODO 2a end
```

<!-- snippet: exercise/workflow.go TODO 2b -->
Lines 56-64 do the same for the charge:

```go
	// TODO 2b start
	// Same move, above the ChargePayment call. This is the one that costs real money.
	if err = workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, RefundPaymentIfCharged, order).Get(ctx, nil)
	})
	// TODO 2b end
```

**Learn.** Each undo is appended on the lines after the step it undoes. When
the step returns an error, the function returns first, and the append never
runs. For a step that failed cleanly that is harmless. For a step that landed
its effect and then failed, nothing will ever reverse it.

Register first and the two cases swap costs. An undo for a step that never
landed runs against nothing, which Challenge 3 makes safe. An undo for a step
that did land is already on the stack. One is a no-op, the other is a stranded
charge.

**Do.**

- [ ] TODO 2a: move the `append` for `ReleaseInventoryIfReserved` above the
  `ReserveInventory` call.
- [ ] TODO 2b: move the `append` for `RefundPaymentIfCharged` above the
  `ChargePayment` call.

<details>
<summary>Show solution</summary>

```go
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, RefundPaymentIfCharged, order).Get(ctx, nil)
	})
	if err = workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, nil); err != nil {
		return OrderResult{}, err
	}
```

TODO 2a is the same move, with `ReleaseInventoryIfReserved` and
`ReserveInventory`.

</details>

# Run it

Restart the Worker so it builds your change: `Ctrl+C` in the
[button label="Worker" background="#444CE7"](tab-1) tab, then:

```bash,run
go run ./worker
```

In the [button label="Client" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter -charge-times-out
```

```bash,nocopy
[ledger] reserve:ORD-XXXXXX
[ledger] charge:ORD-XXXXXX
[ledger] refund:ORD-XXXXXX
[ledger] release:ORD-XXXXXX
```

The same four lines as a failed dispatch. `charge` is there even though
`ChargePayment` failed, because the money moved first. `refund` is there because
the undo was on the stack before the gateway ever saw the request.

# Checkpoint

Click **Check**. It passes when an order whose reservation fails still calls
`ReleaseInventoryIfReserved`, and a charge that lands and then fails is
refunded. If it fails, the message names the TODO still to do.

---

**Please share your feedback so we can make better content for you.** The
**Feedback** tab takes a few seconds, and it is how we find out which parts of
this landed.
