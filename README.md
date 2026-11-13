# temporal-devdays-go

Every Temporal DevDays Workshop in Go: what exists, where each piece lives, and
what the Port cost.

All created via code. Use or repurpose the content to fit your next Dev Day.

**Play any lab in your browser:** see the
[landing page](https://temporalio.github.io/temporal-devdays-go/).

## Workshops

| Workshop | Level | Stack | Deck | Code + Track | Lab | Duration | Total port |
|---|---|---|---|---|---|--:|--:|
| Activity Heartbeats | Advanced | Go · Activities · Heartbeats | `decks/go-master/segments/16-heartbeating.md` | `workshops/heartbeats-shipment/` | pending | 10 min | pending |
| Child Workflows | Foundations | Go · Child Workflows · fan out, fan in | `decks/go-master/segments/14-child-workflows.md` | `workshops/child-workflows/` | pending | 15 min | pending |
| Continue-As-New | Advanced | Go · Event History · Continue-As-New | `decks/go-master/segments/15-continue-as-new.md` | `workshops/continue-as-new-subscription/` | pending | 10 min | pending |
| Nexus: Decouple a Monolith | Advanced | Go · Nexus | `decks/nexus-decouple-monolith/` | `workshops/nexus-decouple-monolith/` | pending | 90 min | pending |
| The Saga Pattern | Advanced | Go · Compensations | `decks/saga-order-fulfillment/` | `workshops/saga-order-fulfillment/` | pending | 45 min | pending |
| Temporal Essentials | Foundations | Go · Workflows, Activities, Workers, Signals, Queries, Updates, timeouts and retries, Sagas, Nexus | `decks/go-master/` | Deck only; code slides proven in `snippets/` | per section, linked in the Deck | — | pending |
| Your First Workflow | Foundations | Go · Workflows · Activities · Workers · Task Queues · Event History | — | `workshops/first-workflow-money-transfer/` | pending | 15 min | pending |

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
