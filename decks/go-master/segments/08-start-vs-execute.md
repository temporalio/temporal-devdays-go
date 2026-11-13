---
layout: section
---

<ZiggyHero scene="launch">

# Starting a Workflow

One call starts it. Wait for it now, or keep the run and come back.

</ZiggyHero>

<!--
Source: PPT slides 55-58.

The original slides showed two calls for the two ways to launch a Workflow. Go
has one: `client.ExecuteWorkflow` starts the Workflow and returns a
`client.WorkflowRun` straight away. The two ways are what you do with that run:
call `Get` on it at once, or keep it and reach the Workflow while it runs.

Short segment, and the distinction matters the moment anyone wants to send a
Signal.
-->

---
layout: code-walkthrough
heading: "Get() now: fire and wait"
notes:
  - "<code>ExecuteWorkflow</code> starts the Workflow and returns as soon as the start is recorded"
  - "<code>ID</code> is yours to choose: a business ID, unique while the Workflow runs"
  - The Task Queue the Worker polls. A typo here and the Workflow waits forever
  - "<code>run.Get</code> blocks until the Workflow completes. Nothing below it runs until then"
---

::code::

```go {all|1|2|3|10-13}
run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
	ID:        "order-" + order.OrderID,
	TaskQueue: "order-task-queue",
}, OrderWorkflow, order)
if err != nil {
	return err
}

var result OrderResult
if err := run.Get(ctx, &result); err != nil {
	return err
}
log.Printf("Result: %s", result.Status)
```

<!--
Source: PPT slide 56.
Code: snippets/startexecute/client.go, run by
snippets/startexecute/startexecute_test.go.

Use this when you want to submit a Workflow and wait for the final answer, and
nothing else.

- **Build 1 - `ExecuteWorkflow`.** The name says execute; the behaviour is
  start. It returns once Temporal has recorded the Workflow, in milliseconds,
  whatever the Workflow does next. The test checks that a Workflow is still
  Running when the call has already returned.
- **Build 2 - the Workflow ID.** Make it meaningful, such as the order ID.
  Leave it empty and the SDK makes up a UUID, and you lose the ability to find
  the Workflow by business key. A second `ExecuteWorkflow` with the ID of a
  Workflow that is still running does not start another one: by default it
  hands back the run already going, so a retried HTTP request cannot start the
  same order twice. Set `WorkflowExecutionErrorWhenAlreadyStarted` to get a
  `*serviceerror.WorkflowExecutionAlreadyStarted` error instead. The test checks
  both.
- **Build 3 - the Task Queue.** Must match the Worker's, character for
  character.
- **Build 4 - the wait.** `run.Get` blocks until the Workflow finishes and
  decodes the result. Called straight after the start, this is "fire and
  wait": simple, and while it waits this code has nothing else to do with the
  Workflow.

Worth pre-empting: the client waiting has nothing to do with the Workflow's
durability. Kill the client here and the Workflow carries on. You have only lost
your ability to read the result from *this* process.
-->

---
layout: code-walkthrough
heading: "Get() later: fire, and keep the run"
notes:
  - Returns as soon as the Workflow is <strong>started</strong>, not when it finishes
  - "The run's ID is how you reach a running Workflow. <code>QueryWorkflow</code> reads its state"
  - "<code>SignalWorkflow</code> pushes data in: the human in the loop"
  - "<code>run.Get</code> waits for the result whenever you want it, from any process"
---

::code::

```go {all|1-4|9-17|19|24-28}
run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
	ID:        "approval-" + order.OrderID,
	TaskQueue: "order-task-queue",
}, ApprovalWorkflow, order)
if err != nil {
	return err
}

value, err := c.QueryWorkflow(ctx, run.GetID(), run.GetRunID(), GetStatusQuery)
if err != nil {
	return err
}
var status string
if err := value.Get(&status); err != nil {
	return err
}
log.Printf("Status: %s", status)

err = c.SignalWorkflow(ctx, run.GetID(), run.GetRunID(), ApproveSignal, "alice")
if err != nil {
	return err
}

var result OrderResult
if err := run.Get(ctx, &result); err != nil {
	return err
}
log.Printf("Result: %s", result.Status)
```

<!--
Source: PPT slides 57-58, merged: the original repeated the same slide twice.
Code: snippets/startexecute/client.go, run by
snippets/startexecute/startexecute_test.go, which checks it prints
`Status: awaiting approval` and then `Result: approved by alice`.

Use this when you need to interact with the Workflow while it runs: the
human-in-the-loop case.

- **Build 1 - the start.** The same call as the previous slide. What changes is
  what you do next: keep `run`, and do not call `Get` yet.
- **Build 2 - Query.** `QueryWorkflow` reads the current state without changing
  it, by Workflow ID and Run ID. The Queries segment shows the Workflow side.
- **Build 3 - Signal.** `SignalWorkflow` pushes data into the running
  Workflow. The Signals segment shows the Workflow side.
- **Build 4 - `Get`.** Waits until completion. Call it at any point, or never.

The fact that usually lands hardest: you can rebuild the run from the Workflow
ID alone, in a different process, next week: `c.GetWorkflow(ctx, workflowID,
"")`, and Query, Signal or `Get` through it. An empty Run ID means "the latest
run". The run is not a socket or a session. It is a name.
-->
