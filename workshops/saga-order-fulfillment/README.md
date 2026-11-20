# The Saga Pattern (Go)

**Level:** Advanced. For teams already running Temporal.
**Duration:** 45 minutes: about 20 of Deck, then the lab.
**Track:** [`saga-go`](https://play.instruqt.com/temporal/tracks/saga-go) ·
**Deck:** [`decks/saga-order-fulfillment/`](../../decks/saga-order-fulfillment/)

An order reserves stock, charges a card and books a courier, three systems with
no transaction across them. When the courier fails, the first two have already
happened. The learner turns that order Workflow into a Saga: every forward step
paired with an undo, the undos run newest first, and each undo safe to run when
there is nothing to undo.

There is no built-in Saga helper. The compensation stack is a slice of funcs,
and a `defer` walks it backwards on a context from
`workflow.NewDisconnectedContext`, so a cancelled order still unwinds. The undos
run with their own retry policy (`undoActivityOptions`): a forward step gives up
after two attempts so the order can unwind, but an undo that gives up strands
the money, so it retries with backoff for up to five minutes.

| # | Challenge | Learner does | Check passes when |
|---|---|---|---|
| 1 | Unwinding an order that failed halfway | `-fail-dispatch`, then TODO 1 in `workflow.go` | undos run newest first, a clean order runs none, one failing undo does not strand the rest, an undo retries past a brief outage, a cancelled order still unwinds, and the Workflow still fails |
| 2 | Register before you run | `-charge-times-out`, then TODO 2a and 2b in `workflow.go` | a failed reservation still calls its undo, and a charge that lands then fails is refunded |
| 3 | Make the undo safe to repeat | `-fail-reserve`, then TODO 3a and 3b in `activities.go` | a release with nothing reserved does nothing, and each undo run twice undoes once |

Every Check is gated (ADR 0003): it runs the `TestChallenge1_...` to
`TestChallengeN_...` tests in `saga_test.go` against their own dev server, and
fails with a message naming the file, the function and the TODO. The single
Challenge of the source Workshop was split at its three TODO groups, which were
already three separate ideas, so each Check can tell them apart.

## Layout

```
exercise/          the learner's terminals start here; TODOs 1, 2a, 2b, 3a, 3b undone
  workflow.go      OrderFulfillmentWorkflow: the compensation slice and the unwind
  activities.go    three forward steps, two undos; each undo returns what it did
  ledger.go        what the downstream systems did, printed as [ledger] lines
  recorded.go      reads an order's Event History: each Activity, the undos that
                   ran, and any undo owed but never run (the starter prints it)
  order.go         the Order, its result, and the scenario switches
  saga_test.go     one group of TestChallenge<N>_ tests per Challenge
  worker/          go run ./worker
  starter/         go run ./starter [-fail-reserve | -fail-dispatch | -charge-times-out]
solution/          the same module path, identical except inside the TODO regions
instruqt/          the Track: track.yml, config.yml, track_scripts/, 01..03/
sandbox/
  challenges.tsv   each Challenge's file and the TODOs its Skip splices in
```

The Sandbox Image is the shared one built from [`../_sandbox/`](../_sandbox/)
(ADR 0004). Task Queue: `fulfillment-tasks`.

After the Workflow closes, the starter reads its Event History and prints what
Temporal recorded: every Activity and how it ended (an undo's result says
whether it undid something or was a no-op), the undos that completed, in order,
and any undo owed for a step that ran but never undone. So the Client tab shows
whether compensation worked, whichever Worker ran the steps, and the
assignments point at the same events in the Temporal UI. Every failure scenario
ends Failed on purpose; the summary says so. The `[ledger]` lines in the Worker
terminal tell the same story from the downstream systems' side.

## Run it locally

Go 1.26 and the [Temporal CLI](https://docs.temporal.io/cli).

```bash
temporal server start-dev                          # terminal 1
cd workshops/saga-order-fulfillment/exercise       # or solution
go run ./worker                                    # terminal 2; restart it after every edit
go run ./starter                                   # terminal 3: a clean order
go run ./starter -fail-dispatch                    # courier fails after two steps landed
go run ./starter -charge-times-out                 # the charge lands, then reports failure
go run ./starter -fail-reserve                     # inventory fails before anything lands
go test ./...                                      # starts its own dev server
```

Measured on the Solution: the Worker's ledger, then the Client's summary line.

```
go run ./starter                    reserve -> charge -> dispatch
                                    Workflow: Completed
go run ./starter -fail-dispatch     reserve -> charge -> refund -> release
                                    Undos completed, in order: RefundPaymentIfCharged, ReleaseInventoryIfReserved
go run ./starter -charge-times-out  reserve -> charge -> refund -> release
                                    Undos completed, in order: RefundPaymentIfCharged, ReleaseInventoryIfReserved
go run ./starter -fail-reserve      [compensate] nothing reserved for ORD-..., no-op
                                    ReleaseInventoryIfReserved  completed nothing reserved, no-op
```

## Prove it

From the repo root:

```bash
(cd workshops/saga-order-fulfillment/solution && go vet ./... && go test -count=1 ./...)
(cd workshops/saga-order-fulfillment/exercise && go vet ./...)
bash scripts/test-checks.sh workshops/saga-order-fulfillment
node scripts/check-snippets.mjs workshops/saga-order-fulfillment
```

## Instruqt Track

Every Challenge has the same four tabs, so tab buttons never shift:

| Tab | Type | Target |
|---|---|---|
| `tab-0` Exercise | service | code-server on 8080, `/?folder=/root/workshop/saga-order-fulfillment` |
| `tab-1` Worker | terminal | `exercise/` |
| `tab-2` Client | terminal | `exercise/` |
| `tab-3` Temporal UI | service | 8233 |

Every Challenge opens fresh terminal tabs, but a Worker started in the last one
keeps running, detached. A second Worker would join `fulfillment-tasks`, the
Service would split orders between them, and the ledger lines (and each
Worker's in-memory ledger, which the undos' guards read) would split with them.
So each Challenge's `setup-workshop` stops any Worker left running: `go run
./worker` and its child binary, which runs from a `go-build` temp folder or,
once cached, from the Go build cache. It sends SIGTERM, waits up to 10 seconds,
then SIGKILLs, and never fails the Challenge start. Each assignment's run step
says the last Worker was stopped and to start a fresh one. Skip splices only
that Challenge's TODOs in with `workshop-solve`; a running Worker keeps its old
code until it is restarted.

`config.yml` names the image by `:latest` until the first publish, which re-pins
it by digest. `track.yml` has no `id:` or `checksum:` yet; the first push writes
them, and that result is what gets committed.
