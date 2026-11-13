---
slug: the-payments-side
id: xlesumsiyh3h
type: challenge
title: 4. The payments side
teaser: Swap the Activity call for a Nexus client, point it at the Endpoint, and delete
  the coupling.
notes:
- type: text
  contents: |-
    # How much code changes when you cross a team boundary?

    The compliance check is about to leave this process entirely. Different
    Namespace, different Task Queue, different deployment, different team.

    Count the lines that change at the call site.
- type: text
  contents: |-
    # One deletion is the whole point

    Somewhere in the Payments Worker, one line registers Compliance code inside
    the Payments process. Deleting it is the decoupling. Everything else is
    wiring.
tabs:
- id: gcyddxdtd41a
  title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/nexus-decouple-monolith
  port: 8080
- id: 7u0gncl8hkx4
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
- id: uihtj0ned5sx
  title: Terminal
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
- id: qv5qw2mpu8gq
  title: Payments Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
- id: xe5evcujussd
  title: Compliance Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
difficulty: intermediate
timelimit: 1800
enhanced_loading: null
---

# Swap the Activity for a Nexus client

Click the [button label="Exercise" background="#444CE7"](tab-0) tab, open
`exercise/payments/workflows.go`, and follow TODO 4a, 4b and 4c. Each one sits on the
code it changes: the import to delete, the call to replace, and the review caller to
write.

<!-- snippet: exercise/payments/workflows.go TODO 4a -->
Lines 18-25 are the import that couples the two teams:

```go
	// TODO 4a start

	// Delete this import. This is the coupling: the Workflow below calls the
	// Compliance team's Activity, and Activities run on THIS Worker, so the
	// Compliance team's code executes inside the Payments process. TODO 4b
	// replaces the one place that uses it.
	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/compliance"
	// TODO 4a end
```

<!-- snippet: exercise/payments/workflows.go TODO 4b -->
Lines 65-81 are step 2, the compliance check:

```go
	// TODO 4b start
	// Replace this Activity call with a Nexus call:
	//
	//	complianceClient := workflow.NewNexusClient(shared.ComplianceEndpoint, shared.ComplianceServiceName)
	//	err := complianceClient.ExecuteOperation(ctx, shared.CheckComplianceOperation, compReq,
	//		workflow.NexusOperationOptions{ScheduleToCloseTimeout: 10 * time.Minute},
	//	).Get(ctx, &result)
	//
	// Ten minutes covers the whole call including retries, which is what lets
	// it outlive the Compliance Worker going away. You prove that in
	// Challenge 5.
	//
	// Note what the Workflow names: an Endpoint, a Service and an Operation.
	// No Namespace, no Task Queue, no address. The Registry resolves those,
	// so this Workflow does not change when Compliance moves.
	err := workflow.ExecuteActivity(ctx, compliance.CheckCompliance, compReq).Get(ctx, &result)
	// TODO 4b end
```

<!-- snippet: exercise/payments/workflows.go TODO 4c -->
Lines 122-131 are the body of `ReviewCallerWorkflow`:

```go
	// TODO 4c start
	// Same client as 4b, but call shared.SubmitReviewOperation with the
	// request, a ScheduleToCloseTimeout of 10 * time.Second, and read the
	// answer into result.
	//
	// Seconds, not minutes: submitReview is a SYNC Operation, so the handler
	// has to answer inside the Nexus handler deadline.
	err := temporal.NewNonRetryableApplicationError(
		"TODO 4c: send the review through the Nexus Endpoint (payments/workflows.go)", "NotImplemented", nil)
	// TODO 4c end
```

Same Operation name, same input, same result type. Underneath, the compliance check
leaves this process entirely.

Two Operations, two timeouts. `checkCompliance` is asynchronous, so its budget covers the
whole call including retries, which is what lets it outlive the Compliance Worker going
away in challenge 5. `submitReview` is synchronous, so the handler answers inside the
Nexus handler deadline.

