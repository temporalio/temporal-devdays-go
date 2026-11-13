---
layout: section
---

<ZiggyHero scene="children">

# Child Workflows

Decomposing orchestration without losing the guarantees.

</ZiggyHero>

<!--
Source: PPT slides 99-107.
-->

---
layout: default
---

# When one Workflow isn't enough

- Some processes are too complex, too long, or too parallel for a single Workflow
- Splitting logic into functions or Activities only goes so far
- What if the sub-process itself needs durability, retries, and its own history?
- **Enter Child Workflows.**

<!--
Source: PPT slide 100.

Frame this as a natural scaling question. Developers hit it when a Workflow
starts doing too much: the history gets long, the logic gets tangled, or they
need the same sub-process reused across several parents. Child Workflows answer
"how do I decompose orchestration without losing Temporal's guarantees?"
-->

---
layout: default
---

# What is a Child Workflow?

- A Workflow launched by another Workflow
- Has its own execution, its own history, its own Retry Policy
- A first-class Temporal citizen. Not a function call, not an Activity
- Can be monitored, cancelled, and queried independently

<!--
Source: PPT slide 101.

The distinction to hammer home: a Child Workflow is not syntactic sugar. It is a
full Workflow Execution that happens to be started by a parent. That
independence is what makes it powerful, and what separates it from calling a
function.

In Go there is no child marker at all. A Child Workflow is an ordinary
registered Workflow function; whether it is a child is decided entirely by who
starts it. The same Workflow can be a child here and started from a Client
somewhere else.
-->

---
layout: default
---

# Child Workflow vs Activity: the decision

- **Use an Activity when** you're interacting with external systems, doing
  I/O, or calling an API
- **Use a Child Workflow when** you're orchestrating a multi-step sub-process,
  need independent retry logic, or are reusing complex logic across Workflows

<div v-click class="mt-8 text-2xl font-light">

Rule of thumb: if it needs its own history, it's a Child Workflow.

</div>

<!--
Source: PPT slide 102.

The slide developers most need. The instinct is to reach for Activities for
everything. The test: does this thing need to survive independently? Does it
have enough internal complexity to deserve its own audit trail? If yes, it is a
Child Workflow.

Other signals that point to a child: the work needs its own Task Queue and its
own Workers, or you want to see it as its own execution in the UI. And the
default worth stating plainly: if it is one unit of work, use an Activity. A
Child Workflow is heavier, a full execution with its own history.
-->

---
layout: default
---

# Parent-child lifecycle

What happens to the child when the parent closes is its `ParentClosePolicy`.

<div class="grid grid-cols-3 gap-4 mt-6">
  <ParentClosePolicy v-click policy="TERMINATE" />
  <ParentClosePolicy v-click policy="ABANDON" />
  <ParentClosePolicy v-click policy="REQUEST_CANCEL" />
</div>

<!--
Source: PPT slide 103. The diagrams are new to this Deck.

The parent starts the child, and the child gets its own Workflow ID. Each click
reveals one policy: the same parent, the same child, the same moment the parent
closes, and only the child's lane changes.

- **Click 1 - `TERMINATE`, the default.** The child is cut off when the parent
  closes. No cleanup code runs.
- **Click 2 - `ABANDON`.** The child keeps running on its own. ABANDON is the
  one people misunderstand: the child keeps running even though the parent is
  gone. That is intentional, and powerful for fire-and-forget sub-processes.
- **Click 3 - `REQUEST_CANCEL`.** The child is sent a cancellation request and
  can clean up gracefully, the same cancellation the Saga section uses to run
  compensations, before it ends as cancelled.

The policy applies whenever the parent closes, whether it completed, failed,
timed out or was terminated. Not only on failure. In code these are
`enumspb.PARENT_CLOSE_POLICY_TERMINATE`, `PARENT_CLOSE_POLICY_ABANDON` and
`PARENT_CLOSE_POLICY_REQUEST_CANCEL`, from `go.temporal.io/api/enums/v1`, set on
`workflow.ChildWorkflowOptions`.
-->

