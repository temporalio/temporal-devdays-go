# Your First Workflow (Go)

**Level:** Foundations. No Temporal experience needed.
**Duration:** about 15 minutes of lab.
**Track:** `your-first-workflow-go` (not yet published) ·
**Deck:** none yet

A Port of `workshops/first-workflow-money-transfer` in
[temporal-devdays-ts](https://github.com/temporalio/temporal-devdays-ts): same
narrative, same two Challenges. It is the opening Track of a Dev Day for a room
new to Temporal. `MoneyTransferWorkflow` runs two Activities, `Withdraw` then
`Deposit`. The lab teaches four ideas: Workflow and Activity, Worker and Task
Queue, Event History, and durable execution through a crash.

| # | Challenge | Change | Check passes when |
|---|---|---|---|
| 1 | Run It | none | a completed transfer waited at least 3 s for its first Workflow Task, so it was started before any Worker ran |
| 2 | Fix It Live | TODO 1, `activities.go` | TODO 1 is fixed, and a completed transfer's `Deposit` ran on attempt 2 or later while its `Withdraw` ran on attempt 1 |

Challenge 2 has one crash, and the fix happens during it. The learner presses
Ctrl+C on a Worker whose deposit keeps failing, fixes the bug, and starts the
Worker again. The same transfer finishes without a second withdraw.

## The bug

The Workshop needs a bug that compiles, passes `go vet`, and fails only on the
large-deposit branch, so the $100 transfer in Challenge 1 never meets it. The
large-deposit audit note makes its slice with `make([]string, 0, 2)`, length 0
and capacity 2, then writes `note[0]`. That panics with
`runtime error: index out of range [0] with length 0`. A slice's length is a
run-time value, so neither the compiler nor `go vet`
flags it. The fix is `make([]string, 2)`.

A panic, not a returned error, on purpose: the Go SDK recovers a panicking
Activity and reports it as a failed attempt, so the learner sees that a panic
does not take the Worker down and Temporal retries it like any other failure.

The Port leaves out these parts of the learn.temporal.io tutorial: the refund
compensation, non-retryable errors, the attempt limit, and the dev-server crash.
Sagas, retries and timeouts have their own Workshops.

## Layout

```
exercise/                 learner starts here; TODO 1 between `// TODO 1 start` / `// TODO 1 end`
  shared.go               TaskQueue, LargeDeposit, PaymentDetails
  workflow.go             MoneyTransferWorkflow
  activities.go           Withdraw, Deposit (TODO 1)
  worker/main.go          go run ./worker
  starter/main.go         go run ./starter [-amount 2500]
  moneytransfer_test.go   TestChallenge1_..., TestChallenge2_...: the coding half of
                          the Check (ADR 0003); TestSolution_... proves a large
                          transfer end to end
  TODOS.md                the TODO
solution/                 identical, TODO 1 done; same module path
instruqt/                 the Track: track.yml, config.yml, track_scripts/, 01-run-it/, 02-fix-it-live/
sandbox/
  bin/transfer-history    the observe-only half of each Check: reads Event Histories
                          with the temporal CLI and jq
  challenges.tsv          which TODOs each coding Challenge owns
```

The Sandbox Image is built from [`../_sandbox/`](../_sandbox/), which also
installs `sandbox/bin/`. Task Queue: `money-transfer`.

## Run it locally

Go 1.26 and the Temporal CLI.

```bash
temporal server start-dev                                 # terminal 1
cd workshops/first-workflow-money-transfer/exercise       # or solution
go run ./starter                                          # terminal 3; waits for a Worker
go run ./worker                                           # terminal 2
go run ./starter -amount 2500                             # hits the bug in exercise/
go test ./...                                             # starts its own dev server
```

`go test ./...` fails `TestChallenge2_...` and `TestSolution_...` in
`exercise/` until TODO 1 is done. That is the Check working.

## How the Checks prove behavior

- **Challenge 1** reads histories only. `transfer-history waited` looks for a
  completed `MoneyTransferWorkflow` whose first **Workflow Task Started** is at
  least 3 s after **Workflow Execution Started**: no Worker was polling when it
  began.
- **Challenge 2** runs `workshop-check exercise 2` first:
  `TestChallenge2_LargeDepositSucceeds` runs the learner's `Deposit` for $2,500
  in the SDK's test Activity environment and fails with the panic until TODO 1
  is fixed. Then `transfer-history retried` looks for a completed transfer whose
  `Deposit` **Activity Task Started** carries attempt 2 or more and whose
  `Withdraw` carries attempt 1.

## Prove it

From the repo root:

```bash
(cd workshops/first-workflow-money-transfer/solution && go vet ./... && go test -count=1 ./...)
(cd workshops/first-workflow-money-transfer/exercise && go vet ./...)
bash scripts/test-checks.sh workshops/first-workflow-money-transfer
node scripts/check-snippets.mjs workshops/first-workflow-money-transfer

cd workshops && docker buildx build --platform linux/arm64 -f _sandbox/Dockerfile \
  -t temporal-devdays-go-sandbox:local --load . && cd ..
bash scripts/test-track-local.sh workshops/first-workflow-money-transfer \
  temporal-devdays-go-sandbox:local
```

## Instruqt Track

Same four tabs as every gated Track: Exercise (code-server on 8080, rooted at
`/root/workshop/first-workflow-money-transfer`), Worker and Client (terminals in
`exercise/`), Temporal UI (8233). `config.yml` points at the shared
`ghcr.io/temporalio/temporal-devdays-go-sandbox` image (ADR 0004); re-pin it by
digest when publishing.

The Worker is plain `go run ./worker`, with no file watcher. The lesson is that
a Worker runs the code it started with until you restart it, and a restart on
every save would hide that.

Release in two steps. Push with `maintenance: true` and run
`instruqt track test`. Once it passes, set `maintenance: false` and push again
to take the Track live. Never leave it in maintenance.

### UI screenshots

None yet. The assignments describe each Temporal UI step in words. Screenshots
taken from a run of this Exercise (the Workflow and Activity types are
`MoneyTransferWorkflow`, `Withdraw` and `Deposit`) can go in `instruqt/assets/`.
Instruqt reads every file in a Challenge folder as a script, so images can't
live there.

## Known issues

- `worker/main.go` calls `worker.SetStickyWorkflowCacheSize(0)`. With a cache,
  the Server offers the next Workflow Task to the crashed Worker's sticky queue
  first, and the history can show a **Workflow Task Timed Out** after the crash.
  That event confuses a first-time learner. Without a cache, the restarted
  Worker gets the task at once.
- Ctrl+C is a crash, not a Shutdown: the Worker installs no signal handler and
  blocks forever after `Start`, so SIGINT ends the process without telling
  Temporal.
- `go run ./worker` builds and runs a child binary named `worker`. The solve
  and cleanup scripts stop it with `pkill -x worker`; signalling `go run` alone
  leaves the Worker running. `go run` prints `signal: interrupt` (or
  `terminated`) when its child dies; that is expected.
- `worker/main.go` prints a failed Activity attempt as one line through a
  `slog.Handler` that rewrites the SDK's `Activity error.` and
  `Activity panic.` log records. The SDK's own report is a full error, with a
  stack trace for a panic, every 2 seconds, and it hides the attempt count.
- After the crash in Challenge 2, the **No workers polling** banner does not
  come back for a few minutes: the Service remembers the dead Worker's last
  poll. Challenge 2 points at Pending Activities instead, where `Deposit` stays
  **Scheduled** and its attempt count stops.
- Sorting the Event History **Ascending** turns its Auto Refresh off. Both
  Challenges sort only after the transfer has completed.
- Challenge 2's Skip runs a stand-in Worker, a copy of the Solution in
  `/tmp/standin` whose TODO 1 always panics for a large deposit, so it produces
  the retried history even if the learner already fixed TODO 1.
