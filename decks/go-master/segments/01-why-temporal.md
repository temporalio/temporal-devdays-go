---
layout: section
---

<ZiggyHero scene="juggle">

# A simple feature

Let customers send money. How hard can it be?

</ZiggyHero>

<!--
The arc of this segment: a feature that fits in two lines, everything
production does to it, the plumbing you write in response, and then the same
two lines on Temporal. Everything after this segment is a consequence of that
last part.

One example the whole way through: a bank transfer. The Workflows segment and
the first lab build on the same transfer, so nothing here is thrown away.
-->

---
layout: default
---

# The feature request

<div class="request">
<p class="request-head">As a customer, I want to send money to another account.</p>
<div class="steps">
<div class="step"><span class="n">1</span><span><strong>Withdraw</strong> $500 from Alice</span></div>
<span class="arrow">→</span>
<div class="step"><span class="n">2</span><span><strong>Deposit</strong> $500 to Bob</span></div>
</div>
<p class="request-done"><strong>Done when:</strong> the money left one account and arrived in the other.</p>
</div>

<style scoped>
.request {
  margin-top: 1.5rem;
  border: 1px solid var(--temporal-border-hi);
  border-radius: 0.6rem;
  padding: 1.5rem 1.75rem;
}
.request-head {
  font-size: 1.3rem;
  margin: 0 0 1.5rem !important;
  color: var(--temporal-text-strong);
}
.steps {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}
.step {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.8rem;
  border: 1px solid var(--temporal-border-hi);
  border-radius: 0.5rem;
  padding: 0.9rem 1rem;
  font-size: 1.05rem;
}
.step .n {
  display: grid;
  place-items: center;
  width: 1.8rem;
  height: 1.8rem;
  border-radius: 999px;
  background: rgba(89, 253, 160, 0.15);
  color: var(--temporal-green);
  font-weight: 600;
}
.arrow {
  font-size: 1.6rem;
  color: var(--temporal-text);
}
.request-done {
  margin: 1.5rem 0 0 !important;
  font-size: 1rem;
}
</style>

<!--
The opening story is a bank transfer.

Read it like a product manager would. One sentence, two steps, one acceptance
criterion. Ask the room how long they would estimate. Someone will say an
afternoon. Hold onto that number.
-->

---
layout: code-stack
heading: The happy path
proseMax: 30
---

::code::

```go
func SendMoney(ctx context.Context, t Transfer) error {
	if err := bank.Withdraw(ctx, t.FromAccount, t.Amount); err != nil {
		return err
	}
	return bank.Deposit(ctx, t.ToAccount, t.Amount)
}
```

::default::

Two calls, and both of them are the feature. The demo works, the tests pass,
the PR is approved. Ship it?

<!--
Illustrative, not compiled (it parses; the tests check). `bank` is the bank's
API client.

Let it sit. This is what everyone writes first, and it is correct: it is
exactly the business requirement. Ask "ship it?" and wait. Someone always says
"what about errors?" That is the cue for the next slide.
-->

---
layout: default
---

# Then it meets production

<div class="failures">
<div v-click class="failure"><p class="what">The bank's API is down</p><p class="then">Every transfer fails for ten minutes. Or hangs, and never returns.</p></div>
<div v-click class="failure"><p class="what">The process crashes between the two lines</p><p class="then">Alice was debited. Bob was never credited.</p></div>
<div v-click class="failure"><p class="what">A power outage, or a deploy</p><p class="then">The node goes away mid-transfer, and so does everything in memory.</p></div>
<div v-click class="failure"><p class="what">A race condition</p><p class="then">Alice double-taps Send. Two requests, two transfers.</p></div>
<div v-click class="failure"><p class="what">A bug in the deposit code</p><p class="then">Every transfer since the release is stuck halfway.</p></div>
<div v-click class="failure"><p class="what">Insufficient funds</p><p class="then">Not an outage. Retrying will never fix it.</p></div>
</div>

<style scoped>
.failures {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
  margin-top: 1.25rem;
}
.failure {
  border: 1px solid var(--temporal-border-hi);
  border-left: 3px solid #ff6b6b;
  border-radius: 0.5rem;
  padding: 0.8rem 1rem;
}
.what {
  font-weight: 600;
  color: var(--temporal-text-strong);
  margin: 0 0 0.4rem !important;
  font-size: 1rem;
}
.then {
  margin: 0 !important;
  font-size: 0.88rem;
}
</style>

<!--
One click per failure. For each, ask the room whether they have been paged for
it. The second one always gets a story.

