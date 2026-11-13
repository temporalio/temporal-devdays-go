---
layout: default
class: "!p-0"
---

<div class="absolute inset-0 flex flex-col items-center justify-center gap-6">

  <div class="flex items-center gap-2">
    <img
      v-click
      :src="'/mia.jpg'"
      alt="Mia, a black and tan rescue dog"
      class="mia-photo w-[248px] h-[248px] rounded-full object-cover"
      style="object-position: 50% 30%"
    />
    <img
      :src="'/nikolay.jpg'"
      alt="Nikolay Advolodkin"
      class="w-[248px] h-[248px] rounded-full object-cover"
    />
  </div>

  <h1 class="intro-name">Nikolay Advolodkin</h1>

  <div class="flex flex-col items-center gap-1">
    <p class="intro-line">Staff Developer Advocate @ Temporal</p>
    <p class="intro-line">Dog Dad &amp; Roller Skater</p>
  </div>

  <div class="flex items-center gap-10">
    <a class="intro-social" href="https://x.com/nikolay_a00">
      <img :src="'/icon-x.png'" alt="X" class="w-[34px] h-[34px] rounded-md" />
      <span>@nikolay_a00</span>
    </a>
    <a class="intro-social" href="https://www.linkedin.com/in/nikolayadvolodkin/">
      <img :src="'/icon-linkedin.png'" alt="LinkedIn" class="w-[34px] h-[34px] rounded-md" />
      <span>/in/nikolayadvolodkin/</span>
    </a>
  </div>

</div>

<style scoped>
/* v-click only fades opacity, which leaves Mia's 248px reserved and pushes the
   headshot off-centre before the click. Collapsing her instead keeps the first
   state centred and lets her pop in. */
.mia-photo.slidev-vclick-hidden {
  display: none;
}
.intro-name {
  font-size: 2.9rem;
  font-weight: 200;
  letter-spacing: -0.02em;
  line-height: 1;
  color: var(--temporal-text-strong);
  margin: 0;
}
.intro-line {
  font-size: 1.15rem;
  color: var(--temporal-text);
  margin: 0;
}
.intro-social {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 1.1rem;
  color: var(--temporal-text);
  text-decoration: none;
  /* The theme underlines every link with a border-bottom; not wanted here. */
  border-bottom: none;
}
.intro-social:hover {
  color: var(--temporal-green);
}
</style>

<!--
Intro slide, rebuilt from slide 2 of the Agentic AI Summit 2026 deck. Same
elements: both circular photos, name, role, the Dog Dad line, and the two social
handles with their icons.

Keep it short. Thirty seconds. The room came for Nexus, not for a bio.

Mia appears again on the lab slide at the end of this segment, which is a nice
callback if you want one.
-->

---
layout: full-image
image: /nexus-intro/04.webp
alt: "Temporal-based services in an application"
---

<!--
Set the scene: several Temporal services in one application, each owned by a
different team.

Rendered page 4 of the [External] Temporal Nexus Overview Slides 2026 deck.
-->

---
layout: full-image
image: /nexus-intro/05.webp
alt: "How do they communicate?"
---

<!--
The question the whole segment answers. Let it sit for a beat.

Rendered page 5 of the [External] Temporal Nexus Overview Slides 2026 deck.
-->

---
layout: full-image
image: /nexus-intro/06.webp
alt: "Before: child workflow and activity wrapper"
---

<!--
The two things people reach for today: a child workflow, or an activity that
wraps a call to another service.

Rendered page 6 of the [External] Temporal Nexus Overview Slides 2026 deck.
-->

---
layout: full-image
image: /nexus-intro/07.webp
alt: "Before: no clean API contract"
---

<!--
Child workflow first. The killer line is the last one: blast radius is not
isolated. That is the same argument this workshop's payments monolith makes.

Rendered page 7 of the [External] Temporal Nexus Overview Slides 2026 deck.
-->