The Workflow names an Endpoint and a contract, never a Namespace, a Task Queue, a
hostname, or a port. Move Compliance tomorrow, update the Endpoint, and this Workflow does
not change. The name is one constant, `ComplianceEndpoint` in
`exercise/shared/types.go`.

# Delete the coupling

Open `exercise/payments/worker.go` and follow TODO 5b, then TODO 5a. Once they are gone
the Payments Worker has no way to run the Compliance team's code, even by accident:
it is not even compiled into the binary.

<!-- snippet: exercise/payments/worker.go TODO 5b -->
Lines 28-36 are the end of `NewWorker`:

```go
	// TODO 5b start
	// Once the Workflow calls Compliance over Nexus, this Worker has no reason
	// to run the Compliance team's code. Delete the registration below, then
	// the import (TODO 5a).
	//
	// That deletion is the decoupling. Everything before it was wiring; this
	// is the line where the two teams stop sharing a process.
	w.RegisterActivity(compliance.CheckCompliance)
	// TODO 5b end
```

<!-- snippet: exercise/payments/worker.go TODO 5a -->
Lines 12-17 are the import it needs:

```go
	// TODO 5a start

	// Delete this import once TODO 5b is done. It compiles the Compliance
	// team's code into the Payments binary.
	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/compliance"
	// TODO 5a end
```

# Run it decoupled

Two Workers now, one per team. Start Compliance first.

Click the [button label="Compliance Worker" background="#444CE7"](tab-4) tab:

```bash,run
go run ./compliance-worker
```

Click the [button label="Payments Worker" background="#444CE7"](tab-3) tab:

```bash,run
go run ./payments-worker
```

Click the [button label="Terminal" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter
```

# Read the result carefully

```bash,nocopy
  TXN-A   Result: COMPLETED             Risk: LOW
  TXN-B   Result: STILL RUNNING         MEDIUM risk, waiting on a human
  TXN-C   Result: DECLINED_COMPLIANCE   Risk: HIGH
```

TXN-B changed. In the monolith a plain Activity returned a MEDIUM verdict and moved on.
Now the check runs inside a real Workflow that parks for human review. Same transaction,
same business rules.

TXN-A and TXN-C are slower too, because the handler Workflow sleeps ten seconds before
returning. That is the window you use in challenge 5. If the starter reports TXN-A as
still running, give it a moment and check the UI.

Leave TXN-B parked. Challenge 5 releases it.

# If it hangs

A payment Workflow stuck in `Running` with no compliance Workflow on the other side means
the Endpoint name does not match. Check the Payments Worker tab for:

```bash,nocopy
msg="Task processing failed with error" ... Error="BadScheduleNexusOperationAttributes: endpoint \"...\" not found"
```

A wrong name does not fail the Workflow. The server rejects the command and the Workflow
Task retries forever, so you get silence instead of an error. Compare
`ComplianceEndpoint` with the Registry:

```bash,run
temporal operator nexus endpoint list
```

# See the boundary

Click the [button label="Temporal UI" background="#444CE7"](tab-1) tab. In
`payments-namespace`, open the newest `PaymentProcessingWorkflow` and find **Nexus
Operation Scheduled**, **Nexus Operation Started** and **Nexus Operation Completed** in
the Event History. There is no **Activity Task Scheduled** for `CheckCompliance` any more.

Switch the Namespace selector to `compliance-namespace`. Three `ComplianceWorkflow`
executions are there, `compliance-TXN-A` to `compliance-TXN-C`. That Namespace was empty
in challenge 1.

Click **Check**. It runs both Workers against its own Temporal dev server, a payment and a
review included, so it takes about a minute.

# What you know now

- Swapping an Activity for a Nexus client is a one-call change at the call site.
- The Workflow names the Endpoint and the contract. The Registry resolves that name into
  a Namespace and a Task Queue.
- The work moved to another Namespace and the business logic never changed.

---

Tell us what worked in the **Feedback** tab. It takes a few seconds.