- **The bank is down or slow.** The call returns an error, or worse, never returns.
- **A crash between the lines.** The money left Alice and never reached Bob.
  This is the one to land; the next slide is only this.
- **Power outage, or a deploy.** A rolling deploy is a planned crash. Whatever
  was in memory is gone.
- **A race.** A double tap, a mobile retry on a flaky connection, a load
  balancer replaying a request: two transfers.
- **A bug.** Not the network: your code. You fix it in an hour. What happens to
  the transfers that failed in that hour?
- **Insufficient funds.** The odd one out. Retrying makes it worse, not better.
  Some errors must never be retried.
-->

---
layout: center
class: text-center
---

<p style="font-size: 2.4rem; font-weight: 200; line-height: 1.3">The withdrawal succeeded.<br />Then the process died.</p>

<p v-click style="font-size: 3rem; font-weight: 300; margin-top: 2rem; color: #ff6b6b">$500 left Alice. It never reached Bob.</p>

<!--
Say the first line, pause, click. Nothing in the two-line version remembers
that the withdrawal happened. When the process comes back, it has no idea
there is a transfer to finish.

This is not an edge case at scale. Every deploy kills processes mid-request.
-->

---
layout: default
---

# And the questions start

<div class="questions">
<div class="q"><strong>When do we retry?</strong><span>A 503? A timeout? A connection reset?</span></div>
<div class="q"><strong>When don't we?</strong><span>Insufficient funds. A closed account.</span></div>
<div class="q"><strong>How many times?</strong><span>Three? Ten? Until it works?</span></div>
<div class="q"><strong>How long between attempts?</strong><span>Exponential backoff? With jitter? Capped at what?</span></div>
<div class="q"><strong>Where do we remember how far we got?</strong><span>So a restart does not withdraw twice.</span></div>
<div class="q"><strong>Who finds out when it never succeeds?</strong><span>And how do they fix it?</span></div>
</div>

<style scoped>
.questions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem 1.5rem;
  margin-top: 1.25rem;
}
.q {
  border: 1px solid var(--temporal-border-hi);
  border-radius: 0.5rem;
  padding: 0.75rem 1rem;
}
.q strong {
  display: block;
  color: var(--temporal-text-strong);
  font-size: 1.02rem;
}
.q span {
  font-size: 0.88rem;
}
</style>

<!--
Every one of these is a real decision, and the two-line function answers none
of them. Read the left column fast, then slow down on the last two: they are
not about retries at all, and they are the ones that turn a function into a
system.

None of these questions is about sending money. All of them have to be answered
before you can ship it.
-->

---
layout: code-stack
heading: So you add a retry
proseMax: 30
---

::code::

```go {all|8-11|12-14|15-16|2-3}
func withRetry(call func() error) error {
	delay := time.Second
	for attempt := 1; ; attempt++ {
		err := call()
		if err == nil {
			return nil
		}
		var nsf *InsufficientFundsError
		if errors.As(err, &nsf) {
			return err
		}
		if attempt == 5 {
			return err
		}
		time.Sleep(delay + rand.N(250*time.Millisecond))
		delay = min(delay*2, 30*time.Second)
	}
}
```

::default::

Eighteen lines to answer three of the questions. All of it in memory: when the
process dies, so do the attempt count, the delay, and the transfer.

<!--
Illustrative, not compiled. `rand` is `math/rand/v2`.

- **Build 1 - when not to retry.** Insufficient funds is final. Every new
  error type the bank invents is another line here.
- **Build 2 - how many times.** Five, because five felt right. Nobody tuned it.
- **Build 3 - how long between.** Doubling, capped at 30 seconds, with jitter so
  a thousand failed transfers do not all retry in the same millisecond. This is
  the part people are proud of. Every service in the company has its own copy,
  slightly different.
- **Build 4 - where the state lives.** A local variable. The retry loop is the
  first thing to die when the process does.

And it answers none of the last three questions on the previous slide.
-->

---
layout: code-walkthrough
heading: The plumbing version
notes:
  - A <strong>distributed lock</strong>, so a double-tap never runs one transfer twice. Die holding it and the transfer waits out the TTL
  - Your own <strong>state table</strong>. Reload progress so a restart skips finished steps
  - A <strong>checkpoint</strong> around every step. Crash between the withdraw and the save, and Alice is debited twice
  - The retry helper from the last slide, round every call
  - A <strong>timeout</strong> on every call, so a hung bank cannot hang you
  - An <strong>audit trail</strong> you write, and nobody reads until the incident
  - Failures parked for a human, in a <strong>dead-letter queue</strong> you also run
  - And a <strong>recovery cron</strong> for the transfers that get stuck anyway
