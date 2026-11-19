# Activity Heartbeats (Go)

**Level:** Advanced. Requires Temporal Essentials.
**Duration:** about 10 minutes of lab.
**Track:** `heartbeats-go`

One Challenge, one concept. A packing Activity works through 40 items, 0.3 s
each, under a 20 s start-to-close timeout, with up to 10 attempts. The learner
crashes its Worker with Ctrl+C partway through, twice:

- **Without Heartbeats:** the retry waits out the attempt's timeout and starts
  again at item 1 (`no Heartbeat recorded, starting at item 1 of 40`).
- **With Heartbeats:** the retry starts as soon as a replacement Worker is up,
  about 2 s after the kill (the Heartbeat Timeout), and resumes from the last
  recorded Heartbeat.

## The checkpoint is a few items behind the last one printed

Learners report "I stopped the Worker near item 10, but the last Heartbeat
showed about 8 packed." That is the SDK, not a defect, and the lab now says so
where it happens. Measured on a local dev server, SIGKILL right after the
Worker logged `packed item 10 of 40`, three runs before and three after the
change, every time:

- the Server's pending Activity (`temporal workflow describe -o json`,
  `pendingActivities[].heartbeatDetails`, the UI's **Heartbeat Details**) held
  `[6]`, on attempt 2, Scheduled, last failure `activity Heartbeat timeout`;
- the retry logged `last recorded Heartbeat says 6 packed, resuming at item 7
  of 40`;
- the Client printed `Attempt 2 resumed at item 7: the last recorded Heartbeat
  said 6 packed.`, and the `ActivityTaskCompleted` result recorded
  `"FinishedOnAttempt":2,"ResumedFrom":6`.

The SDK throttles Heartbeats: it sends at most one every 0.8 × the Heartbeat
Timeout (2 s → 1.6 s, about every 5 items at 0.3 s each) and keeps only the
newest details between sends. A killed process sends nothing more, so items 7
to 10 never reached Temporal and are packed twice.

Three things made that read as a bug, and are fixed:

1. The resume log printed a packed count as "starting at item 6" while the
   per-item lines are 1-based, so it read as if item 6 would be repacked. The
   start log now says `no Heartbeat recorded, starting at item 1 of 40` or
   `last recorded Heartbeat says K packed, resuming at item K+1 of 40`.
2. The assignment had learners restart the Worker at once and explained
   throttling only at the end. It now has them stop, read the pending
   Activity's Heartbeat Details in the UI, and explains the throttle there.
3. Only console text named the checkpoint. The Event History does not hold
   Heartbeat details (attempt 2's `ActivityTaskStarted` last failure is a
   Heartbeat timeout with no details), so `ShipmentResult` now carries
   `ResumedFrom`, and the `ActivityTaskCompleted` result records it. While the
   retry is pending, the authoritative checkpoint is the pending Activity's
   Heartbeat Details.

| # | Challenge | Change | Check passes when |
|---|---|---|---|
| 1 | Activity Heartbeats | TODO 1a in `workflows.go`; 1b and 1c in `activities.go` | a silent attempt is retried within seconds, every item is Heartbeated, and a retry resumes from the last Heartbeat |

Cancellation through the Heartbeat is a closing note in the assignment, not a
step: the cancel request arrives in the reply to a Heartbeat, the SDK cancels
the Activity's `ctx`, and `PackShipment`'s `select` on `ctx.Done()` returns
`ctx.Err()`.

## Ctrl+C is a crash

`worker/main.go` calls `w.Run(nil)`, not `w.Run(worker.InterruptCh())`.
`InterruptCh` would catch Ctrl+C and stop the Worker gracefully, which fails
the running attempt so it retries at once, and the lab would show nothing.
With no handler installed, the Go runtime exits on SIGINT at once, without
telling Temporal, the way a lost node does, and the attempt stays Started.

`go run ./worker` forks a child binary named `worker`. Ctrl+C in a terminal
reaches both. A script that crashes the Worker must signal the child by name
(`pkill -KILL -x worker`), not `go run`. A process started in the background
by a non-interactive shell inherits SIGINT as ignored, so `kill -INT` does
nothing to it; use SIGKILL from scripts.

