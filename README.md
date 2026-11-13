# temporal-devdays-go

Every Temporal DevDays Workshop in Go: what exists, where each piece lives, and
what the Port cost.

All created via code. Use or repurpose the content to fit your next Dev Day.

**Play any lab in your browser:** see the
[landing page](https://temporalio.github.io/temporal-devdays-go/).

## Workshops

| Workshop | Level | Stack | Deck | Code + Track | Lab | Duration | Port cost |
|---|---|---|---|---|---|--:|--:|
| Activity Heartbeats | Advanced | Go · Activities · Heartbeats | [`segments/16-heartbeating.md`](decks/go-master/segments/16-heartbeating.md) | [`workshops/heartbeats-shipment/`](workshops/heartbeats-shipment/) | [Open](https://play.instruqt.com/temporal/invite/pcvx5aqrwygq) | 10 min | [$8.28](docs/ports/heartbeats-go.md) |
| Child Workflows | Foundations | Go · Child Workflows · fan out, fan in | [`segments/14-child-workflows.md`](decks/go-master/segments/14-child-workflows.md) | [`workshops/child-workflows/`](workshops/child-workflows/) | [Open](https://play.instruqt.com/temporal/invite/gyynuoe3u1sh) | 15 min | [$11.54](docs/ports/child-workflows-go.md) |
| Continue-As-New | Advanced | Go · Event History · Continue-As-New | [`segments/15-continue-as-new.md`](decks/go-master/segments/15-continue-as-new.md) | [`workshops/continue-as-new-subscription/`](workshops/continue-as-new-subscription/) | [Open](https://play.instruqt.com/temporal/invite/dziwe8cptxda) | 10 min | [$4.60](docs/ports/continue-as-new-go.md) |
| Nexus: Decouple a Monolith | Advanced | Go · Nexus | [`decks/nexus-decouple-monolith/`](decks/nexus-decouple-monolith/) | [`workshops/nexus-decouple-monolith/`](workshops/nexus-decouple-monolith/) | [Open](https://play.instruqt.com/temporal/invite/xx12iuebqz2e) | 90 min | [$34.24](docs/ports/nexus-decouple-monolith-go.md) |
| The Saga Pattern | Advanced | Go · Compensations | [`decks/saga-order-fulfillment/`](decks/saga-order-fulfillment/) | [`workshops/saga-order-fulfillment/`](workshops/saga-order-fulfillment/) | [Open](https://play.instruqt.com/temporal/invite/hy8hpfzptxk8) | 45 min | [$16.29](docs/ports/saga-go.md) |
| Temporal Essentials | Foundations | Go · Workflows, Activities, Workers, Signals, Queries, Updates, timeouts and retries, Sagas, Nexus | [`decks/go-master/`](decks/go-master/) | Deck only; code slides proven in [`snippets/`](decks/go-master/snippets/) | per section, linked in the Deck | — | [$97.62](docs/ports/temporal-essentials-go.md) |
| Your First Workflow | Foundations | Go · Workflows · Activities · Workers · Task Queues · Event History | — | [`workshops/first-workflow-money-transfer/`](workshops/first-workflow-money-transfer/) | [Open](https://play.instruqt.com/temporal/invite/r8azhqt26ino) | 15 min | [$8.21](docs/ports/first-workflow-go.md) |

Lab links are the Social Invites: one public link per Track for social posts,
kept apart from Dev Day invites. Port costs are agent cost only, a floor; the
seven Ports were run in parallel in one session, and that session's own work
(scaffold, Sandbox Image, end-to-end Track runs, publishing) cost a further
[$21.36](docs/ports/go-devdays-session.md). **Total: $202.14.**

Every Workshop here is a Port of its TypeScript version in
[temporal-devdays-ts](https://github.com/temporalio/temporal-devdays-ts).

## How this repo is organised

- `CONTEXT.md`: the words we use (Workshop, Port, Track, Check, Skip...).
- `docs/adr/`: the decisions, and why.
- `workshops/<slug>/`: each Workshop's `exercise/`, `solution/`, `instruqt/`
  Track and `sandbox/` Check helpers.
- `workshops/_sandbox/`: the one Sandbox Image every Track boots, and the
  shared Check and Skip helpers.
- `decks/`: Slidev Decks. `go-master` is Temporal Essentials and the pool of
  segments the other Decks import.
- `docs/ports/`: what each Port cost.
- `site/`: the landing page.

## Proving a Workshop

```bash
(cd workshops/<slug>/solution && go test -count=1 ./...)
(cd workshops/<slug>/exercise && go vet ./...)
scripts/test-checks.sh workshops/<slug>      # every Check fails before its TODOs, passes after
node scripts/check-snippets.mjs              # assignment snippets match the code
```

Needs Go 1.26, `jq`, and the `temporal` CLI on `PATH`.
