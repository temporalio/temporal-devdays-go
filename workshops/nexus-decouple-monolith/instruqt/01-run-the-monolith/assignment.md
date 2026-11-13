---
slug: run-the-monolith
type: challenge
title: 1. Run the monolith
teaser: Run three payments through one Worker, then find the coupling in the code.
notes:
- type: text
  contents: |-
    # Two teams, one program

    A payment cannot go through until the Compliance team clears it. Today that
    check is an ordinary function call, because both teams' code is compiled into
    one program and runs as one process.

    What that costs:

    - A bug in Compliance's code stops Payments too.
    - Compliance cannot ship a fix unless Payments ships at the same time.
    - Nothing in the code marks where one team ends and the other begins.

    Temporal gives each team a Namespace of their own to run in. Compliance has
    one, and it is empty, because all of their code runs inside Payments'.

    **Author:** [Nikolay Advolodkin](https://www.linkedin.com/in/nikolayadvolodkin/), Staff Developer Advocate
- type: text
  contents: |-
    # Ziggy is warming up your sandbox

    A Temporal dev server is starting, and two Namespaces are being created,
    one for Payments and one for Compliance. Only one of them has anything in it.
tabs:
- title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/nexus-decouple-monolith
  port: 8080
- title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
- title: Terminal
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
- title: Payments Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
- title: Compliance Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
difficulty: basic
timelimit: 900
enhanced_loading: null
---

# The setup

You work at a bank. Every payment takes three steps.

1. Validate the payment. Payments team.
2. Check compliance. Compliance team.
3. Execute the payment. Payments team.

Nothing moves until Compliance says yes.

The blue buttons below are clickable. Click one to jump to that tab.

# Start the Worker

Click the [button label="Payments Worker" background="#444CE7"](tab-3) tab and start it:

```bash,run
go run ./payments-worker
```

The Go build cache is already warm, so this takes seconds. Wait for:

```bash,nocopy
  Payments Worker started on: payments-processing
  Namespace: payments-namespace
```

# Run three payments

Click the [button label="Terminal" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter
```

```bash,nocopy
  TXN-A   Result: COMPLETED             Risk: LOW
  TXN-B   Result: COMPLETED             Risk: MEDIUM
  TXN-C   Result: DECLINED_COMPLIANCE   Risk: HIGH
```

# Check the Temporal UI

Click the [button label="Temporal UI" background="#444CE7"](tab-1) tab. Switch the
Namespace selector at the top to `payments-namespace`.

Three Workflows, every one of them **Completed**. Refresh if you do not see them yet.

| Workflow ID | Amount | What happened |
|---|---|---|
| `payment-TXN-A` | $250 | paid |
| `payment-TXN-B` | $12,000 | paid |
| `payment-TXN-C` | $75,000 | declined by Compliance |

Open `payment-TXN-C`, find the **Input and Results** panel, and open **Results**:

```json,nocopy
{
  "explanation": "Transaction amount exceeds $50,000 threshold. Requires enhanced due diligence review.",
  "riskLevel": "HIGH",
  "status": "DECLINED_COMPLIANCE",
  "success": false,
  "transactionId": "TXN-C"
}
```

`success` is `false`, there is no `error` field, and the Workflow status is **Completed**
rather than Failed. A declined payment is a business outcome, so the Workflow returned it
and finished normally. A failed Workflow is a defect. This result should look identical
after you move the check across a team boundary in challenge 4.

# Find the coupling

Click the [button label="Exercise" background="#444CE7"](tab-0) tab. The file tree holds
`exercise/` and `solution/`; you work in `exercise/` the whole time.

Open `exercise/payments/worker.go`. `NewWorker` builds the Payments Worker, and near the
bottom it registers one Activity that is not Payments' at all:

```go,nocopy
w.RegisterActivity(compliance.CheckCompliance)
```

Same process, same deployment. Compliance cannot ship a fix unless Payments ships too.
You delete that line in challenge 4, and that deletion is the decoupling.

Now open `exercise/payments/workflows.go` and find step 2:

```go,nocopy
err := workflow.ExecuteActivity(ctx, compliance.CheckCompliance, compReq).Get(ctx, &result)
```

An in-process Activity call across a team boundary. Both files import the `compliance`
package, so Compliance's code is compiled into the Payments binary.

# See the empty Namespace

Click the [button label="Temporal UI" background="#444CE7"](tab-1) tab and switch the
Namespace selector to `compliance-namespace`.

No Workflows. Compliance's code is executing inside the Payments Worker. If **Workers**
lists anything, it is on a `temporal-sys-` Task Queue, which is Temporal Server's own and
runs in every Namespace. You start Compliance's Worker in challenge 3.

Switch back to `payments-namespace`, open `payment-TXN-C`, and find **Activity Task
Scheduled** for `CheckCompliance` in the Event History. That event is the one you spend the
rest of this lab replacing.

Click **Check** when your three transactions have finished. It reads the three Workflows
from this lab's Temporal Service.

# What you know now

- One Worker runs both teams, so one bug stops both.
- The compliance check is an ordinary Activity call.
- A declined payment completes. It does not fail.
- Compliance has a Namespace and it is empty.

---

Tell us what worked in the **Feedback** tab. It takes a few seconds.
