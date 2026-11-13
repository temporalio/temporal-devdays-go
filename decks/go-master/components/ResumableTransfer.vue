<!--
  ABOUTME: An on-call console for one bank transfer to a mistyped account: the
  ABOUTME: presenter clicks the engineer's actions and watches status and Event History.

    <ResumableTransfer />

  Slide clicks step through the three engineer actions (the slide needs
  `clicks: 3`), so arrow keys and a clicker work; each button jumps to its
  action. The action just performed is outlined, so the button always matches
  the terminal and the Events beside it; before the first click, a dashed
  outline points at action 1. Each step shows the CLI command an
  engineer would run, its output, and the Events it adds. Workflow Task events are left out, as the UI's compact
  view does. Everything here matches snippets/resumable/resumable.go and
  what snippets/resumable/resumable_test.go checks: one attempt on a
  non-retryable failure, a Query that adds nothing to history,
  and one Workflow run from start to finish.
-->
<script setup lang="ts">
import { computed } from 'vue'
import { useNav, useSlideContext } from '@slidev/client'

type Row = { event: string; detail: string; tone: 'ok' | 'fail' | 'plain' | 'signal' }
type Stage = 'idle' | 'parked' | 'done'
type Step = {
  stage: Stage
  status: string
  account: string
  rows: Row[]
  fresh: number
  term?: { cmd: string; out: string }
  say: string
}

const ID = 'transfer-4471'
const STARTED: Row[] = [
  { event: 'WorkflowExecutionStarted', detail: 'TransferWorkflow', tone: 'plain' },
  { event: 'ActivityTaskScheduled', detail: 'ExecuteTransfer → 99999999', tone: 'plain' },
  { event: 'ActivityTaskStarted', detail: 'attempt 1', tone: 'plain' },
  { event: 'ActivityTaskFailed', detail: 'AccountNotFoundError · non-retryable', tone: 'fail' },
]
const GOOD_FIX: Row[] = [
  { event: 'WorkflowExecutionSignaled', detail: 'retryWithCorrection "12345678"', tone: 'signal' },
  { event: 'ActivityTaskScheduled', detail: 'ExecuteTransfer → 12345678', tone: 'plain' },
  { event: 'ActivityTaskStarted', detail: 'attempt 1', tone: 'plain' },
  { event: 'ActivityTaskCompleted', detail: 'ExecuteTransfer', tone: 'ok' },
  { event: 'WorkflowExecutionCompleted', detail: '"Transfer of 500 to 12345678 completed"', tone: 'ok' },
]
const signal = (input: string) =>
  `temporal workflow signal --workflow-id ${ID} --name retryWithCorrection --input '"${input}"'`

// One entry per click: click 0 is the incident, then one engineer action per click.
const STEPS: Step[] = [
  {
    stage: 'idle', status: '?', account: '99999999', rows: [], fresh: 0,
    say: 'A customer sends <strong>$500</strong> to account <code>99999999</code>. They mistyped it. You are on call.',
  },
  {
    stage: 'parked', status: '?', account: '99999999', rows: STARTED, fresh: STARTED.length,
    term: { cmd: `temporal workflow start --type TransferWorkflow --workflow-id ${ID} ...`, out: `Running execution: WorkflowId ${ID}` },
    say: 'One attempt, not three: an unknown account is <strong>non-retryable</strong>. The Workflow did not fail. It is still <strong>Running</strong>, parked. Your alert fires.',
  },
  {
    stage: 'parked', status: 'AWAITING_CORRECTION', account: '99999999', rows: STARTED, fresh: 0,
    term: { cmd: `temporal workflow query --workflow-id ${ID} --name getStatus`, out: 'QueryResult  "AWAITING_CORRECTION"' },
    say: 'It tells you exactly where it is stuck. Look at the history: <strong>no new Events</strong>. A Query is a read.',
  },
  {
    stage: 'done', status: 'COMPLETED', account: '12345678', rows: [...STARTED, ...GOOD_FIX], fresh: GOOD_FIX.length,
    term: { cmd: signal('12345678'), out: 'Signal workflow succeeded' },
    say: 'The same Workflow wakes up and retries with your fix. <strong>One run, one history</strong>: the failure, your correction and the success, in order. The customer never re-entered anything.',
  },
]

// The buttons name the action each click performs, and jump to it when pressed.
const ACTIONS = ['Send the transfer', 'Query getStatus', 'Signal fix: "12345678"']

const { $clicks, $page } = useSlideContext()
const nav = useNav()
const at = computed(() => Math.min(Math.max($clicks.value, 0), STEPS.length - 1))
const step = computed(() => STEPS[at.value])
const go = (clicks: number) => nav.go($page.value, clicks)
const badge = computed(() => ({ idle: 'Not started', parked: 'Running', done: 'Completed' })[step.value.stage])
</script>