Unlike a file-watching dev runner, `go run ./worker` does not restart on save.
The assignment tells the learner to Ctrl+C and rerun after each change.

## Layout

```
exercise/             learner starts here; TODOs between `// TODO 1x start` / `// TODO 1x end`
  shipment.go         Task Queue, sizes, input and result types
  activities.go       PackShipment (TODO 1b, 1c)
  workflows.go        ShipmentWorkflow (TODO 1a)
  heartbeats_test.go  TestChallenge1_*, one test per part of TODO 1, in order; it IS the Check (ADR 0003)
  worker/, starter/   go run ./worker, go run ./starter
  TODOS.md            every TODO, in order
solution/             identical module path and files, TODOs done
instruqt/             the Track: track.yml, config.yml, track_scripts/, 01-heartbeats/
sandbox/
  challenges.tsv      which TODOs the Challenge owns, one line per file (both lines are Challenge 1)
```

Module `github.com/temporalio/temporal-devdays-go/workshops/heartbeats-shipment`
in both trees. Task Queue: `shipment-tasks`.

## Run it locally

Go 1.26 and the Temporal CLI.

```bash
temporal server start-dev                                 # terminal 1
cd workshops/heartbeats-shipment/exercise                 # or solution
go run ./worker                                           # terminal 2; Ctrl+C it around item 10, then rerun
go run ./starter                                          # terminal 3, shipment-demo
go run ./starter mine                                     # shipment-mine
go test ./...                                             # starts its own dev server
```

## How the Check proves behavior

None of it reads the source. Every test runs the learner's `ShipmentWorkflow`
on a real dev server (started once in `TestMain` from the `temporal` CLI on
`PATH`), on a Worker of its own with an interceptor that records every
`RecordHeartbeat` call before the SDK throttles it. The tests run in order, so
the first missing part is the one reported. About 9 s on the Solution, 17 s on
the untouched Exercise.

- **1a** registers a stand-in `PackShipment` whose attempt 1 returns
  `activity.ErrResultPending`, so it never reports back, the way a dead Worker
  does. With an 8 s execution timeout, only a Heartbeat Timeout gets attempt 2
  started in time.
- **1b** runs the real `PackShipment` on 3 items and expects Heartbeats
  `1 2 3`.
- **1c** registers a stand-in whose attempt 1 Heartbeats `16` and fails, and
  whose attempt 2 is the real `PackShipment` on 18 items: it expects the first
  Heartbeat of attempt 2 to be 17, all 18 items packed, and `ResumedFrom` 16.
  A second test expects a first attempt to start with 0 packed. Calling `GetHeartbeatDetails` without
  `HasHeartbeatDetails` returns `temporal.ErrNoData` on a first attempt, and
  that is reported against 1c by name.

## Prove it

From the repo root:

```bash
(cd workshops/heartbeats-shipment/solution && go vet ./... && go test -count=1 ./...)
(cd workshops/heartbeats-shipment/exercise && go vet ./...)
bash scripts/test-checks.sh workshops/heartbeats-shipment
node scripts/check-snippets.mjs workshops/heartbeats-shipment

cd workshops && docker buildx build --platform linux/arm64 -f _sandbox/Dockerfile \
  -t temporal-devdays-go-sandbox:local --load . && cd ..
bash scripts/test-track-local.sh workshops/heartbeats-shipment temporal-devdays-go-sandbox:local
```

## Instruqt Track

Same four tabs as every gated Track: Exercise (code-server on 8080, rooted at
`/root/workshop/heartbeats-shipment`), Worker and Client (terminals in
`exercise/`), Temporal UI (8233). The image is the shared
`ghcr.io/temporalio/temporal-devdays-go-sandbox` (ADR 0004).

## Known issues

- `config.yml` points at `:latest` until the first publish, when it is pinned
  by digest. Re-pin it after every rebuild.
- `track.yml` has no `id:` or `checksum:` yet. The first push writes them;
  commit what it writes.
- TODO 1 spans two files, so `sandbox/challenges.tsv` has two lines for
  Challenge 1 and the solve script splices both.
- Order matters for a learner who runs early: a Heartbeat Timeout (1a) with no
  Heartbeats (1b) fails every attempt after 2 s. The assignment says to do all
  three parts before running.
