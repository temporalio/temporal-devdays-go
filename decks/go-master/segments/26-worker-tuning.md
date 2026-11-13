---
layout: section
---

<ZiggyHero scene="tune">

# Worker tuning

</ZiggyHero>

<!--
Source: PPT slides 187-206.

The defaults on these slides are the Go SDK's, checked against
`go.temporal.io/sdk` v1.51.0. If you deliver this against a newer SDK, re-check
them; the snippets package compiles every option name the slides use
(snippets/tuning/options.go).
-->

---
layout: default
---

# How a Worker works

Workers poll a Task Queue in Temporal, execute Tasks, and respond with the result.

<img src="/images/worker-internals.png" alt="Inside a Temporal Worker: Workflow and Activity Task pollers poll their Task Queues in Temporal Cloud, hand Tasks to handlers with execution slots, cache Workflows, and respond with the results" class="h-[340px] mx-auto rounded" />

<!--
Source: PPT slides 188-191.

Walk the numbers on the diagram:

1. Workflow Task pollers long-poll the Workflow Task Queue.
2. A Task arrives and is handed to the Workflow Task handler, which needs a free
   execution slot.
3. If this Worker has the Workflow cached, it loads the state from the cache
   rather than replaying the whole history.
4. After the Task, the Workflow state goes back in the cache.
5. The Worker responds with the commands the Workflow produced.

Activity Tasks work the same way, minus the cache.

In a Go Worker, everything on the left of the diagram is goroutines. Each
poller is a goroutine holding a long-poll open. Each Activity runs on its own
goroutine, and the Go scheduler spreads them across every core the process may
use. Each Workflow runs as a set of coroutines (the root one plus any
`workflow.Go` started) that the SDK runs one at a time, in a fixed order, so
replay always takes the same path.

The original slides 189-190 asked "what happens if a Worker goes down while
executing a Workflow?" Answer: another Worker picks up the next Workflow Task,
replays the Event History to rebuild state, and carries on. Nothing is lost,
because nothing lived only in that Worker.
-->

---
layout: default
---

# The Workflow event loop

<img src="/images/workflow-event-loop.png" alt="The Workflow event loop: initialize, then repeatedly progress Signals and Updates until blocked, progress the Workflow until blocked, execute Queries, and finally complete or continue as new" class="h-[400px] mx-auto rounded" />

<!--
Source: PPT slide 192.