---
layout: code-walkthrough
heading: Spawning a child, waiting for the result
notes:
  - "<code>WithChildOptions</code>: the child's options travel on the context the call uses"
  - The child's Workflow ID. Plan this scheme; IDs must be unique
  - What happens to the child if this parent closes first
  - Pass the child Workflow function itself, the same way the Client does, and its input. Serializable, like every Workflow argument
  - "<code>Get</code> blocks until the child completes, and decodes its result"
---

::code::

```go {all|2|3|4|8|8}
func OrderWorkflow(ctx workflow.Context, order Order) (string, error) {
	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:        "shipment-" + order.OrderID,
		ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
	})

	var shipment Shipment
	err := workflow.ExecuteChildWorkflow(ctx, ShipmentWorkflow, order.Shipment).Get(ctx, &shipment)
	if err != nil {
		return "", err
	}
	return shipment.TrackingNumber, nil
}
```

<!--
Source: PPT slide 104.
Code: snippets/child/child.go, run by snippets/child/child_test.go.

- **Build 1 - the options.** Go passes Workflow options on the
  `workflow.Context`: `WithChildOptions` returns a context that every child
  started from it uses. The same shape sets Activity Options.
- **Build 2 - the ID.** `"shipment-" + order.OrderID` is deterministic and
  meaningful: support can find the child by order number. A random ID would
  have to come from `workflow.SideEffect` to replay safely, and you would lose
  the ability to find it. Leave `WorkflowID` empty and the SDK generates one.
- **Build 3 - the policy.** ABANDON here, so a parent that fails after
  dispatching does not kill a shipment that is already moving. Leave it unset
  and you get TERMINATE.
- **Build 4 - the reference and the input.** `ShipmentWorkflow`, the function,
  not a string. Its name becomes the Workflow Type. The arguments after it are
  passed to the child, the same convention as `client.ExecuteWorkflow`; they
  are checked against the function's parameters when the call is made, and a
  mismatch comes back as the error from `Get`, not from the compiler. The
  child must be
  registered on a Worker like any other Workflow.
- **Build 5 - the result.** `ExecuteChildWorkflow` returns a
  `workflow.ChildWorkflowFuture` at once; `Get` blocks until the child
  completes and decodes its result into `shipment`.

Waiting on `GetChildWorkflowExecution().Get(ctx, nil)` instead returns as soon
as the child has started; you can call `Get` for the result later, Signal the
child, or leave it alone. That is the fan-out shape on the next slide.
-->

---
layout: default
---

# Practical use cases

- **Fan-out / fan-in.** Spawn N children in parallel, collect the results
- **Per-entity Workflows.** One child per user, order, or customer
- **Reusable sub-processes.** The same Child Workflow called from several Parent Workflows
- **Isolating long-running steps.** Keep the parent's history lean

<!--
Source: PPT slide 105.

Fan-out/fan-in gets the most excitement. The shape in Go is a loop that calls
`workflow.ExecuteChildWorkflow` for each item and keeps the futures in a slice,
then a second loop that calls `Get` on each one. Every child is already
running before the first `Get` blocks. Fifty children processing a batch each,
then a parent that aggregates. A compelling demo if
you have one ready.
-->

---
layout: default
---

# Gotchas

- Deeply nested child chains are hard to debug. Keep hierarchies shallow
- Each child has its own history limit. Continue-As-New applies to children too
- Don't use children just to avoid writing clean Activity code
- Workflow IDs must be unique. Plan your ID scheme carefully

<!--
Source: PPT slide 106.

The last one bites in production. If you spawn one child per entity, the ID
scheme (`"order-" + orderID + "-fulfillment"`) is part of your system design, not an
afterthought. A collision fails the start, and it happens at 3am.

Tease Continue-As-New here. It is next.
-->

---
layout: default
---

# What you got

- Composable orchestration without sacrificing durability
- Sub-processes with their own history, retries, and lifecycle
- Independent monitoring and cancellation per child
- The building block for fan-out, per-entity, and reusable Workflow patterns

<!--
Source: PPT slide 107.

Close on composability. Temporal gives you the primitives; Child Workflows are
how you build architecture out of them.
-->
