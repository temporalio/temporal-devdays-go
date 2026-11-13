---
layout: section
---

<ZiggyHero scene="heartbeat">

# Heartbeats

Long-running Activities that say they're alive, and remember how far they got.

</ZiggyHero>

<!--
Source: PPT slides 122-131.
-->

---
layout: default
---

# Why Heartbeats?

Think of it like a contractor renovating your house.

- **Without Heartbeats,** you hand over the keys on Monday and agree to check
  back in two weeks. If they vanish on Tuesday, you won't know until the
  deadline, and you've lost nearly two weeks you could have spent hiring
  someone else.
- **With Heartbeats,** the contractor checks in every day. Miss a check-in and
  you know immediately, and can act.

<!--
Source: PPT slide 123.

The mental model: a Heartbeat is the Activity saying "I'm still alive and
making progress." If the Heartbeats stop, Temporal assumes the worst and acts.
You get visibility and automatic failure detection for long-running work.

In Temporal terms: `StartToCloseTimeout` is how long Temporal waits before
deciding an Activity is dead. For a two-hour Activity, that means a crashed
Worker goes unnoticed for two hours. Heartbeats give you a two-hour Activity
with a thirty-second failure detection time.
-->

---
layout: default
---

# Business benefits

- **Faster failure recovery.** Less downtime, less customer impact
- **Resume from a checkpoint,** not from scratch
- **Cancellations are actually honoured**

<!--
Source: PPT slide 124.

- **Faster recovery:** instead of waiting out a 12-hour `StartToCloseTimeout`
  before retrying a crashed job, a 30-second `HeartbeatTimeout` means you
  notice and retry within seconds.
- **Resume from a checkpoint:** processing a million records and the Worker dies
  at 750,000? The Heartbeat details let the next attempt start at 750,001. That
  is real money saved in compute.
- **Cancellation:** if an order is cancelled mid-processing, the Heartbeat is
  how the cancellation reaches the running Activity. Without Heartbeats, the
  Activity keeps running and keeps consuming resources until it finishes on its
  own.
-->

---
layout: code-walkthrough
heading: Heartbeats in Go
notes:
  - "<code>activity.GetInfo</code>: the Activity's execution info, read from its <code>ctx</code>"
  - Which attempt this is. 1 on the first try, 2 on the first retry
  - "The last Heartbeat's saved details. <code>HasHeartbeatDetails</code> is false on attempt 1; on a retry, decode the progress so far"
  - "The Heartbeat: <strong>still alive, still working</strong>, and here is how far I got. <code>RecordHeartbeat</code>"
---

::code::

```go {all|2|3|4-9|15}
func PackShipment(ctx context.Context, input ShipmentInput) (int, error) {
	info := activity.GetInfo(ctx)
	activity.GetLogger(ctx).Info("packing", "attempt", info.Attempt)
	start := 0
	if activity.HasHeartbeatDetails(ctx) {
		if err := activity.GetHeartbeatDetails(ctx, &start); err != nil {
			return 0, err
		}
	}

	for index := start; index < input.ItemCount; index++ {
		if err := packItem(ctx, input.ShipmentID, index); err != nil {
			return 0, err
		}
		activity.RecordHeartbeat(ctx, index+1)
	}

	return input.ItemCount, nil
}
```

<!--
Source: PPT slides 125-128. One build per slide in the original.
Code: snippets/heartbeat/activities.go, run by
snippets/heartbeat/heartbeat_test.go against a real dev server.

- **Build 1 - `activity.GetInfo(ctx)`.** The original fetched an execution
  context object. In Go that context is the `context.Context` every Activity
  takes first: the functions in `go.temporal.io/sdk/activity` (`GetInfo`,
  `RecordHeartbeat`, `GetLogger`) read the current Activity out of it.
- **Build 2 - the attempt.** Only logged here, but useful: an Activity can
  behave differently on a retry, for example by checking whether a charge
  already went through.
- **Build 3 - resume.** If this is:
  - attempt 1, `HasHeartbeatDetails` is false, so `start` stays at item 0
  - attempt 2 or later, `GetHeartbeatDetails` decodes wherever the previous
    attempt got to into `start`
- **Build 4 - the Heartbeat.** Tells Temporal "still alive" and records the
  progress. If the Worker crashes, the next attempt reads it back in build 3.

Heartbeats are throttled by the SDK, so calling it every item is fine; it will
not hit the Server every loop (at most once per 80% of the Heartbeat Timeout).
When the Activity returns an error, the SDK sends the latest details with the
failure, so a retry resumes exactly where the last attempt got to: the test
checks it. If the whole Worker crashes, the details the Server holds may be a
little behind the last call, so the work must tolerate redoing an item or two.
Idempotent item processing, in other words.

Two things the caller must do, or none of this works:

- set `HeartbeatTimeout` in the `workflow.ActivityOptions`. Without it,
  Temporal is not watching for missed Heartbeats, and a dead Worker is only
  noticed when `StartToCloseTimeout` expires.
- keep `StartToCloseTimeout` long enough for the whole job. The Heartbeat
  Timeout is the short one.

`GetHeartbeatDetails` decodes whatever you passed with the Data Converter, into
the pointers you give it, in order. An `int` comes back as an `int`; a struct
comes back with its exported fields. Keep the details small: they are stored by
the Server like everything else.
-->

---
layout: code-walkthrough
heading: Heartbeating and cancellation
notes:
  - Heartbeat with progress, every item
  - "A cancelled Activity finds out <strong>at a Heartbeat</strong>. The SDK cancels its context; look at <code>ctx.Err()</code>"
  - Clean up, then return the error so Temporal records the cancellation
---

::code::

```go {all|6|8|9-10}
func ProcessItems(ctx context.Context, input ProcessInput) error {
	for index, item := range input.Items {
		if err := processItem(ctx, item); err != nil {
			return err
		}
		activity.RecordHeartbeat(ctx, index+1)

		if err := ctx.Err(); err != nil {
			releaseLocks(input)
			return err
		}
	}
	return nil
}
```

<!--
Source: PPT slides 129-130.
Code: snippets/heartbeat/activities.go, run by
snippets/heartbeat/heartbeat_test.go.

The original had a second, empty heartbeat call "to check for cancellation."
Go does not need one: when the Server tells the Worker the Activity was
cancelled, the SDK cancels the Activity's `ctx`, so `ctx.Done()` closes and
`ctx.Err()` returns `context.Canceled`. That delivery rides on the Heartbeat,
which is why an Activity that never heartbeats can never be cancelled.

- **Build 1 -** progress on every item, as before.
- **Build 2 -** nothing interrupts your goroutine. Your code has to look: check
  `ctx.Err()`, or `select` on `ctx.Done()`. Passing `ctx` to anything that
  takes one (`http.NewRequestWithContext`, a database driver) cancels the
  in-flight call too.
- **Build 3 -** clean up and return the error. Returning `ctx.Err()` from a
  cancelled Activity is what Temporal records as the cancellation. Swallowing
  it and returning `nil` makes the Activity look like it succeeded. Cleanup
  that itself needs a context should use a fresh one
  (`context.WithoutCancel(ctx)`), since `ctx` is already done.

The caller decides how long to wait for this cleanup: with
`WaitForCancellation: true` in the `workflow.ActivityOptions`, the Workflow
waits until the Activity has actually stopped.
-->
