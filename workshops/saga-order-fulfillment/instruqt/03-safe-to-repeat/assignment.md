---
slug: safe-to-repeat
type: challenge
title: 3. Make the undo safe to repeat
teaser: The undo now runs before you know the step worked. Make it cope with nothing
  to undo, and with running twice.
notes:
- type: text
  contents: |-
    # The inventory service was down. The undo released stock anyway.

    Registering the undo first means it can run when the step never landed.
    Temporal also retries a compensation like any other Activity, so an undo
    can run twice even when it succeeds.

    An undo that cannot tell whether there is anything to undo does damage of
    its own.
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
timelimit: 600
---

# Goal

Make both undos do nothing when there is nothing to undo, so they are safe to
run when the step never landed, and safe to run twice.

**Why it matters:** a release with nothing reserved adds stock that does not
exist. A refund with nothing charged sends money the wrong way.

# Watch the undo overreach

Start the Worker again. In the
[button label="Worker" background="#444CE7"](tab-1) tab:

```bash,run
go run ./worker
```

In the [button label="Client" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter -fail-reserve
```

The inventory service is down and nothing was reserved. The undo was registered
first, so it runs anyway:

```bash,nocopy
[ledger] release:ORD-XXXXXX
```

That releases stock nobody ever held.

# Guard each undo (TODO 3a, TODO 3b)

**Observe.** Open `exercise/activities.go` in the
[button label="Exercise" background="#444CE7"](tab-0) tab.

<!-- snippet: exercise/activities.go TODO 3a before=1 -->
Lines 51-55 are the top of `ReleaseInventoryIfReserved`:

```go
func ReleaseInventoryIfReserved(ctx context.Context, order Order) error {
	// TODO 3a start
	// Guard this with isReserved(order.OrderID). When nothing is reserved, print
	// "[compensate] nothing reserved for <order ID>, no-op" and return nil.
	// TODO 3a end
```

<!-- snippet: exercise/activities.go TODO 3b before=1 -->
Lines 62-66 are the top of `RefundPaymentIfCharged`:

```go
func RefundPaymentIfCharged(ctx context.Context, order Order) error {
	// TODO 3b start
	// Same guard, with isCharged(order.OrderID). Refunding a charge that never
	// landed, or refunding it twice, sends money the wrong way.
	// TODO 3b end
```

**Learn.** The `If` in each name is a promise the body has to keep. There are
two ordinary reasons an undo runs with nothing to undo:

1. **The effect never landed.** You registered the undo before the step, so a
   step that fails cleanly leaves its undo on the stack anyway.
2. **Temporal retried it.** Compensations are Activities, with the same retry
   policy as everything else. A refund whose reply is lost is called again.

`isReserved` and `isCharged` in `exercise/ledger.go` stand in for the check a
real system makes: reading the reservation's status, or the gateway's own
idempotency key.

**Do.**

- [ ] TODO 3a: in `ReleaseInventoryIfReserved`, return early when
  `isReserved(order.OrderID)` is false.
- [ ] TODO 3b: in `RefundPaymentIfCharged`, return early when
  `isCharged(order.OrderID)` is false.

<details>
<summary>Show solution</summary>

```go
	if !isReserved(order.OrderID) {
		fmt.Printf("[compensate] nothing reserved for %s, no-op\n", order.OrderID)
		return nil
	}
```

TODO 3b is the same guard with `isCharged`, printing `nothing charged`.

</details>

# Run it

Restart the Worker so it builds your change: `Ctrl+C` in the
[button label="Worker" background="#444CE7"](tab-1) tab, then:

```bash,run
go run ./worker
```

In the [button label="Client" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter -fail-reserve
```

The Worker prints a no-op instead of a ledger entry:

```bash,nocopy
[compensate] nothing reserved for ORD-XXXXXX, no-op
```

In the [button label="Temporal UI" background="#444CE7"](tab-3) tab,
`ReleaseInventoryIfReserved` is still in the history and still shows
**Activity Task Completed**. The guard did not stop the call. It changed what
the call did, which is what idempotency means here.

# What you know now

- A distributed transaction has no rollback. You write the undo yourself.
- Compensations run newest first, or you undo into an inconsistent state.
- They run on a disconnected context, so a cancelled order still unwinds.
- Register the undo before the step, because a step can land its effect and
  still report failure.
- So every undo must tolerate running when there is nothing to undo, and running
  more than once.
- One failing undo must not strand the rest, and the Workflow still fails.

# Checkpoint

Click **Check**. It passes when a release with nothing reserved does nothing,
and running each undo twice undoes once. If it fails, the message names the
function and the TODO.

---

**Please share your feedback so we can make better content for you.** The
**Feedback** tab takes a few seconds, and it is how we find out which parts of
this landed.
