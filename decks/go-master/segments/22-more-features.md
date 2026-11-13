---
layout: section
---

<ZiggyHero scene="orbit">

# More Features

Beyond the fundamentals: more of what the platform does for you.

</ZiggyHero>

<!--
Ported from the More Features section of the Essentials Deck, with every API
checked against the Go SDK (go.temporal.io/sdk v1.51.0).

A tour, not a lab: features people ask about once the fundamentals land. Each
gets a "what it is, why it matters" slide, then its code where code helps.
Nothing after this section depends on it. The divider names no features, so
adding one to the section means adding its slides, not editing this one.
-->

---
layout: default
---

# Schedules

<WhatWhy>
  <template #what>

A Schedule starts a Workflow on a timetable: a cron expression, a fixed
interval, or a calendar. It lives on the Temporal Service, not on a machine.

  </template>
  <template #why>
    <li><strong>Nothing to babysit.</strong> No cron host whose outage silently skips a run</li>
    <li><strong>Every run is a Workflow Execution</strong>, with retries, Event History and a place in the UI</li>
    <li><strong>Operable.</strong> Pause, trigger now, backfill a missed window, or update it, from code, the CLI or the UI</li>
    <li><strong>An overlap policy</strong> decides what happens when the last run is still going</li>
  </template>
</WhatWhy>

<!--
Source: docs.temporal.io/develop/go/schedules.

The comparison to land is cron. A crontab entry runs a command on one machine;
if that machine is down at 9:00, the run is gone and nobody is told. A
Schedule is state on the Temporal Service, and what it starts is an ordinary
Workflow Execution, so everything from the morning applies: retries, history,
visibility.
-->

---
layout: code-walkthrough
heading: Schedules in Go
notes:
  - A Schedule lives on the Temporal Service under its own ID. No cron box to keep alive
  - "The action: start this Workflow, on this Task Queue, each time the Schedule fires"
  - "The spec: when. Cron expressions, intervals, or calendars"
  - "The policy: what to do if the last run is still going. <code>SKIP</code> is the default"
---

::code::

```go {all|1-2|3-6|7-9|10}
handle, err := c.ScheduleClient().Create(ctx, client.ScheduleOptions{
	ID: "daily-report",
	Action: &client.ScheduleWorkflowAction{
		Workflow:  ReportWorkflow,
		TaskQueue: "reports",
	},
	Spec: client.ScheduleSpec{
		CronExpressions: []string{"0 9 * * *"},
	},
	Overlap: enumspb.SCHEDULE_OVERLAP_POLICY_SKIP,
})
```

<!--
Source: docs.temporal.io/develop/go/schedules.
Code: snippets/features/schedules.go, run by snippets/features/features_test.go,
which creates this Schedule, checks its next run is at 09:00 UTC, and triggers
it once to watch a Workflow run.

`c` is the `client.Client` from `client.Dial`; `client` is the SDK package.
`enumspb` is `go.temporal.io/api/enums/v1`.

Each run is a normal Workflow Execution with its own history, retries and
visibility. Set `ID` on the action to choose the Workflow ID (it defaults to a
UUID); Temporal appends the scheduled time to it, so runs do not collide.

Everything after creation goes through the `client.ScheduleHandle`, or
`c.ScheduleClient().GetHandle(ctx, "daily-report")` later: `Pause`, `Unpause`,
`Trigger` to run once now, `Backfill` to run for a missed window, `Update`,
`Describe` and `Delete`. The same verbs are on `temporal schedule` in the CLI.
Creating a Schedule whose ID already exists returns
`temporal.ErrScheduleAlreadyRunning`; check for it with `errors.Is` rather than
describing first.

`Intervals: []client.ScheduleIntervalSpec{{Every: time.Hour}}` is the
alternative to a cron string, and `Calendars` covers "every Wednesday at 20:30".

Prefer Schedules over the older `CronSchedule` field on
`client.StartWorkflowOptions`. The docs recommend them: they can be paused,
updated and backfilled.
-->

---
layout: default
---

# Serverless Workers

<WhatWhy>
  <template #what>

Workers that run on serverless compute (AWS Lambda, GCP Cloud Run). Temporal
starts them when Tasks arrive and lets them stop when the work drains.
_Public Preview._

  </template>
  <template #why>
    <li><strong>No Worker fleet</strong> to provision, scale, patch or keep running idle</li>
    <li><strong>Scales with demand</strong>, down to zero, so you pay for compute only while work is in flight</li>
    <li><strong>Same SDK, same code.</strong> Workflows and Activities register exactly as they do today</li>
    <li><strong>Spillover capacity.</strong> Share a Task Queue with long-lived Workers and absorb bursts</li>
  </template>