---
layout: full-image
image: /nexus-intro/08.webp
alt: "Before: overly permissive namespace access"
---

<!--
Activity wrapper next. Broad namespace access is the cost, and mTLS cert
management is the part nobody budgets for.

Rendered page 8 of the [External] Temporal Nexus Overview Slides 2026 deck.
-->

---
layout: full-image
image: /nexus-intro/09.webp
alt: "Before: overly permissive namespace access, continued"
---

<!--
Same slide, built out: three hops, two of them wrappers, and full namespace
write access at both ends.

Rendered page 9 of the [External] Temporal Nexus Overview Slides 2026 deck.
-->

---
layout: full-image
image: /nexus-intro/10.webp
alt: "Before: bespoke gateways to connect workflows"
---

<!--
The third option: build a gateway. Note it is not durable, which is the point
that sets up Nexus.

Rendered page 10 of the [External] Temporal Nexus Overview Slides 2026 deck.
-->

---
layout: full-image
image: /nexus-intro/11.webp
alt: "Now: Nexus endpoints and service contracts, open source and Temporal Cloud"
---

<!--
The payoff. Endpoints and service contracts, in open source and Cloud.

Rendered page 11 of the [External] Temporal Nexus Overview Slides 2026 deck.
-->

---
layout: full-image
image: /nexus-intro/12.webp
alt: "Now: Nexus on Temporal Cloud"
---

<!--
What Cloud adds on top: global registry, multi-cloud routing, mTLS mesh,
access controls, metrics, audit logs.

Rendered page 12 of the [External] Temporal Nexus Overview Slides 2026 deck.
-->

---
layout: default
clicks: 10
---

# One payment, end to end

<NexusFlow :step="$clicks" />

