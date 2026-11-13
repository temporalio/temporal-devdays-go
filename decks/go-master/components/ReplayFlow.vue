<!--
  ABOUTME: Event History and Replay, a click at a time: Worker A runs a Workflow and crashes,
  ABOUTME: Worker B replays the code against the Event History, then carries on for real.

  The Activity panel shows GetDistance running on its own click, then struck
  through during replay, where its result is read from history instead.

  Follows the example on docs.temporal.io/encyclopedia/event-history: a
  distance lookup, a 30-minute timer, a bill. Set `clicks:` to
  STEPS.length - 1. The code is Go; error checks are left out of the Workflow
  so each step is one line, and the first line of CODE says so.

  Workflow Task events are left out to keep the history readable; the presenter
  notes say so. An event appears on the click in its `at`; a step's `matched`
  lists the events its Command is matched against during replay.
-->
<script setup lang="ts">
import { computed } from 'vue'
import { useSlideContext } from '@slidev/client'

type Worker = { name: string; state: 'running' | 'crashed' | 'replaying' }
type Step = {
  line: number
  activity?: 'running' | 'skipped'
  worker?: Worker
  command?: { text: string; mode: 'sent' | 'matched' }
  matched?: number[]
  say: string
}

const CODE = [
  '// error checks left out to fit',
  'func PizzaWorkflow(ctx workflow.Context, o Order) (Bill, error) {',
  '  var distance int',
  '  f := workflow.ExecuteActivity(ctx, GetDistance, o.Address)',
  '  f.Get(ctx, &distance)',
  '  workflow.Sleep(ctx, 30*time.Minute)',
  '  var bill Bill',
  '  f = workflow.ExecuteActivity(ctx, SendBill, o, distance)',
  '  err := f.Get(ctx, &bill)',
  '  return bill, err',
  '}',
]

const ACTIVITY = [
  'func GetDistance(ctx context.Context, addr string) (int, error) {',
  '  return maps.Distance(ctx, Store, addr)',
  '}',
]

const EVENTS = [
  { name: 'WorkflowExecutionStarted', detail: 'input: order', at: 1 },
  { name: 'ActivityTaskScheduled', detail: 'GetDistance', at: 1 },
  { name: 'ActivityTaskStarted', at: 2 },
  { name: 'ActivityTaskCompleted', detail: 'result: 15', at: 2 },
  { name: 'TimerStarted', detail: '30m', at: 3 },
  { name: 'TimerFired', at: 3 },
  { name: 'ActivityTaskScheduled', detail: 'SendBill', at: 7 },
  { name: 'ActivityTaskStarted', at: 7 },
  { name: 'ActivityTaskCompleted', detail: 'result: bill', at: 7 },
  { name: 'WorkflowExecutionCompleted', at: 8 },
]

const A_RUN: Worker = { name: 'Worker A', state: 'running' }
const B_REPLAY: Worker = { name: 'Worker B', state: 'replaying' }
const B_RUN: Worker = { name: 'Worker B', state: 'running' }

const STEPS: Step[] = [
  {
    line: -1,
    say: 'Workflow code on the left. On the right, the Event History the Temporal Service keeps for this Workflow Execution.',
  },
  {
    line: 3,
    worker: A_RUN,
    command: { text: 'ScheduleActivityTask', mode: 'sent' },
    say: 'Worker A runs the Workflow code. workflow.ExecuteActivity does not run GetDistance: it sends a Command, and the Service records ActivityTaskScheduled.',
  },
  {
    line: 4,
    activity: 'running',
    worker: A_RUN,
    say: 'A Worker picks up the Activity Task and runs GetDistance, a real call to a maps API. It returns 15, and the Service records ActivityTaskStarted and ActivityTaskCompleted.',
  },
  {
    line: 5,
    worker: A_RUN,
    command: { text: 'StartTimer', mode: 'sent' },
    say: 'workflow.Sleep sends StartTimer. Thirty minutes later the Service records TimerFired.',
  },
  {
    line: -1,
    worker: { name: 'Worker A', state: 'crashed' },
    say: 'Worker A crashes. Every variable in its memory is gone. The Event History is not: the Service persisted it.',
  },
  {
    line: 3,
    worker: B_REPLAY,
    command: { text: 'ScheduleActivityTask', mode: 'matched' },
    matched: [1, 3],
    activity: 'skipped',
    say: 'Worker B fetches the history and runs the code from the top. The Command matches an Event, so it is not sent: distance = 15 comes from history, and GetDistance does not run again.',
  },
  {
    line: 5,
    worker: B_REPLAY,
    command: { text: 'StartTimer', mode: 'matched' },
    matched: [4, 5],
    say: 'The timer already fired, so workflow.Sleep returns at once. No second thirty-minute wait.',
  },
  {
    line: 7,
    worker: B_RUN,
    command: { text: 'ScheduleActivityTask', mode: 'sent' },
    say: 'Past the end of the history there is nothing left to match. Replay is over, Commands are sent for real, and SendBill runs.',
  },
  {
    line: 9,
    worker: B_RUN,
    command: { text: 'CompleteWorkflowExecution', mode: 'sent' },
    say: 'The Workflow completes. Same code, same history, same state, despite the crash.',
  },
]