<template>
  <div class="rt">
    <div class="console">
      <section class="panel card">
        <h3>Transfer <span>{{ ID }}</span></h3>
        <div class="route">
          <code>10000001</code><span class="arrow">→</span><code :class="{ wrong: step.stage !== 'idle' && step.account === '99999999' }">{{ step.account }}</code>
          <span class="amount">$500</span>
        </div>
        <div class="states">
          <span class="pill" :class="step.stage">{{ badge }}</span>
          <span class="status">getStatus: <code>{{ step.status }}</code></span>
        </div>
      </section>

      <div class="buttons">
        <button
          v-for="(label, i) in ACTIONS"
          :key="label"
          :class="{ current: i === at - 1, next: at === 0 && i === 0, done: i < at - 1 }"
          @click.stop="go(i + 1)"
        >
          <span class="n">{{ i + 1 }}</span>{{ label }}
        </button>
        <button class="reset" @click.stop="go(0)">Reset</button>
      </div>

      <section class="panel term">
        <h3>Terminal</h3>
        <template v-if="step.term">
          <div class="cmd">$ {{ step.term.cmd }}</div>
          <div class="out">{{ step.term.out }}</div>
        </template>
        <div v-else class="empty">Pick an action.</div>
      </section>
    </div>

    <section class="panel history">
      <h3>Temporal UI <span>Event History</span></h3>
      <div v-if="!step.rows.length" class="empty">No events yet.</div>
      <ol>
        <li v-for="(row, i) in step.rows" :key="i" :class="[row.tone, { fresh: i >= step.rows.length - step.fresh }]">
          <span class="event">{{ row.event }}</span>
          <span class="detail">{{ row.detail }}</span>
        </li>
      </ol>
    </section>

    <p class="say" v-html="step.say" />
  </div>
</template>

<style scoped>
.rt {
  display: grid;
  grid-template-columns: 1fr 1.15fr;
  grid-template-rows: 1fr auto;
  gap: 0.6rem 0.9rem;
  height: 100%;
}
.console { display: flex; flex-direction: column; gap: 0.5rem; min-width: 0; }
.panel {
  border: 1px solid var(--temporal-border); border-radius: 0.5rem;
  background: var(--temporal-bg-elev); padding: 0.5rem 0.65rem;
}
h3 {
  margin: 0 0 0.4rem; font-size: 0.62rem; font-weight: 600; letter-spacing: 0.08em;
  text-transform: uppercase; color: var(--temporal-text);
}
h3 span {
  margin-left: 0.3rem; text-transform: none; letter-spacing: 0; font-weight: 400;
  font-family: 'Noto Sans Mono', ui-monospace, monospace;
}
.empty { font-size: 0.66rem; opacity: 0.6; font-style: italic; }

.route { display: flex; align-items: center; gap: 0.45rem; font-size: 0.85rem; }
.route code { font-size: 0.8rem !important; }
.route code.wrong { color: #f9a8d4; text-decoration: wavy underline #f472b6; }
.arrow { opacity: 0.6; }
.amount { margin-left: auto; font-weight: 600; color: var(--temporal-green); }
.states { display: flex; align-items: center; gap: 0.6rem; margin-top: 0.45rem; font-size: 0.66rem; }
.status code { font-size: 0.64rem !important; }
.pill {
  padding: 0.1rem 0.5rem; border-radius: 999px; font-size: 0.58rem; font-weight: 600;
  letter-spacing: 0.05em; text-transform: uppercase; border: 1px solid var(--temporal-border);
}
.pill.parked { color: #fcd34d; border-color: #fcd34d; }
.pill.done { color: var(--temporal-green); border-color: var(--temporal-green); }

.buttons { display: grid; grid-template-columns: 1fr 1fr; gap: 0.35rem; }
button {
  padding: 0.35rem 0.5rem; border-radius: 0.4rem; font-size: 0.66rem; font-weight: 500;
  color: var(--temporal-text-strong); background: rgb(154 128 255 / 18%);
  border: 1px solid var(--temporal-lavender); cursor: pointer; text-align: left;
  transition: background 150ms ease, opacity 150ms ease;
}
button:hover { background: rgb(154 128 255 / 35%); }
button .n { display: inline-block; width: 1.1em; opacity: 0.6; font-variant-numeric: tabular-nums; }
button.current { box-shadow: 0 0 0 2px var(--temporal-green); }
button.next { outline: 2px dashed var(--temporal-green); outline-offset: 1px; }
button.done { opacity: 0.4; }
button.reset { text-align: center; background: none; border-color: var(--temporal-border); font-size: 0.6rem; }

.term { background: #0b1020; flex: 1 1 auto; min-height: 4.2rem; }
.cmd, .out {
  font-family: 'Noto Sans Mono', ui-monospace, monospace; font-size: 0.56rem; line-height: 1.45;
  white-space: pre-wrap; word-break: break-all;
}
.cmd { color: #59fda0; }
.out { opacity: 0.85; margin-top: 0.2rem; }

.history { min-height: 0; overflow: hidden; }
ol { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 0.18rem; }
li {
  /* The theme spaces list items for prose; these are table rows. */
  margin: 0 !important;
  display: flex; gap: 0.4rem; align-items: baseline; white-space: nowrap; overflow: hidden;
  padding: 0.16rem 0.45rem; border-radius: 0.3rem;
  border-left: 3px solid var(--temporal-border); background: rgb(255 255 255 / 3%);
  transition: background 200ms ease;
}
li .event { font-size: 0.56rem; font-weight: 600; flex: 0 0 auto; }
li .detail { font-family: 'Noto Sans Mono', ui-monospace, monospace; font-size: 0.52rem; opacity: 0.8; overflow: hidden; text-overflow: ellipsis; }
li.ok { border-left-color: var(--temporal-green); }
li.fail { border-left-color: #f472b6; }
li.fail .event { color: #f9a8d4; }
li.signal { border-left-color: #fcd34d; }
li.fresh { background: rgb(94 234 212 / 12%); }

.say { grid-column: span 2; margin: 0; font-size: 0.72rem; line-height: 1.4; min-height: 2rem; }
.say :deep(code) { font-size: 0.66rem !important; }
.say :deep(strong) { color: var(--temporal-green); font-weight: 500; }
</style>
