---
layout: section
---

<ZiggyHero scene="signal">

# Go Signals

Asynchronous write / fire-and-forget

</ZiggyHero>

<!--
Source: PPT slides 71-79.

The human-in-the-loop primitive. If the room is here for approvals, reviews, or
anything with a person in it, this is the segment they came for.

The original split the Signal into an interface slide and an implementation
slide. In Go a Signal is a name the sender and the Workflow share, plus a
channel the Workflow receives from, so it is one walkthrough.
-->

---
layout: code-walkthrough
heading: Go Signals
notes:
  - "<code>ApproveWireSignal</code> is the name Temporal routes on. Sender and Workflow share the constant"
  - Signals return <strong>nothing</strong>. They deliver data in, they do not read it out
  - Workflow state is plain local variables inside the Workflow function
  - "<code>GetSignalChannel</code> gives the Workflow a channel for that name"
  - "<code>Receive</code> blocks, durably, until a Signal arrives, and decodes its argument into the state"
---

::code::

```go {all|1|1|4|6|7}
const ApproveWireSignal = "approveWire"

func WireTransferWorkflow(ctx workflow.Context, wire WireTransfer) (string, error) {
	var approver string

	approvals := workflow.GetSignalChannel(ctx, ApproveWireSignal)
	approvals.Receive(ctx, &approver)

	return fmt.Sprintf("%s approved by %s", wire.WireID, approver), nil
}
```

<!--
Source: PPT slides 72-79.
Code: snippets/signals/signals.go, run by snippets/signals/signals_test.go.

The scenario: a wire transfer over the bank's review threshold cannot leave
until a compliance officer signs off. The Workflow holds the wire; the officer's
review tool sends `approveWire` with their name.

- **Build 1 - the name.** An exported constant, so the code that sends the
  Signal uses the same string. `"approveWire"` is what Temporal routes on. The
  argument is whatever the sender passes: it is serialized by the Data
  Converter and recorded in Event History as a `WorkflowExecutionSignaled`
  Event; on replay, `Receive` hands the Workflow the same value again.
- **Build 2 - no return value.** Builds 1 and 2 point at the same line on
  purpose: a Signal is a name and an argument, nothing more. It is
  fire-and-forget: the sender gets an acknowledgement that it was recorded, not
  a result. If you need a result, that is an Update, two segments from now.
- **Build 3 - state.** A plain local variable. Temporal rebuilds it by replay,
  so it survives a Worker restart without you storing anything.
- **Build 4 - `GetSignalChannel`.** A `workflow.ReceiveChannel` for one Signal
  name. Signals that arrive before anything receives are buffered on the
  channel, so nothing is lost either way. Signals still unread when the Workflow
  returns are dropped, and the SDK logs a warning that names them.
- **Build 5 - `Receive`.** The Go equivalent of the original's
  `Workflow.await { approved }`. The Workflow is blocked here: no CPU, no Worker
  slot, and it can be evicted from the Worker's cache and rebuilt by replay
  later. It can wait over a weekend for compliance to come back. If the Worker
  restarts, it replays to this line and waits again.

For a Workflow that must keep doing other work while it waits, receive in a
`workflow.Go` coroutine, or put the channel in a `workflow.Selector` next to a
Timer. If that coroutine runs Activities or Timers before it changes state,
have the Workflow wait on `workflow.Await(ctx, func() bool { return
workflow.AllHandlersFinished(ctx) })`, or on the state it sets, before it
returns. Receive-and-block is the simplest shape and the one to teach first.

Worth saying: the Server accepts a Signal even if no Worker is running. It waits
in Event History until a Worker picks the Workflow up.

The demo that sells this, if you have a terminal:
`temporal workflow signal --workflow-id <id> --name approveWire --input '"alice"'`.
-->
