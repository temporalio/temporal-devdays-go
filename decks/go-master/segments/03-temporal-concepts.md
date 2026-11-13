---
layout: section
---

<ZiggyHero scene="replay">

# Temporal concepts

The four pieces, and why replay works.

</ZiggyHero>

<!--
Source: PPT slides 22-28. Carried over from the original Essentials Deck.

Seven short slides. This is the vocabulary everything after it assumes, so
check the room is with you before the code starts.
-->

---
layout: default
---

# The big picture

<TemporalArchitecture />

<!--
Source: PPT slide 23.

The original had three bullets: the Temporal Server tracks state, schedules work
and records history; the Worker is the process you deploy that executes code;
your business logic is orchestration plus action.

- **Temporal Client** - a library in your application that sends control
  requests (start, Signal, cancel, Query) to the Server over gRPC. It runs no
  Workflow or Activity code.
- **Temporal Server** - records Event History, schedules Tasks on Task Queues,
  and persists state to a database. Frontend is the API, History owns Workflow
  state, Matching owns the Task Queues.
- **Workers** - your processes. They long-poll the Server for Tasks and run your
  Workflow and Activity code.

The three points worth making, in the order people ask:

1. **Nothing of yours runs on the Server.** It schedules and remembers. It never
   executes your code or sees your dependencies. That is the answer to the
   security review.
2. **The arrows only point outward.** Workers poll the Server. Temporal never
   opens a connection into your network, so Workers can sit in a private subnet
   behind NAT with no inbound rule, no ingress and no VPN.
3. **The database is the product.** Event History is the thing you would
   otherwise build: a state table, a job runner, a retry column, a
   reconciliation script, and someone on call for all of it.

If asked about Temporal Cloud versus self-hosted: the Server box is the only
thing that changes hands. Client and Workers are identical either way, which is
why moving between them is a connection string and a credential.

If asked "what happens if the Server goes down": running Workflows do not lose
state, because state is in the database, not in memory. Workers keep retrying
their polls and pick up where they left off. Nothing is lost; things just pause.
-->

---
layout: default
---

# Two kinds of code

<div class="kinds">
<div class="kind">
<p class="col-eyebrow">Workflow</p>
<p class="lede"><strong>Orchestrates the business logic.</strong> Sequencing, decisions, waiting. Never touches the outside world directly.</p>
<dl>
<dt>Determinism</dt><dd>⚠️ Must be deterministic</dd>
<dt>I/O</dt><dd>⚠️ No network, DB, or file I/O</dd>
<dt>Time</dt><dd><code>workflow.Now()</code>, <code>workflow.Sleep()</code></dd>
<dt>Randomness</dt><dd>Recorded once with <code>workflow.SideEffect()</code></dd>
<dt>Logging</dt><dd><code>workflow.GetLogger(ctx)</code></dd>
<dt>Retries</dt><dd>Off; failed Workflow Tasks retry</dd>
<dt>Duration</dt><dd>Indefinite</dd>
<dt>On failure</dt><dd><strong>Replays from history</strong></dd>
</dl>
</div>
<div class="kind">
<p class="col-eyebrow">Activity</p>
<p class="lede"><strong>Does the actual work</strong> that touches the outside world. An API call, a database write, sending an email.</p>
<dl>
<dt>Determinism</dt><dd>Can be non-deterministic</dd>
<dt>I/O</dt><dd>All I/O allowed</dd>
<dt>Time</dt><dd>Any clock</dd>
<dt>Randomness</dt><dd><code>math/rand</code>, <code>uuid.New()</code></dd>
<dt>Logging</dt><dd><code>activity.GetLogger(ctx)</code></dd>
<dt>Retries</dt><dd><code>RetryPolicy</code> in <code>ActivityOptions</code></dd>
<dt>Duration</dt><dd>Typically short-lived</dd>
<dt>On failure</dt><dd><strong>Restarts from scratch</strong></dd>
</dl>
</div>
</div>

<style>
.kinds { display: grid; grid-template-columns: 1fr 1fr; gap: 1.2rem; margin-top: 0.2rem; }
.kind {
  padding: 0.9rem 1.1rem 1rem; border-radius: 0.7rem;
  border: 1px solid var(--temporal-border); background: var(--temporal-bg-elev);
}
.kind .lede { min-height: 4.2em; margin: 0 0 0.7rem !important; font-size: 0.95rem !important; line-height: 1.4 !important; }
.kind dl {
  display: grid; grid-template-columns: max-content 1fr; gap: 0.32rem 0.9rem;
  margin: 0; padding-top: 0.6rem; border-top: 1px solid var(--temporal-border);
}
.kind dt { font-size: 0.72rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--temporal-text-muted); align-self: center; }
.kind dd { margin: 0; font-size: 0.82rem; color: var(--temporal-text); }
.kind dd code { font-size: 0.74rem !important; }
</style>

<!--
Source: PPT slide 24, "Workflow code cannot interact with the outside world
directly", merged with PPT slide 37 (`{{SLIDE:workflows-activities-comparison}}`),
the Workflows vs. Activities table that used to close the design-patterns
segment. The original table carried another SDK's names; they are Go here.