What one Workflow Task does inside the Worker. Signals, Updates and Queries
arrive from outside and are handled at defined points. Handlers run first, then
the Workflow's coroutines run until every one of them is blocked on something
(a Timer, an Activity's Future, a Channel, a Selector, `workflow.Await`), then
Queries are answered. Repeat until the Workflow completes or continues as new.

"Until blocked" is why a CPU-heavy loop in Workflow code is so harmful: nothing
else in that Workflow runs until it blocks. Other Workflows on the Worker run
on their own goroutines, so they carry on, but this one holds its Workflow Task
slot. After one second without yielding, the deadlock detector gives up and
fails the Workflow Task ("[TMPRL1101] Potential deadlock detected: workflow
goroutine "root" didn't yield for over 1s"); the Task is retried and fails the
same way until the code changes. The snippet test runs exactly that and
checks the Workflow is still open afterwards.
-->

---
layout: section
---

<ZiggyHero scene="lift">

# Worker resources

</ZiggyHero>

<!--
Source: PPT slide 193.
-->

---
layout: default
---

# Theory of constraints

<img src="/images/theory-of-constraints.png" alt="Adding capacity upstream of a bottleneck sends more work to it, so performance at the bottleneck gets worse and system throughput goes down" class="h-[380px] mx-auto rounded" />

<!--
Source: PPT slide 194.

The warning before the tuning knobs: adding capacity anywhere but the
bottleneck makes things worse. More Workflow slots in front of a saturated
database means more Activities queueing on that database. Find the constraint
first. The three places it usually lives are the next slide.
-->

---
layout: section
---

<ZiggyHero scene="trio">

# Memory · IO · Compute

</ZiggyHero>

<!--
Source: PPT slide 195.

Every Worker is bounded by one of these three. The next four slides take them in
turn.
-->

---
layout: default
---

# Compute settings

| Setting                                   |   Go default |
| ----------------------------------------- | -----------: |
| `MaxConcurrentWorkflowTaskExecutionSize`  |        1,000 |
| `MaxConcurrentActivityExecutionSize`      |        1,000 |
| `MaxConcurrentLocalActivityExecutionSize` |        1,000 |
| `spec.containers[].resources.limits.cpu`  | yours to set |

<!--
Source: PPT slides 196-197.

These are execution slots: how many Tasks of each kind this Worker will run at
once. They are fields of `worker.Options`, passed to `worker.New`.

Two Go-specific notes:

- Every Activity runs on its own goroutine, and the Go scheduler runs them in
  parallel across `GOMAXPROCS` threads. For Activities that mostly wait on IO,
  1,000 concurrent is cheap: a blocked goroutine costs a few kilobytes. But
  1,000 CPU-heavy Activities (hashing, image processing, a big
  `json.Unmarshal`) all compete for the same cores, every one of them slows
  down, and their heartbeats arrive late. Give CPU-heavy work its own Worker on
  its own Task Queue with a low `MaxConcurrentActivityExecutionSize`, sized to
  the cores you have.
- Workflow Tasks for different Workflows run in parallel too, so CPU-heavy
  Workflow code uses every core rather than queueing behind one. Since Go 1.25,
  `GOMAXPROCS` follows the container's CPU limit on Linux, so the limit you set
  in the last row is what the scheduler sees.
-->

---
layout: code-stack
heading: Execution slot auto-tuning
proseMax: 36
---

::code::

```go
tuner, err := worker.NewResourceBasedTuner(worker.ResourceBasedTunerOptions{
	TargetMem:    0.8,
	TargetCpu:    0.8,
	InfoSupplier: sysinfo.SysInfoProvider(),
})
if err != nil {
	return nil, err
}
w := worker.New(c, "order-task-queue", worker.Options{
	Tuner: tuner,
})
w.RegisterWorkflow(OrderWorkflow)
w.RegisterActivity(SubmitOrder)
```

::default::

Part of Worker autotuning, which aims to get the most out of each Worker's
resources. Slots are handed out by how close the Worker is to a resource
limit: the percentage of CPU and memory in use.

<!--
Source: PPT slide 198.
Code: snippets/tuning/worker.go, run by snippets/tuning/tuning_test.go.

`sysinfo.SysInfoProvider()` comes from `go.temporal.io/sdk/contrib/sysinfo`, a
small module beside the SDK that reads CPU and memory use (from cgroups inside
a container). The tuner needs it, which keeps that dependency out of Workers
that never tune.

`Tuner` is mutually exclusive with the `MaxConcurrent*ExecutionSize` options on
the previous slide. Pick one approach: `worker.New` panics if you set both.

For most workloads Temporal's docs recommend fixed-size slots, tuned to your
hardware: they perform better and behave more predictably. The resource-based
tuner suits fluctuating workloads where each Task uses little CPU or memory but
may wait a long time (HTTP calls, other IO), and works as protection against
out-of-memory crashes when per-Task use is unpredictable. Work with very high
or highly variable per-Task cost belongs on fixed slots, tuned by hand.

One catch: `temporal_worker_task_slots_available` only reports for fixed-size
slots. With a resource-based tuner, watch `temporal_worker_task_slots_used`
instead.

To mix them, build a composite tuner with a slot supplier per Task type:
`worker.NewCompositeTuner(worker.CompositeTunerOptions{...})` with
`WorkflowSlotSupplier` from `worker.NewFixedSizeSlotSupplier(1000)`,
`ActivitySlotSupplier` from `worker.NewResourceBasedSlotSupplier(...)`, plus
`LocalActivitySlotSupplier` and `NexusSlotSupplier` for the Nexus Operations
a `temporalnexus` Worker handles.
-->

---
layout: default
---

# Memory settings

| Setting                                     |              Go default |
| ------------------------------------------- | ----------------------: |
| `worker.SetStickyWorkflowCacheSize`         | 10,000, per process     |
| `GOMEMLIMIT`                                | none                    |
| `spec.containers[].resources.limits.memory` | yours to set            |

<!--
Source: PPT slides 199-200.

`worker.SetStickyWorkflowCacheSize` is the sticky cache from the observability
segment. Each cached Workflow holds its state in memory, so the right number
depends on how big your Workflows are. In Go it is one cache for the whole
process, shared by every Worker in it, and the setting must be made before the
first Worker starts. Watch `temporal_sticky_cache_size` against container
memory and adjust.

`GOMEMLIMIT` is the Go runtime's soft memory limit. Set it a little below the
container limit (say 90%) and the garbage collector works harder as the heap
approaches it, instead of letting the heap grow until the container is killed.
It does not shrink the cache for you: if the live Workflows do not fit, lower
the cache size.
-->

---
layout: default
---

# IO settings

| Setting                                         |        Go default |
| ----------------------------------------------- | ----------------: |
| `WorkflowTaskPollerBehavior` (maximum)          |                 2 |
| `ActivityTaskPollerBehavior` (maximum)          |                 2 |
| `TaskQueueActivitiesPerSecond`                  |         unlimited |
| Namespace APS                                   | set per Namespace |

<!--
Source: PPT slides 201-202.

Pollers are the IO side: how many long-polls the Worker keeps open to the
server, each one a goroutine. The default is 2 of each. The older
`MaxConcurrentWorkflowTaskPollers` and `MaxConcurrentActivityTaskPollers`
options set the same simple maximum; the poller behaviours
(`worker.NewPollerBehaviorSimpleMaximum`) are the newer way.

`worker.NewPollerBehaviorAutoscaling(worker.PollerBehaviorAutoscalingOptions{})`
lets the server tell the Worker how many pollers it needs (between 1 and 100
by default, starting at 5). Worth trying before hand-tuning the maximum.

`TaskQueueActivitiesPerSecond` is a server-side limit across every Worker on
the Task Queue, which is how you protect a fragile downstream API.
`WorkerActivitiesPerSecond` is the per-Worker equivalent. Namespace APS
(Actions Per Second) is a Temporal Cloud limit on your Namespace, set by your
account.
-->

---
layout: default
---

# Failures

- **Network failures**
- **Worker crash** (for example, out of memory)

<!--
Source: PPT slide 203.

Both are what Temporal is designed to survive, and neither loses a Workflow. But
both show up in the metrics first: poll failures and request latency for the
network, and a sawtooth in `temporal_sticky_cache_size` for a Worker that keeps
being OOM-killed and restarted. In Go, an out-of-memory crash is usually the
kernel killing the container at its memory limit (exit code 137, no Go stack
trace), which is what `GOMEMLIMIT` on the previous slide helps you avoid.
-->

---
layout: default
---

# What is sync match?

**`poll_success_sync_count`**

A sync match is when a Task is assigned to a Worker **immediately**. If no Worker
is available, Temporal Cloud persists the Task in its database first (an "async
match").

- **Sync match rate** is the proportion of Tasks assigned immediately
- 100% means lower latency and a more responsive system
- A low rate often means there aren't enough Workers for the incoming load
- Also consider scaling on **schedule-to-start latency**, a related metric

<!--
Source: PPT slide 204. From the reference guide on monitoring, alerting and
health.
-->

---
layout: default
---

# Compute settings to metrics

- `MaxConcurrentWorkflowTaskExecutionSize` → `temporal_worker_task_slots_available`
  with `worker_type="WorkflowWorker"`
- `MaxConcurrentActivityExecutionSize` → `temporal_worker_task_slots_available`
  with `worker_type="ActivityWorker"`
- CPU-intensive logic → `temporal_workflow_task_execution_latency`

Don't forget to measure your machine's own CPU use, for example
`container_cpu_usage_seconds_total`.

<!--
Source: PPT slide 205.

How to read this: each setting on the left has a metric on the right that tells
you whether it is the constraint. Slots available near zero means raise the
slot count, or add Workers. High execution latency with slots to spare means
the CPU is the constraint, and more slots will only make it worse (the Theory of
constraints slide).

For CPU-heavy Activities in Go, also watch the scheduler: the Prometheus Go
collector can export `go_sched_latencies_seconds` (from the standard library's
`runtime/metrics`), how long runnable goroutines wait for a thread. When it climbs, goroutines are
queueing for CPU before it shows up in Temporal's metrics.
-->

---
layout: default
---

# Key takeaways

Scaling Workers is case by case, but some starting rules of thumb:

<div class="text-[0.85rem] leading-snug">

- **Schedule-to-start latency** rising, Workers at capacity: scale up
- **Task slots available** at zero for a while: scale up
- **Task Queue backlog**: watch it through the control-plane APIs
- **CPU and memory**: the signal for CPU-heavy Activity Workers
- **High volume**: book a tuning session with your Temporal solutions engineer

</div>

<!--
Source: PPT slide 206.

The bullets are the short form; say the rest:

- Schedule-to-start latency (or sync match rate) is where most resource
  contention shows up first. Rising latency with Workers at capacity means
  scale up the pool.
- Task slots available dropping to zero for a while means the Workers are fully
  used.
- Task Queue backlog gives more direct control. There is a community-built,
  KEDA-based autoscaler that reads it.
- CPU and memory are a good scaling metric for Activity Workers doing CPU-heavy
  work, though less directly tied to Temporal's own load.
-->