<!--
The whole workshop on one slide, before any code. Ten clicks, one payment
(TXN-A from the lab's starter), and the yellow dot is the data.

The magenta key badges are the three things a Nexus request cannot be sent
without, and each one glows on the click where it is used (clicks 2, 3, 4):
the caller's contract + Endpoint name, the Endpoint's target Namespace + Task
Queue, and a Worker polling that queue with the handler registered. Miss any
one and the call never lands.

- [click] ValidatePayment, an ordinary Activity on the Payments Worker.
- [click] The Workflow calls checkCompliance through its Nexus client:
  ExecuteOperation(...).Get(ctx, &result). Same call shape as
  ExecuteActivity, which is the point.
- [click] The Endpoint is a name. It resolves to compliance-namespace and the
  compliance-risk Task Queue. Payments never sees either.
- [click] The Compliance Worker's handler starts ComplianceWorkflow and hands
  back an operation token immediately. That is what makes it async: the
  handler is done in milliseconds, the Operation stays open.
- [click] ComplianceWorkflow runs in its own Namespace, on its own Worker.
- [click] The CheckCompliance Activity scores the risk. LOW for TXN-A.
- [click] ComplianceWorkflow completes with a ComplianceResult.
- [click] The result comes back by completion callback, not back through the
  Endpoint. That is why the dashed line takes a different path.
- [click] Approved, so ExecutePayment runs.
- [click] PaymentResult, COMPLETED. Two Namespaces, two Workers, one call.
-->

---
layout: section
---

# Nexus Fundamentals

Every term you need, with the code that uses it.

<!--
Framing line for the block. ~15 seconds. Six concepts, one slide each, and
every one of them is in the lab: they write five and read the Endpoint.
-->

---
layout: code-walkthrough
heading: Service
subheading: "A named set of Operations: the <strong>contract</strong> both teams build against"
notes:
  - "<code>nexus-rpc/sdk-go</code> holds the contract types. Both teams import from it"
  - "The Service name. A string both sides agree on, in a package both import"
  - "<code>NewOperationReference[In, Out]</code> declares one Operation: its name, its input and its output"
---

::code::

```go {all|1|3|5-7}
import "github.com/nexus-rpc/sdk-go/nexus"

const ComplianceServiceName = "ComplianceNexusService"

var CheckComplianceOperation = nexus.NewOperationReference[
	ComplianceRequest, ComplianceResult,
]("checkCompliance")
```

<!--
Code Walkthrough: one click per note.

Source: shared/service.go lines 8-18, with the doc comments and the second
Operation (submitReview, next slide) trimmed, and the reference rewrapped for
line length.

- **Build 1 - Line 1: the contract package.** github.com/nexus-rpc/sdk-go is
  the Nexus RPC library. It is a separate module from the Temporal SDK, which
  depends on it, so it is already in the lab's go.mod.
- **Build 2 - Line 3: the Service.** A Service is the set of Operations that
  both teams work against. In Go it is a name plus the Operation references
  next to it, all in one shared package: Payments imports it to call,
  Compliance imports it to name its handlers. If someone asks:
  "ComplianceNexusService" is the Service name on the wire. It has to match on
  both sides, and it is not the Endpoint name.
- **Build 3 - Lines 5-7: an Operation.** NewOperationReference[In, Out]
  declares one callable Operation with its input and output types, and no
  implementation. Do not explain sync vs async here. That is its own slide.
-->

---
layout: code-walkthrough
heading: Operation
subheading: "One callable unit on a Service: a name, an input type and an output type"
notes:
  - "<code>checkCompliance</code>: a name, an input type, an output type"
  - "<code>submitReview</code>: same shape. The contract declares <strong>every</strong> Operation, called yet or not"
  - "Fast or slow is only a <strong>comment</strong> here. The contract cannot say it: the <strong>handler</strong> decides"
---

::code::

```go {all|3-4|7-8|2,6}
var (
	// Slow path. Backed by a Workflow.
	CheckComplianceOperation = nexus.NewOperationReference[
		ComplianceRequest, ComplianceResult]("checkCompliance")

	// Fast path. Answers a running Workflow.
	SubmitReviewOperation = nexus.NewOperationReference[
		ReviewRequest, ComplianceResult]("submitReview")
)
```

<!--
Code Walkthrough: one click per note.

Source: shared/service.go lines 15-23. The two comments are the lab's doc
comments, shortened; the references are rewrapped for line length.

- **Build 1 - Lines 3-4: the slow Operation.** A risky payment may need a
  human, so the answer can be minutes or hours away. But all the contract
  records is the name and the two types.
- **Build 2 - Lines 7-8: the fast Operation.** By the time it is called, the
  compliance check is running and parked waiting for a decision. The rule to
  land: the contract is exhaustive. submitReview is only used once a payment
  needs human review, and it still has to be declared from day one.
- **Build 3 - Lines 2 and 6: who decides fast or slow.** Nothing in
  NewOperationReference marks an Operation as sync or async. That is the
  handler's choice: temporalnexus.NewWorkflowRunOperation makes it async,
  nexus.NewSyncOperation makes it sync (the Sync vs async slide). The caller
  cannot tell and does not need to: it calls ExecuteOperation the same way for
  both. The handler team can even switch one from sync to async later without
  changing the contract.
- Worth naming out loud: the compiler checks every use of these names, but
  nothing checks at compile time that Compliance registered a handler for
  each one. An Operation the handler's Service does not have reaches the
  Worker and fails with a non-retryable "Not Found". The same goes for
  version skew across teams: if Payments ships a contract with a new
  Operation before Compliance deploys a handler for it, that is what the
  caller sees. In the lab, the Check's tests are what prove both sides agree.
-->

---
layout: code-walkthrough
heading: Endpoint
subheading: "A reverse proxy that routes Nexus requests to <strong>one</strong> target Namespace and Task Queue"
notes:
  - "Registers a reverse proxy in the Nexus Registry. Live as soon as it is created"
  - "<code>--name</code>: the <strong>only</strong> thing the caller ever knows"
  - "<code>--target-namespace</code>: the Compliance team's Namespace"
  - "<code>--target-task-queue</code>: must match the queue the Compliance Worker polls"
---

::code::

```bash {all|1|2|3|4}
temporal operator nexus endpoint create \
  --name compliance-endpoint \
  --target-namespace compliance-namespace \
  --target-task-queue compliance-risk
```

<!--
Code Walkthrough: one click per note.

Source: the Temporal CLI. The three names are the constants in
shared/types.go (ComplianceEndpoint, ComplianceNamespace,
ComplianceTaskQueue). The lab runs this exact command when it starts, and
learners read the result back with `temporal operator nexus endpoint list` in
challenge 3.

- **Build 1 - Line 1: create the Endpoint.** An Endpoint is a reverse proxy
  that routes Nexus requests from a caller Workflow to one target Namespace and
  Task Queue, not to multiple backends. Adding it to the Nexus Registry deploys
  it immediately. This slide is the same whatever the Workers are written in:
  the CLI talks to the Temporal Service, not to your code. Say so, it lands the
  "Nexus is a platform feature" point.
- **Build 2 - Line 2: the name.** Callers only need the Endpoint name. The
  target Namespace, Task Queue and internal implementation stay hidden.
- **Build 3 - Line 3: the target Namespace.** Namespace + Task Queue as the two
  halves of an address is the reframing the room needs, even though they
  already know both terms.
- **Build 4 - Line 4: the target Task Queue.** One Endpoint can serve several
  Services on the same queue. The quiet failure: point --target-task-queue at
  a queue nobody polls and the caller Workflow does not fail. The pending Nexus
  Operation records each failed attempt and keeps retrying until
  ScheduleToCloseTimeout runs out, visible in the UI and in
  `temporal workflow describe`. It is the most common first mistake, so tell
  them where to look.
-->

---
layout: code-walkthrough
heading: Caller
subheading: "The Workflow that calls an Operation. It knows the contract and an Endpoint name, nothing else"
notes:
  - "<strong>Before</strong>: an Activity call. Same Worker, same Namespace, one blast radius"
  - "<strong>After</strong>: a Nexus client. An Endpoint name and a Service name, nothing else"
  - "The Endpoint is a <strong>name</strong> the Registry resolves, not an address"
  - "The call site. <code>ScheduleToCloseTimeout</code> keeps it alive through a handler outage"
---

::code::

```go {all|1-2|4-5|5|7-11}
err := workflow.ExecuteActivity(ctx,
	compliance.CheckCompliance, compReq).Get(ctx, &result)

complianceClient := workflow.NewNexusClient(
	shared.ComplianceEndpoint, shared.ComplianceServiceName)

err := complianceClient.ExecuteOperation(ctx,
	shared.CheckComplianceOperation, // Operation from the contract
	compReq,                         // its input: a ComplianceRequest
	workflow.NexusOperationOptions{ScheduleToCloseTimeout: 10 * time.Minute},
).Get(ctx, &result)
```

<!--
Code Walkthrough: one click per note.

Source: lines 1-2 are the exercise's payments/workflows.go line 80 (the
monolith). Lines 4-11 are the solution's payments/workflows.go lines 69-72.
Both workflow.ExecuteActivity and workflow.NewNexusClient come from
go.temporal.io/sdk/workflow; long lines are rewrapped and the two comments are
added for the slide.

