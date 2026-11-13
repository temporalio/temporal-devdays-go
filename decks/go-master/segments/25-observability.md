---
layout: section
---

<ZiggyHero scene="observe">

# Observability

Knowing it's healthy before someone tells you it isn't.

</ZiggyHero>

<!--
Source: PPT slides 168-186.

Most of this segment is Temporal Cloud and SDK metrics, which are the same
however your Workers are built. The Go-specific parts are at the end: giving
the Client a Prometheus-backed `MetricsHandler` and a `Logger`, and logging from
Workflows and Activities.
-->

---
layout: default
---

# Observability objectives

- Ensuring high availability and maintaining application health
- Diagnosing and troubleshooting performance bottlenecks
- Detecting failures and errors
- Optimizing Worker performance
- Making informed operational decisions

<!--
Source: PPT slide 169.
-->

---
layout: default
---

# SDK vs Cloud metrics

<div class="cols-2">
<div>
<p class="col-eyebrow">SDK metrics</p>

**Emitted by your Clients and Workers**

- Performance of your code on your Workers
- Error rates at the Client and Worker level
- Cache use within Workers
- Resource use and availability on Worker hosts

</div>
<div>
<p class="col-eyebrow">Cloud metrics</p>

**Emitted by Temporal Cloud**

- Your workload, and Worker fleet health against Task Queues
- Service request latency and counts
- Sync vs async Task matching rates
- Service availability and error rates
- Throttling and rate limits imposed by the server

</div>
</div>

<!--
Source: PPT slide 170.

The closing line from the original: SDK and Cloud metrics together give a
holistic view of your Temporal application. Effective monitoring usually means
correlating the two to diagnose performance issues. A slow Workflow is either
slow in your code (SDK metrics) or slow getting to your code (Cloud metrics),
and you need both to tell which.
-->

---
layout: section
---

<ZiggyHero scene="gauge">

# Key SDK metrics

</ZiggyHero>

<!--
Source: PPT slide 171.

The Go SDK names every metric with a `temporal_` prefix, so
`worker_task_slots_available` is `temporal_worker_task_slots_available` in your
Prometheus. The slides use the full names. One Go-specific twist: through the
Prometheus naming scope (the code slide at the end of this segment), every
latency becomes a histogram with a `_seconds` suffix, so
`temporal_activity_execution_latency` is scraped as
`temporal_activity_execution_latency_seconds_bucket` and friends. Dashboards
built for the Go SDK expect that.
-->

---
layout: default
---

# Worker throughput and Task execution

- **Worker slot availability** · `temporal_worker_task_slots_available`
  - **The goal:** always have headroom. If this sits at zero, the Worker is
    saturated and cannot accept new Tasks, and processing is delayed
- **Activity latency** · `temporal_activity_execution_latency`
  - **The insight:** time spent inside your Activity code
  - **Action:** persistent spikes point to inefficient code, slow database
    queries, or a slow downstream API

<!--
Source: PPT slide 172.
-->

---
layout: default
---

# Workflow performance

**The 10-second rule:** by default a Workflow Task must complete within 10
seconds. If it doesn't, check two metrics:

- `temporal_workflow_task_replay_latency`: how long it takes to rebuild state from history
- `temporal_workflow_task_execution_latency`: how long it takes to run new decisions

If these are high, your history is probably too large. **Use Continue-As-New**
to break it up and cut what the Worker has to process.

<!--
Source: PPT slide 173.

In Go there is a second common cause of high execution latency: CPU work in
Workflow code that never blocks. A Workflow's coroutines run one at a time, and
the Workflow Task cannot finish until each of them is blocked on a Future, a
Channel, a Selector or `workflow.Await`. The Worker's deadlock detector stops
Workflow code that runs for one second without yielding, and fails the Workflow
Task with "[TMPRL1101] Potential deadlock detected: workflow goroutine "root"
didn't yield for over 1s". The Workflow stays open and the Task is retried, so
it fails again until you fix the code (or move the work into an Activity).
`DeadlockDetectionTimeout` on `worker.Options` raises the limit; setting
`TEMPORAL_DEBUG=true` lifts it while you sit in a debugger.
-->