---

::code::

```go {all|2-6|14|18-31|19-23|20-21|30|7-12|36}
func SendMoney(ctx context.Context, t Transfer) (err error) {
	lock, err := locks.Acquire(ctx, t.TransferID, 30*time.Second)
	if err != nil {
		return err
	}
	defer lock.Release(ctx)
	defer func() {
		if err != nil {
			db.MarkFailed(ctx, t.TransferID, err)
			deadLetters.Push(ctx, t.TransferID)
		}
	}()

	state, err := db.LoadTransfer(ctx, t.TransferID)
	if err != nil {
		return err
	}
	if state.Step < StepWithdrawn {
		err = withRetry(func() error {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			return bank.Withdraw(ctx, t.FromAccount, t.Amount)
		})
		if err != nil {
			return err
		}
		if err = db.SaveStep(ctx, t.TransferID, StepWithdrawn); err != nil {
			return err
		}
		audit.Append(ctx, t.TransferID, "withdrawn")
	}
	// ...and the same block again for the deposit
	return nil
}

// recovery.go: a cron that finds transfers stuck halfway and re-runs them
```

<!--
Illustrative, not compiled: the happy path, written to survive production
without Temporal. "The Temporal version" a few slides on is the same transfer;
the two are meant to be compared. The deposit's block is elided to one comment
so the slide fits; it is another fourteen lines of the same.

Each bubble is one of the questions, answered by hand:

- **Build 1 - the lock.** The race condition. Two requests for one transfer;
  with the lock, only one runs. A pod that crashes holding it stalls the
  transfer for the TTL.
- **Build 2 - the state table.** "Where do we remember how far we got?" In a
  table you design, migrate and back up.
- **Build 3 - the checkpoint.** Do the step, then record it. The gap between
  the withdrawal and `SaveStep` is exactly the crash from three slides ago,
  moved a few lines down. Closing it needs idempotency keys on the bank calls
  too.
- **Build 4 - the retry.** The helper from the last slide.
- **Build 5 - the timeout.** A `context.WithTimeout` round every call. Pick a
  number for each one.
- **Build 6 - the audit trail.** So someone can answer "what happened to my
  money?" Rows you write yourself.
- **Build 7 - failure handling.** "Who finds out when it never succeeds?" A
  deferred function that parks the failure on a queue, and someone to drain it.
- **Build 8 - recovery.** For everything above that still goes wrong: a second
  program whose only job is finding the first one's stuck work.

The bug from the failures slide is still not handled: a transfer that hit the
bad deposit code is now in the dead-letter queue, and someone replays it by
hand after the fix ships.
-->

---
layout: default
---

# Most of the code is not the feature

<div class="grid grid-cols-2 gap-10 mt-8">
<div>
<p class="loc-eyebrow">Happy path</p>
<p class="loc-count">6 <span>lines</span></p>
<p><strong>Two calls are the feature.</strong> It does not survive a single failure.</p>
</div>
<div v-click>
<p class="loc-eyebrow">Production-ready</p>
<p class="loc-count">65+ <span>lines</span></p>
<p><strong>The same two lines of feature,</strong> plus a lock, a state table, an audit log, a dead-letter queue, a recovery cron, and a retry helper to maintain.</p>
</div>
</div>

<style scoped>
.loc-eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.8rem;
  color: var(--temporal-green);
  margin-bottom: 0.25rem;
}
.loc-count {
  font-size: 4rem;
  font-weight: 200;
  line-height: 1;
  margin: 0 0 1rem !important;
  color: var(--temporal-text-strong);
}
.loc-count span {
  font-size: 1.4rem;
  color: var(--temporal-text);
}
</style>

<!--
The numbers are the last three slides: the happy path is 6 lines; the plumbing
version is 34 on the slide plus the 14 it elides for the deposit, and the retry
helper 18. That is before the cron, the queue consumer and three table schemas,
hence the plus.

Remember the afternoon estimate from the feature request? This is the version
that takes a sprint, and it is still the part of the system nobody wants to
touch.
-->

---
layout: center
class: text-center
---

<p style="font-size: 2.6rem; font-weight: 200; line-height: 1.2">What if you only wrote the happy path?</p>

<p v-click style="font-size: 3.2rem; font-weight: 300; margin-top: 2rem; color: var(--temporal-green)">That is Temporal.</p>