- **Build 1 - Lines 1-2: before.** The Activity call they already have. The
  compliance code runs in the same Worker and the same Namespace as Payments,
  and the compliance package is compiled into the Payments binary, so one bad
  deploy takes both down.
- **Build 2 - Lines 4-5: after.** The Nexus client that replaces it. It takes
  an Endpoint name and the Service name from the shared contract.
- **Build 3 - Line 5: the Endpoint.** ComplianceEndpoint is
  "compliance-endpoint". Land the boundary precisely: the Workflow names an
  Endpoint and a contract, and nothing else. An Endpoint is a name the
  Registry resolves: not a Namespace, not a Task Queue, not an address. Move
  Compliance tomorrow and this Workflow does not change.
- **Build 4 - Lines 7-11: the call site.** Same input, same output, different
  architecture. CheckComplianceOperation is the reference from the contract,
  and ExecuteOperation checks the input against its input type before
  scheduling, so a wrong input fails fast. This is the beat of the whole
  workshop: one call changes here, and the deployment topology underneath it
  does not follow. ScheduleToCloseTimeout covers the whole call, retries
  included, which is what keeps the Operation alive across a handler outage.
  Ten minutes, deliberately longer than the handler's 8-minute review window.
-->

---
layout: code-walkthrough
heading: Handler
subheading: "The code that implements the Service, on the other team's Worker"
notes:
  - "<code>NewWorkflowRunOperation</code>: an Operation backed by a Workflow"
  - "Named after the reference in the contract, so caller and handler agree"
  - "The Workflow it starts, on this Worker's Task Queue. It returns at once"
  - "A business ID. A retried Operation <strong>re-attaches</strong> instead of starting a duplicate"
  - "<code>nexus.NewService</code> binds the handlers to the contract. The Worker registers it"
