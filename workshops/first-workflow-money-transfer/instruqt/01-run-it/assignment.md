---
slug: run-it
id: nnn4srwbmd5p
type: challenge
title: 1. Run It
teaser: Start a money transfer with no Worker running and watch it wait. Start the
  Worker and watch it finish.
notes:
- type: text
  contents: |-
    # Who runs your code, and who remembers how far it got?

    A money transfer is two steps: take $100 out of account A-100, put it into
    account B-200. That is simple until the process running it dies between
    the two steps.

    Temporal splits the job in two. Your Worker runs your code, and the
    Temporal Service records every step it took. This Challenge shows you each
    half in turn.

    **Author:** [Nikolay Advolodkin](https://www.linkedin.com/in/nikolayadvolodkin/), Staff Developer Advocate
- type: text
  contents: |-
    # Five words for this lab

    **Workflow.** A Workflow is the steps of the job in order: withdraw, then
    deposit. You write it as a plain Go function, and Temporal records its
    progress so it survives a crash.

    **Activity.** An Activity is one step that calls the outside world, such as
    the bank. Activities can fail, so Temporal retries them.

    **Worker.** The Worker is your process. It runs your Workflow and Activity
    code.

    **Task Queue.** The Task Queue is where the Temporal Service leaves work for
    Workers to pick up.

    **Event History.** The Event History is the Service's record of every step a
    Workflow took.
tabs:
- id: c6brntizrryh
  title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/first-workflow-money-transfer
  port: 8080
- id: b5wrwvlda8bc
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/first-workflow-money-transfer/exercise
- id: ofkkl9xfwg0g
  title: Client
  type: terminal
  hostname: workshop
  workdir: /root/workshop/first-workflow-money-transfer/exercise
- id: mrkw05yiou9p
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: basic
timelimit: 600
enhanced_loading: null
---

# Goal

Start a transfer before any Worker is running, watch it wait, then start the
Worker and watch it finish.

**Why it matters:** your code can stop and start, and the transfer is never
lost. The Temporal Service holds it until a Worker is there to run it.

# Your tabs

- [button label="Exercise" background="#444CE7"](tab-0) VS Code in the browser, open on this Workshop. You read
  the code under `exercise/` here, and fix a bug in the next Challenge. It saves as you type.
- [button label="Worker" background="#444CE7"](tab-1) a terminal in `exercise/`. Run the Worker here. It does
  not reload on save: you stop it with Ctrl+C and start it again by hand.
- [button label="Client" background="#444CE7"](tab-2) a terminal in `exercise/`. Run the starter here to start
  a transfer.
- [button label="Temporal UI" background="#444CE7"](tab-3) the Temporal UI of the local dev server. It lists
  every transfer, and shows each one's status and Event History.

# Step 1 of 4: Read the code

You change nothing in this Challenge. In the
[button label="Exercise" background="#444CE7"](tab-0) tab, open
`exercise/workflow.go` (`Ctrl+P`, or `Cmd+P` on a Mac, then the file name).
The Workflow is the whole transfer. This is the body of `MoneyTransferWorkflow`:

```go,nocopy
	var withdrawID, depositID string
	if err := workflow.ExecuteActivity(ctx, Withdraw, details).Get(ctx, &withdrawID); err != nil {
		return "", err
	}
	if err := workflow.ExecuteActivity(ctx, Deposit, details).Get(ctx, &depositID); err != nil {
		return "", err
	}
	return "Transfer complete: " + withdrawID + ", " + depositID, nil
```

The Workflow does not call the bank, or even `Withdraw`, itself.
`workflow.ExecuteActivity` asks Temporal to run that Activity, and `.Get` waits
for its result before the Workflow moves on. `workflow.WithActivityOptions`,
just above, sets each Activity's timeout and retry policy. The Activities are
plain Go functions in `exercise/activities.go`. Each one prints a line and
pretends to call the bank.

# Step 2 of 4: Start a transfer with no Worker

No Worker is running yet. In the [button label="Client" background="#444CE7"](tab-2)
tab, start a $100 transfer:

```bash,run
go run ./starter
```

```bash,nocopy
Started transfer-1f19b3: $100 from A-100 to B-200.
Find it in the Temporal UI: Workflows › transfer-1f19b3
Waiting for the result...
```

The starter waits. Your ID ends in different characters; use the one it
printed.

Open the [button label="Temporal UI" background="#444CE7"](tab-3) tab. It
opens on the **Workflows** list, with your transfer in it as **Running**. Click
your transfer's ID. If it isn't in the list, click **Refresh** next to the
title.

The transfer's page opens on the **Timeline** tab. This lab uses the
**Event History** tab instead, next to it, because it lists each event by name.

**You should see:**

- **Running** next to the ID.
- A yellow **No workers polling** banner that names the `money-transfer` Task
  Queue.
- **Event History 2**: the transfer has only two events so far. Click the tab
  to see them, **Workflow Execution Started** and **Workflow Task Scheduled**.

The transfer exists, and the Service has put its first piece of work on the
`money-transfer` Task Queue. No Worker is polling that queue, so nothing runs,
and nothing is lost.

# Step 3 of 4: Start the Worker

In the [button label="Worker" background="#444CE7"](tab-1) tab:

```bash,run
go run ./worker
```

```bash,nocopy
Worker polling the money-transfer Task Queue. Ctrl+C crashes it.
[activity] withdraw attempt 1: $100 from A-100
[activity] deposit attempt 1: $100 to B-200
```

The Worker picked up the waiting work at once. Back in the
[button label="Client" background="#444CE7"](tab-2) tab, the starter has its
result:

```bash,nocopy
Transfer complete: W-1f19b3, D-1f19b3
```

Leave the Worker running.

# Step 4 of 4: Read the Event History

In the [button label="Temporal UI" background="#444CE7"](tab-3) tab, the
transfer's page updates by itself: it now says **Completed**, and the banner
is gone. Click the **Event History** tab.

The history lists the newest event first. Click **Descending** so it reads
**Ascending**, oldest first, and read it from the top. Sorting turns
**Auto Refresh** off, which is fine: this transfer has finished.

| Event | What happened |
|---|---|
| **Workflow Execution Started** | The starter asked for a transfer. |
| **Workflow Task Scheduled**, **Started**, **Completed** | The Worker ran the Workflow code until it asked for `Withdraw`. |
| **Activity Task Scheduled**, **Started**, **Completed** | The Worker ran `Withdraw`. Its result is stored here. |
| **Activity Task Scheduled**, **Started**, **Completed** | The same for `Deposit`. |
| **Workflow Execution Completed** | The result went back to the starter. |

Temporal names the Activities and the Workflow after their Go functions:
`Withdraw`, `Deposit` and `MoneyTransferWorkflow`. The **Details** column
shows each Activity's name next to **Activity Task Scheduled**, and its attempt
number next to **Activity Task Started**.

**You should see:** a gap of several seconds in the **Timestamp** column
between **Workflow Task Scheduled** (event 2) and **Workflow Task Started**
(event 3). That is how long the transfer waited for a Worker.

Every step the Worker took is now in the Service, with its result. In the next
Challenge, this record lets a transfer survive a crashed Worker.

# Checkpoint

Click **Check**. It passes once a transfer has completed after waiting for a
Worker. If you started the Worker first, stop it with **Ctrl+C**, run the
starter again, wait a few seconds, then start the Worker.

---

**Please share your feedback so we can make better content for you.** The
**Feedback** tab takes a few seconds, and it is the only way we find out which
parts of this landed.