<!--
The thesis of the day. Say the question and let it hang, then click.

If someone says it sounds like magic, agree that it feels that way and promise
the mechanism: Event History and replay, in the concepts and Workflows
segments. Developers distrust magic, so do not leave it there.
-->

---
layout: code-walkthrough
heading: The Temporal version
notes:
  - Each step is its own <strong>Activity</strong>. Just the call. No lock, no retry loop, no checkpoint
  - "<strong>Timeouts</strong> are configuration: a hung bank call is abandoned and retried after 10 seconds"
  - "<strong>When not to retry</strong> is configuration too. Everything else retries, with exponential backoff, by default"
  - The Workflow is the <strong>happy path</strong>. The same two calls
---

::code::

```go {all|2-8|13|14-16|19-22}
// activities.go
func Withdraw(ctx context.Context, t Transfer) error {
	return bank.Withdraw(ctx, t.FromAccount, t.Amount)
}

func Deposit(ctx context.Context, t Transfer) error {
	return bank.Deposit(ctx, t.ToAccount, t.Amount)
}

// workflows.go
func TransferWorkflow(ctx workflow.Context, t Transfer) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			NonRetryableErrorTypes: []string{"InsufficientFundsError"},
		},
	})

	if err := workflow.ExecuteActivity(ctx, Withdraw, t).Get(ctx, nil); err != nil {
		return err
	}
	return workflow.ExecuteActivity(ctx, Deposit, t).Get(ctx, nil)
}
```

<!--
Code: snippets/why/activities.go and snippets/why/workflows.go, run by
snippets/why/why_test.go: a bank that fails the deposit twice still completes
the transfer with one withdrawal, and insufficient funds fails at once without
a retry.

Deliver it slowly, with the plumbing version in their heads.

- **Build 1 - the Activities.** The two bank calls, and nothing else. If the
  bank fails, the Activity returns the error; that is all of its error
  handling.
- **Build 2 - the timeout.** `StartToCloseTimeout` bounds each attempt. A call
  that hangs is abandoned and retried; the Activity's `ctx` is cancelled, so a
  bank client that honours it stops too.
- **Build 3 - when not to retry.** One field. A Go error is reported by its
  type name, so returning a `*InsufficientFundsError` matches the string
  `"InsufficientFundsError"`. The defaults answer the rest: a 1 second first
  retry, doubling, capped at 100 seconds, unlimited attempts.
- **Build 4 - the Workflow.** The happy path, call for call. The error check
  is ordinary Go, and the only error it sees is one the Retry Policy gave up
  on.

The claim to make out loud, because it separates Temporal from a retry library:
this survives the process dying. Kill the Worker after the withdrawal and
another Worker carries on with the deposit. The withdrawal is not repeated,
because it is already recorded in the Workflow's Event History.

Do not explain `ActivityOptions` and `ExecuteActivity` here. They get their own
walkthrough in the
Workflows segment.
-->

---
layout: default
---

# Where did the failures go?

<table class="where">
<thead><tr><th>In production</th><th>With Temporal</th></tr></thead>
<tbody>
<tr><td>The bank's API is down, or hangs</td><td>Retried with backoff until it is back; <code>StartToCloseTimeout</code> abandons a hung call</td></tr>
<tr><td>A crash between the two lines</td><td>Another Worker resumes the Workflow at the deposit. The withdrawal is not repeated</td></tr>
<tr><td>A power outage, or a deploy</td><td>Progress lives in the Temporal Service, not in your process</td></tr>
<tr><td>A double-tapped Send</td><td>Use the transfer ID as the Workflow ID: the second start is rejected</td></tr>
<tr><td>A bug in the deposit code</td><td>The deposit keeps retrying. Ship the fix, and the next attempt runs it</td></tr>
<tr><td>Insufficient funds</td><td>Listed as non-retryable: it fails at once</td></tr>
</tbody>
</table>

<style scoped>
.where {
  margin-top: 1rem;
  font-size: 0.9rem;
}
.where td:first-child {
  white-space: nowrap;
  color: var(--temporal-text-strong);
}
</style>

<!--
Go back to the failures slide, row by row. None of the right-hand column is
code you wrote.

- **Bank down.** The Retry Policy. The bank can be down for an hour; the
  transfer simply finishes late.
- **Crash between the lines.** Every completed Activity is recorded in Event
  History. A new Worker replays the Workflow, sees the withdrawal already
  happened, and schedules only the deposit. That replay is the mechanism behind
  the "only the happy path" promise.