---
layout: default
---

# Cache optimization and Worker memory

**The sticky cache:** your Worker caches Workflow state so it doesn't have to
download the whole Event History from Temporal for every Task.

- `temporal_sticky_cache_total_forced_eviction`: should stay stable and low
- `temporal_sticky_cache_size`: watch it to right-size the Worker's memory

<!--
Source: PPT slide 174.

In Go the cache size is `worker.SetStickyWorkflowCacheSize`, and it is one
cache for the whole process: every Worker in it shares the 10,000 Workflows it
holds by default. Call it before the first Worker starts. Forced evictions
climbing means the cache is too small for the number of open Workflows this
process handles: each eviction costs a full replay next time.
-->

---
layout: default
---

# Poller health and connectivity

- `temporal_num_pollers`: total active long-polls from your Worker to Temporal
- `temporal_long_request_latency`: a long-poll should stay open for nearly a
  minute while it waits for work. Low values here often mean your network is
  cutting connections early
- `temporal_cloud_v1_poll_timeout_count`: on an idle Task Queue, timeouts are
  expected and healthy. Under steady load, poll success rate (success over
  success plus timeouts) should be above 90%

<!--
Source: PPT slide 175.

The counterintuitive one is the last: on an idle Task Queue, almost every poll
should time out empty. That is a healthy connection doing its job. Under load,
the docs target a poll success rate above 90% (above 95% for high volume); a
low success rate under load often means more pollers than the work needs.

The original slide used the `temporal_cloud_v0_*` name. The v0 (PromQL)
metrics were retired when that endpoint shut down on 2026-10-05; every Cloud
metric in this segment uses the v1 OpenMetrics name.
-->

---
layout: default
---

# The gold-standard metrics

- **Schedule-to-start latency** · `temporal_activity_schedule_to_start_latency`,
  `temporal_workflow_task_schedule_to_start_latency`
  - How long a Task sits in the Task Queue before a Worker picks it up. Keep p90
    under 700ms. If it climbs, **scale your Workers**
- **Sync match rate** (calculated): ideally above 95%. The Task was handed
  straight to a waiting poller without ever touching the database
- **Service latency and poll success** (calculated): the handshake between your
  network and Temporal Cloud is seamless

<!--
Source: PPT slide 176.

If you only alert on one thing, alert on schedule-to-start. It is the single
best signal that you do not have enough Workers.
-->

---
layout: section
---

<ZiggyHero scene="cloud">

# Cloud-side observability

</ZiggyHero>

<!--
Source: PPT slide 177.
-->

---
layout: default
---

# Cloud metrics: resource exhaustion

`temporal_cloud_v1_resource_exhausted_error_count`

- **What it measures:** requests rejected because one resource (a Namespace,
  Task Queue, or Workflow ID) got a burst larger than it could absorb
- **Why it matters:** a hot resource. SDKs retry these, so progress is rarely
  hit, but persistent non-zero values are unexpected
- **Not the same as throttling:** for APS/RPS limits, watch
  `temporal_cloud_v1_total_action_throttled_count` and
  `temporal_cloud_v1_service_request_throttled_count`
- **Actions:** use the `operation` label to find the hot RPC, smooth the burst

<!--
Source: PPT slide 178.

The original slide described this metric as APS/RPS throttling. The OpenMetrics
reference says the opposite: it counts per-resource bursts and "does not include
throttling due to Namespace limits". A common example: starting the same
Workflow ID more than once per second.
-->

---
layout: default
---

# Cloud metrics: frontend service errors

`temporal_cloud_v1_service_error_count`