const { $clicks } = useSlideContext()
const n = computed(() => Math.min(Math.max($clicks.value, 0), STEPS.length - 1))
const step = computed(() => STEPS[n.value])
// Let a long Command name like CompleteWorkflowExecution wrap between its words.
const wrapAtWords = (name: string) => name.replace(/(?=[A-Z])/g, '\u200B').slice(1)
const visible = computed(() => EVENTS.map((e, i) => ({ ...e, i })).filter((e) => e.at <= n.value))
</script>

<template>
  <div class="rp">
    <div class="panes">
      <div class="pane worker-box" :class="step.worker?.state">
        <div class="worker-head">
          <span class="worker-name">{{ step.worker?.name ?? 'Worker' }}</span>
          <span v-if="step.worker" class="worker-state">
            {{ step.worker.state === 'crashed' ? '✕ crashed' : step.worker.state === 'replaying' ? '↺ replaying' : '● running' }}
          </span>
        </div>
        <span class="pane-title sub">Workflow code</span>
        <pre class="code"><code><span
          v-for="(l, i) in CODE"
          :key="i"
          class="ln"
          :class="{ cur: i === step.line, replay: i === step.line && step.worker?.state === 'replaying' }"
        >{{ l }}
</span></code></pre>
        <div class="act" :class="step.activity">
          <div class="act-head">
            <span class="pane-title">Activity code</span>
            <span v-if="step.activity === 'skipped'" class="act-tag skipped">↺ not run: result read from history</span>
          </div>
          <pre class="code"><code><span
            v-for="(l, i) in ACTIVITY"
            :key="i"
            class="ln"
            :class="{ cur: step.activity === 'running' && i === 1 }"
          >{{ l }}
</span></code></pre>
        </div>
      </div>

      <div class="cmd-lane">
        <div v-if="step.command" :key="n" class="cmd" :class="step.command.mode">
          <div class="cmd-box">
            <span class="cmd-label">Command</span>
            <code>{{ wrapAtWords(step.command.text) }}</code>
          </div>
          <svg v-if="step.command.mode === 'sent'" viewBox="0 0 100 16" class="cmd-arrow" aria-label="sent to the Service">
            <line x1="2" y1="8" x2="90" y2="8" />
            <path d="M88,2 L98,8 L88,14 z" />
          </svg>
          <span v-else class="cmd-mode">↺ matched,<br />not sent</span>
        </div>
      </div>

      <div class="pane history">
        <div class="pane-head">
          <span class="pane-title h">Event History</span>
          <span class="persisted">persisted by the Temporal Service</span>
        </div>
        <ol class="events">
          <li
            v-for="e in visible"
            :key="e.i"
            class="ev"
            :class="{ fresh: e.at === n && !step.matched, matched: step.matched?.includes(e.i) }"
          >
            <span class="num">{{ e.i + 1 }}</span>
            <span class="name">{{ e.name }}</span>
            <span v-if="e.detail" class="detail">{{ e.detail }}</span>
          </li>
          <li v-if="!visible.length" class="empty">nothing recorded yet</li>
        </ol>
      </div>
    </div>

    <p class="say">{{ step.say }}</p>
  </div>
</template>

<style scoped>
.rp {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}
.panes {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) 6.4rem minmax(0, 1fr);
  gap: 0.5rem;
}
.pane {
  border: 1px solid var(--temporal-border-hi);
  border-radius: 0.6rem;
  background: var(--temporal-bg-elev);
  padding: 0.6rem 0.8rem;
  min-height: 0;
  transition: border-color 200ms, opacity 200ms;
}
/* The Worker box is a process: blue while it is up, red once it crashes. */
.worker-box {
  --worker: #8fb6ff;
  border: 2px solid var(--worker);
  background: rgba(143, 182, 255, 0.07);
  transition: border-color 250ms, background 250ms;
}
.worker-box.crashed {
  --worker: #ff6b8a;
  background: rgba(255, 107, 138, 0.1);
}
.worker-box.crashed .code { opacity: 0.45; }
.worker-head {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  margin-bottom: 0.4rem;
}
.worker-name {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--worker);
}
.worker-state {
  font-size: 0.62rem;
  font-weight: 500;
  color: var(--worker);
}
.pane-title.sub {
  display: block;
  margin-bottom: 0.15rem;
}
.pane-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 0.5rem;
  margin-bottom: 0.45rem;
}
.pane-title {
  font-size: 0.6rem;
  font-weight: 500;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--temporal-text-muted);
}
.pane-title.h { color: var(--temporal-green); }
.persisted { font-size: 0.58rem; color: var(--temporal-text-muted); }

