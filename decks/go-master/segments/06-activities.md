---
layout: section
---

<ZiggyHero scene="work">

# Go Activities

The work. Ordinary Go code that is allowed to fail.

</ZiggyHero>

<!--
Source: PPT slides 51-61.

The relief slide. After the Workflow rules, Activities have almost none.

The original had an interface slide and an implementation slide here too. In
Go an Activity is a function (or a method), so it is one walkthrough.
-->

---
layout: default
---

# What is an Activity?

The unit of work that talks to the outside world. Anything that can fail on its
own, and should retry on its own.

- **A plain Go function.** No annotation, no interface to implement.
- **Outside the Workflow's rules.** Full I/O, any library, any clock.
- **Retried by the platform.** It returns an error; Temporal decides what
  happens next.
- **Recorded.** Its result goes into Event History; replay reads it back.
- Keep it a thin wrapper

<!--
Source: PPT slide 51.

The framing slide before the anatomy. Two minutes at most. The code on the next
slide does the real teaching, so resist explaining it twice.

The bullets are deliberately about what an Activity *is*, not about the shape
of the file. The anatomy walkthrough that follows takes the shape apart one
line at a time, and these would steal its beats.

The recorded bullet is the one worth dwelling on, because it is the one that is
not obvious: the recorded result is what makes replay cheap and what makes a
crashed Workflow recoverable. An Activity that already completed does not run
again.

"Thin wrapper": the Activity calls the API or the database and returns. The
decisions (what next, what if it failed) stay in the Workflow, where they are
durable.

Good question to take here if the room is quiet: "so what stops me putting
everything in Activities?" Answer: nothing, and it would work. You would just
have no orchestration left, and a Workflow that is one Activity call is a
Workflow that cannot make a decision. That is the design-patterns segment, coming up.
-->

---
layout: code-walkthrough
heading: Go Activity anatomy
notes:
  - No annotation, no interface. A plain function a Worker registers is the whole story
  - The function name is the Activity Type. It is what the Workflow passes to <code class='ref'>ExecuteActivity</code>
  - "A <code>context.Context</code> first, then <strong>serializable</strong> arguments; a result and an <code>error</code> last"
  - Finally, non-deterministic code! Full I/O, any library, any external call
  - Returning an error triggers Temporal's retry logic. No retry code of your own
---

::code::

```go {all|2|2|2|3-17|12-14}
// activities.go
func SubmitOrder(ctx context.Context, order Order) (OrderResult, error) {
	body, err := json.Marshal(order)
	if err != nil {
		return OrderResult{}, err
	}
	resp, err := http.Post(OrdersAPI+"/submit", "application/json", bytes.NewReader(body))
	if err != nil {
		return OrderResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return OrderResult{}, fmt.Errorf("orders API returned %d", resp.StatusCode)
	}
	var result OrderResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}
```

<!--
Source: PPT slides 52-61.
Code: snippets/activities/activities.go, run (against a local test server
standing in for the Orders API) by snippets/activities/activities_test.go.

- **Build 1 - no annotation.** The function becomes an Activity when a Worker
  registers it: `w.RegisterActivity(SubmitOrder)`. Activities that share a
  database pool or an HTTP client are usually methods on a struct, and
  registering a pointer to the struct registers every exported method.
- **Build 2 - the name.** The Activity Type defaults to the function name, or
  the method name for a struct. The Workflow passes the function itself to
  `workflow.ExecuteActivity`, so a typo in the Workflow is a compile error, not
  a runtime surprise. The test reads `SubmitOrder` back out of the
  `ActivityTaskScheduled` Event.
- **Build 3 - the signature.** Builds 1 to 3 point at the same line on purpose.
  `context.Context` comes first: it carries the Activity's deadline and
  cancellation, and `activity.GetInfo(ctx)` and `activity.RecordHeartbeat(ctx,
  ...)` read it. The arguments and the result cross the wire as JSON, so pass
  data, never a client, a channel, or a function. The last result is an
  `error`.
- **Build 4 - just code.** You are outside the Workflow's rules. This is
  ordinary Go: `net/http`, a database driver, the filesystem, `time.Now()`,
  `math/rand`, goroutines, all fine. In production, build the request with
  `http.NewRequestWithContext(ctx, ...)` so a cancelled Activity stops its
  call; the slide keeps `http.Post` short.
- **Build 5 - just return the error.** A non-2xx response is an error, and that
  is the entire error handling. The Activity has no attempt counter, no
  backoff, no retry loop. It returns the error, and the platform retries
  according to the policy the Workflow set. The test makes the API answer 503
  once and sees the Activity succeed on attempt 2. Every error is retried by
  default, so a permanent failure such as a 400 should return
  `temporal.NewNonRetryableApplicationError(...)` or be listed in the Retry
  Policy's `NonRetryableErrorTypes`. The error-handling segment shows how.

Two things to say before moving on:

- Assume every Activity runs more than once. A timeout can fire on an attempt
  that actually succeeded. Make them idempotent.
- Each Activity runs on its own goroutine in the Worker, so a blocking call or
  heavy CPU work in one does not stall the others. The
  limit is how many run at once: `MaxConcurrentActivityExecutionSize` on the
  Worker, which the Worker tuning segment covers.

The sentence to land: the only thing that makes this an Activity is a Worker
that registers it. There is nothing in the function itself.
-->