- **What it measures:** errors reported by the Temporal Cloud frontend service API
- **Why it matters:** indicates communication issues between your Clients and Temporal Cloud
- **Signs of trouble:** success rate below 99.9%
- **Actions:** check the metric's `operation` and `status_code` tags to find the
  type of error, and verify network connectivity between your Clients and
  Temporal Cloud

<!--
Source: PPT slide 179. The original used the v0 name
`temporal_cloud_v0_frontend_service_error_count`; its v1 equivalent is
`temporal_cloud_v1_service_error_count`.
-->

---
layout: section
---

<ZiggyHero scene="plug">

# Observability: integrations

</ZiggyHero>

<!--
Source: PPT slide 181.
-->

---
layout: default
---

# Metrics endpoints

<div class="cols-2">
<div>
<p class="col-eyebrow">OpenMetrics endpoint</p>

A Prometheus-compatible, **scrapable** endpoint. Ideal for pull-based tools
(Grafana Cloud, Prometheus, and so on).

</div>
<div>
<p class="col-eyebrow">PromQL endpoint · retired</p>

The older Prometheus **query API** endpoint. Shut down on 2026-10-05: anything
still pointing at it gets no data. Move it to OpenMetrics.

</div>
</div>

<!--
Source: PPT slide 182.

OpenMetrics exposes Cloud metrics for every Namespace in your account, as
pre-computed rates at `https://metrics.temporal.io/v1/metrics`, authenticated
with a Metrics Read-Only API key. SDK metrics are separate: you turn them on in
your Worker code (later in this segment).

The PromQL column is here for rooms with dashboards that went quiet in
October 2026: that is why. Drop it once nobody asks.
-->

---
layout: default
---

# Integration: Grafana Cloud

<div class="cols-2">
<div>
<p class="col-eyebrow">What you get</p>

- A pre-built, always-updated dashboard for Workflow health, Task Queue depth, and usage
- Dimensions: Namespace, Task Queue, Workflow Type, operation
- Adaptive Telemetry for cost control, and Grafana Assistant for natural-language queries

</div>
<div>
<p class="col-eyebrow">Setup</p>

1. Create a service account API key with the Metrics Read-Only role
2. In Grafana Cloud, open Connections and find the Temporal Cloud tile
3. Configure a scrape job with the API key
4. No mTLS certificates or self-hosted Prometheus needed

</div>
</div>

<!--
Source: PPT slide 183.
-->

---
layout: default
---

# Integration: Datadog

<div class="cols-2">
<div>
<p class="col-eyebrow">What you get</p>

- A ready-made dashboard for health, efficiency, and performance
- Built-in monitors: p99 latency, Workflow failures, replication lag
- Standard, not custom, metrics: lower Datadog billing

</div>
<div>
<p class="col-eyebrow">Setup</p>

1. Create a service account with the Metrics Read-Only role, and an API key
2. In Datadog, install the **Temporal Cloud - OpenMetrics** integration
3. Add your account and paste the API key
4. No certificates needed

</div>
</div>

<!--
Source: PPT slide 184. The original described the deprecated integration
(root CA upload, certificates). The current one uses an API key and the
`temporal.cloud.v1_*` metric prefix. The old tile stopped receiving data when the
PromQL endpoint shut down on 2026-10-05.
-->

---
layout: default
---

# Integration: self-hosted Prometheus

<div class="cols-2">
<div>
<p class="col-eyebrow">What you get</p>

- Every Cloud metric, scraped from the OpenMetrics endpoint
- Temporal's OpenMetrics Grafana dashboard, on GitHub
- Feeds an OpenTelemetry Collector or your own alerting

</div>
<div>
<p class="col-eyebrow">Setup</p>

1. Create a Metrics Read-Only service account and API key
2. Add a Prometheus scrape job for the OpenMetrics endpoint, with the key
   as a Bearer token
3. Import the OpenMetrics dashboard into Grafana

</div>
</div>

