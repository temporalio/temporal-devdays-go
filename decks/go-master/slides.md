---
theme: ./theme
title: "Temporal Essentials: Go"
info: |
  ## Temporal Essentials, Go
  The Foundations Deck and the pool of conceptual slides every Go
  DevDays Workshop draws from: the case for Temporal, local development, the
  four entities, design patterns, Workflows, Activities, Workers, Clients,
  Signals, Queries and Updates; then Child Workflows, Continue-As-New,
  Heartbeats, Sagas, versioning, data conversion, Schedules, Serverless
  Workers, AI integrations, Priority and Fairness, timeouts and retries, error
  handling, observability and Worker tuning.

  Pull the segments a Workshop needs with `src:` and leave the rest behind.
author: Nikolay Advolodkin
keywords: temporal,go,golang,workflows,activities,workers,signals,queries,updates,patterns,saga,observability
colorSchema: dark
fonts:
  sans: "Inter"
  mono: "Noto Sans Mono"
  weights: "200,300,400,500,600"
  italic: false
mdc: true
themeConfig:
  footer: "Temporal Essentials | Go"
  toc:
    - id: why
      label: Why Temporal
    - id: local-dev
      label: Local development
    - id: concepts
      label: Temporal concepts
    - id: workflows
      label: Workflows
    - id: activities
      label: Activities
    - id: design
      label: Design patterns
    - id: workers
      label: Workers
    - id: client
      label: Clients
    - id: lab-first-workflow
      label: "Lab: The First Workflow"
    - id: start-execute
      label: Starting a Workflow
    - id: signals
      label: Signals
    - id: queries
      label: Queries
    - id: updates
      label: Updates
    - id: quiz
      label: Quiz
    - id: child
      label: Child Workflows
    - id: can
      label: Continue-As-New
    - id: heartbeats
      label: Heartbeats
    - id: saga
      label: "Compensations: the Saga pattern"
    - id: patterns
      label: Patterns and anti-patterns
    - id: claim-check
      label: The Claim-Check pattern
    - id: versioning
      label: Versioning
    - id: data
      label: Serialization and encryption
    - id: more-features
      label: More features
    - id: timeouts
      label: Timeouts and Retry Policies
    - id: errors
      label: Error handling
    - id: observability
      label: Observability
    - id: tuning
      label: Worker tuning
layout: cover
variant: planet-teal
---

# Temporal Essentials

## Go

<!--
Temporal Essentials, a Port of the Deck of the same name in the sibling
DevDays repos, which was itself ported from the master PowerPoint used for a
three-session Foundations Dev Day. Every snippet here is idiomatic Go against
go.temporal.io/sdk v1.51.0, and the code on each slide is proven by the
runnable packages in `snippets/` (`go test ./...`).

The opening follows decks/_template: cover, presenter, agenda
(`segments/00-open.md`). Every topic segment after it is self-contained and
starts with its own section divider. To assemble a Workshop, copy the `src:`
lines you want into that Workshop's `slides.md` and drop the rest. The lab
slides are here in `slides.md` rather than in segments, because they belong to a
delivery, not to a topic. Check each lab's minutes before delivering, and add
session breaks if the Dev Day runs over several days.

The through-line, if you do run several segments together: you write the happy
path, and the platform supplies the resilience. Every pattern here is a
consequence of that one trade.
-->

---
src: ./segments/00-open.md
---

---
src: ./segments/01-why-temporal.md
---

---
src: ./segments/02-local-development.md
---

---
src: ./segments/03-temporal-concepts.md
---

---
src: ./segments/05-workflows.md
---

---
src: ./segments/06-activities.md
---

---
src: ./segments/04-design-patterns.md
---

---
src: ./segments/06b-event-history.md
---

---
src: ./segments/07-workers.md
---

---
src: ./segments/07b-client.md
---

---
layout: exercise
minutes: 15
heading: "Lab: The First Workflow"
---

<CopyLink href="https://play.instruqt.com/temporal/tracks/your-first-workflow-go" />

<!--
Track: your-first-workflow-go (workshops/first-workflow-money-transfer).

The lab for Workflows, Activities, Workers and the Client: take an order app
that breaks under failure and make it durable. The minutes are the Track's time limit, not a measured
time; trim them to fit the day.

In the lab, the learner runs `go run ./worker` in one terminal and
`go run ./starter` in another, from the Exercise folder.

Leave this slide up while people open the lab. It is the only URL they have to
type, so give them a moment on it.
-->

---
src: ./segments/08-start-vs-execute.md
---

---
src: ./segments/09-signals.md
---

---
src: ./segments/10-queries.md
---

---
src: ./segments/11-updates.md
---

---
src: ./segments/12-quiz.md#1
# Only the "Quiz time!" divider. Drop `#1` to bring back the ten questions.
---

---
layout: section
---

<ZiggyHero scene="magic" src="/images/ziggy-wizard.png">

# Advanced Temporal Topics

Child Workflows, Continue‑As‑New, Heartbeats, Sagas, and beyond.

</ZiggyHero>

<!--
The turn from the fundamentals to the patterns. Everything after this slide
assumes the room is comfortable with Workflows, Activities, Signals, Queries
and Updates.

The tardigrade is Temporal's mascot, chosen because tardigrades survive almost
anything, which is the Durable Execution pitch in one animal. The wizard hat and
wand were added for an earlier DevDays Deck on top of the official art (from the
"Temporal in Space" blog post); there is no official wizard version, so check
with brand before using it outside a workshop.
-->

