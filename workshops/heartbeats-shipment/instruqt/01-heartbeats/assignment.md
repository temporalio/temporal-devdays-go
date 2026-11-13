---
slug: heartbeats
id: eijvtxe91fpw
type: challenge
title: 1. Activity Heartbeats
teaser: Crash a Worker mid-job and watch the retry wait, then start over. Add Heartbeats
  and watch it pick up in seconds, where it left off.
notes:
- type: text
  contents: |-
    # The Worker dies 10 items into a 40 item job. Now what?

    Packing a shipment takes a while. Halfway through, the process doing it
    disappears: a deploy, an out-of-memory kill, a lost node.

    How long before Temporal notices? And where does the retry start?

    **Author:** [Nikolay Advolodkin](https://www.linkedin.com/in/nikolayadvolodkin/), Staff Developer Advocate
- type: text
  contents: |-
    # Three words for this lab

    **Heartbeat.** A call an Activity makes while it works: "still here, and
    this is how far I got."

    **Heartbeat Timeout.** How long Temporal waits between Heartbeats before it
    treats the attempt as lost and retries it.

    **Heartbeat details.** The progress sent with the last Heartbeat. The next
    attempt can read it back.
tabs:
- id: osrvsb2su1ff
  title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/heartbeats-shipment
  port: 8080
- id: ff0cyia5spar
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/heartbeats-shipment/exercise
- id: 6ay6xl86geva
  title: Client
  type: terminal
  hostname: workshop
  workdir: /root/workshop/heartbeats-shipment/exercise
- id: mmbiclv9id3l
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

Kill a Worker in the middle of a packing job, without Heartbeats and then with
them, and compare how long the retry takes to start and where it starts.

**Why it matters:** without Heartbeats, a crash holds the work until the whole
attempt times out, and the retry redoes every item. In a real system that is
lost time and duplicated side effects.

# Your tabs

- [button label="Exercise" background="#444CE7"](tab-0) VS Code in the browser, open on this Workshop. You edit
  the code under `exercise/`. It saves as you type.
- [button label="Worker" background="#444CE7"](tab-1) a terminal in `exercise/`. Run `go run ./worker` here.
  After you change the code, press **Ctrl+C** and run it again to load the change.
- [button label="Client" background="#444CE7"](tab-2) a terminal in `exercise/`. Run `go run ./starter` here to
  start a Workflow.
- [button label="Temporal UI" background="#444CE7"](tab-3) the Temporal Web UI of the local dev server:
  the shipment Workflow, each packing attempt, and the Heartbeat details it saved.

# Run it without Heartbeats

`PackShipment` in `exercise/activities.go` packs 40 items, one every 0.3
seconds, about 12 seconds in all. `exercise/workflows.go` gives each attempt a
20-second start-to-close timeout.

In the [button label="Worker" background="#444CE7"](tab-1) tab:

```bash,run
go run ./worker
```

```bash,nocopy
Shipment Worker polling shipment-tasks. Ctrl+C to crash it.
```

In the [button label="Client" background="#444CE7"](tab-2) tab, start the job:

```bash,run
go run ./starter
```

Switch to the [button label="Worker" background="#444CE7"](tab-1) tab and watch
it count: `packed item 1 of 40`, `packed item 2 of 40`. Around item 10, press
**Ctrl+C**. Nothing in the Worker catches the signal, so the process exits on
the spot, with no chance to tell Temporal, the way a crashed node does.

Open `shipment-demo` in the [button label="Temporal UI" background="#444CE7"](tab-3)
tab (click refresh if it isn't listed) and look at **Pending Activities**:
attempt 1, still **Started**, and no Heartbeat. Temporal has no idea the Worker
is gone.

Wait about 20 seconds and look again. Attempt 1 has timed out, with
`activity StartToClose timeout` as its last failure, and the Activity is now
on **attempt 2, Scheduled**. It stays on attempt 2 for as long as no Worker is
running. An attempt only counts against the 10 when it fails, and it can only
fail once a Worker has started it. Until then it waits in the Task Queue.
**Temporal does not burn retries while your Workers are down.**

Start a replacement in the [button label="Worker" background="#444CE7"](tab-1) tab:

```bash,run
go run ./worker
```

It picks up attempt 2 as soon as it starts polling. If you start it before
attempt 1's 20 seconds have run out, it packs nothing until they do. Either
way it prints:

```bash,nocopy
15:28:39 packing shipment-demo, attempt 2, starting at item 0
```

**Slow, and from item 0.** The items the crashed Worker had already packed get
packed a second time.

# Step 1 of 3: notice a silent attempt (TODO 1a)

**Observe.** Open `exercise/workflows.go` in the
[button label="Exercise" background="#444CE7"](tab-0) tab. Press `Ctrl+P`
(`Cmd+P` on a Mac), type the file name, then `Ctrl+G` to jump to a line.

<!-- snippet: exercise/workflows.go TODO 1a before=3 -->
Lines 13-17 set the options for every packing attempt:

```go
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// A ceiling for the whole shipment, which takes about 12 seconds.
		StartToCloseTimeout: 20 * time.Second,
		// TODO 1a start: give up on an attempt when no Heartbeat arrives for 2 seconds
		// TODO 1a end
```

**Learn.** The only timeout is the 20-second start-to-close, a ceiling on the
whole attempt. Nothing notices an attempt that has gone silent, so a stuck
attempt is retried only after the full 20 seconds. A Heartbeat Timeout does:
if that long passes with no Heartbeat, Temporal treats the attempt as lost and
the Retry Policy starts the next one.

**Do.**

- [ ] Between the TODO 1a markers, add a `HeartbeatTimeout` of 2 seconds to the
  `workflow.ActivityOptions`.

<details>
<summary>Show solution</summary>

<!-- snippet: solution/workflows.go TODO 1a -->
Lines 16-18 of `solution/workflows.go`:

```go
		// TODO 1a start: give up on an attempt when no Heartbeat arrives for 2 seconds
		HeartbeatTimeout: 2 * time.Second,
		// TODO 1a end
```

</details>

Don't run it yet. A Heartbeat Timeout with no Heartbeats makes every attempt
fail after 2 seconds, even a healthy one. Step 2 sends them.

# Step 2 of 3: send a Heartbeat after each item (TODO 1b)

**Observe.** Open `exercise/activities.go`, the Activity that packs.

<!-- snippet: exercise/activities.go TODO 1b before=12 -->
Lines 29-42 are the packing loop:

```go
	for item := packed; item < input.ItemCount; item++ {
		select {
		case <-ctx.Done():
			// Cancellation reaches a running Activity through its Heartbeat:
			// the SDK cancels ctx. Stop and let the error out.
			say("interrupted at item %d", packed)
			return ShipmentResult{}, ctx.Err()
		case <-time.After(TimePerItem):
		}
		packed = item + 1
		say("packed item %d of %d", packed, input.ItemCount)

		// TODO 1b start: Heartbeat the number of items packed so far
		// TODO 1b end
```

**Learn.** The loop logs each item, but Temporal hears nothing until the
Activity returns. A Heartbeat is the "still here" signal the timeout from
step 1 waits for. It can also carry details, and the count of packed items is
the detail a retry needs.

**Do.**

- [ ] Between the TODO 1b markers, Heartbeat `packed` with
  `activity.RecordHeartbeat`. The `activity` package is already imported.

<details>
<summary>Show solution</summary>

<!-- snippet: solution/activities.go TODO 1b -->
Lines 46-48 of `solution/activities.go`:

```go
		// TODO 1b start: Heartbeat the number of items packed so far
		activity.RecordHeartbeat(ctx, packed)
		// TODO 1b end
```

</details>

# Step 3 of 3: resume where the last attempt stopped (TODO 1c)

**Observe.** Stay in `exercise/activities.go`.

<!-- snippet: exercise/activities.go TODO 1c before=3 -->
Lines 20-25 are the top of the same Activity:

```go
func PackShipment(ctx context.Context, input ShipmentInput) (ShipmentResult, error) {
	attempt := activity.GetInfo(ctx).Attempt

	// TODO 1c start: start from the count the last attempt Heartbeated, or 0 on a first attempt
	packed := 0
	// TODO 1c end
```

**Learn.** Every attempt starts `packed` at 0, so a retry packs every item
again, even the ones the failed attempt finished. A retry can read the details
of the last Heartbeat with `activity.GetHeartbeatDetails(ctx, &packed)`. A
first attempt has none, and reading them then returns an error, so ask
`activity.HasHeartbeatDetails(ctx)` first.

**Do.**

- [ ] Between the TODO 1c markers, start `packed` from the last Heartbeat's
  count, and keep 0 when there is none.

<details>
<summary>Hint</summary>

The same shape, reading an optional setting into a variable that already holds
its default:

```go
port := 8080
if v, ok := os.LookupEnv("PORT"); ok {
	if _, err := fmt.Sscan(v, &port); err != nil {
		return err
	}
}
```

</details>

<details>
<summary>Show solution</summary>

<!-- snippet: solution/activities.go TODO 1c -->
Lines 23-30 of `solution/activities.go`:

```go
	// TODO 1c start: start from the count the last attempt Heartbeated, or 0 on a first attempt
	packed := 0
	if activity.HasHeartbeatDetails(ctx) {
		if err := activity.GetHeartbeatDetails(ctx, &packed); err != nil {
			return ShipmentResult{}, err
		}
	}
	// TODO 1c end
```

</details>

The editor saves as you type. The Worker runs the code it started with, so
restart it to pick up your change.

# Run it with Heartbeats

In the [button label="Worker" background="#444CE7"](tab-1) tab, press
**Ctrl+C** if the Worker is still running, and start it again:

```bash,run
go run ./worker
```

Run the same experiment. In the
[button label="Client" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter
```

Watch the [button label="Worker" background="#444CE7"](tab-1) tab and press
**Ctrl+C** around item 10.

Pending Activities in the [button label="Temporal UI" background="#444CE7"](tab-3)
tab now shows a last Heartbeat and its details, the count. Start a replacement
right away:

```bash,run
go run ./worker
```

As soon as it is up, a few seconds after the crash:

```bash,nocopy
15:29:08 packing shipment-demo, attempt 2, starting at item 6
```

And the Client tab finishes with:

```bash,nocopy
Packed 40 items, finished on attempt 2.
An earlier attempt was interrupted. This one finished the job.
```

**Fast, and from where it left off.** Your number depends on when you killed it.

It resumes a few items behind the last one the dead Worker printed. The SDK
**throttles** Heartbeats: with a 2-second timeout it sends at most one every
1.6 seconds and keeps only the newest in between. A crash sends nothing more,
so whatever was packed after the last Heartbeat left the process is packed
again. Heartbeat as often as you like; the SDK decides what goes over the wire.

# Checkpoint

Click **Check**. It passes when a silent attempt is retried within seconds,
every item is Heartbeated, and a retry resumes from the last Heartbeat. If it
fails, the message names the file and the part still to do.

One more thing Heartbeats carry: **cancellation**. Start a job again:

```bash,run
go run ./starter
```

While it packs, open `shipment-demo` in the
[button label="Temporal UI" background="#444CE7"](tab-3) tab and choose
**Request Cancellation** from the Workflow's actions, top right. The Worker
prints `interrupted at item N`: the cancellation arrived in the reply to a
Heartbeat, the SDK cancelled the Activity's `ctx`, and the `select` in
`PackShipment` took the `<-ctx.Done()` case and returned `ctx.Err()`. The
Client tab says the shipment ended Canceled. An Activity that never Heartbeats
never learns it was cancelled, and keeps working.

---

**Please share your feedback so we can make better content for you.** The
**Feedback** tab takes a few seconds, and it is the only way we find out which
parts of this landed.