</WhatWhy>

<!--
Source: docs.temporal.io/serverless-workers. AWS Lambda and GCP Cloud Run
support are in Public Preview; check the docs before delivering.

The case: running a Worker means choosing hosting, sizing compute, writing
scaling policies and building a deploy pipeline before the first Workflow runs
in production. A Serverless Worker hands invocation and scaling to Temporal and
the cloud provider. It fits bursty, event-driven and low-volume work best;
steady high-throughput work is usually cheaper on long-lived Workers.
-->

---
layout: code-walkthrough
heading: Serverless Workers in Go
notes:
  - "<code>RunWorker</code> is the function's whole <code class=\"ref\">main</code>. Temporal invokes it when Tasks arrive and no Worker is free"
  - A Deployment Version is required. Every Serverless Worker uses Worker Versioning
  - Same registration as any Worker. A Task Queue and the Activities
  - "<code>RegisterWorkflow</code>: Workflows are compiled into the same binary. Nothing to bundle, so a cold start is a process start"
---

::code::

```go {all|2|3-4|6,8|7}
func main() {
	lambdaworker.RunWorker(worker.WorkerDeploymentVersion{
		DeploymentName: "orders",
		BuildID:        "build-1",
	}, func(opts *lambdaworker.Options) error {
		opts.TaskQueue = "orders"
		opts.RegisterWorkflow(OrderWorkflow)
		opts.RegisterActivity(ChargePayment)
		return nil
	})
}
```

<!--
Source: the `go.temporal.io/sdk/contrib/aws/lambdaworker` module (v0.2.0) and
docs.temporal.io/serverless-workers. The module is versioned separately from
the SDK and is in Public Preview; APIs may change before the stable release.
Code: snippets/features/lambda/main.go. `go vet` compiles it against the
module; it only runs inside Lambda, so no test runs it.

How it works: a server component watches the Task Queue. When a Task arrives
and no Worker is free to take it, it invokes your Lambda. Each invocation dials
Temporal, starts a Worker, processes Tasks, and shuts down gracefully before
the Lambda deadline (`ShutdownDeadlineBuffer` on the options sets the margin).
No work, no compute.

Nothing to bundle: Workflow code is ordinary Go compiled into the function's
binary, so build it for Lambda (`GOOS=linux GOARCH=arm64 go build`) like any
other Go function. The module sets Lambda-sized Worker defaults (low
`MaxConcurrent*` limits) and loads the connection from `temporal.toml` and
`TEMPORAL_*` environment variables through `contrib/envconfig`; credentials can
come from AWS Secrets Manager. Set `opts.ClientOptions` or `opts.WorkerOptions`
in the callback to override anything. Workflows default to the `PINNED`
versioning behavior. Setup the slide leaves out: an IAM role Temporal can
assume to invoke the function, and a compute provider on the Worker Deployment
Version.
-->

---
layout: default
---

# AI integrations

<p class="ai-lede">
Durable Execution for agents: the agent loop runs as a Workflow, and every
model and tool call runs as an Activity. Agents survive crashes, retry
rate-limited models, and never pay twice for a call that already finished.
</p>

<div class="ai-groups">
  <section>
    <h3>Agent frameworks</h3>
    <div class="chips">
      <span>Google ADK · <code>contrib/googleadk</code></span>
      <span>Any model client, called from an Activity</span>
    </div>
  </section>
  <section>
    <h3>Agent observability</h3>
    <div class="chips">
      <span>LangSmith</span>
      <span>Braintrust</span>
      <span>Laminar</span>
      <span>Parseable</span>
    </div>
  </section>
  <section>
    <h3>Gateways and memory</h3>
    <div class="chips">
      <span>OpenRouter</span>
      <span>xMemory</span>
    </div>
  </section>
  <section class="own">
    <h3>Not on the list?</h3>
    <p>Most of these are built on the SDK's <strong>Plugin system</strong>. You can build your own the same way.</p>
  </section>
</div>

