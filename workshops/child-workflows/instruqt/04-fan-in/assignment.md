---
slug: fan-in
id: ubs9pxaqnrzu
type: challenge
title: 4. Fan In
teaser: Wait for every child and collect the results. Five 3-second children, done
  in about 3 seconds.
notes:
- type: text
  contents: |-
    # Five children, three seconds each. How long does the parent take?

    Every child you started in the last Challenge left a Future behind. Wait
    on all of them and the parent stays open until every child is done, while
    they all run at the same time.
tabs:
- id: ktoyb5snzkx9
  title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/child-workflows
  port: 8080
- id: kp5czbq68yf3
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/child-workflows/exercise
- id: fhduhffzstjd
  title: Client
  type: terminal
  hostname: workshop
  workdir: /root/workshop/child-workflows/exercise
- id: pan08smhrcfn
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: basic
timelimit: 600
enhanced_loading: null
---

# Goal

Make `PackWorkflow` wait for every child and return their boxes, so no child
is terminated and the order comes back fully packed.

**Why it matters:** the order is only packed when every box is. A parent that
returns early ships an empty order and throws the packing work away.

# Wait for every child (TODO 1)

**Observe.** Open `exercise/workflows/pack.go` in the
[button label="Exercise" background="#444CE7"](tab-0) tab. Press `Ctrl+P`
(`Cmd+P` on a Mac), type the file name, then `Ctrl+G` to jump to a line.

<!-- snippet: exercise/workflows/pack.go TODO 1 -->
Lines 32-37 are the end of `PackWorkflow`, right after the fan-out loops:

```go
	// TODO 1 start
	// Fan in: call Get on every Future in boxes, in item order, and return the
	// boxes they hold. Until you do, the parent returns at once and the Parent
	// Close Policy above terminates every child.
	return []string{}, nil
	// TODO 1 end
```

**Learn.** `return []string{}, nil` returns before any child has finished, and
that is why the last Challenge's children were terminated. `boxes` is the
`[]workflow.ChildWorkflowFuture` the fan-out loop filled, one Future per child,
in item order. `Get(ctx, &value)` blocks until that child finishes, decodes its
result into `value`, and returns its error, if any.

Calling `Get` on each Future in turn waits for every child, and it does not
make them run one after another: they were all started before the first `Get`,
so the loop takes as long as the slowest child, not the sum. The results land
in item order whatever order the children finish in.

`Get` blocks the Workflow the deterministic way, through the SDK's own
scheduler, so it replays the same every time. That is why Workflow code waits
on Futures, not on goroutines, channels or a `sync.WaitGroup`. To act on each
child the moment it finishes instead, in completion order, add the Futures to
a `workflow.NewSelector`.

**Do.** Between the TODO 1 markers:

- [ ] Call `Get` on every Future in `boxes`, in item order, and keep each box.
- [ ] Return the boxes instead of an empty slice, and return the error if a child failed.

Leave the fan-out loops as they are: they start every child before you wait for any.

<details>
<summary>Hint</summary>

The same shape, on numbers:

```go
sizes := []workflow.Future{ /* ... */ }
values := make([]int, len(sizes))
for i, f := range sizes {
	if err := f.Get(ctx, &values[i]); err != nil { // one value per Future, same order
		return nil, err
	}
}
```

</details>

<details>
<summary>Show solution</summary>

```go
	packed := make([]string, len(boxes))
	for i, box := range boxes {
		if err := box.Get(ctx, &packed[i]); err != nil {
			return nil, err
		}
	}
	return packed, nil
```

</details>

# Run it

A running Worker keeps the code it was built from. If one is still running,
stop it with `Ctrl+C`. Then, in the
[button label="Worker" background="#444CE7"](tab-1) tab, start it on your new
code:

```bash,run
go run ./worker
```

```bash,nocopy
Worker polling child-workflows. Ctrl+C to stop.
```

In the [button label="Client" background="#444CE7"](tab-2) tab:

```bash,run
go run ./starter pack
```

```bash,nocopy
Packed 5 of 5 items in 3.1s: [box:mug, box:book, box:lamp, box:plant, box:clock]
```

Five children of 3 seconds each, in about 3 seconds, not 15. They ran in
parallel.

Open the newest `pack-ORD-...` Workflow in the
[button label="Temporal UI" background="#444CE7"](tab-3) tab. If it isn't
there, click the refresh button. All five **Start Child Workflow Execution
Initiated** events come first, then five **Child Workflow Execution
Completed**. This time every child is **Completed**, none Terminated.

# What you saw

- One child, started with `workflow.ExecuteChildWorkflow` and waited for with
  `Get`, with its own history.
- A chain five deep, where one error came up through five
  `ChildWorkflowExecutionError` wrappers.
- A fan out that kept one `ChildWorkflowFuture` per child, and
  `PARENT_CLOSE_POLICY_TERMINATE` ending the children the parent did not wait
  for.
- A fan in that called `Get` on every Future, five children in the time of one.

# Checkpoint

Click **Check**. It passes when TODO 1's tests pass and five children have
completed under one parent in the Temporal dev server. If it fails, the message
names the file and the TODO still to do, or what to run.

If you used **Skip**, it stopped your Worker to run the new code. Start it
again with `go run ./worker` to keep going.

---

**Please share your feedback so we can make better content for you.** The
**Feedback** tab takes a few seconds, and it is the only way we find out which
parts of this landed.
