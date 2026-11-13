---
slug: run-a-child
id: ems2rdrgqpds
type: challenge
title: 1. Run a Child
teaser: One Workflow starts another and waits for its result. Find both in the Temporal
  UI.
notes:
- type: text
  contents: |-
    # What do you get when a Workflow starts a Workflow?

    Not a function call, and not an Activity. A second Workflow Execution, with
    its own Workflow ID, its own Event History and its own retries. The parent
    can wait for its result, or start it and move on.

    **Author:** [Nikolay Advolodkin](https://www.linkedin.com/in/nikolayadvolodkin/), Staff Developer Advocate
- type: text
  contents: |-
    # Three words for this lab

    **Child Workflow.** A Workflow started by another Workflow, the parent.

    **Fan out.** Start many children at once, without waiting for any of them.

    **Fan in.** Wait for all of them and collect their results.
tabs:
- id: g0euryo6sdzf
  title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/child-workflows
  port: 8080
- id: gnmbhwyshsuk
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/child-workflows/exercise
- id: zutejjesizui
  title: Client
  type: terminal
  hostname: workshop
  workdir: /root/workshop/child-workflows/exercise
- id: o0yd8rertvcx
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: basic
timelimit: 300
enhanced_loading: null
---

# Goal

Run an order whose Workflow starts a shipment Workflow as its child, and find
the two separate histories in the Temporal UI.

**Why it matters:** a child is a full Workflow Execution. It can be found,
retried and read on its own, apart from the order that started it.

# Your tabs

- [button label="Exercise" background="#444CE7"](tab-0) VS Code in the browser, open on this Workshop. You edit
  the code under `exercise/`. It saves as you type.
- [button label="Worker" background="#444CE7"](tab-1) a terminal in `exercise/`. Run `go run ./worker` here. A
  running Worker keeps the code it was built from: after you change code, stop
  it with `Ctrl+C` and run it again.
- [button label="Client" background="#444CE7"](tab-2) a terminal in `exercise/`. Run `go run ./starter` here to start
  a Workflow.
- [button label="Temporal UI" background="#444CE7"](tab-3) the Temporal UI of the local dev server:
  every order Workflow and the child Workflows it started, each with its own Event History.

# Read the code

Nothing to change in this Challenge. Open `exercise/workflows/order.go` in the
[button label="Exercise" background="#444CE7"](tab-0) tab to read the parent.
Press `Ctrl+P` (`Cmd+P` on a Mac) and type the file name. Here is the body of
`OrderWorkflow`:

```go,nocopy
	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: "shipment-" + order.OrderID,
	})

	var tracking string
	err := workflow.ExecuteChildWorkflow(ctx, ShipmentWorkflow, order.OrderID).Get(ctx, &tracking)
	return tracking, err
```

`workflow.ChildWorkflowOptions` travels on the Workflow Context, the same way
Activity options do: here it gives the child its own Workflow ID. With no Task
Queue set, the child runs on its parent's.

`workflow.ExecuteChildWorkflow` starts `ShipmentWorkflow` as a child and returns
a Future straight away. `Get` blocks until the child returns, and decodes its
result into `tracking`. You pass the child Workflow function itself, the same
way a Client does.

# Run it

In the [button label="Worker" background="#444CE7"](tab-1) tab:

```bash,run
go run ./worker
```

```bash,nocopy
Worker polling child-workflows. Ctrl+C to stop.
```

Leave it running. Each Challenge opens fresh terminals, so you start it again
at the top of every one.

In the [button label="Client" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter child
```

```bash,nocopy
Parent  order-ORD-FC5B56  completed
Child   shipment-ORD-FC5B56  returned TRACK-ORD-FC5B56
```

# Find both Workflows

Open the [button label="Temporal UI" background="#444CE7"](tab-3) tab. Two
Workflows are listed, not one: `order-ORD-...` and `shipment-ORD-...`. If they
aren't there, click the refresh button: the list does not always update on its
own.

Click `order-ORD-...` and read its Event History:

| Event | What it means |
|---|---|
| **Start Child Workflow Execution Initiated** | the parent asked for a child |
| **Child Workflow Execution Started** | the child exists, with its own ID |
| **Child Workflow Execution Completed** | the child returned `TRACK-...` |

Click the child's link in those events. The shipment has a history of its own,
with the 2 second timer the parent never saw. Its page names the parent.

# Checkpoint

Click **Check**. It passes once a shipment child has completed in the Temporal
dev server. If it fails, the message says what to run.

---

**Please share your feedback so we can make better content for you.** The
**Feedback** tab takes a few seconds, and it is the only way we find out which
parts of this landed.
