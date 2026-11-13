# Nexus: Decouple a Monolith (Go)

**Level:** Advanced. 90 minutes live, with a 40-minute lab block.

Payments and Compliance share one Worker and one Namespace, so a bug in compliance code
at 3 AM takes payments down with it. Learners put a Nexus boundary between the two teams
without changing a line of business logic: define the shared Service contract, implement
the handlers on the Compliance side, register them on a Worker, then swap the caller
from an Activity to a Nexus client. Then they stop the Compliance Worker mid-payment and
watch the payment Workflow wait instead of fail.

- Deck: [`decks/nexus-decouple-monolith/`](../../decks/nexus-decouple-monolith/)
- Track: `nexus-decouple-monolith-go`, defined in [`instruqt/`](instruqt/)

## Layout

```
exercise/ and solution/      one Go module each, same module path
  shared/                    the contract: types, names, the Service's Operations   TODO 1
  compliance/                checker, Activity, ComplianceWorkflow (given)
    handler.go               checkCompliance (worked example), submitReview         TODO 2
    worker.go                NewWorker                                              TODO 3
  payments/                  gateway, Activities (given)
    workflows.go             PaymentProcessingWorkflow, ReviewCallerWorkflow        TODO 4a-4c
    worker.go                NewWorker                                              TODO 5a-5b
  payments-worker/           go run ./payments-worker
  compliance-worker/         go run ./compliance-worker
  starter/                   go run ./starter         three payments, TXN-A to TXN-C
  review-starter/            go run ./review-starter  approves TXN-B through Nexus
  tests/                     the Checks' tests, TestChallenge<N>_...
sandbox/
  challenges.tsv             coding Challenge -> file -> TODOs, for scripts/test-checks.sh
  bin/                       observe-only Checks, installed into the Sandbox Image
instruqt/                    track.yml, config.yml, track_scripts/, five Challenges
```

Each team's Worker is built by a `NewWorker` in its own package, so a `main` package
only connects, starts it and waits for Ctrl+C. That is also what lets the tests run the
learner's real Workers against their own dev server.

## Run it

You need Go 1.26 and the `temporal` CLI.

```bash
temporal server start-dev
temporal operator namespace create --namespace payments-namespace
temporal operator namespace create --namespace compliance-namespace
temporal operator nexus endpoint create --name compliance-endpoint \
  --target-namespace compliance-namespace --target-task-queue compliance-risk

cd solution                      # or exercise/, which starts as the monolith
go run ./compliance-worker       # one terminal
go run ./payments-worker         # another
go run ./starter                 # TXN-A completes, TXN-C is declined, TXN-B parks for review
go run ./review-starter          # approves TXN-B through Nexus
```

Every program reads `TEMPORAL_ADDRESS` (default `localhost:7233`).

## Challenges

| # | Slug | Format | Files touched | TODOs | Check |
|---|------|--------|---------------|-------|-------|
| 01 | run-the-monolith | observe | none | | `nexus-check-monolith` |
| 02 | the-shared-contract | code | `shared/service.go` | 1 | `workshop-check ... 2` |
| 03 | the-compliance-side | code | `compliance/handler.go`, `compliance/worker.go` | 2, 3 | `workshop-check ... 3` |
| 04 | the-payments-side | code | `payments/workflows.go`, `payments/worker.go` | 4a-4c, 5a-5b | `workshop-check ... 4` |
| 05 | durability-and-human-review | observe | none | | `nexus-check-outage` |

`CheckComplianceOperation` in `compliance/handler.go` is **given, not a TODO**. It is the
harder of the two Operations and the one whose shape carries the lesson, so learners read
it as a worked example and then write `submitReview`, the opposite shape on the same
Service.

TODO 4 and TODO 5 are lettered because each is one idea in several places, and every
letter sits on the code it changes. 4a deletes the `compliance` import, 4b replaces the
Activity call with a Nexus call, 4c writes the review caller; 5b deletes the
`RegisterActivity` line and 5a its import. 4b and 4c build the same Nexus client with
different budgets, `10 * time.Minute` for the async `checkCompliance` and
`10 * time.Second` for the sync `submitReview`, and that pair is the lesson.