---

::code::

```go {all|1|2|3|7|12-14}
var CheckComplianceOperation = temporalnexus.NewWorkflowRunOperation(
	shared.CheckComplianceOperation.Name(),
	ComplianceWorkflow,
	func(ctx context.Context, input shared.ComplianceRequest,
		options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
		return client.StartWorkflowOptions{
			ID: workflowIDFor(input.TransactionID),
		}, nil
	},
)

service := nexus.NewService(shared.ComplianceServiceName)
service.Register(CheckComplianceOperation, SubmitReviewOperation)
w.RegisterNexusService(service)
```

<!--
Code Walkthrough: one click per note.

Source: lines 1-10 are compliance/handler.go lines 33-44, comments trimmed and
the function signature rewrapped. Lines 12-14 are NewNexusService in the same
file (lines 74-75) and compliance/worker.go line 25, with error handling cut.

- **Build 1 - Line 1: an async Operation.** The handler is the side
  implementing the Service. It runs in a separate Worker: different process,
  different Namespace, different deploy schedule. Caller and Handler are the
  pair the rest of the workshop uses constantly; this slide and the last one
  are where they get named. NewWorkflowRunOperation starts a Workflow and
  hands back an operation token. Temporal delivers the result whenever the
  Workflow finishes.
- **Build 2 - Line 2: the name.** The handler names itself after the
  contract's reference, so the two sides cannot drift apart on a typo.
- **Build 3 - Line 3: the Workflow.** The input type comes from
  ComplianceWorkflow's signature. The Task Queue is left out on purpose: it
  defaults to the Task Queue the Operation is handled on, which is the one the
  Endpoint already targets.
- **Build 4 - Line 7: the Workflow ID.** Business-meaningful and deduplicated:
  a retried Operation attaches to the same compliance check rather than
  starting a second one.
- **Build 5 - Lines 12-14: bind and register.** nexus.NewService takes the
  Service name from the contract, Register adds one handler per Operation, and
  the Worker needs w.RegisterNexusService(service), or it starts fine and
  silently never answers. That one line is TODO 3 in the lab.
-->

---
layout: code-walkthrough
heading: Sync vs async
subheading: "Sync returns the result in the request. Async starts work and delivers the result later"
notes:
  - "<strong>Async</strong>: a Workflow-run Operation. The answer arrives whenever the Workflow finishes"
  - "<strong>Sync</strong>: a plain function. Must return its result inside the 10-second handler deadline"
  - "<code>GetClient(ctx)</code> gives the handler a Temporal Client"
  - "Same ID formula, so it finds the Workflow <code>checkCompliance</code> started"
  - "An Update hands over the decision and returns the result"
---

::code::

