---
slug: the-compliance-side
id: oz2zruycdpxf
type: challenge
title: 3. The compliance side
teaser: Implement the fast handler, register the Service on a Worker, and read the
  Endpoint.
notes:
- type: text
  contents: |-
    # A compliance check can wait on a human. How do you put that behind an RPC?

    Some checks clear in milliseconds. A $12,000 international transfer waits
    for an officer to click approve, and that can take an afternoon.

    A synchronous handler has ten seconds.
- type: text
  contents: |-
    # This is the hardest challenge in the lab

    A handler in a different shape from its worked example, the Worker
    registration, and the Endpoint that routes to it.

    The finished code sits in solution/ in the same file tree. Use it if you
    stall.
tabs:
- id: n4f1lu0nzpth
  title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/nexus-decouple-monolith
  port: 8080
- id: u9lotcjf5j52
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
- id: wstgfr5unkex
  title: Terminal
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
- id: yupgbtejhnea
  title: Payments Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
- id: srras0cec857
  title: Compliance Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
difficulty: intermediate
timelimit: 1800
enhanced_loading: null
---

# Everything Compliance owns

Three pieces, all on this team's side of the boundary: the handlers that answer Nexus
calls, the Worker that runs them, and the Endpoint that tells Temporal where to route.

# Read the worked example first

Click the [button label="Exercise" background="#444CE7"](tab-0) tab and open
`exercise/compliance/handler.go`. `CheckComplianceOperation` is written for you, and the
Operation you are about to write is the other half of the same pattern.

| Operation | Shape | Why |
|---|---|---|
| `checkCompliance` *(given)* | `temporalnexus.NewWorkflowRunOperation(...)` | Slow. Starts a Workflow and returns a token for it. |
| `submitReview` *(yours)* | `nexus.NewSyncOperation(...)` | Fast. Answers inside the call, in milliseconds. |

Written as a sync Operation, `checkCompliance` would be cut off at the ten-second handler
deadline, and every retry would start a second compliance check for the same payment.

# Write the other handler

Follow TODO 2 in the same file. The signature is given; the body is yours.

<!-- snippet: exercise/compliance/handler.go TODO 2 before=2 -->
Lines 51-80 are `SubmitReviewOperation`:

```go
	shared.SubmitReviewOperation.Name(),
	func(ctx context.Context, input shared.ReviewRequest, options nexus.StartOperationOptions) (shared.ComplianceResult, error) {
		// TODO 2 start
		// Implement submitReview, the other half of the pattern.
		//
		// This is the FAST call, the opposite shape to the worked example
		// above. When it arrives, the compliance check is already running and
		// parked waiting for a decision. All this Operation does is find that
		// Workflow and hand it the yes or no, which takes milliseconds, so it
		// answers inside the call. That is why it is a nexus.NewSyncOperation
		// and not a temporalnexus.NewWorkflowRunOperation: slow work starts a
		// Workflow and returns a token, fast work answers in the call. Same
		// Service, two Operations, two shapes.
		//
		// Replace the return below with:
		//   1. a Temporal Client: temporalnexus.GetClient(ctx), already
		//      connected to this team's Namespace
		//   2. c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{...}) with
		//        WorkflowID:   workflowIDFor(input.TransactionID)  (the SAME ID
		//                      the worked example starts the Workflow with)
		//        UpdateName:   ReviewUpdate
		//        Args:         []any{input.Approved, input.Explanation}
		//        WaitForStage: client.WorkflowUpdateStageCompleted
		//   3. handle.Get(ctx, &result) on what it returns, and return result
		//
		// Stuck? "Ask AI" on https://docs.temporal.io:
		//   "how do I send a Workflow Update from a Nexus Operation handler in Go?"
		return shared.ComplianceResult{}, nexus.NewHandlerErrorf(nexus.HandlerErrorTypeNotImplemented,
			"TODO 2: submitReview is not implemented yet (compliance/handler.go)")
		// TODO 2 end
```

# Register the Service

Open `exercise/compliance/worker.go` and follow TODO 3. The Worker already knows its
Workflow and its Activity. One registration is missing, the one that makes this team
callable by Payments at all.

