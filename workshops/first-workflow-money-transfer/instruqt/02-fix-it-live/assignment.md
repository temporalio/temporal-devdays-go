---
slug: fix-it-live
type: challenge
title: 2. Fix It Live
teaser: A large transfer hits a bug after the money is withdrawn. Crash the Worker,
  fix the bug, and watch the same transfer finish.
notes:
- type: text
  contents: |-
    # The money is out of A-100. The deposit keeps failing. Now what?

    Large deposits have a bug. The withdraw has already gone through, so the
    money has left one account and not arrived in the other.

    Without Temporal, this is a support ticket and a manual fix-up. With
    Temporal, the transfer waits. You fix the code and restart the Worker, and
    the same transfer carries on from the deposit.
tabs:
- title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/first-workflow-money-transfer
  port: 8080
- title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/first-workflow-money-transfer/exercise
- title: Client
  type: terminal
  hostname: workshop
  workdir: /root/workshop/first-workflow-money-transfer/exercise
- title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: basic
timelimit: 900
---

# Goal

Start a $2,500 transfer that hits a bug, crash the Worker, fix the bug, and
watch the same transfer finish without withdrawing twice.

**Why it matters:** bugs reach production. When one does, the transfers that
hit it are not lost or half-done. They wait, and they finish once you deploy
the fix.

# Step 1 of 4: Start a transfer that hits the bug

Your Worker from the last Challenge should still be running in the
[button label="Worker" background="#444CE7"](tab-1) tab. If it isn't, start it
there:

```bash,run
go run ./worker
```

In the [button label="Client" background="#444CE7"](tab-2) tab, start a $2,500
transfer:

```bash,run
go run ./starter -amount 2500
```

```bash,nocopy
Started transfer-7ebcf9: $2500 from A-100 to B-200.
Find it in the Temporal UI: Workflows › transfer-7ebcf9
Waiting for the result...
```

The [button label="Worker" background="#444CE7"](tab-1) tab shows the withdraw
go through, then the deposit fail, again and again:

```bash,nocopy
[activity] withdraw attempt 1: $2500 from A-100
[activity] deposit attempt 1: $2500 to B-200
[activity] failed: runtime error: index out of range [0] with length 0. Temporal will retry.
[activity] deposit attempt 2: $2500 to B-200
[activity] failed: runtime error: index out of range [0] with length 0. Temporal will retry.
```

The deposit panics, and the Worker keeps running. The SDK catches a panic in an
Activity and reports it to Temporal as a failed attempt, like a returned error.

In the [button label="Temporal UI" background="#444CE7"](tab-3) tab, go back
to the **Workflows** list (**Workflows** in the left sidebar) and click the ID
the starter just printed. It is the top row. On the transfer's page, click the
**Pending Activities** tab. It updates by itself every few seconds.

**You should see:**

- `Deposit`, with an **Attempt** count that goes up as you watch, out of
  **UNLIMITED**.
- **Last Failure**: the `index out of range` panic from the Worker tab.
- **Running** next to the ID, not Failed.

The Workflow's retry policy has no limit on attempts and waits at most 2
seconds between them, so Temporal retries the deposit until someone fixes it.
The withdraw is not pending: it already finished, and its result is in the
Event History.

# Step 2 of 4: Crash the Worker

In the [button label="Worker" background="#444CE7"](tab-1) tab, press **Ctrl+C**.
This Worker has no handler for Ctrl+C, so the process exits at once and does
not tell Temporal.

Watch the **Pending Activities** tab in the
[button label="Temporal UI" background="#444CE7"](tab-3) tab for 10 seconds.

**You should see:**

- `Deposit` marked **SCHEDULED**, and its **Attempt** count no longer going up.
  The next attempt waits on the Task Queue for a Worker, as your first transfer
  did in Challenge 1.
- **Running** next to the ID. The Temporal Service holds the transfer, not the
  Worker.

The **No workers polling** banner from Challenge 1 does not come back yet. The
Service remembers a Worker's last poll for a few minutes after it goes away.

# Step 3 of 4: Fix the bug (TODO 1)

