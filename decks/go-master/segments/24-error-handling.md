---
layout: section
---

<ZiggyHero scene="shield">

# Error handling

What fails, how it surfaces, and who retries it.

</ZiggyHero>

<!--
Source: PPT slides 156-167.

The original showed each code example twice, once per SDK, side by side. This
Workshop is Go only, so every pair became one Go snippet.
-->

---
layout: default
---

# Error handling concepts (1)

<v-clicks>

- You can categorize failures as **platform** or **application**
  - **Platform:** caused by things beyond your application code's control
  - **Application:** caused by problems with application code or input data
  - Tell them apart by asking: does detecting and fixing it need knowledge of the application?
- You can also classify them by how likely they are to happen again
  - **Transient:** unlikely to recur. Retry with a short delay
  - **Intermittent:** likely to recur. Retry with a longer, increasing delay
  - **Permanent:** guaranteed to recur. Needs manual intervention

</v-clicks>

<!--
Source: PPT slide 157. Eight builds in the original; `v-clicks` gives one per
bullet.

The second classification maps directly onto a Retry Policy. Transient is the
default policy doing its job. Intermittent is a longer `InitialInterval` and a
higher `BackoffCoefficient` in the `temporal.RetryPolicy`. Permanent is a
non-retryable `ApplicationError` (`temporal.NewNonRetryableApplicationError`):
retrying it is only burning money.
-->

---
layout: default
---

# Error handling concepts (2)

<v-clicks>

- **Idempotency** is a general concern for distributed systems
  - Will multiple invocations of your operation change application state for the worse?
  - It matters for Activities, because they may run more than once
  - Temporal strongly recommends making your Activities idempotent
- An `error` returned from Workflow code **fails the Workflow Execution**
- A **panic** fails the **Workflow Task**, which Temporal retries
- `temporal.NewApplicationError` is the error to return on purpose

</v-clicks>

<!--
Source: PPT slide 158. The original's second half was about Go's error return
values and panics, so this half is the original's own point.

- `return ..., err` from a Workflow function ends the Workflow Execution as
  Failed, whatever the error. A plain `errors.New` or `fmt.Errorf` error is
  converted to an `ApplicationError` on the way out. Return
  `temporal.NewApplicationError(msg, errType)` when you want to choose the type
  a caller matches on. Either way it is a deliberate business decision.
- A panic (a nil pointer, an index out of range) is treated as a bug. So is
  non-determinism the Worker detects on replay. The Workflow Task fails and is
  retried, and the Workflow stays open so you can fix the code and redeploy.
  The UI shows it as a failing Workflow Task, not a failed Workflow.
- To make a panic fail the Workflow instead, set
  `worker.Options{WorkflowPanicPolicy: worker.FailWorkflow}`. The default is
  `worker.BlockWorkflow`, which is the behaviour above.

snippets/errorhandling/errorhandling_test.go proves all three.
-->

---
layout: default
---

# The Temporal error model

<v-clicks>

- Temporal is polyglot: a Workflow and its Activities can be written in different languages
- If an Activity returns an error, it must reach the Workflow
- That has to work whichever languages the Workflow and Activity use
- So, as with data, errors cross language boundaries
- Errors are serialized in a language-neutral format (protobuf)

</v-clicks>

<!--
Source: PPT slide 159.

Temporal's error model can seem counterintuitive at first. Normally, when a
called function returns an error, you expect to get that exact error back in
the caller. Because of Temporal's design, you get a wrapper instead, and the
original error is in its chain.

Each SDK uses its own native error handling; Temporal converts the errors into
its own failure types: Application Failures, Task failures, and so on. Under the
hood, Temporal talks gRPC, which uses protobufs, and it has a set of protobuf
messages for failures that preserve context (message, type, stack trace, cause
chain) across process and language boundaries.
-->

---
layout: default
---

# How Temporal represents failures

<v-clicks>

- Every failure in Temporal is represented as a **Temporal failure**
- In Go they are error types in `go.temporal.io/sdk/temporal`, and you inspect them with `errors.As`
- An error returned by an Activity reaches the Workflow as an **`*ActivityError`**; a failed Child Workflow as a **`*ChildWorkflowExecutionError`**
- `Unwrap` gives what actually happened: `*ApplicationError`, `*TimeoutError`, `*CanceledError`
- You can check for and handle it in your Workflow, if you want