- **Power outage.** Same mechanism. Workers are stateless; kill all of them and
  start them again.
- **Double tap.** Two running Workflows can never share an ID. The patterns
  section comes back to this.
- **A bug.** Unlimited retries are a feature here: the stuck transfers wait
  for your fix instead of landing in a dead-letter queue.
- **Insufficient funds.** The one line of configuration from the last slide.

The honest caveat: Activities run at least once, so the bank calls still need
an idempotency key. That has its own slide later in the day.
-->

---
layout: default
---

# Retries are configuration

<table class="policy">
<thead><tr><th>The question</th><th>Retry Policy field</th><th>Default</th></tr></thead>
<tbody>
<tr><td>How long before the first retry?</td><td><code>InitialInterval</code></td><td>1 second</td></tr>
<tr><td>How fast does the wait grow?</td><td><code>BackoffCoefficient</code></td><td>2.0</td></tr>
<tr><td>What is the longest wait?</td><td><code>MaximumInterval</code></td><td>100 × the initial interval</td></tr>
<tr><td>How many times?</td><td><code>MaximumAttempts</code></td><td>0, meaning unlimited</td></tr>
<tr><td>When don't we retry?</td><td><code>NonRetryableErrorTypes</code></td><td>None</td></tr>
</tbody>
</table>

<p class="policy-note">Declared once, by the caller. No loop, no helper, no state to lose.</p>

<style scoped>
.policy {
  margin-top: 1rem;
  font-size: 0.95rem;
}
.policy-note {
  margin-top: 1.25rem !important;
}
</style>

<!--
Source: docs.temporal.io/encyclopedia/retry-policies (default values).

Every question from "And the questions start" that was about retries, mapped to
one field. The defaults are sensible, which is why the Temporal version set
only one of them.

Point out that this is not a library default living in your process: the
Temporal Service tracks the attempt count and the next retry time, so a retry
schedule survives every Worker restarting.

`StartToCloseTimeout` is not on this table because it is not part of the Retry
Policy, but it is what decides that an attempt has failed by hanging.
-->

---
layout: default
---

# Just the code you already write

<div class="grid grid-cols-2 gap-10 mt-6">
<div>

- **No DSL.** No YAML, no JSON state machine, no visual designer.
- **Plain Go.** Functions, `if`, loops, `error` values, `defer`.
- **Your tools.** Your IDE, your debugger, your test runner, your CI.

</div>
<div>

- **Your packages.** Activities call any Go module, any API, any database.
- **Your architecture.** Start one Workflow from the service you already run.
- **One feature at a time.** No rewrite to adopt it.

</div>
</div>

<!--
The objection this answers before anyone raises it: "so I have to learn a new
workflow language." No. A Workflow is an exported Go function that takes a
`workflow.Context`, and an
Activity is any function.

The rule that comes with it, said now so it is not a surprise later: Workflow
code must be deterministic, which is why the bank calls live in Activities.
The design-patterns segment draws that line properly.

Adoption is the point of the last row. Teams usually start with one painful
flow (a transfer, an onboarding, a nightly batch), not a migration.
-->

---
layout: default
---

# And out of the box

<div class="features">
<div class="feature"><strong>Durable timers</strong><span><code>workflow.Sleep(ctx, 30*24*time.Hour)</code> survives every restart</span></div>
<div class="feature"><strong>Signals, Queries, Updates</strong><span>Talk to a Workflow while it runs</span></div>
<div class="feature"><strong>Full visibility</strong><span>Every run, its inputs, and its history in the Web UI</span></div>
<div class="feature"><strong>Schedules</strong><span>Start Workflows on a schedule, without a cron box</span></div>
<div class="feature"><strong>Child Workflows and Sagas</strong><span>Compose flows, and undo them when they fail</span></div>
<div class="feature"><strong>Versioning</strong><span>Change code while transfers are still running</span></div>
</div>

<style scoped>
.features {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
  margin-top: 1.25rem;
}
.feature {
  border: 1px solid var(--temporal-border-hi);
  border-left: 3px solid var(--temporal-green);
  border-radius: 0.5rem;
  padding: 0.8rem 1rem;
}
.feature strong {
  display: block;
  color: var(--temporal-text-strong);
  margin-bottom: 0.3rem;
}
.feature span {
  display: block;
  font-size: 0.88rem;
  line-height: 1.4;
}
.feature code {
  font-size: 0.8rem !important;
}
</style>

<!--
A tour of the agenda, not a feature list to explain. Each box is a section
later today; name it and move on.

