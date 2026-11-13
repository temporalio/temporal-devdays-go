---
layout: default
---

# Best practices

- **Make all compensations idempotent.** They may run when the forward Activity never
  executed, or run multiple times on retry. Use idempotency keys for safe re-execution.
- **Register compensations before Activity Execution.** Cleanup runs even if the
  Activity fails after partial completion; the compensation no-ops if nothing happened.
- **Use idempotency keys for forward Activities.** Pass a unique identifier, such as
  a Workflow ID, so retries do not create duplicate side effects.
- **Set `StartToCloseTimeout` on compensation Activities.** Avoid `ScheduleToCloseTimeout`
  on compensations, and Workflow-level timeouts. Let compensations retry until they
  succeed.

<!--
Source: Temporal docs, Saga pattern page (Best practices).

Two minutes for both best practices slides. These are the docs page's best
practices, in its order.

Say which ones the lab code shows and which it deliberately does not:
- Shown: idempotent compensations (the `IfReserved` / `IfCharged` guards),
  register before execution, `StartToCloseTimeout` on every
  Activity.
- Not shown: idempotency keys (`ChargePayment` fails non-retryable instead), and
  retrying compensations until they succeed. The lab caps every Activity at two
  attempts so failures show up in seconds. Say this out loud: it is the line
  most likely to be copied into production by mistake.
-->

---
layout: default
---

# Best practices, continued

- **Keep compensation payloads small.** Pass references (IDs, URLs) instead of full
  data objects to avoid exceeding the 2 MB payload limit.
- **Log compensation failures but continue.** If a compensation fails, log the error
  and continue with the remaining compensations. In production, alert for manual
  intervention on persistent failures.
- **Return the original error after compensating.** Always return the original
  error after running compensations so the Workflow reports the correct failure
  reason.

<!--
Source: Temporal docs, Saga pattern page (Best practices).

In the lab code: log and continue is the `if cerr := ...` check inside the
unwind loop, and returning the original error is the deferred function leaving
the named `err` result alone. The lab
passes the whole `Order`, a handful of small fields; pass a reference once the
input can grow.
-->

---
layout: default
---

# Common pitfalls

<div class="dense">

- **Non-idempotent compensations.** Compensations may run when the forward Activity
  never executed, or multiple times on retry. All compensations must be idempotent.
- **Forgetting to register a compensation.** If a step succeeds but its compensation
  was never registered, a later failure leaves that step's effects in place.
- **Compensations that can fail permanently.** If a compensation fails with a
  non-retryable error, the Saga cannot fully roll back. Use generous Retry Policies.
- **Large payloads in compensation state.** Passing large objects through the
  compensation chain can exceed the 2 MB payload limit. Use references instead.

</div>

<style>
.dense :is(p, li) { font-size: 1.2rem; line-height: 1.45; }
</style>

<!--
Source: Temporal docs, Saga pattern page (Common pitfalls).

One minute. The docs page's pitfalls.
-->
