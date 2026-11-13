# Temporal Essentials: Go

The Foundations Deck for Go, and the pool of conceptual Go slides every other
Go Workshop Deck draws from: the case for Temporal, local development, the four
entities, design patterns, Workflows, Activities, Workers, Clients, Signals,
Queries and Updates; then Child Workflows, Continue-As-New, Heartbeats, Sagas,
patterns and anti-patterns, claim-check, versioning, data conversion and
encryption, Schedules and more, timeouts and retries, error handling,
observability and Worker tuning.

It is a Port of the Temporal Essentials Deck in the sibling DevDays repos: the
same narrative, timings, click order and speaker notes, with every code slide
rewritten as idiomatic Go for `go.temporal.io/sdk` v1.51.0.

## Viewing it

```bash
npm ci
npm run dev            # http://localhost:3030, presenter view at /presenter
npx slidev build       # static site in dist/
```

## Shared slides

Each file under `segments/` is self-contained and opens with its own section
divider, so a Workshop Deck assembles the ones it needs:

```yaml
---
src: ../go-master/segments/09-signals.md
---
```

Copy `layouts/code-walkthrough.vue` across with it. That layout is what gives
the annotated slides their one-caption-per-click arrow, and it lives in this
Deck rather than in the vendored theme. Copy the `pre, code { tab-size: 4 }`
rule from `styles/index.css` too: the Go on the slides is gofmt-indented with
tabs.

The lab slides live in `slides.md`, at the end of the section each lab
practises, and link the Go Tracks:

| Lab | Track | After |
|---|---|---|
| Your First Workflow | `your-first-workflow-go` | Clients (`07b-client.md`) |
| Child Workflows | `child-workflows-go` | Heartbeats (`16-heartbeating.md`) |
| Heartbeats | `heartbeats-go` | Heartbeats |
| Continue-As-New | `continue-as-new-go` | Heartbeats |
| The Saga Pattern | `saga-go` | The Saga section (`17-saga-07-related.md`) |
| Decouple a Monolith with Nexus | `nexus-decouple-monolith-go` | Patterns and anti-patterns (`17b-...`) |

## Every code slide is proven

`snippets/` is a Go module that holds each slide's code as a runnable package.
Its tests fail when a slide and its code drift apart, when a Go block stops
parsing, when a walkthrough's bubbles stop matching its highlights, or when the
code stops doing what the speaker notes say. See
[`snippets/README.md`](snippets/README.md).

```bash
(cd snippets && go vet ./... && go test -count=1 ./...)   # needs `temporal` on PATH
npm run dev                    # terminal 1
npm run check:walkthroughs     # terminal 2: every rendered bubble sits on its lines
```

CI runs both on any change to the Deck
(`.github/workflows/go-master-deck.yml`).