**Observe.** Open `exercise/activities.go` in the
[button label="Exercise" background="#444CE7"](tab-0) tab. Press `Ctrl+P`
(`Cmd+P` on a Mac), type the file name, then `Ctrl+G` to jump to a line.

<!-- snippet: exercise/activities.go TODO 1 -->
Lines 28-35 are the large-deposit audit note in `Deposit`:

```go
	// TODO 1 start: write the audit note for a large deposit without crashing
	if details.Amount >= LargeDeposit {
		note := make([]string, 0, 2)
		note[0] = fmt.Sprintf("Large deposit of $%d", details.Amount)
		note[1] = "from " + details.SourceAccount
		fmt.Println("[activity] audit:", strings.Join(note, " "))
	}
	// TODO 1 end
```

**Learn.** A slice has a length, the elements it holds, and a capacity, the
room set aside for more. `make([]string, 0, 2)` makes a slice with length 0 and
capacity 2: room for two strings, but none there yet. `note[0]` reads past the
length, so Go panics with `index out of range [0] with length 0`. The code
compiles and `go vet` passes, because a slice's length is only known when the
code runs. Every deposit of $1,000 or more panics. The $100 transfer in
Challenge 1 never took this branch, so it worked.

**Do.**

- [ ] Between the TODO 1 markers, make the slice long enough to hold the two
  lines the audit note writes.

<details>
<summary>Hint</summary>

The second argument to `make` is the length, and the optional third is the
capacity:

```go
a := make([]int, 0, 5) // length 0, capacity 5: a[0] panics, append(a, 1) works
b := make([]int, 5)    // length 5, capacity 5: b[0] to b[4] exist, all zero
```

</details>

<details>
<summary>Show solution</summary>

<!-- snippet: solution/activities.go TODO 1 -->
Lines 28-35 of `solution/activities.go`:

```go
	// TODO 1 start: write the audit note for a large deposit without crashing
	if details.Amount >= LargeDeposit {
		note := make([]string, 2)
		note[0] = fmt.Sprintf("Large deposit of $%d", details.Amount)
		note[1] = "from " + details.SourceAccount
		fmt.Println("[activity] audit:", strings.Join(note, " "))
	}
	// TODO 1 end
```

</details>

# Step 4 of 4: Start the fixed Worker

The editor saved your fix, but nothing runs it yet. A Worker runs the code it
started with. In the [button label="Worker" background="#444CE7"](tab-1) tab,
start it again. `go run` compiles your fix first:

```bash,run
go run ./worker
```

```bash,nocopy
Worker polling the money-transfer Task Queue. Ctrl+C crashes it.
[activity] deposit attempt 6: $2500 to B-200
[activity] audit: Large deposit of $2500 from A-100
```

The new Worker ran the deposit's next attempt. **It did not run the withdraw
again.** In the [button label="Client" background="#444CE7"](tab-2) tab, the
starter that has been waiting all along gets its result:

```bash,nocopy
Transfer complete: W-7ebcf9, D-7ebcf9
```

Your attempt number depends on how long you took.

In the [button label="Temporal UI" background="#444CE7"](tab-3) tab, the
transfer now says **Completed**, and **Pending Activities** shows 0. Click the
**Event History** tab, then **Descending** so it reads **Ascending**.

**You should see:** two **Activity Task Started** events. In the **Details**
column, the one after `Withdraw` says **Attempt 1**, and the one after
`Deposit` says the attempt that worked. The failed attempts are not events.
Temporal keeps only the last failure on the pending Activity, so retries do
not fill the history.

# Checkpoint

Click **Check**. It passes when TODO 1 is fixed and a large transfer has
completed after its deposit failed, with the withdraw run only once. If it
fails, the message says what is missing.

# What you just did

- A **Workflow** is the steps of a job. **Activities** are the steps that can fail.
- A **Worker** runs your code. The **Temporal Service** keeps the Event History.
- Work waits on a **Task Queue** until a Worker picks it up, even across a crash.
- A failed Activity is **retried**, and the steps before it are not run again.
- You fixed a bug while the transfer was running, and the same transfer
  finished.

---

**Please share your feedback so we can make better content for you.** The
**Feedback** tab takes a few seconds, and it is the only way we find out which
parts of this landed.