```go {all|1-5|7-25|11|13|12-23}
var CheckComplianceOperation = temporalnexus.NewWorkflowRunOperation(
	shared.CheckComplianceOperation.Name(),
	ComplianceWorkflow,
	startOptions, // returns the business Workflow ID, last slide
)

var SubmitReviewOperation = nexus.NewSyncOperation(
	shared.SubmitReviewOperation.Name(),
	func(ctx context.Context, input shared.ReviewRequest,
		options nexus.StartOperationOptions) (shared.ComplianceResult, error) {
		c := temporalnexus.GetClient(ctx)
		handle, err := c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
			WorkflowID:   workflowIDFor(input.TransactionID),
			UpdateName:   ReviewUpdate,
			Args:         []any{input.Approved, input.Explanation},
			WaitForStage: client.WorkflowUpdateStageCompleted,
		})
		if err != nil {
			return shared.ComplianceResult{}, err
		}
		var result shared.ComplianceResult
		err = handle.Get(ctx, &result)
		return result, err
	},
)
```

<!--
Code Walkthrough: one click per note.

Source: compliance/handler.go lines 33-69, comments trimmed, the TODO 2
markers removed and the function signature rewrapped. The options function on
line 4 is the one from the last slide, named here instead of repeated.

- **Build 1 - Lines 1-5: async.** The shape difference is the slide: async is
  temporalnexus.NewWorkflowRunOperation, a Workflow behind the Operation; sync
  is nexus.NewSyncOperation around a plain function. The constructor is the
  only thing that marks one from the other.
- **Build 2 - Lines 7-25: sync.** The function returns the result directly.
  The 10 seconds is the maximum time a sync Nexus handler has to fully process
  a request and return a result, measured from the calling History Service.
  Sync Operations complete as part of the start request itself: the caller's
  Nexus Machinery in the Temporal Service holds that request open, not the
  caller Workflow or its Worker.
- **Build 3 - Line 11: the client.** temporalnexus.GetClient(ctx) returns a
  Temporal Client connected to the handler's own Namespace.
- **Build 4 - Line 13: find the Workflow.** The same ID formula as
  checkCompliance, so Payments never needs to know a Compliance Workflow ID.
- **Build 5 - Lines 12-23: the Update.** ReviewUpdate is the Update name,
  "review", that ComplianceWorkflow handles with (approved bool, explanation
  string). WaitForStage Completed makes UpdateWorkflow return once the
  handler has run, and its handler returns immediately, so this finishes well
  inside the deadline.
-->

---
layout: default
class: "!p-0"
---

<div class="absolute inset-0 grid grid-cols-2">

  <div class="relative overflow-hidden bg-black">
    <video
      :src="'/mia-programming.mp4'"
      autoplay
      loop
      muted
      playsinline
      class="absolute inset-0 w-full h-full object-cover"
    />
  </div>

  <div class="flex flex-col justify-center items-start gap-5 px-14 bg-glow">
    <p class="eyebrow">TIME TO BUILD</p>
    <h1 class="cta-headline">Open the lab</h1>
    <a class="lab-link" href="https://play.instruqt.com/temporal/tracks/nexus-decouple-monolith-go">
      play.instruqt.com/temporal/tracks/nexus-decouple-monolith-go
    </a>
  </div>

</div>

<style scoped>
.cta-headline {
  font-size: 3rem;
  font-weight: 200;
  letter-spacing: -0.02em;
  line-height: 1.05;
  color: var(--temporal-text-strong);
  margin: 0;
}
.lab-link {
  font-family: var(--slidev-code-font-family, 'Noto Sans Mono', ui-monospace, monospace);
  font-size: 1rem;
  color: var(--temporal-green);
}
</style>

<!--
Last slide of segment 0. Leave it up while people open the lab: the link on
the slide is the Track, and paste it in chat too. Give them a moment to open
it. For a Dev Day, swap it for that day's Invite.

While it starts, say what they are about to run: a Temporal dev server, two
Namespaces, one per team, and the Nexus Endpoint between them. Every command
they type is `go run ./<something>` from the exercise folder.
-->
