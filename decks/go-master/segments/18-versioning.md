---
layout: section
---

<ZiggyHero scene="versions">

# Workflow Versioning

Changing code that running Workflows are still replaying.

</ZiggyHero>

<!--
Source: PPT slides 68-70, `{{SLIDE:versioning-before}}` and
`{{SLIDE:versioning-after}}`. The replay test, retirement and Worker
Versioning slides are new to the Go Deck.

Frame the problem before the code, or the solution looks like ceremony: a
Workflow started last Tuesday is still running, and it replays against
*today's* code. Change the order of Activity calls and the replay no longer
matches the Event History. That is a non-determinism error, and it wedges the
Workflow.
-->

---
layout: code-stack
heading: "Before: the original Workflow"
proseMax: 26
---

::code::

```go
func MyWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	if err := workflow.ExecuteActivity(ctx, ActivityA).Get(ctx, nil); err != nil {
		return err
	}
	return workflow.ExecuteActivity(ctx, ActivityB).Get(ctx, nil)
}
```

::default::

Simple and clean: A, then B. Now the requirement changes: new executions should
run C instead of B.

<!--
Source: PPT slide 69. The original used scheduleToCloseTimeout; this uses
StartToCloseTimeout to match every other Activity call in the Deck.
Code: snippets/versioning/before/workflow.go.

Set the trap deliberately. Ask what happens if you just edit the last line and
redeploy.

