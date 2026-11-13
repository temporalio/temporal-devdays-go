---
slug: fan-out
id: grgwdvtroo8f
type: challenge
title: 3. Fan Out
teaser: Start one child per item, all at once, then close the parent without waiting.
  Watch every child get terminated.
notes:
- type: text
  contents: |-
    # The parent starts five children and returns. What happens to them?

    `workflow.ExecuteChildWorkflow` returns a Future at once. Call `Get` on it
    straight away and you wait for one child. Keep the Futures instead, one per
    item, and five children run at the same time.

    This parent then returns before any child finishes. Watch what its Parent
    Close Policy does to them.
tabs:
- id: ml5ikatmfuqu
  title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/child-workflows
  port: 8080
- id: uqgdfu4csfwi
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/child-workflows/exercise
- id: l7jxeiiwvy1r
  title: Client
  type: terminal
  hostname: workshop
  workdir: /root/workshop/child-workflows/exercise
- id: 2rr54qj3gpee
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

Fan out five children, close the parent without waiting for them, and see
what happens to the children.

**Why it matters:** a parent that returns early does not leave its children
running. By default it takes them down with it, and the work is lost.

# Read the code

<!-- snippet: exercise/workflows/pack.go TODO 1 before=19 -->
**This code is for review only.** It ends with TODO 1: don't do it yet, it is
the next Challenge. Open `exercise/workflows/pack.go` in the
[button label="Exercise" background="#444CE7"](tab-0) tab. Press `Ctrl+P`
(`Cmd+P` on a Mac), type the file name, then `Ctrl+G` to jump to a line.
Read lines 13-37, the fan-out:

```go
	// Fan out: start one child per item, without waiting for any of them to finish.
	// ExecuteChildWorkflow returns a Future straight away; keep one per item.
	boxes := make([]workflow.ChildWorkflowFuture, len(order.Items))
	for i, item := range order.Items {
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("pack-%s-%d", order.OrderID, i),
			// The default, written out: when this parent closes, every child
			// still running is terminated.
			ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_TERMINATE,
		})
		boxes[i] = workflow.ExecuteChildWorkflow(childCtx, PackItemWorkflow, item)
	}
	// Wait for every child to start, not to finish.
	for _, box := range boxes {
		if err := box.GetChildWorkflowExecution().Get(ctx, nil); err != nil {
			return nil, err
		}
	}

	// TODO 1 start
	// Fan in: call Get on every Future in boxes, in item order, and return the
	// boxes they hold. Until you do, the parent returns at once and the Parent
	// Close Policy above terminates every child.
	return []string{}, nil
	// TODO 1 end
```

The loop calls `workflow.ExecuteChildWorkflow` once per item and keeps each
`ChildWorkflowFuture` in `boxes`, without calling `Get`, so all five children
are started back to back. Each child gets its own Context, with its own
Workflow ID and a `ParentClosePolicy`.

A `ChildWorkflowFuture` holds two things: `GetChildWorkflowExecution()`, which
is ready once the child has **started**, and `Get`, which is ready once it has
**finished**. The second loop waits only for the first. Each child sleeps 3
seconds and returns a box. Then the parent reaches TODO 1 and returns an empty
slice.

# Run it

Each Challenge opens fresh terminals, so start the Worker again. In the
[button label="Worker" background="#444CE7"](tab-1) tab:

```bash,run
go run ./worker
```

```bash,nocopy
Worker polling child-workflows. Ctrl+C to stop.
```

In the [button label="Client" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter pack
```

```bash,nocopy
Packed 0 of 5 items in 0.1s: []
```

Zero boxes. Open the newest `pack-ORD-...` Workflow in the
[button label="Temporal UI" background="#444CE7"](tab-3) tab. If it isn't
there, click the refresh button: the list does not always update on its own.
Five **Start Child Workflow Execution Initiated** events, five
**Child Workflow Execution Started**, then the parent **completed**.

Go back to the Workflow list. The five children, `pack-ORD-...-0` to `-4`, are
all **Terminated**.

The parent closed while its children were still running.
`PARENT_CLOSE_POLICY_TERMINATE`, the default, ended every one of them. Starting
children is only half the pattern.

# Checkpoint

Click **Check**. It passes once five children have been terminated this way.
If it fails, the message says what to run.

---

**Please share your feedback so we can make better content for you.** The
**Feedback** tab takes a few seconds, and it is the only way we find out which
parts of this landed.