<!--
Source: PPT slide 185. The original title was "self-hosted Prometheus and
Grafana"; Grafana is the dashboard on top of the Prometheus data source.

The original setup (mTLS certificates, a CA upload, PromQL as a Grafana data
source) was for the deprecated PromQL endpoint. The current path is a scrape
job against `https://metrics.temporal.io/v1/metrics` with a Bearer API key,
and the dashboard is `temporal_cloud_openmetrics.json` in temporalio/dashboards.
The migration guide at
docs.temporal.io/cloud/metrics/openmetrics/migration-guide has the full
`scrape_configs` example. OpenMetrics v1 values are pre-computed rates, so use
`sum()`, not `rate()` or `increase()`, in queries.
-->

---
layout: code-walkthrough
heading: "Integration: SDK metrics in Go"
notes:
  - Telemetry is configured on the <strong>Client</strong>. Every Worker built from this <code>client.Dial</code> inherits it
  - The client's <code>Logger</code> receives every SDK, Workflow and Activity log line
  - Serve a Prometheus <code class="ref">/metrics</code> endpoint from the Worker Process on port 9000
  - <code>NewPrometheusNamingScope</code> gives every latency the <code class="ref">_seconds</code> suffix Prometheus expects
---

::code::

```go {all|15-19|17|2-5|18}
func Dial(hostPort string) (client.Client, error) {
	reporter, err := prometheus.Configuration{
		ListenAddress: "0.0.0.0:9000",
		TimerType:     "histogram",
	}.NewReporter(prometheus.ConfigurationOptions{})
	if err != nil {
		return nil, err
	}
	scope, _ := tally.NewRootScope(tally.ScopeOptions{
		CachedReporter:  reporter,
		Separator:       prometheus.DefaultSeparator,
		SanitizeOptions: &sdktally.PrometheusSanitizeOptions,
	}, time.Second)

	return client.Dial(client.Options{
		HostPort:       hostPort,
		Logger:         log.NewStructuredLogger(slog.Default()),
		MetricsHandler: sdktally.NewMetricsHandler(sdktally.NewPrometheusNamingScope(scope)),
	})
}
```

<!--
Source: PPT slide 186.
Code: snippets/observability/metrics.go, run by
snippets/observability/observability_test.go.

What you get: schedule-to-start latency, poll success, sync match; Task slots
available, sticky cache size and evictions; request latency and failures; and
Activity Execution and Workflow Execution metrics. Everything on the SDK
metrics slides, with the same `temporal_` names.

The imports, for anyone copying this: `github.com/uber-go/tally/v4`, its
`prometheus` reporter package, `sdktally "go.temporal.io/sdk/contrib/tally"`
and `go.temporal.io/sdk/log`. The `contrib/tally` package is a small module of
its own beside the SDK.

- **Build 1 - the Client.** In Go, telemetry belongs to the Client. A
  `worker.New(c, ...)` takes the client's Logger and MetricsHandler, so dial
  once at startup and build every Worker in the process from that one client.
  The same client also reports the metrics of the Workflows it starts.

- **Build 2 - `Logger`.** Anything that implements `log.Logger` works here.
  `log.NewStructuredLogger` adapts the standard library's `slog`, so Worker
  logs go wherever the rest of your service's `slog` output goes, JSON handler
  and all. Leave it out and the SDK logs to stdout with its default logger.

- **Build 3 - the exporter.** The tally Prometheus reporter starts its own
  HTTP listener and serves `/metrics`. From here, scrape it with Prometheus,
  point a Datadog Agent at it, or run an OpenTelemetry Collector. For push
  instead of pull, build the handler from an OpenTelemetry meter with
  `go.temporal.io/sdk/contrib/opentelemetry` instead.

- **Build 4 - the naming scope.** `NewMetricsHandler` adapts the tally scope
  to the SDK's `client.MetricsHandler`. `NewPrometheusNamingScope` appends
  `_seconds` to every timer, so the latencies on the earlier slides arrive as
  `temporal_activity_execution_latency_seconds` and so on. The snippet test
  scrapes port 9000 and finds every metric this segment names.

