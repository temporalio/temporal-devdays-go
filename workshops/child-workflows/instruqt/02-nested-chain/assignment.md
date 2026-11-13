---
slug: nested-chain
type: challenge
title: 2. A Deeply Nested Chain
teaser: Children can start children. Go five levels deep, fail the bottom one, and
  find out why from the top.
notes:
- type: text
  contents: |-
    # The bottom of a five-level chain fails. What does the top see?

    A child can start its own children, as deep as you like. Every level is a
    full Workflow Execution with its own history.

    When the bottom one fails, every parent above it fails too. The real reason
    is five clicks away.
tabs:
- title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/child-workflows
  port: 8080
- title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/child-workflows/exercise
- title: Client
  type: terminal
  hostname: workshop
  workdir: /root/workshop/child-workflows/exercise
- title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: basic
timelimit: 300
---

# Goal

Run a chain of Workflows five levels deep where the bottom one fails, and
follow the failure from the top to its real cause.

**Why it matters:** every level you nest is one more hop between an alert on
the top Workflow and the error that caused it.

# Read the code

Nothing to change in this Challenge. Open `exercise/workflows/chain.go` in the
[button label="Exercise" background="#444CE7"](tab-0) tab. `ChainWorkflow`
starts **itself** as a child, one level shallower each time, until it reaches
depth 0:

```go,nocopy
	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s/%d", workflow.GetInfo(ctx).WorkflowExecution.ID, request.Depth-1),
	})
	var below int
	next := ChainRequest{Depth: request.Depth - 1, FailAtBottom: request.FailAtBottom}
	if err := workflow.ExecuteChildWorkflow(ctx, ChainWorkflow, next).Get(ctx, &below); err != nil {
		return 0, err
	}
	return below + 1, nil
```

Each child's ID is its parent's ID plus one segment, so the ID shows how deep
it is. At depth 0, with `--fail`, it returns
`temporal.NewApplicationError(..., "BottomOfChain")`. Every level above passes
its child's error straight up with `return 0, err`.

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
go run ./starter chain --fail
```

```bash,nocopy
Chain chain-6F682D failed. The failure, one wrapper per level:

  ChildWorkflowExecutionError  from chain-6F682D/4
  ChildWorkflowExecutionError  from chain-6F682D/4/3
  ChildWorkflowExecutionError  from chain-6F682D/4/3/2
  ChildWorkflowExecutionError  from chain-6F682D/4/3/2/1
  ChildWorkflowExecutionError  from chain-6F682D/4/3/2/1/0
  ApplicationError: BottomOfChain: the bottom of the chain failed
```

One failure, wrapped five times. A parent that returns its child's error fails
with it, and the error it fails with is a `*temporal.ChildWorkflowExecutionError`
around the child's own. The starter walks down the chain with `Unwrap()` until
it reaches the `*temporal.ApplicationError` at the bottom. `errors.As` finds
that in one call, when all you want is the cause.

# Follow it in the UI

Open the [button label="Temporal UI" background="#444CE7"](tab-3) tab. Six
Workflows failed, one per level. If they aren't listed, click the refresh
button.

Click the top one, `chain-...` with no `/` in its ID. Its history ends with
**Child Workflow Execution Failed**. The reason it shows is a child failure,
not the cause. Click through to the child, and its child, until you reach
`.../0`, where the `ApplicationError` was returned.

That is the cost of depth: five hops to read one error. **Keep hierarchies
shallow.** Use a child when the sub-process needs its own history, not to make
the code look tidier.

# Checkpoint

Click **Check**. It passes once a five-level chain has failed in the Temporal
dev server. If it fails, the message says what to run.

---

**Please share your feedback so we can make better content for you.** The
**Feedback** tab takes a few seconds, and it is the only way we find out which
parts of this landed.
