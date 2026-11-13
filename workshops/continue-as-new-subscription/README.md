# Continue-As-New (Go)

**Level:** Advanced. Requires Temporal Essentials.
**Duration:** about 10 minutes of lab.
**Track:** `continue-as-new-go` (not yet published)

One Challenge, one concept. A subscription bills 30 cycles. Without
Continue-As-New the Temporal Service terminates it when its Event History passes
the limit. The learner adds Continue-As-New on the Service's suggestion
(`workflow.GetInfo(ctx).GetContinueAsNewSuggested()`), returning
`workflow.NewContinueAsNewError(...)` with the billing count, and runs it again:
it finishes across a chain of runs.

| # | Challenge | Change | Check passes when |
|---|---|---|---|
| 1 | Continue-As-New | TODO 1, `subscription/workflow.go` | the subscription rolls over on the Service's suggestion and bills every cycle |

The Sandbox's dev server lowers the history limits with two dynamic-config
values, so both are visible in seconds:

| Setting | Default | Track | Test |
|---|---|---|---|
| `limit.historyCount.suggestContinueAsNew` | 4,096 | 40 | 20 |
| `limit.historyCount.error` (terminates the run) | 51,200 | 200 | 100 |

Measured on the Track's values: the suggestion first appears at Event 42, the
unfixed run is terminated at Event 201 after cycle 18 (about 19 s), and the fixed
run bills 30 cycles across 8 runs in about 30 s.

## Layout

```
exercise/                 learner starts here; one Go module
  subscription/
    workflow.go           SubscriptionWorkflow; TODO 1 between `// TODO 1 start` / `// TODO 1 end`
    activities.go         ChargeSubscription
    shared.go             Task Queue, cycle length, input and result types
    workflow_test.go      TestChallenge1_... IS the Track's Check (ADR 0003)
  worker/main.go          go run ./worker
  starter/main.go         go run ./starter [tag]
  TODOS.md                every TODO, in order
solution/                 identical module path, TODO done
instruqt/                 the Track: track.yml, config.yml, track_scripts/, 01-roll-it-over/
sandbox/
  challenges.tsv          which Challenge owns which TODOs (for scripts/test-checks.sh)
```

The Sandbox Image is built from [`../_sandbox/`](../_sandbox/). Task Queue:
`subscription-tasks`.

## Run it locally

Go 1.26 and the Temporal CLI.

```bash
temporal server start-dev \
  --dynamic-config-value limit.historyCount.suggestContinueAsNew=40 \
  --dynamic-config-value limit.historyCount.error=200       # terminal 1
cd workshops/continue-as-new-subscription/exercise          # or solution
go run ./worker                                             # terminal 2; Ctrl+C and rerun after a change
go run ./starter                                            # terminal 3, subscription-demo
go run ./starter mine                                       # subscription-mine
go test ./...                                               # starts its own dev server
```

Without the two flags the default limits apply and nothing visible happens in a
30-cycle run.

## How the Check proves behavior

It does not read the source. `TestMain` starts a real `temporal server
start-dev` (the `temporal` binary on `PATH`, through `testsuite.StartDevServer`)
with the Test limits above, and the test bills 12 cycles. A cycle adds about 11
Events, so the fixed Workflow rolls over every 2 cycles:

- `*temporal.TerminatedError` (after cycle 9): the run never rolled over. The
  message names TODO 1 and `GetContinueAsNewSuggested()`.
- `*temporal.TimeoutError` (60 s execution timeout): it rolled over without
  carrying `CyclesCompleted`.
- Otherwise it expects 12 cycles and the first run's status to be
  `CONTINUED_AS_NEW`.

The passing test takes about 13 s.

## Prove it

From the repo root:

```bash
(cd workshops/continue-as-new-subscription/solution && go vet ./... && go test -count=1 ./...)
(cd workshops/continue-as-new-subscription/exercise && go vet ./...)
bash scripts/test-checks.sh workshops/continue-as-new-subscription
node scripts/check-snippets.mjs workshops/continue-as-new-subscription

cd workshops && docker buildx build --platform linux/arm64 -f _sandbox/Dockerfile \
  -t temporal-devdays-go-sandbox:local --load . && cd ..
bash scripts/test-track-local.sh workshops/continue-as-new-subscription temporal-devdays-go-sandbox:local
```

## Instruqt Track

Same four tabs as every gated Track: Exercise (code-server on 8080, rooted at
`/root/workshop/continue-as-new-subscription`), Worker and Client (terminals in
`exercise/`), Temporal UI (8233).

## Known issues

- `config.yml` names `ghcr.io/temporalio/temporal-devdays-go-sandbox:latest`.
  Pin it by digest at publish, and re-pin after every rebuild.
- `track.yml` has no `id:` or `checksum:`, and the challenge and its tabs have
  no `id:`. The first push writes them; commit what it writes.
- The lab's limits live in two places: `instruqt/track_scripts/setup-workshop`
  (40 / 200, quoted in the assignment) and `subscription/workflow_test.go`
  (20 / 100). Change the assignment's numbers and expected output if you change
  the first.
- The test's execution timeout is 60 s although a passing run takes about 13 s,
  so a freshly booted Sandbox has room. Only the "count reset" failure waits
  that long; it stays inside `workshop-check`'s 180 s limit.
- The Worker does not restart on save. The assignment tells the learner to
  `Ctrl+C` and rerun `go run ./worker`; Skip changes the file but not a running
  Worker.
- The starter replaces any subscription with the same ID
  (`WORKFLOW_ID_CONFLICT_POLICY_TERMINATE_EXISTING`), so a rerun never collides
  with the last one.