</v-clicks>

<!--
Source: PPT slide 160.

In Go the failures are ordinary `error` values that wrap one another, so the
standard library does the work: `errors.As(err, &appErr)` walks the whole
chain, and `temporal.IsTimeoutError`, `temporal.IsCanceledError` and
`temporal.IsApplicationError` are shorthands for the common checks.

On the client, a Workflow that failed makes `run.Get(ctx, &result)` return a
`*temporal.WorkflowExecutionError`. That is the client's wrapper, and its cause
is the Workflow's failure, such as the `ApplicationError` the Workflow
returned. `errors.As` reaches it through the wrapper.

Cancellation can arrive wrapped, as an `*ActivityError` or
`*ChildWorkflowExecutionError` whose cause is a `*CanceledError`. To ask "is
this a cancellation?" whatever the wrapping, use `temporal.IsCanceledError(err)`.

Don't define your own types to stand in for these. The error to return on
purpose is the one `temporal.NewApplicationError` (or its `WithOptions` and
`NonRetryable` variants) builds.
-->

---
layout: code-walkthrough
heading: Raising errors from an Activity
notes:
  - A message for humans. It shows up in the UI and in the Workflow's <code class="ref">ActivityError</code> chain
  - A <strong>type</strong> for code. Retry Policies and Workflows match on it
  - Tell Temporal not to retry. This is a business rule, not a blip
---

::code::

```go {all|4|5|6}
func ChargePayment(ctx context.Context, order Order) (string, error) {
	if !cardHasFunds(ctx, order) {
		return "", temporal.NewApplicationErrorWithOptions(
			"payment failed: insufficient funds",
			"PaymentError",
			temporal.ApplicationErrorOptions{NonRetryable: true},
		)
	}
	return gateway.Charge(ctx, order)
}
```

<!--
Source: PPT slide 161.
Code: snippets/errorhandling/activities.go, run by
snippets/errorhandling/errorhandling_test.go, which counts one attempt.

- **Build 1 -** the message.
- **Build 2 -** the type. Any other error an Activity returns becomes an
  `ApplicationError` on the way out, typed with its Go type's name (an
  `errors.New` error has no type) and retryable. Returning an
  `ApplicationError` yourself lets you choose the type, attach `Details`, and
  mark it non-retryable.
- **Build 3 -** `NonRetryable: true`. The shorthand is
  `temporal.NewNonRetryableApplicationError(message, errType, cause)`. The
  alternative, when the caller should decide, is
  `NonRetryableErrorTypes: []string{"PaymentError"}` in the `RetryPolicy` of
  the Workflow's Activity options.

Everything else, a timeout or a connection reset from `gateway.Charge`, is
returned as it is, and the Retry Policy handles it. That is the right default.
-->

---
layout: code-walkthrough
heading: Handling an ActivityError in a Workflow
notes:
  - Every Activity failure arrives as an <code>ActivityError</code>
  - "<code>Unwrap</code> it to see what actually happened"
  - Timed out, canceled, or a custom application error with its <code>Type()</code>
  - Nothing matched. Return it, and the Workflow fails with the full cause chain
---

::code::

```go {all|5-6|8|9-14|17}
func PaymentWorkflow(ctx workflow.Context, order Order) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, paymentOptions)
	var receipt string
	err := workflow.ExecuteActivity(ctx, ChargePayment, order).Get(ctx, &receipt)
	var activityErr *temporal.ActivityError
	if errors.As(err, &activityErr) {
		var appErr *temporal.ApplicationError
		switch cause := activityErr.Unwrap(); {
		case temporal.IsTimeoutError(cause):
			// the Activity timed out
		case temporal.IsCanceledError(cause):
			// the Activity was canceled
		case errors.As(cause, &appErr) && appErr.Type() == "PaymentError":
			return requestNewCard(ctx, order)
		}
	}
	return receipt, err
}
```

<!--
Source: PPT slide 162.
Code: snippets/errorhandling/workflows.go, run by
snippets/errorhandling/errorhandling_test.go.

