---
layout: eyebrow-hero
eyebrow: OVERVIEW
---

# The Saga pattern

Split a change across services into steps, and give every step an **undo**. If a
later step fails, run the undos for the steps that already happened, in reverse.

<!--
Source: Temporal docs, Saga pattern page (Overview).

Thirty seconds. Every slide from here to Related follows the Saga pattern page in
the Temporal docs, in its order and its words. Only the code is different: it is
the lab's order fulfillment Workflow, from the Saga Workshop in this repo
(workshops/saga-order-fulfillment).
-->

---
layout: default
---

# Problem

- In distributed systems, you need to maintain data consistency across multiple
  services or databases **without traditional ACID transactions**.
- When a multi-step business process fails partway through, you must **undo the
  effects of completed steps** to maintain system consistency.
- Traditional **two-phase commit does not scale** well and creates tight coupling
  between services.

<!--
Source: Temporal docs, Saga pattern page (Problem).

One minute. Make it concrete with the lab's order: the stock is reserved, the
card is charged, and the courier says there is nobody to pick it up. Two effects
landed in systems the Workflow does not own, and nothing takes them back on its
own. Ask the room what their system does next. The honest answer in most shops
is "a ticket."

Durable execution guarantees the orchestration survives. It does not, and
cannot, un-charge a card. The undo is business logic somebody has to write.
-->