<style>
.ai-lede {
  max-width: 46rem;
  font-size: 1rem !important;
  line-height: 1.5;
  color: var(--temporal-text-strong);
}
.ai-groups {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}
.ai-groups section {
  border-radius: 0.7rem;
  padding: 0.8rem 1rem;
  background: var(--temporal-bg-elev);
  border: 1px solid var(--temporal-border-hi);
}
.ai-groups h3 {
  margin: 0 0 0.55rem !important;
  font-size: 0.66rem !important;
  font-weight: 500;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--temporal-green);
}
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}
.chips span {
  font-size: 0.78rem;
  padding: 0.2rem 0.65rem;
  border-radius: 999px;
  border: 1px solid var(--temporal-lavender);
  background: rgb(154 128 255 / 14%);
  color: #fff;
}
.own p {
  margin: 0 !important;
  font-size: 0.85rem !important;
  color: var(--temporal-text-strong);
}
</style>

<!--
Source: docs.temporal.io's AI integrations pages, as of 2026-11-13. The list
grows, and not every integration is available in every SDK; check what the Go
SDK has before delivering and add what is new.

Why agents need this: an agent is a long-running loop of model calls and tool
calls. Models are rate-limited and flaky, tools call real systems, and a crash
halfway through a 20-step task is expensive to start over. An LLM call is I/O,
so it belongs in an Activity, which is the rule from the Design Patterns
section. The integrations do that wiring for you.

The one to show a Go room is Google ADK: `go.temporal.io/sdk/contrib/googleadk`
runs an ADK agent's model and tool calls as Activities, so the agent loop is a
Workflow and every call is retried and recorded. Without a framework the Go
shape is the Design Patterns one: call the model provider's Go client from an
Activity, and keep the loop in the Workflow; that is the second chip. The
observability tools and gateways are services, maintained by those projects;
check which of them have a Go integration today before promising one.
-->

---
layout: default
---

# Priority and Fairness

<WhatWhy>
  <template #what>

Two controls over the order Tasks leave a shared Task Queue. **Priority**:
levels 1 to 5, and higher-priority Tasks always go first. **Fairness**:
within a level, tenants take turns in proportion to their weight.

  </template>
  <template #why>
    <li><strong>Urgent work cuts the line.</strong> A payment no longer waits behind a bulk import</li>
    <li><strong>No noisy neighbours.</strong> One big tenant's backlog can't starve the small ones</li>
    <li><strong>One Worker pool</strong> serves every tier, with no Task Queue per tenant to configure and route</li>
    <li><strong>New tenants are just new keys.</strong> No new queues or Workers to deploy</li>
  </template>
</WhatWhy>

<!--
Source: docs.temporal.io/develop/task-queue-priority-fairness and
docs.temporal.io/design-patterns/fairness.

The classic workaround for both problems is a Task Queue per tier or per
tenant. That means a Worker pool, routing logic and idle capacity for every
one of them. Priority and Fairness do the same job inside one queue.

Priority is on by default, on Temporal Cloud and self-hosted. Fairness is
enabled per Namespace, and on Temporal Cloud it is a paid feature.
-->

---
layout: code-walkthrough
heading: Priority and Fairness in Go
notes:
  - "<code>PriorityKey</code>: 1 to 5, lower dispatches first, default 3. Payments cut ahead of batch work"
  - "<code>FairnessKey</code>: one virtual queue per tenant, dispatched in proportion to <code>FairnessWeight</code>. A big tenant can't starve the rest"
---

::code::

```go {all|5|6-7}
func OrderWorkflow(ctx workflow.Context, order Order) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		Priority: temporal.Priority{
			PriorityKey:    1,
			FairnessKey:    order.TenantID,
			FairnessWeight: 2,
		},
	})
	var receipt string
	err := workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, &receipt)
	return receipt, err
}
```

<!--
Source: docs.temporal.io/develop/task-queue-priority-fairness.
Code: snippets/features/features.go, run by snippets/features/features_test.go,
which runs this Workflow and finds all three values on the
ActivityTaskScheduled Event.

- **Priority** is strict. All priority-1 Tasks dispatch before the first
  priority-2 Task.
- **Fairness** works within a priority level. Each key is a virtual queue, and
  keys take turns in proportion to their weights: weights of 5, 3 and 2 give
  roughly 50%, 30% and 20% of dispatches while all three have a backlog.

The Activity options are set inside the Workflow here because the fairness key
comes from the input. Set `Priority` the same way on
`client.StartWorkflowOptions` or `workflow.ChildWorkflowOptions`. Activities and
Child Workflows inherit their Workflow's Priority unless they set their own.

Neither is a rate limit or an isolation boundary. For hard isolation, use a
separate Task Queue with its own Workers.
-->