## Checks gate progress (ADR 0003)

Every Check fails until its Challenge's work is done, and says what is missing and where.

- **Coding Challenges (2, 3, 4)** run `workshop-check`: `go vet`, then the tests named
  `TestChallenge1_` to `TestChallengeN_` in `tests/`. The package starts one dev server
  from the `temporal` CLI with both Namespaces and the `compliance-endpoint` Endpoint, as
  the lab has. Challenge 2's test checks each Operation's name and types. Challenge 3's
  run `compliance.NewWorker`, confirm it polls for Nexus tasks, and call both Operations
  through the Endpoint from test-owned caller Workflows. Challenge 4's run both teams'
  `NewWorker`s: the payment must schedule a Nexus Operation and no `CheckCompliance`
  Activity, the review must go through `submitReview`, and a probe Workflow proves the
  Payments Worker can no longer run `CheckCompliance`.
- **Observe-only Challenges (1, 5)** read the lab's own dev server with the temporal CLI
  and jq. `nexus-check-monolith` wants `payment-TXN-A` to `-C` finished with the
  monolith's results. `nexus-check-outage` wants all three finished, at least one whose
  **Nexus Operation Started** came 5 or more seconds after **Nexus Operation Scheduled**
  (it waited out the outage), and a completed `review-TXN-B`.
- A Nexus Endpoint name that does not exist hangs rather than fails: the server rejects
  the command and the Workflow Task retries forever. The tests watch for that retry loop
  and name `ComplianceEndpoint` in `shared/types.go`.

Prove the Checks before changing anything they gate:

```bash
(cd solution && go vet ./... && go test -count=1 ./...)
(cd exercise && go vet ./...)
bash ../../scripts/test-checks.sh .
node ../../scripts/check-snippets.mjs .
```

## The Track

- **Tabs, in this order:** Exercise (code-server on 8080, rooted at the Workshop so
  `exercise/` and `solution/` are both in the tree), Temporal UI (8233), Terminal,
  Payments Worker, Compliance Worker. `tab-N` in an assignment is a position, so
  reordering means remapping every button. One editor tab, first, so its iframe has real
  dimensions when code-server boots.
- **The Endpoint is created by the Track's setup** (ADR 0004), after the dev server is
  healthy, along with both Namespaces. Challenge 3 has learners read it with
  `temporal operator nexus endpoint list` and match its target Task Queue to
  `ComplianceTaskQueue`.
- **Skip never overwrites a tree.** Coding Challenges splice their TODOs in with
  `workshop-solve`; observe-only ones run the programs. No pristine copy or restore guard
  is needed.
- **Workers are stopped by name.** `go run ./payments-worker` forks a child binary named
  `payments-worker`, so `pkill -f payments-worker` stops both.
- **No architecture-diagram tab.** That diagram embeds code and is served by a small
  static server; the Sandbox Image has neither a Go copy of it nor a server for it.

New Tracks have no `id:` or `checksum:`; the first `instruqt track push` writes them,
and `config.yml` is re-pinned to the Sandbox Image's digest at publish.

## Gotchas

- **`PaymentGateway` fails 10% of the time on purpose.** Tests and Checks assert on final
  Workflow state, never on a clean first attempt.
- **The review is time sensitive.** `submitReview` is a sync Operation with a 10-second
  budget, and the review Update is rejected until the compliance check has produced its
  MEDIUM verdict. The tests wait for the check's timer before reviewing; Challenge 5's
  solve polls for `compliance-TXN-B` and retries.
- **After an outage the Nexus Operation retries with backoff**, so a payment can take up
  to a minute to resume once the Compliance Worker is back.
- **A Worker's Nexus poller outlives it for a while** in `temporal task-queue describe`.
  The tests use a fresh dev server, so this never misleads a Check.