Metrics are not traces. For distributed tracing across Client, Workflow and
Activity, add the tracing interceptor from `go.temporal.io/sdk/contrib/opentelemetry`
(or `contrib/opentracing`, `contrib/datadog`) to `client.Options.Interceptors`:
it creates spans and propagates the trace context through Temporal's headers.
Search Attributes and the Temporal UI cover the other question operators ask,
"which Workflows are in this state right now?"
-->

---
layout: code-walkthrough
heading: Logging from a Workflow
notes:
  - <code>workflow.GetLogger</code> comes from the Workflow context, not the <code class="ref">log</code> or <code class="ref">slog</code> packages
  - Log with key-value pairs. The Worker adds the WorkflowID, RunID, WorkflowType and TaskQueue
  - Not repeated on replay. This line runs again on every replay, but logs once
---

::code::

```go {all|2|3|3,12}
func OrderWorkflow(ctx workflow.Context, orderID string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Order received", "OrderID", orderID)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})
	var receipt string
	if err := workflow.ExecuteActivity(ctx, ChargeCard, orderID).Get(ctx, &receipt); err != nil {
		return "", err
	}
	logger.Info("Order charged", "OrderID", orderID, "Receipt", receipt)
	return receipt, nil
}
```

<!--
Source: new for this Deck; no PPT slide. Ported alongside the observability
segment.
Code: snippets/observability/workflows.go, run by
snippets/observability/observability_test.go.

- **Build 1 - the logger.** Workflow code must survive replay, so it does not
  write to the `log` or `slog` packages directly. `workflow.GetLogger(ctx)`
  hands back the client's Logger from the previous slide, wrapped for this
  Workflow Execution.

- **Build 2 - key-value pairs.** Pass alternating keys and values rather than
  building a string with `fmt.Sprintf`. Every entry also carries `WorkflowID`,
  `RunID`, `WorkflowType`, `Namespace`, `TaskQueue` and `Attempt`, so you can
  filter one execution's lines out of a busy Worker's logs.

- **Build 3 - replay.** When a Worker replays this Workflow (after a cache
  eviction, or on a different Worker), both `logger.Info` lines execute again.
  The logger drops calls made during replay, so each line appears once per
  execution, not once per replay. The snippet test proves it by replaying the
  finished Workflow's Event History with `worker.NewWorkflowReplayer`: the
  Workflow runs from the top, and logs nothing. (`EnableLoggingInReplay` on
  `worker.Options` turns replay logging on, for debugging.)
-->

---
layout: code-walkthrough
heading: Logging from an Activity
notes:
  - <code>activity.GetLogger</code> takes the Activity's <code>ctx</code>
  - Each line carries the Activity type, the attempt number and the Workflow it belongs to
---

::code::

```go {all|2|3-5}
func ChargeCard(ctx context.Context, orderID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Charging card", "OrderID", orderID)
	receipt := "receipt-" + orderID
	logger.Info("Card charged", "Receipt", receipt)
	return receipt, nil
}
```

<!--
Source: new for this Deck; no PPT slide. Ported alongside the observability
segment.
Code: snippets/observability/activities.go, run by
snippets/observability/observability_test.go.

- **Build 1 - the logger.** The `context.Context` the Worker passes to an
  Activity carries the running Activity's identity; `activity.GetLogger(ctx)`
  reads it. Like the Workflow logger, it ends up at the client's Logger.

- **Build 2 - context.** Each entry carries `ActivityType`, `ActivityID`,
  `Attempt`, `TaskQueue`, and the parent `WorkflowID`, `RunID` and
  `WorkflowType`. Activities are not replayed: each retry is a new attempt, and
  logs again with the attempt number, which is exactly what you want when you
  debug a flaky downstream call.
-->