The answer: every Workflow that already ran B replays, reaches this line,
expects B in the Event History, finds a command for C, and fails with a
non-determinism error ("[TMPRL1100] nondeterministic workflow: history event is
ActivityTaskScheduled ... ActivityB, replay command is ScheduleActivityTask ...
ActivityC"). It does not fail loudly at deploy time. It fails later, on the
next Workflow Task, for executions you had forgotten about. The Workflow is not
failed or terminated: by default the Worker's panic policy blocks it and its
Workflow Task keeps retrying, so a fix deployed later lets it carry on.

Only executions that are still open are at risk. A Workflow that completed
before the deploy never replays again.
-->

---
layout: code-walkthrough
heading: "After: the versioned Workflow"
notes:
  - "<code>GetVersion</code> marks one decision point, named by a string you choose"
  - New executions get version 1 and take this branch
  - Executions that started before the change get DefaultVersion and keep the old one
---

::code::

```go {all|10|11-12|14}
func MyWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	if err := workflow.ExecuteActivity(ctx, ActivityA).Get(ctx, nil); err != nil {
		return err
	}

	v := workflow.GetVersion(ctx, "use-activity-c", workflow.DefaultVersion, 1)
	if v == 1 {
		return workflow.ExecuteActivity(ctx, ActivityC).Get(ctx, nil)
	}
	return workflow.ExecuteActivity(ctx, ActivityB).Get(ctx, nil)
}
```

<!--
Source: PPT slide 70, ported to Go's GetVersion.
Code: snippets/versioning/patched/workflow.go, run by
snippets/versioning/versioning_test.go.

Walk it carefully: this is the segment people most often half-understand.

- **Build 1 - `GetVersion`.** `"use-activity-c"` is the change ID, a
  human-readable label for this one decision point. It is not a global version
  number and there is no ordering between change IDs. Pick a name that says
  what changed. The last two arguments are the oldest and newest versions this
  code supports: `workflow.DefaultVersion` (-1) means "before this change
  existed", and 1 is the change.
- **Build 2 - the new branch.** Every execution that starts after this deploy
  gets the maximum, 1, and takes it.
- **Build 3 - the old branch.** Every execution already in flight when the
  deploy landed gets `DefaultVersion`, because the SDK can see from its Event
  History that it never recorded this change.

The mechanism, in one sentence: on a new execution `GetVersion` records a
`Version` marker in the Event History and returns the maximum; on replay it
returns whatever the marker says, or `DefaultVersion` if there is none. The
branch taken is the branch that was always taken.

A Go detail if anyone asks: the version is fixed per execution the first time
`GetVersion` runs, so calling it again with the same change ID, in a loop or in
another function, returns the same value. A second change to the same spot
becomes version 2 of the same change ID: `GetVersion(ctx, "use-activity-c",
workflow.DefaultVersion, 2)` and a `switch`.
-->

---
layout: code-walkthrough
heading: Retiring the old branch
notes:
  - "<code>GetVersion</code> with a minimum of 1: still records the marker, no branch left"
  - Only the new path is left. The <code class="ref">ActivityB</code> branch is gone
---

::code::

```go {all|7|8}
func MyWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, options)
	if err := workflow.ExecuteActivity(ctx, ActivityA).Get(ctx, nil); err != nil {
		return err
	}

	workflow.GetVersion(ctx, "use-activity-c", 1, 1)
	return workflow.ExecuteActivity(ctx, ActivityC).Get(ctx, nil)
}
```

<!--
Ported from the PPT slide 70 notes, which described these steps in prose.
Code: snippets/versioning/deprecated/workflow.go.

The part people skip, and the reason this looks like ceremony until it does
not: adding `GetVersion` is step one of three.

- **Build 1 - raise the minimum.** Deploy this once no execution that started
  before the change is still running. The minimum supported version is now 1,
  so `DefaultVersion` is no longer accepted. The call still records the marker
  on new executions and still replays executions that have it.
- **Build 2 - the old branch is gone.** Only C remains.

Why the middle step exists, and the tests behind this Deck show it: if you
were wrong and an execution on the B branch is still open, it does not quietly
run C. Its next Workflow Task fails with an error that names the change:
"Workflow code removed support of version -1 for "use-activity-c" changeID.
The oldest supported version is 1". That is a message you can act on, instead
of a generic activity-type mismatch.

Step three: once every execution that ran the old branch is gone, delete the
`GetVersion` line too. The SDK skips a `Version` marker that no `GetVersion`
call consumes, so executions that recorded it still replay; the tests behind
this Deck replay one to prove it.

How to know when to move on: list open executions of the type in the Temporal
UI or with `temporal workflow list`, and wait for the old ones to finish. Leave
the `if` in forever and you have built a museum of every change you ever made.
-->

---
layout: code-walkthrough
heading: Catch it before you deploy
notes:
  - A real execution's Event History, exported as JSON and checked in
  - The replayer runs the new code against it. No Server, no Activities
  - It returns an error if the new code would decide differently
---

::code::

```go {all|6|1-2|4-6}
replayer := worker.NewWorkflowReplayer()
replayer.RegisterWorkflow(MyWorkflow)

// Returns an error if the new code would make
// different decisions from this history
err := replayer.ReplayWorkflowHistoryFromJSONFile(nil, historyFile)
```

<!--
New to the Go Deck.
Code: snippets/versioning/replay.go, run by snippets/versioning/versioning_test.go,
which exports each history with the CLI exactly as described here.

The non-determinism error from three slides ago shows up in production, days
later. This moves it into `go test`.

- **Build 1 - the history.** Export any execution with
  `temporal workflow show --workflow-id <id> --output json > testdata/x.json`
  (or download it from the Temporal UI) and check it in. The useful ones are
  long-running executions that are open right now.
- **Build 2 - replay.** `NewWorkflowReplayer` is a Worker with no Task Queue:
  register the Workflow code exactly as the Worker does, and it runs that code
  against the recorded Events. Activities are not executed; their results come
  from the history. The `nil` is the logger.
- **Build 3 - the verdict.** `nil` if the code makes the same decisions, and a
  `[TMPRL1100] nondeterministic workflow` error if it does not.

Put it in a `_test.go` file over a `testdata/` folder of real histories and the
unsafe edit from the "Before" slide fails the build instead of the Workflow.
`ReplayWorkflowHistory` takes a `*historypb.History` you fetched yourself.
-->

---
layout: code-walkthrough
heading: Worker Versioning
notes:
  - The Worker joins a Worker Deployment
  - "<code>DeploymentName</code> stays the same. <code>BuildID</code> changes with every build"
  - A <strong>pinned</strong> execution runs to completion on the version it started on
---

::code::

```go {all|2-3|5-6|8}
w := worker.New(c, "orders", worker.Options{
	DeploymentOptions: worker.DeploymentOptions{
		UseVersioning: true,
		Version: worker.WorkerDeploymentVersion{
			DeploymentName: "orders",
			BuildID:        "2.0.0",
		},
		DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
	},
})
```

<!--
New to the Go Deck. Checked against go.temporal.io/sdk v1.51.0.
Code: snippets/versioning/replay.go, run by snippets/versioning/versioning_test.go.

Patching with `GetVersion` changes the code so old and new executions can
share one Worker. Worker Versioning is the other approach: run old and new
Workers side by side, and let Temporal route each execution to the version it
belongs to.

- **Build 1 - opt in.** `DeploymentOptions` on `worker.Options`, with
  `UseVersioning: true`.
- **Build 2 - the version.** `DeploymentName` names the service and stays
  constant; `BuildID` names this build (a git hash, a semver, a build number,
  often stamped in with `-ldflags`).
- **Build 3 - the behaviour.** The default for every Workflow this Worker
  registers; `workflow.RegisterOptions{VersioningBehavior: ...}` overrides it
  per Workflow. `VersioningBehaviorPinned` keeps an execution on the version it
  started on, so it never sees new code and never needs a `GetVersion`.
  `VersioningBehaviorAutoUpgrade` moves it to the current version, which brings
  patching back.

A new version takes no new executions until you make it current:
`temporal worker deployment set-current-version --deployment-name orders
--build-id 2.0.0`. Existing pinned executions stay where they are, so the old
Workers keep running until those drain.

When to pick which: pinning suits short Workflows and frequent deploys. A
Workflow that runs for months would keep old Workers alive for months, so
long-running ones use `GetVersion`, or Continue-As-New onto the new version.
-->
