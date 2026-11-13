---
layout: comparison
leftEyebrow: WHEN TO USE
leftTitle: A good fit when
leftFeatures:
  - "You need consistency across multiple services or databases"
  - "Two-phase commit is too slow or unavailable"
  - "You can define compensating actions for each step"
  - "Eventual consistency is acceptable"
  - "Transactions may span hours or days"
rightEyebrow: WHEN TO USE
rightTitle: Not a good fit for
rightFeatures:
  - "Operations that require strong ACID consistency"
  - "Single-service transactions that can use a local database transaction"
  - "Processes where compensations cannot be defined"
  - "Operations that must appear atomic to external observers"
---

<!--
Source: Temporal docs, Saga pattern page (When to use).

One minute. "When to use" from the docs page. Read the right-hand column more
slowly than the left: it is the one people skip.
-->

---
layout: comparison
leftEyebrow: BENEFITS AND TRADE-OFFS
leftTitle: Benefits
leftFeatures:
  - "Eventual consistency without distributed locks"
  - "Each service can use its own database and transaction model"
  - "Durable execution guarantees compensations run even after Worker failures"
  - "Scales better than two-phase commit"
rightEyebrow: BENEFITS AND TRADE-OFFS
rightTitle: Trade-offs
rightFeatures:
  - "Only eventual consistency: intermediate states are visible to other processes"
  - "Compensation Activities must be idempotent"
  - "Compensation logic must be maintained alongside forward logic"
  - "Some operations may not have meaningful compensations"
---

<!--
Source: Temporal docs, Saga pattern page (Benefits and trade-offs).

One minute. "Benefits and trade-offs" from the docs page.

The benefit to land is the third: the compensations slice lives in Workflow state,
so it survives a Worker dying mid-unwind. The trade-off to land is the first: the
order was reserved and charged for a few hundred milliseconds and the world saw
it. A Saga promises it does not stay that way.
-->

---
layout: default
---

# Comparison with alternatives

<div class="compare-table">

| Approach             | Consistency   | Rollback mechanism         | Coupling   | Scalability    |
| :------------------- | :------------ | :------------------------- | :--------- | :------------- |
| Saga (orchestration) | Eventual      | Compensating transactions  | Loose      | High           |
| Two-phase commit     | Strong (ACID) | Distributed lock/rollback  | Tight      | Low            |
| Saga (choreography)  | Eventual      | Event-driven compensations | Very loose | High           |
| Local transaction    | Strong (ACID) | Database rollback          | None       | Single service |

</div>

<style>
.compare-table { font-size: 0.8em; }
.compare-table td, .compare-table th { padding-top: 0.35em; padding-bottom: 0.35em; }
</style>

<!--
Source: Temporal docs, Saga pattern page (Comparison with alternatives).

One minute. The table from the docs page, verbatim. The lab is the first row:
one Workflow orchestrates every step and every compensation.

The row worth a sentence is choreography: each service reacts to the previous
one's events. Looser coupling, but "which compensations have run for this
order?" has no single place to look. With orchestration, the answer is the
Workflow's History.
-->