/* The theme sizes every pre and code for code slides; this panel needs its own. */
.code,
.code code,
.ln {
  font-size: 0.64rem !important;
  line-height: 1.8 !important;
}
.code {
  margin: 0 !important;
  padding: 0 !important;
  background: none !important;
  font-family: "Noto Sans Mono", monospace;
  white-space: pre;
  overflow: hidden;
}
.ln {
  display: block;
  color: var(--temporal-text-muted);
  border-left: 2px solid transparent;
  padding-left: 0.4rem;
  transition: background 200ms, color 200ms;
}
.ln.cur {
  color: #fff;
  background: rgba(154, 128, 255, 0.18);
  border-left-color: var(--temporal-lavender);
}
.ln.cur.replay {
  background: rgba(89, 253, 160, 0.14);
  border-left-color: var(--temporal-green);
}

.act {
  margin-top: 0.6rem;
  padding-top: 0.5rem;
  border-top: 1px dashed var(--temporal-border-hi);
  transition: opacity 200ms;
}
.act.skipped { opacity: 0.55; }
.act.skipped .ln { text-decoration: line-through; text-decoration-color: rgba(89, 253, 160, 0.6); }
.act-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 0.25rem;
}
.act-tag { font-size: 0.6rem; font-weight: 500; }
.act-tag.skipped { color: var(--temporal-green); }
.cmd-lane {
  display: flex;
  align-items: center;
  justify-content: center;
}
.cmd {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 0.35rem;
  width: 100%;
  animation: cmd-in 350ms ease-out;
}
.cmd-box {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  padding: 0.35rem 0.45rem;
  border-radius: 0.4rem;
  border: 1px solid var(--temporal-lavender);
  background: rgba(154, 128, 255, 0.16);
}
.cmd.matched .cmd-box {
  border-color: var(--temporal-green);
  background: rgba(89, 253, 160, 0.12);
}
.cmd-label {
  font-size: 0.5rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--temporal-text-muted);
}
.cmd code {
  font-size: 0.56rem !important;
  padding: 0 !important;
  background: none !important;
  color: #fff;
  line-height: 1.3;
}
.cmd-arrow { width: 100%; height: 0.8rem; overflow: visible; }
.cmd-arrow line { stroke: var(--temporal-lavender); stroke-width: 2.5; }
.cmd-arrow path { fill: var(--temporal-lavender); }
.cmd-mode {
  font-size: 0.56rem;
  line-height: 1.3;
  text-align: center;
  color: var(--temporal-green);
}
@keyframes cmd-in {
  from { opacity: 0; transform: translateX(-8px); }
}

.events {
  list-style: none;
  margin: 0 !important;
  padding: 0 !important;
  display: flex;
  flex-direction: column;
  gap: 0.18rem;
}
.ev {
  margin: 0 !important;
  display: flex;
  align-items: baseline;
  gap: 0.45rem;
  font-family: "Noto Sans Mono", monospace;
  font-size: 0.64rem;
  line-height: 1.6;
  padding: 0.05rem 0.4rem;
  border-radius: 0.25rem;
  border-left: 2px solid var(--temporal-border-hi);
  background: rgba(255, 255, 255, 0.03);
  color: var(--temporal-text-muted);
  transition: background 200ms, color 200ms, border-color 200ms;
}
.ev .num { width: 1.1rem; opacity: 0.6; }
.ev .detail { margin-left: auto; opacity: 0.8; }
.ev.fresh {
  border-left-color: var(--temporal-lavender);
  background: rgba(154, 128, 255, 0.18);
  color: #fff;
}
.ev.matched {
  border-left-color: var(--temporal-green);
  background: rgba(89, 253, 160, 0.16);
  color: #fff;
}
.empty {
  font-size: 0.62rem;
  font-style: italic;
  color: var(--temporal-text-muted);
}

.say {
  margin: 0 !important;
  min-height: 2.7em;
  font-size: 0.85rem !important;
  line-height: 1.4;
  color: var(--temporal-text-strong);
}
</style>
