---
layout: default
clicks: 9
---

# Inside one Workflow run

<WorkflowRun />

<!--
Back to concepts after the two anatomy slides: the room has now seen
`workflow.ExecuteActivity` and a Future's `Get`, so this is what that call does
on the wire. One click per step; the line under the diagram
says what each step is, and the colours are the four kinds of thing in play.

1. **Worker action.** A Worker is running a Workflow Task, and your code calls
   `workflow.ExecuteActivity(ctx, ChargeCard, ...)` and then `Get` on the
   Future it returns.
2. **Command.** The call becomes a `ScheduleActivityTask` Command. The Worker
   sends it to the Service when the Workflow Task finishes, not the moment the
   line runs.
3. **Event.** The Service records `ActivityTaskScheduled` in the Event History.
4. **Task.** The Service puts an Activity Task on the Task Queue.
5. **Worker action.** A Worker polls and takes the Activity Task. It may be a
   different Worker, on a different machine.
6. **Worker action.** The Worker runs `ChargeCard`.
7. **Event.** On success, the Worker reports back and the Service records
   `ActivityTaskCompleted`. On failure, it records the failure or schedules a
   retry.
8. **Task.** The Service puts a new Workflow Task on the Task Queue.
9. **Worker action.** A Worker picks up that Workflow Task; `Get` returns with
   the recorded result, and the next line of your code runs. Usually the Worker still has this Workflow in memory and
   just hands it the result. If it does not (it crashed, restarted, or this is
   a different Worker), it first replays the Workflow from the top against the
   Event History to get back to this line. That replay is the next slides.

Two details worth knowing. The Service writes `ActivityTaskStarted` only after
the attempt finishes, not at pickup, so a retrying Activity does not flood the
history with Started Events; that is why it appears at step 7, faded. And the
"next line runs" part depends on a second Workflow Task, not on the Activity
Worker handing control straight back.

Next: why the Service keeps that history, and then the same machinery with a
crash in the middle.
-->

---
layout: default
---

# Event History: why replay works

Every step of a Workflow is recorded as an Event. If a Worker dies, a new one
**replays** the history to rebuild the Workflow's state. Determinism is what
makes that safe, and it is why Workflow code has strict rules.

<div v-click class="mt-8 text-xl font-light">

Event History is like a `git log` for your Workflow Execution. You can always
reconstruct the current state by replaying from the beginning.

</div>

<!--
Source: PPT slide 26.

The previous slide showed the Event History being written, one Event per
step. This is why it is kept: it is what replay reads.

Make replay concrete: the Worker runs your Workflow function from the top. When
it reaches an Activity call, it does not run the Activity again. It finds the
`ActivityTaskCompleted` Event in history and the Future's `Get` returns the
recorded result. The function races forward through history to where it left
off, then carries on for real.

Which is exactly why reading anything from the outside world in Workflow code
is a bug: a config file, an environment variable, a database row. The replay
could take a different branch than the original run did, and the history would
no longer match the code. Temporal calls that a non-determinism error. The SDK
gives you safe versions of the common traps (`workflow.Now`, `workflow.Sleep`,
`workflow.Go`, `workflow.SideEffect`), and the `workflowcheck` analyzer flags
the unsafe ones, but neither can see what your own I/O returns.
-->

---
layout: default
clicks: 8
---

# Event History and Replay

<ReplayFlow />

<!--
Source: docs.temporal.io/encyclopedia/event-history, whose example this
follows: a distance lookup, a 30-minute timer, a bill. It replaces the
"Putting it together" list from PPT slide 27.

One click per step. Clicks 1 and 2 separate the Command from the work: calling
`GetDistance` through `workflow.ExecuteActivity` only asks the Service to
schedule it;
the Activity runs on its own Activity Task, and only then are its Started and
Completed Events written. The beat to land is click 5: Worker B runs the same
code from the top, but the `GetDistance` call does not send its Command. The
Worker finds `ActivityTaskScheduled` in the history, takes 15 from
`ActivityTaskCompleted`, and moves on. Activities are not re-executed during
replay; their results are read back.

Workflow Task events (`WorkflowTaskScheduled`, `Started`, `Completed`, and the
`WorkflowTaskTimedOut` that a crash produces) are left out to keep the history
readable. In the real history they sit between every pair of steps here. Also
left out: `ActivityTaskStarted` is written when the attempt closes, not when it
starts, because it carries the attempt count.

If someone asks what breaks this: replay compares each Command the code
produces with the next one the history expects. Branch on something read from
outside the Workflow, a config file or a database row, and replay may produce
`ScheduleActivityTask` where the history says `StartTimer`. That is a
non-determinism error. It fails the Workflow Task, not the Workflow, and
retries until you fix the code or version it.
-->

---
layout: default
---

# More replay-aware methods

Workflow-safe tools for the things you would normally reach for.

| In a Workflow, use                          | For                                             |
| ------------------------------------------- | ----------------------------------------------- |
| `workflow.Sleep()` `workflow.NewTimer()`    | Waiting between checks. Durable across restarts |
| `workflow.Await()`                          | Waiting for a Signal to flip a flag             |
| `workflow.Go()` and Futures                 | Running Activities in parallel                  |
| `workflow.NewSelector()`                    | Whichever finishes first                        |
| `workflow.Now()` `workflow.SideEffect()`    | Time and randomness that replay identically     |

<!--
Source: PPT slide 28.

The original listed five helpers: sleep, await a condition, wait for all, wait
for any, and a new Timer. The Go mapping:

- `workflow.Sleep(ctx, d)` is a durable Timer. A Workflow can sleep for thirty
  days and it costs nothing while it waits. `workflow.NewTimer(ctx, d)` is the
  same Timer as a Future, for when you want to wait on it alongside something
  else.
- `workflow.Await(ctx, func() bool {...})` returns when the function you pass
  returns true. `workflow.AwaitWithTimeout` returns `false` if time runs out
  first. A Signal handled in a coroutine typically flips the flag it waits on.
- Parallel work needs no special API: `workflow.ExecuteActivity` returns a
  Future without blocking, so start three and `Get` each in turn. For work that
  has to block on its own, `workflow.Go` starts a coroutine, the Workflow's
  replay-safe goroutine.
- `workflow.NewSelector` is the Workflow's `select`: add Futures, channels and
  Timers to it and `Select` returns on whichever is ready first, in the order
  history records.
- `workflow.Now(ctx)` returns the Workflow Task's time. Go has no seeded random
  in a Workflow; draw the value inside `workflow.SideEffect` and it is recorded
  once and read back on replay. A UUID is the same: `uuid.New()` inside a
  `SideEffect`.
- Never `time.Now`, `time.Sleep`, the `go` keyword, a native channel or
  `select`, or ranging over a map (Go randomises map order) in Workflow code.
  The Workflows segment has the full list, and `workflowcheck` finds them.
-->