---
src: ./segments/14-child-workflows.md
---

---
src: ./segments/15-continue-as-new.md
---

---
src: ./segments/16-heartbeating.md
---

---
layout: exercise
minutes: 35
heading: Labs
---

- Child Workflows · 15 min <CopyLink href="https://play.instruqt.com/temporal/tracks/child-workflows-go" />
- Heartbeats · 10 min <CopyLink href="https://play.instruqt.com/temporal/tracks/heartbeats-go" />
- Continue-As-New · 10 min <CopyLink href="https://play.instruqt.com/temporal/tracks/continue-as-new-go" />

<!--
The labs for the three sections just taught, in this order. Each is its own
Track; the minutes are each lab's measured length, and each Track's time limit
is 30.

- **Child Workflows** (workshops/child-workflows): run one child, chain
  children five deep and follow a failure up from the bottom, fan out and watch
  the default Parent Close Policy terminate every child, then write the fan in
  by starting every child first and calling `Get` on each child's Future.
- **Heartbeats** (workshops/heartbeats-shipment): crash a Worker ten items into
  a packing job and watch the retry wait out the start-to-close timeout and
  start again at item 0; then add a `HeartbeatTimeout`,
  `activity.RecordHeartbeat(ctx, packed)` and a resume from
  `activity.GetHeartbeatDetails`, crash it again, and watch the retry start in
  seconds where it left off.
- **Continue-As-New** (workshops/continue-as-new-subscription): a 30-cycle
  subscription is terminated by the Service at its history limit, then rolls
  over on `workflow.GetInfo(ctx).GetContinueAsNewSuggested()` and finishes. The lab's dev server lowers the
  limits to 40 and 200 Events so both happen in seconds.

Every lab runs the same way: `go run ./worker` in one terminal and
`go run ./starter` in another, from the Exercise folder. Track slugs:
child-workflows-go, heartbeats-go, continue-as-new-go.

Leave this slide up while people open the labs.
-->

---
layout: section
---

<ZiggyHero scene="undo">

# Compensations

The Saga pattern

</ZiggyHero>

<!--
Segments 17-saga-01 to 17-saga-07 follow the Temporal docs' Saga pattern page
section by section, ported from the Essentials Deck. The code matches this repo's
Saga Workshop (workshops/saga-order-fulfillment), so the lab that closes the
section uses the names the slides just taught.
-->

---
src: ./segments/17-saga-01-overview.md
---

---
src: ./segments/17-saga-02-solution.md
---

---
src: ./segments/17-saga-03-implementation.md
---

---
src: ./segments/17-saga-04-when-to-register.md
---

---
src: ./segments/17-saga-05-when-to-use.md
---

---
src: ./segments/17-saga-06-practices.md
---

---
src: ./segments/17-saga-07-related.md
---

---
layout: exercise
minutes: 25
heading: "Lab: The Saga Pattern"
---

<CopyLink href="https://play.instruqt.com/temporal/tracks/saga-go" />

<!--
Track: saga-go (workshops/saga-order-fulfillment).

The same lab the Saga Workshop Deck closes on (decks/saga-order-fulfillment):
an order that reserves stock, charges a card and dispatches a courier, and
undoes the first two when the third fails. Learners run `go run ./worker` and
`go run ./starter` from the Exercise folder.

Leave this slide up while people open the lab.
-->

---
src: ./segments/17b-patterns-and-anti-patterns.md
---

---
layout: exercise
minutes: 40
heading: "Lab: Decouple a Monolith with Nexus"
---

<CopyLink href="https://play.instruqt.com/temporal/tracks/nexus-decouple-monolith-go" />

<!--
Track: nexus-decouple-monolith-go (workshops/nexus-decouple-monolith; its own
Deck is decks/nexus-decouple-monolith).

Optional, and the longest lab of the day (40 minutes of a 90-minute Workshop):
run it in its own session or leave it for a follow-up Dev Day. The patterns
segment just argued for clear boundaries between services; Nexus is how
Temporal draws one between teams. The learner takes a payments monolith whose
compliance check runs as a local Activity, registers a Nexus Endpoint with the
Temporal CLI, defines the shared Service contract both teams compile against,
implements the compliance side as a Nexus Operation (`temporalnexus`), and
swaps the payments Workflow's Activity call for a durable cross-team call
through `workflow.NewNexusClient`. Each side runs its own Worker with
`go run ./worker`, and `go run ./starter` sends payments through. The Event
History in the Temporal UI shows the Nexus Operation events.

This Deck has no Nexus segment of its own; the Nexus Workshop's Deck teaches
it. Leave this slide up while people open the lab.
-->

---
src: ./segments/21-claim-check.md
---

---
src: ./segments/18-versioning.md
---

---
src: ./segments/19-serialization.md
---

---
src: ./segments/20-encryption.md
---

---
src: ./segments/22-more-features.md
---

---
src: ./segments/23-timeouts-and-retries.md
---

---
src: ./segments/24-error-handling.md
---

---
src: ./segments/25-observability.md
---

---
src: ./segments/26-worker-tuning.md
---

---
layout: qa
---

<!--
Source: PPT slide 207.
-->

---
layout: end
---

# Thank you!

<!--
Source: PPT slide 208.
-->