This is the split everything else hangs on. Say the rule in one sentence and
come back to it every time a code slide appears: if it can fail for external
reasons, it is an Activity.

Do not read the rows out. Name the two that matter, Determinism and the last
row: "replays from history" versus "restarts from scratch" is the whole
difference in five words. The rest earns its keep in the PDF.

The Retries row, if asked. A Workflow Execution is not retried by default. You
can give it a `RetryPolicy` in `StartWorkflowOptions`, but that is rare and
usually a smell. What does retry is the Workflow Task: if the Workflow panics
because of a bug, Temporal retries the Workflow Task, so you can fix the code
and redeploy while the Workflow waits. Business failures are handled in the
Workflow's own code, or by returning an error, which fails the Workflow. A
retried Activity starts from the top, unless it heartbeats progress.

One nuance if someone challenges the Time row: `workflow.Now(ctx)` and
`workflow.Sleep(ctx, d)` *are* available inside a Workflow, and they are
deterministic because the SDK drives them from Workflow time rather than the
wall clock. It is `time.Now()` and `time.Sleep()` that are banned, not the
concept of time. The same goes for randomness: draw it inside
`workflow.SideEffect`, which records the value in Event History so replay reads
it back instead of drawing again.
-->

---
layout: default
---

# The Worker runs both

Workflow code runs under special constraints.

- **Workflow code must be deterministic.** Same inputs, same history, same
  decisions, every time it runs.
- **The Worker checks this.** A Workflow runs as **coroutines** the SDK
  schedules one at a time, and on replay every Command it makes is compared
  with the Event History. Use `workflow.Go`, `workflow.Now` and
  `workflow.Sleep`, not goroutines, `time.Now` and `time.Sleep`.

<div v-click class="mt-8 text-2xl font-light">

If it can fail for external reasons, it's an Activity.

</div>

<!--
Source: PPT slide 25.

The original said "the Worker enforces this with different mechanisms,
depending on the SDK." In Go the mechanism is the SDK's own scheduler and
replay. The Worker runs each Workflow on coroutines it controls: only one runs
at a time, and it switches between them only where the Workflow blocks on a
`workflow` API (a Future's `Get`, a channel `Receive`, `workflow.Sleep`,
`workflow.Await`). That is why a Workflow uses `workflow.Go` instead of the `go`
keyword, `workflow.Channel` and `workflow.Selector` instead of Go channels and
`select`, and `workflow.Now` and `workflow.Sleep` instead of the `time`
package: those are the versions the scheduler can replay identically.

Go does not stop you importing `net/http` or `os` into a Workflow. What catches
it is replay: if the code makes a different Command from the one Event History
recorded, the Workflow Task fails with a non-determinism error. Two tools catch
it earlier: the `workflowcheck` static analyzer (from
`go.temporal.io/sdk/contrib/tools/workflowcheck`) flags non-deterministic calls
in Workflow code at build time, and replaying production histories in a test
catches the rest. The Versioning segment comes back to both.

That is also why the Workflow refers to Activities as function values and
never calls them. `workflow.ExecuteActivity(ctx, Withdraw, t)` takes the
function only for its name and signature; it schedules an Activity Task. It
appears on the Workflow anatomy slide shortly; plant it now.
-->

---
layout: code-walkthrough
heading: One Worker, both kinds of code
notes:
  - The one Task Queue this Worker polls for work
  - "Workflow code: run on the SDK's <strong>deterministic scheduler</strong>, replayed from history"
  - "Activity code: <strong>plain Go</strong>, on ordinary goroutines, free to touch the world"
  - Polls until it is interrupted. Nothing runs without a Worker
---

::code::

```go {all|1|2|3-4|6}
w := worker.New(c, "transfers", worker.Options{})
w.RegisterWorkflow(TransferWorkflow)
w.RegisterActivity(Withdraw)
w.RegisterActivity(Deposit)

err := w.Run(worker.InterruptCh())
```

<!--
Code: snippets/why/worker.go (compiled; the tests in snippets/why/ build the
same Worker with the same registrations against the dev server).

The previous slide in code. One process, two very different sets of rules, and
you can see the split in the two kinds of registration:

- **Build 1 - the Task Queue.** The Worker asks the Service for Tasks on
  `transfers`, and only those. The Client will start Workflows on the same
  name. `c` is a `client.Client` from `client.Dial`.
- **Build 2 - `RegisterWorkflow`.** The Workflow function is registered by
  name, and the Worker runs it on its own coroutines, one at a time, replaying
  from Event History whenever it needs to rebuild state. That is where the
  rules from the previous slide apply.
- **Build 3 - `RegisterActivity`.** Ordinary functions, run on ordinary
  goroutines in this process. No rules: HTTP, databases, the file system.
  Register a struct pointer instead and every exported method becomes an
  Activity, which is how Activities get their dependencies.
- **Build 4 - `Run`.** Long-polls until the process gets an interrupt, then
  shuts down. The Server never runs your code; if no Worker is polling, Tasks
  wait on the queue.

Keep it short: the Go Workers section takes the whole file apart, connection
and all. This slide is only about the two halves.
-->
