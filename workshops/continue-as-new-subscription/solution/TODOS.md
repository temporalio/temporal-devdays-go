# TODOs

One TODO in this Workshop. The Track's Check fails until it is done.

| TODO | Challenge | File | What to do |
|---|---|---|---|
| 1 | 1. Continue-As-New | `subscription/workflow.go`, `SubscriptionWorkflow` | When `workflow.GetInfo(ctx).GetContinueAsNewSuggested()` is true, return `workflow.NewContinueAsNewError(ctx, SubscriptionWorkflow, ...)` with a `SubscriptionInput` carrying this run's `cyclesCompleted`. |

The TODO sits between a `// TODO 1 start` and a `// TODO 1 end` comment. Change
only the lines between them.
