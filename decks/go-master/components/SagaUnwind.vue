<!--
  ABOUTME: The compensation stack and the Temporal UI event history for one
  ABOUTME: --fail-dispatch run, stepped by the slide's clicks beside the code.

  Pair it with the Workflow code block whose highlight ranges follow STEPS
  below, one range per step, starting with `all` on click 0:

    ```go {all|1-3|20|21|24|25|29|5-11|13-15|13-15|18}

  The stack is a picture of what is still owed, newest on top. The code does
  not remove anything: `append` put the newest undo at the end of the slice,
  the deferred loop walks it backwards, and each undo is drawn as done once it
  has run. The history rows are the Activity events
  Temporal records, in order; Workflow Task events are left out, as the UI's
  compact view does.
-->
<script setup lang="ts">
import { computed } from 'vue'
import { useSlideContext } from '@slidev/client'

type Row = { event: string; detail: string; tone: 'ok' | 'fail' | 'plain' }
type Step = { stack: string[]; done: string[]; running?: string; rows: Row[]; say: string }

const STARTED: Row = { event: 'WorkflowExecutionStarted', detail: 'OrderFulfillmentWorkflow', tone: 'plain' }
const RESERVE: Row = { event: 'ActivityTaskCompleted', detail: 'ReserveInventory', tone: 'ok' }
const CHARGE: Row = { event: 'ActivityTaskCompleted', detail: 'ChargePayment', tone: 'ok' }
const DISPATCH: Row = { event: 'ActivityTaskFailed', detail: 'DispatchCourier', tone: 'fail' }
const REFUND: Row = { event: 'ActivityTaskCompleted', detail: 'RefundPaymentIfCharged', tone: 'ok' }
const RELEASE: Row = { event: 'ActivityTaskCompleted', detail: 'ReleaseInventoryIfReserved', tone: 'ok' }
const FAILED: Row = { event: 'WorkflowExecutionFailed', detail: 'CourierUnavailable', tone: 'fail' }

const R = 'ReleaseInventoryIfReserved'
const F = 'RefundPaymentIfCharged'

const STEPS: Step[] = [
  { stack: [], done: [], rows: [], say: 'One order, started with <code>--fail-dispatch</code>.' },
  { stack: [], done: [], rows: [STARTED], say: 'The Workflow starts with an empty stack.' },
  { stack: [R], done: [], rows: [STARTED], say: '<code>append</code> the release <strong>before</strong> reserving. Nothing has run.' },
  { stack: [R], done: [], rows: [STARTED, RESERVE], say: 'Stock is held. Its undo is already on the stack.' },
  { stack: [R, F], done: [], rows: [STARTED, RESERVE], say: '<code>append</code> the refund before charging. It goes on top.' },
  { stack: [R, F], done: [], rows: [STARTED, RESERVE, CHARGE], say: 'Money has moved.' },
  { stack: [R, F], done: [], rows: [STARTED, RESERVE, CHARGE, DISPATCH], say: 'The courier fails. The Workflow returns the error, and the history shows the failure first.' },
  { stack: [R, F], done: [], running: F, rows: [STARTED, RESERVE, CHARGE, DISPATCH], say: 'The deferred function sees an <code>ActivityError</code> and starts the unwind from the <strong>top</strong> of the stack.' },
  { stack: [R, F], done: [F], rows: [STARTED, RESERVE, CHARGE, DISPATCH, REFUND], say: 'The newest undo runs first: the refund.' },
  { stack: [R, F], done: [F, R], rows: [STARTED, RESERVE, CHARGE, DISPATCH, REFUND, RELEASE], say: 'Then the release. The stack is empty.' },
  { stack: [R, F], done: [F, R], rows: [STARTED, RESERVE, CHARGE, DISPATCH, REFUND, RELEASE, FAILED], say: '<code>err</code> is left as it was: the Workflow ends Failed, with nothing left owed.' },
]

const { $clicks } = useSlideContext()
const step = computed(() => STEPS[Math.min(Math.max($clicks.value, 0), STEPS.length - 1)])
// Newest on top.
const stack = computed(() => [...step.value.stack].reverse())
const fresh = computed(() => {
  const prev = STEPS[Math.max($clicks.value - 1, 0)]
  return $clicks.value > 0 ? step.value.rows.length - prev.rows.length : 0
})
</script>

<template>
  <div class="unwind">
    <section class="panel">
      <h3>Compensation stack <span>compensations</span></h3>
      <div class="stack">
        <div v-if="!stack.length" class="empty">[ ] empty</div>
        <div
          v-for="(name, i) in stack"
          :key="name"
          class="frame"
          :class="{
            top: i === 0 && !step.done.includes(name),
            running: step.running === name,
            done: step.done.includes(name),
          }"
        >
          <span class="name">{{ name }}</span>
          <span class="tag">{{ step.done.includes(name) ? 'ran' : step.running === name ? 'next' : 'owed' }}</span>
        </div>
      </div>
    </section>

    <section class="panel history">
      <h3>Temporal UI <span>Event History</span></h3>
      <div v-if="!step.rows.length" class="empty">No events yet.</div>
      <ol>
        <li
          v-for="(row, i) in step.rows"
          :key="i"
          :class="[row.tone, { fresh: i >= step.rows.length - fresh }]"
        >
          <span class="event">{{ row.event }}</span>
          <span class="detail">{{ row.detail }}</span>
        </li>
      </ol>
    </section>

    <p class="say" v-html="step.say" />
  </div>
</template>

<style scoped>
.unwind { display: flex; flex-direction: column; gap: 0.55rem; height: 100%; }
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
  font-family: 'Noto Sans Mono', ui-monospace, monospace; color: var(--temporal-text);
}
.empty { font-size: 0.66rem; opacity: 0.6; font-style: italic; }

.stack { display: flex; flex-direction: column; gap: 0.3rem; min-height: 3.6rem; }
.frame {
  display: flex; justify-content: space-between; align-items: center; gap: 0.5rem;
  padding: 0.28rem 0.5rem; border-radius: 0.35rem;
  border: 1px solid var(--temporal-border); border-left: 3px solid #a78bfa;
  font-family: 'Noto Sans Mono', ui-monospace, monospace; font-size: 0.64rem;
  transition: opacity 200ms ease, border-color 200ms ease;
}
.frame .tag { font-family: Inter, sans-serif; font-size: 0.56rem; letter-spacing: 0.06em; text-transform: uppercase; opacity: 0.7; }
.frame.top { border-left-color: var(--temporal-green); }
.frame.running { box-shadow: 0 0 0 1.5px var(--temporal-green); }
.frame.done { opacity: 0.4; text-decoration: line-through; border-left-color: var(--temporal-border); }

.history { flex: 1 1 auto; min-height: 0; }
ol { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 0.18rem; }
li {
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
li.fresh { background: rgb(94 234 212 / 12%); }

.say { margin: 0; font-size: 0.72rem; line-height: 1.4; min-height: 2rem; }
.say :deep(code) { font-size: 0.66rem !important; }
.say :deep(strong) { color: var(--temporal-green); font-weight: 500; }
</style>
