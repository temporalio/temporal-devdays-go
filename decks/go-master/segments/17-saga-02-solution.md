---
layout: default
---

# Solution

- You implement each step as a **local transaction** with a corresponding
  **compensation transaction**.
- If any step fails, you execute compensation transactions **in reverse order** to
  undo the effects of all completed steps.
- You register compensations as each step runs, then automatically trigger them when
  errors occur to ensure cleanup happens reliably.

<!--
Source: Temporal docs, Saga pattern page (Solution).

One minute. The worked example is an order whose courier fails after the stock is reserved and the card is charged.
-->

---
layout: default
clicks: 5
---

# The Problem

<SagaFlow panes="without" />

<!--
Source: public/diagrams/saga-flow.html, the "Without compensations" pane.

Two minutes, five clicks: one order whose courier fails, with no
compensations.

Stop on the last step and read the ledger card out loud. `reserve` and
`charge` are marked stranded. The Workflow failed correctly, and two other
systems are still wrong: the stock is held and the card is charged, for a
parcel that will never move. Ask the room what would undo them. A retry will
not: the courier error is non-retryable, and retrying `ChargePayment` would
charge the card twice.

The node names and files are real; the order id is an example. The full page,
with the "With compensations" half as well, is public/diagrams/saga-flow.html:
open /diagrams/saga-flow.html in the dev server.
-->
