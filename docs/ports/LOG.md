# Port log

Start and end of each Port, and the subagent that did it, so
`scripts/port-cost.mjs` can cost each one. UTC.

| Port | Source | Agent id | Start | End |
|---|---|---|---|---|
| Scaffold (main session) | temporal-devdays-ts tooling | — | 2026-11-13T15:35:10Z | 2026-11-13T15:41:43Z |
| Your First Workflow | ts `workshops/first-workflow-money-transfer` | aac8417943e081fcd | 2026-11-13T15:42Z | 2026-11-13T15:54Z |
| Child Workflows | ts `workshops/child-workflows` | a58f4c36d477a7557 | 2026-11-13T15:42Z | 2026-11-13T15:57Z |
| Continue-As-New | ts `workshops/continue-as-new-subscription` | afbe07eaa8490dbc2 | 2026-11-13T15:42Z | 2026-11-13T15:51Z |
| Activity Heartbeats | ts `workshops/heartbeats-shipment` | aa350f8108f0cf7fc | 2026-11-13T15:42Z | 2026-11-13T15:56Z |
| The Saga Pattern | ts `workshops/` + `decks/saga-order-fulfillment` | af00bd6c933760fc4 | 2026-11-13T15:42Z | 2026-11-13T16:04Z |
| Nexus: Decouple a Monolith | edu-nexus-code `typescript/` + ts `decks/nexus-decouple-monolith` | ad332bead930f3b5e | 2026-11-13T15:42Z | 2026-11-13T16:26Z |
| Temporal Essentials Deck | ts `decks/typescript-master` | a3794015bf16aac40 | 2026-11-13T15:42Z | 2026-11-13T16:15Z |

The Essentials Deck agent forked five helpers of its own; `--agent` counts
them through their `parentAgentId`.
