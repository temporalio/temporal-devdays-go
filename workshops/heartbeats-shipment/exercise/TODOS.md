# TODOs

One TODO in three parts, all in Challenge 1. The Track's Check fails until all
three are done, and names the first part still missing.

| TODO | Challenge | File | What to do |
|---|---|---|---|
| 1a | 1 | `workflows.go`, the `workflow.ActivityOptions` in `ShipmentWorkflow` | Add `HeartbeatTimeout: 2 * time.Second`. |
| 1b | 1 | `activities.go`, the `PackShipment` loop | Call `activity.RecordHeartbeat(ctx, packed)` after each item. |
| 1c | 1 | `activities.go`, top of `PackShipment` | When `activity.HasHeartbeatDetails(ctx)`, read the count with `activity.GetHeartbeatDetails(ctx, &packed)`; otherwise keep 0. |

Each part sits between a `// TODO 1x start` and a `// TODO 1x end` comment.
Change only the lines between them.
