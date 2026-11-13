---
slug: roll-it-over
id: lsd1dspybdkk
type: challenge
title: 1. Continue-As-New
teaser: A subscription bills every cycle until the Temporal Service terminates it.
  Add Continue-As-New and watch it finish.
notes:
- type: text
  contents: |-
    # What happens to a Workflow that never stops growing?

    A subscription bills its customer every cycle. Every charge and every timer
    appends to one Event History, and that history has a hard ceiling.

    Past it, the Temporal Service terminates the Workflow. Long before that, it
    tells the Workflow to roll over. Does anything listen?

    **Author:** [Nikolay Advolodkin](https://www.linkedin.com/in/nikolayadvolodkin/), Staff Developer Advocate
- type: text
  contents: |-
    # Three words for this lab

    **Event History.** The append-only record of everything a Workflow Execution
    did. It has a limit on the number of Events.

    **Continue-As-New.** Close this run and start a fresh one atomically: same
    Workflow ID, new Run ID, new short history.

    **Run.** One link in that chain. The next run knows only the input it is
    handed.
tabs:
- id: ivuvxzbkottz
  title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/continue-as-new-subscription
  port: 8080
- id: rsux8apgkf0s
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/continue-as-new-subscription/exercise
- id: ksjplyfvukcq
  title: Client
  type: terminal
  hostname: workshop
  workdir: /root/workshop/continue-as-new-subscription/exercise
- id: lxq4kbnl0ixc
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: advanced
timelimit: 600
enhanced_loading: null
---

# Goal

Run a 30-cycle subscription and watch the Temporal Service terminate it. Then
add Continue-As-New so it rolls over when the Service suggests it, and finishes.

**Why it matters:** any Workflow that lives for months, such as a subscription,
a policy that renews every year, or an account, keeps appending to its history.
Without Continue-As-New it is eventually terminated, mid-flight.

# Your tabs

- [button label="Exercise" background="#444CE7"](tab-0) VS Code in the browser, open on this Workshop. You edit
  the code under `exercise/`. It saves as you type.
- [button label="Worker" background="#444CE7"](tab-1) a terminal in `exercise/`. Run `go run ./worker` here.
  It runs the code as it was when it started: stop it with `Ctrl+C` and run it
  again after you change the code.
- [button label="Client" background="#444CE7"](tab-2) a terminal in `exercise/`. Run `go run ./starter` here to
  start a Workflow.
- [button label="Temporal UI" background="#444CE7"](tab-3) the Temporal UI of the local dev server: the
  subscription Workflow, its Event History, and each new run it continues as.

# The limits

The Temporal Service checks every Event History against these limits:

| | Normally | In this lab |
|---|---|---|
| Suggests Continue-As-New | 4,096 Events or 4 MB | 40 Events |
| Terminates the Workflow | 51,200 Events or 50 MB | 200 Events |

This lab's dev server lowers both so you can watch them in seconds. One billing
cycle adds about 11 Events.

# Run it without Continue-As-New

In the [button label="Worker" background="#444CE7"](tab-1) tab:

```bash,run
go run ./worker
```

```bash,nocopy
Subscription Worker polling subscription-tasks. Ctrl+C to stop.
```

Leave it running.

In the [button label="Client" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter
```

The Worker counts cycles. About 20 seconds later the Client tab prints:

```bash,nocopy
Billing subscription-demo for 30 cycles. Waiting for it to finish...

subscription-demo was terminated by the Temporal Service: Workflow history count exceeds limit.
```

The Worker stopped at cycle 18 of 30.

Open the [button label="Temporal UI" background="#444CE7"](tab-3) tab and click
`subscription-demo`. If it isn't listed, click the refresh button: the list does
not always update on its own. It is **Terminated**, and its history ends at
Event 201 with **Workflow Execution Terminated**.

Now click any **Workflow Task Started** event after Event 40. It reads
**Suggest Continue As New: true**. The Service suggested a rollover on every
Workflow Task from then on, and nothing in the code listened.

# Roll the history over (TODO 1)

**Observe.** Open `subscription/workflow.go` in the
[button label="Exercise" background="#444CE7"](tab-0) tab. Press `Ctrl+P`
(`Cmd+P` on a Mac), type the file name, then `Ctrl+G` to jump to a line.

<!-- snippet: exercise/subscription/workflow.go TODO 1 before=16 -->
Lines 14-31 are the subscription's billing loop:

```go
func SubscriptionWorkflow(ctx workflow.Context, input SubscriptionInput) (SubscriptionResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	cyclesCompleted := input.CyclesCompleted

	for input.MaxCycles == 0 || cyclesCompleted < input.MaxCycles {
		err := workflow.ExecuteActivity(ctx, ChargeSubscription, input.SubscriptionID, cyclesCompleted+1).Get(ctx, nil)
		if err != nil {
			return SubscriptionResult{}, err
		}
		cyclesCompleted++
		if err := workflow.Sleep(ctx, Cycle); err != nil {
			return SubscriptionResult{}, err
		}

		// TODO 1 start: when the Service suggests it, continue as new with what this run has done
		// TODO 1 end
```

**Learn.** Each cycle charges, sleeps and goes round again, and every cycle
adds Events to this one run's Event History. Nothing ever starts a fresh
history, so it grows until the Temporal Service terminates the run. When the
history gets large, the Service suggests Continue-As-New:
`workflow.GetInfo(ctx).GetContinueAsNewSuggested()` turns `true`. Returning the
error from `workflow.NewContinueAsNewError(...)` then closes this run and starts
a new one, with an empty history, under the same Workflow ID.

The next run knows only the input you hand it. Pass `cyclesCompleted`, or every
new run starts counting from 0 again and the subscription never finishes.

**Do.** Between the TODO 1 markers:

- [ ] After each cycle, check `workflow.GetInfo(ctx).GetContinueAsNewSuggested()`.
- [ ] When it is `true`, return `workflow.NewContinueAsNewError(...)` for
  `SubscriptionWorkflow`, with a `SubscriptionInput` that carries what this run
  has done: `SubscriptionID`, `CyclesCompleted` and `MaxCycles`.

The `workflow` package is already imported at the top of the file.

<details>
<summary>Hint</summary>

`workflow.NewContinueAsNewError` takes the context, the Workflow function, and
the next run's arguments, the same way the starter passes the first run's. Return
it as the Workflow's error, with a zero `SubscriptionResult`:

```go
return SubscriptionResult{}, workflow.NewContinueAsNewError(ctx, SubscriptionWorkflow, SubscriptionInput{ /* ... */ })
```

Returning it is what closes the run: nothing after that line runs.

</details>

<details>
<summary>Show solution</summary>

<!-- snippet: solution/subscription/workflow.go TODO 1 -->
Lines 30-38 of `solution/subscription/workflow.go`:

```go
		// TODO 1 start: when the Service suggests it, continue as new with what this run has done
		if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
			return SubscriptionResult{}, workflow.NewContinueAsNewError(ctx, SubscriptionWorkflow, SubscriptionInput{
				SubscriptionID:  input.SubscriptionID,
				CyclesCompleted: cyclesCompleted,
				MaxCycles:       input.MaxCycles,
			})
		}
		// TODO 1 end
```

</details>

# Run it again

In the [button label="Worker" background="#444CE7"](tab-1) tab, press `Ctrl+C`
to stop the Worker, then start it again so it runs your change:

```bash,run
go run ./worker
```

In the [button label="Client" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter
```

About 30 seconds later:

```bash,nocopy
Billing subscription-demo for 30 cycles. Waiting for it to finish...

Done. Billed 30 cycles.
```

The Worker counts straight from cycle 1 to cycle 30.

Refresh the [button label="Temporal UI" background="#444CE7"](tab-3) tab. One
Workflow ID, `subscription-demo`, now has a chain of runs: each closed run is
**Continued as New**, and the last is **Completed**. Open any of them: each
history is short, and each run after the first starts with the
`CyclesCompleted` the run before it handed over.

**Same Workflow ID, new Run ID, fresh Event History.**

# Checkpoint

Click **Check**. It passes when a subscription rolls over on the Service's
suggestion and still bills every cycle it was asked for. If it fails, the
message names the file and the TODO still to do.

Before you ship this pattern:

- The input you hand over is a payload, limited to 2 MB. Carry IDs and counters,
  not large data.
- Clients that address the Workflow ID follow the chain to the newest run, which
  is why the starter's `run.Get(...)` waited across every rollover.
- If the Workflow receives Signals, drain each Signal Channel with
  `ReceiveAsync` before you return the Continue-As-New error, and carry what you
  read into the next run's input. If it has Update handlers, wait for them with
  `workflow.Await(ctx, func() bool { return workflow.AllHandlersFinished(ctx) })`
  first, so a rollover does not cut one off halfway.

---

**Please share your feedback so we can make better content for you.** The
**Feedback** tab takes a few seconds, and it is the only way we find out which
parts of this landed.
