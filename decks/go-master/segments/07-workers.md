---
layout: section
---

<ZiggyHero scene="poll">

# Go Workers

The process that polls. Nothing runs without one.

</ZiggyHero>

<!--
Source: PPT slides 62-70.

Worth saying before the code: the Worker is the only part of Temporal that runs
your code. The Server never sees it. That is the answer to the security question
somebody always asks.
-->

---
layout: default
---

# What is a Worker?

The part of your application that runs your Workflow and Activity code.

- **The Temporal Service is the manager.** It keeps state, schedules work and
  records what happened
- **The Worker does the work.** It polls a Task Queue ("any work for me?"), runs
  your code, and reports the result back
- **It runs on your infrastructure**: servers, containers, serverless. The
  Service never runs your code or calls your databases and APIs
- **It scales horizontally**: one on your laptop, hundreds in production

<!--
Sources: docs.temporal.io/evaluate/understanding-temporal#workers and
docs.temporal.io/encyclopedia/architecture/how-temporal-works#worker.

The misconception to kill: "Temporal runs my code." It does not. The Service
stores Event History and hands out Tasks; the code only ever runs inside a
Worker that you deploy, so your databases and APIs are only ever called from
inside your own network. That is the answer to the security question somebody
always asks. Be precise if it comes up: Workflow and Activity inputs and results
do travel to the Service, because they are stored in Event History. Encrypting
them with a Payload Codec has its own segment later.

Polling is the other word to land. Workers pull work; the Service never pushes
to them. So a Worker needs only outbound access to the Service, no open
inbound port, and adding capacity means starting another Worker on the same
Task Queue. Nothing to register, no load balancer to reconfigure.
-->

---
layout: code-walkthrough
heading: Go Workers
notes:
  - "<code>client.Dial</code> connects to the Temporal Service at <code>localhost:7233</code>. The Worker polls through this Client"
  - Polls one specific Task Queue. It must match the Task Queue the Client starts Workflows on
  - "<code>worker.Options</code> sizes the goroutine pools that run Workflow and Activity Tasks. The defaults are fine to start"
  - "Workflows are registered by <strong>function value</strong>, and run in this process as ordinary Go"
  - Registers the Activity functions this Worker can run
  - "<code>Run</code> polls until <code>InterruptCh</code> fires on Ctrl-C or SIGTERM, then shuts down"
---

::code::

```go {all|2|8|8|9|10|12}
func main() {
	c, err := client.Dial(client.Options{HostPort: "localhost:7233"})
	if err != nil {
		log.Fatalln("unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "order-task-queue", worker.Options{})
	w.RegisterWorkflow(orders.OrderWorkflow)
	w.RegisterActivity(orders.SubmitOrder)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("unable to start worker", err)
	}
}
```

<!--
Source: PPT slides 63-70.
Code: snippets/workers/main.go (compiled and vetted, never run: it would poll
localhost:7233). snippets/workers/main_test.go builds the same Worker against
the test's dev server, runs a Workflow through it, and checks `Run` returns
cleanly when its interrupt channel fires.

The original had seven steps: service stubs, Client, factory, Worker, register
Workflows, register Activities, start. Go folds them into one `client.Dial`,
one `worker.New` and two registrations. This is the `worker/main.go` a learner
starts with `go run ./worker`.

- **Build 1 - the connection.** A Worker polls through an ordinary
  `client.Client`, the same type your application uses to start Workflows; a
  process that does both can share one. Leave `HostPort` empty and it dials
  `localhost:7233` on its own. For Temporal Cloud you add `Namespace`, an API
  key or TLS in `client.Options`; the rest of the file does not change.
- **Build 2 - the Task Queue.** A name, nothing more. The Client puts Tasks on
  it, the Worker takes them off. Typo it on one side and the Workflow starts
  fine and waits forever with no error. Say that out loud; it is the most common
  first-day bug. The test proves it: a start on `order-taks-queue` is accepted
  and still Running two seconds later.
- **Build 3 - the options.** This is where the Worker's concurrency lives. A Go
  Worker runs each Workflow Task and each Activity on a goroutine, bounded by
  `MaxConcurrentWorkflowTaskExecutionSize` and
  `MaxConcurrentActivityExecutionSize`, and fetches work with a handful of
  pollers per queue (`MaxConcurrentWorkflowTaskPollers`,
  `MaxConcurrentActivityTaskPollers`). The defaults suit a laptop and most
  first deployments; the Worker tuning segment is where to change them.
- **Build 4 - Workflows.** Registered by value: `RegisterWorkflow` takes the
  function, and its name becomes the Workflow Type. There is no bundling and no
  separate runtime: the Workflow runs in this process, in coroutines the SDK
  schedules one at a time, which is why the determinism rules are yours to
  keep. The Worker keeps recently run Workflows in a cache shared by every
  Worker in the process (`worker.SetStickyWorkflowCacheSize`), so a Workflow
  that wakes up usually carries on without replaying from the start.
- **Build 5 - Activities.** `RegisterActivity` takes a function, or a pointer
  to a struct to register all of its exported methods. If your Activities need
  a shared HTTP client or database pool, put it on that struct and build it
  here. The Worker only runs Activities it was given.
- **Build 6 - `Run`.** Starts the pollers and blocks. `worker.InterruptCh()`
  is a channel that fires on SIGINT or SIGTERM, so Ctrl-C or a container stop
  starts a graceful shutdown and `Run` returns. Set `WorkerStopTimeout` in
  `worker.Options` so in-flight Activities get time to finish; it defaults to
  zero, which cancels them straight away.

If asked about scale: run as many Workers on this Task Queue as you like. They
are stateless and interchangeable. That is the whole horizontal-scaling story.
-->