- Durable timers: the Workflows segment.
- Signals, Queries, Updates: right after the first lab.
- Visibility: the next two slides, and every demo.
- Schedules: More features.
- Child Workflows and Sagas: the advanced topics.
- Versioning: its own segment.
-->

---
layout: default
---

# What you see in the Web UI

<div class="ui-timeline">
  <div class="row"><span class="lane">Workflow Execution</span><span class="track"><span class="bar ok" style="left: 0%; width: 100%">transfer-4471</span></span></div>
  <div class="row"><span class="lane">Activity</span><span class="track"><span class="bar ok" style="left: 0%; width: 48%">withdraw</span></span></div>
  <div class="row"><span class="lane">Activity</span><span class="track"><span class="bar ok" style="left: 50%; width: 50%">deposit</span></span></div>
</div>

<style scoped>
.ui-timeline {
  margin-top: 2rem;
  border: 1px solid var(--temporal-border-hi);
  border-radius: 0.6rem;
  overflow: hidden;
}
.row {
  display: grid;
  grid-template-columns: 11rem 1fr;
  border-bottom: 1px solid var(--temporal-border-hi);
}
.row:last-child {
  border-bottom: none;
}
.lane {
  padding: 0.7rem 0.9rem;
  font-size: 0.9rem;
  border-right: 1px solid var(--temporal-border-hi);
}
.track {
  position: relative;
  margin: 0.45rem 0.9rem;
}
.bar {
  position: absolute;
  top: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  padding: 0 0.7rem;
  border-radius: 999px;
  font-family: "Noto Sans Mono", monospace;
  font-size: 0.8rem;
  white-space: nowrap;
  overflow: hidden;
}
.bar.ok {
  border: 2px solid var(--temporal-green);
  background: rgba(89, 253, 160, 0.12);
}
</style>

<!--
A redrawing of the Web UI's timeline view for one transfer, the happy path.

The Workflow ID is the transfer ID, so support finds Alice's transfer by the
number on her receipt. Every Activity is a bar with a start, an end, and a
result. This is the audit trail the plumbing version wrote by hand.
-->

---
layout: default
---

# …including when the bank is down

<div class="ui-timeline">
  <div class="row"><span class="lane">Workflow Execution</span><span class="track"><span class="bar ok" style="left: 0%; width: 100%">transfer-4472</span></span></div>
  <div class="row"><span class="lane">Activity</span><span class="track"><span class="bar ok" style="left: 0%; width: 22%">withdraw</span></span></div>
  <div class="row"><span class="lane">Activity</span><span class="track"><span class="bar failed" style="left: 24%; width: 16%">deposit ⓘ</span><span class="bar failed" style="left: 42%; width: 16%">attempt 2</span><span class="bar failed" style="left: 60%; width: 16%">attempt 3</span><span class="bar ok" style="left: 78%; width: 22%">attempt 4</span></span></div>
</div>

<style scoped>
.ui-timeline {
  margin-top: 2rem;
  border: 1px solid var(--temporal-border-hi);
  border-radius: 0.6rem;
  overflow: hidden;
}
.row {
  display: grid;
  grid-template-columns: 11rem 1fr;
  border-bottom: 1px solid var(--temporal-border-hi);
}
.row:last-child {
  border-bottom: none;
}
.lane {
  padding: 0.7rem 0.9rem;
  font-size: 0.9rem;
  border-right: 1px solid var(--temporal-border-hi);
}
.track {
  position: relative;
  margin: 0.45rem 0.9rem;
}
.bar {
  position: absolute;
  top: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  padding: 0 0.7rem;
  border-radius: 999px;
  font-family: "Noto Sans Mono", monospace;
  font-size: 0.8rem;
  white-space: nowrap;
  overflow: hidden;
}
.bar.ok {
  border: 2px solid var(--temporal-green);
  background: rgba(89, 253, 160, 0.12);
}
.bar.failed {
  border: 2px solid #ff6b6b;
  background: rgba(255, 107, 107, 0.15);
}
</style>

<!--
A redrawing, not a screenshot: the real timeline shows one Activity bar with
its attempt count and last failure, one click from the stack trace. It is
drawn as separate attempts here so the room can see the retries.

The bank is down for three attempts. Nobody was paged, nothing went to a
dead-letter queue, and no one wrote a retry loop. The deposit succeeds on the
fourth attempt and the withdrawal ran once. This is the answer to "who finds
out when it never succeeds?": anyone with the Web UI open, with the error in
front of them.
-->