`paymentOptions` is a `workflow.ActivityOptions` with
`StartToCloseTimeout: 30 * time.Second`, off the slide to save room; say it.

`errors.As` is both the check and the conversion: it finds the
`*ActivityError` in the chain and assigns it. A nil `err` matches nothing, so
the success path falls straight through to the last line.

The branch that does real work is the `PaymentError` one: a business outcome,
so the Workflow takes a different path, asking for a new card, rather than
failing. That is what "errors are part of the business logic" means.

`appErr.Type()` is the string the Activity chose on the last slide. It is how a
Workflow tells failures apart without sharing error types with the Activity,
which may live in another service, or another language.

A Child Workflow works the same way: the child's future returns a
`*temporal.ChildWorkflowExecutionError`, and its cause is the child's own
error.
-->

---
layout: code-stack
heading: Failing a Workflow with ApplicationError
proseMax: 32
---

::code::

```go
func CriticalWorkflow(ctx workflow.Context, order Order) (OrderResult, error) {
	// Returning an error fails the Workflow Execution
	return OrderResult{}, temporal.NewApplicationError(
		"unrecoverable state",
		"CriticalError",
	)
}
```

::default::

An `error` returned from Workflow code **fails the Workflow Execution**. A
panic fails only the **Workflow Task**, and Temporal retries it until the code
is fixed.

<!--
Source: PPT slide 163.
Code: snippets/errorhandling/workflows.go, run by
snippets/errorhandling/errorhandling_test.go.

Say the consequence of getting this backwards. A panic, a nil map write, an
index out of range, does not fail the Workflow. It sits there, retrying its
Workflow Task, because Temporal assumes it is a bug you will deploy a fix for.
If you mean "this Workflow has failed," say so by returning an error, and make
it an `ApplicationError` so the caller can match on its type.

On the client, the caller waiting on `run.Get` gets a
`*temporal.WorkflowExecutionError`, and `errors.As` finds this
`*temporal.ApplicationError` in its chain, with `Type()` `"CriticalError"`.
-->

---
layout: default
---

# Raising and handling errors

- An error returned from an **Activity** fails that attempt
  - It is retried according to the policy, unless it is **non-retryable**
  - Only the final failure is an Event: `ActivityTaskFailed`, with the error details
- An `error` returned from a **Workflow** fails the Workflow Execution
- A **panic** in a Workflow fails the **Workflow Task**, and is retried
- Return **non-retryable** errors when you don't want an Activity retried

<!--
Source: PPT slide 164.

The summary slide for the last four. If the room has a UI open, fail an
Activity and show the `ActivityTaskFailed` Event: the message, the type, the
stack trace and the retry state are there. The attempt count is on the
`ActivityTaskStarted` Event just before it. Failed attempts before the last one
leave no Event (Timeouts and Retry Policies segment).

A panic in an Activity is different from one in a Workflow: the Worker recovers
it and fails that attempt with a `*temporal.PanicError`, which is retried like
any other error.
-->

---
layout: default
---

# Recovering from failure

Temporal gives you a few options for recovering from persistent failure.

- **Cancel** a Workflow Execution: graceful, allows cleanup before closing
- **Terminate** a Workflow Execution: forceful, no cleanup
- **Reset** a Workflow Execution: continue from an earlier point in its Event History
- The application may also support **rolling back** to a previous state
  - Often with the **Saga pattern**: a series of related operations, each
    depending on the one before
  - On failure, compensating transactions revert the changes

<!--
Source: PPT slide 167.

In Go, cancellation reaches the Workflow through its `workflow.Context`: the
context is done (`ctx.Err()` is `workflow.ErrCanceled`), and whatever it was
waiting on returns a `*temporal.CanceledError`, or an error wrapping one. Check
it with `temporal.IsCanceledError(err)`, run cleanup Activities on
`workflow.NewDisconnectedContext(ctx)` so they are not cancelled too, and
return the error.

Reset is the one people do not know about and love once they do: after
deploying a fix for a bug that failed a batch of Workflows, reset them to just
before the failing Event and they re-run on the new code.

`temporal workflow reset --workflow-id <id> --type LastWorkflowTask` is the
common form.
-->
