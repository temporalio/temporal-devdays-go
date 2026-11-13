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
`workflow.NewDisconnectedContext`, so a cancelled order still unwinds.

| # | Challenge | Learner does | Check passes when |
|---|---|---|---|
| 1 | Unwinding an order that failed halfway | `-fail-dispatch`, then TODO 1 in `workflow.go` | undos run newest first, a clean order runs none, one failing undo does not strand the rest, a cancelled order still unwinds, and the Workflow still fails |
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
  activities.go    three forward steps, two undos
  ledger.go        what the downstream systems did, printed as [ledger] lines
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

The ledger lives in the Worker process, so the Worker terminal is where the
learner reads what each order did. The Temporal UI shows the same story as
Activities in the Event History; the assignments ask about both after every run.

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

Measured on the Solution, from the Worker terminal:

```
go run ./starter                    reserve -> charge -> dispatch
go run ./starter -fail-dispatch     reserve -> charge -> refund -> release
go run ./starter -charge-times-out  reserve -> charge -> refund -> release
go run ./starter -fail-reserve      [compensate] nothing reserved ..., no-op
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

`go run ./worker` runs a child binary named `worker`. Each Challenge's setup stops
both, by name, so a Worker left on old code never serves the next Challenge.
Skip splices only that Challenge's TODOs in with `workshop-solve`; a running
Worker keeps its old code until it is restarted.

`config.yml` names the image by `:latest` until the first publish, which re-pins
it by digest. `track.yml` has no `id:` or `checksum:` yet; the first push writes
them, and that result is what gets committed.