<!-- snippet: exercise/compliance/worker.go TODO 3 before=5 -->
Lines 13-33 are `NewWorker`:

```go
func NewWorker(c client.Client) (worker.Worker, error) {
	w := worker.New(c, shared.ComplianceTaskQueue, worker.Options{})
	w.RegisterWorkflow(ComplianceWorkflow)
	w.RegisterActivity(CheckCompliance)

	// TODO 3 start
	// A Worker only handles work it has been told about. This one already
	// knows its Workflow and its Activity. What it does not know yet is the
	// Nexus Service, the thing that makes this team callable by Payments at
	// all.
	//
	// Build the Service with NewNexusService() (compliance/handler.go),
	// return the error if there is one, and register it on w with
	// w.RegisterNexusService(...).
	//
	// Leave it out and the Worker still starts. Payments' calls just go
	// unanswered. That is the failure you are guarding against.
	//
	// Stuck? "Ask AI" on https://docs.temporal.io:
	//   "what do I register on a Go Worker that handles Nexus Operations?"
	// TODO 3 end
```

Click the [button label="Terminal" background="#444CE7"](tab-2) tab and make sure it
compiles:

```bash,run
go vet ./...
```

Silence means it passed. The Payments side still calls the Activity, which is valid. You
change that in challenge 4.

# Read the Endpoint

The Endpoint is the routing rule: a name, a target Namespace, and a target Task Queue. It
is the one piece of Nexus that lives outside your code, so this lab created it when it
started, with:

```bash,nocopy
temporal operator nexus endpoint create --name compliance-endpoint --target-namespace compliance-namespace --target-task-queue compliance-risk
```

Look it up in the Registry:

```bash,run
temporal operator nexus endpoint list
```

```bash,nocopy
  Name                     compliance-endpoint
  Target.Worker.Namespace  compliance-namespace
  Target.Worker.TaskQueue  compliance-risk
```

`Target.Worker.TaskQueue` must match `ComplianceTaskQueue` in
`exercise/shared/types.go` exactly. Point it at a queue no Worker polls and the calls are
never answered.

# Start the Worker

Click the [button label="Compliance Worker" background="#444CE7"](tab-4) tab:

```bash,run
go run ./compliance-worker
```

The startup banner prints whether or not you registered anything, so it proves nothing.
The Task Queue does. Click the [button label="Terminal" background="#444CE7"](tab-2) tab
and ask who is polling it for Nexus work:

```bash,run
temporal task-queue describe --namespace compliance-namespace --task-queue compliance-risk --task-queue-type nexus
```

Under **Pollers** your Worker's identity appears with a `LastAccessTime` of `now`. A Worker
without the Service registered starts happily and never polls for Nexus tasks at all.

Then click the [button label="Temporal UI" background="#444CE7"](tab-1) tab, switch the
Namespace selector to `compliance-namespace`, and open **Workers**. Yours is on
`compliance-risk`. In challenge 1 no Worker of yours was here at all.

| What you see | What it means |
|---|---|
| `go vet` names `handler.go` | TODO 2 does not compile yet. |
| No Pollers for the `nexus` type | You skipped `w.RegisterNexusService` in TODO 3. |
| `tried to register an operation with no name` | TODO 1 is unfinished. |
| Worker starts, calls never answered | The Task Queue does not match the Endpoint's target. |

Click **Check**. It runs this Worker against its own Temporal dev server and calls both
Operations through an Endpoint, so it takes about thirty seconds. Challenge 4 starts both
Workers fresh, so you can stop this one with **Ctrl+C** afterwards.

# What you know now

- Slow work starts a Workflow and returns. Fast work answers in the call.
- A Workflow-run Operation with a business ID makes a retry re-attach instead of
  starting a duplicate.
- A Worker starts fine with nothing registered, so a clean banner proves nothing.
- The Endpoint maps a name to a Namespace and a Task Queue.
- Compliance now runs in its own Namespace, with its own Worker.

---

Tell us what worked in the **Feedback** tab. It takes a few seconds.
