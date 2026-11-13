<!--
  ABOUTME: A click-through sequence diagram of one timeout or retry scenario,
  ABOUTME: with the Temporal Event History it produces filling in beside it.

  Usage (clicks must equal the scenario's `say.length - 1`):

    ---
    layout: default
    clicks: 8
    ---

    # Start-To-Close Timeout

    <TimeoutSequence scenario="start-to-close" />

  Scenarios live in ./timeout-scenarios.ts. Rows are laid out up front and
  revealed by `visibility`, so the diagram never reflows between clicks: an
  arrow appears where it will stay. Timeout windows are bars in the two gutter
  columns, spanning the rows they cover; they pulse while open and take their
  outcome colour when they close.
-->
<script setup lang="ts">
import { computed } from 'vue'
import { useSlideContext } from '@slidev/client'
import { SCENARIOS, type Row, type Window } from './timeout-scenarios'

const props = defineProps<{ scenario: string }>()
const s = computed(() => {
  const found = SCENARIOS[props.scenario]
  if (!found) throw new Error(`Unknown TimeoutSequence scenario: ${props.scenario}`)
  return found
})

const { $clicks } = useSlideContext()
const click = computed(() => Math.min(Math.max($clicks.value, 0), s.value.say.length - 1))

// Lane centres, as a percentage of the lane area.
const CENTRE = [16.67, 50, 83.33]

function arrow(row: Extract<Row, { kind: 'msg' }>) {
  const a = CENTRE[row.from]
  const b = CENTRE[row.to]
  return { left: `${Math.min(a, b)}%`, width: `${Math.abs(b - a)}%`, rtl: b < a }
}

function windowState(w: Window) {
  if (click.value < w.at) return 'hidden'
  if (w.close && click.value >= w.close.at) return `closed ${w.close.tone}`
  return 'open'
}

// An open window grows with the diagram: it ends at the last row revealed so
// far, not at the row where it will eventually close.
const lastShown = computed(() => {
  let last = -1
  s.value.rows.forEach((row, i) => {
    if (row.at <= click.value) last = i
  })
  return last
})
function windowRows(w: Window) {
  const end = Math.max(w.from, Math.min(w.to, lastShown.value))
  return `${w.from + 2} / ${end + 3}`
}

const events = computed(() => s.value.events.filter((e) => e.at <= click.value))
const isNew = (at: number) => at === click.value && click.value > 0
</script>

<template>
  <div class="ts">
    <div class="diagram">
      <code v-if="s.code" class="opts">{{ s.code }}</code>
      <div class="grid" :style="{ gridTemplateRows: `auto repeat(${s.rows.length}, 1fr)` }">
        <!-- header: lane titles above the lifelines -->
        <div class="time head">t</div>
        <div class="gutter head" />
        <div class="gutter head" />
        <div class="lanes head">
          <span v-for="(lane, i) in s.lanes" :key="i" class="lane-title" :style="{ left: `${CENTRE[i]}%` }">{{ lane }}</span>
        </div>

        <template v-for="(row, r) in s.rows" :key="r">
          <div class="time" :class="{ shown: row.at <= click }" :style="{ gridRow: r + 2 }">{{ row.t }}</div>
          <div class="lanes" :style="{ gridRow: r + 2 }">
            <span v-for="i in 3" :key="i" class="lifeline" :style="{ left: `${CENTRE[i - 1]}%` }" />
            <div
              v-if="row.kind === 'msg'"
              class="msg"
              :class="[row.tone ?? 'plain', { shown: row.at <= click, fresh: isNew(row.at), rtl: arrow(row).rtl }]"
              :style="{ left: arrow(row).left, width: arrow(row).width }"
            >
              <span class="label">{{ row.label }}</span>
              <span v-if="row.sub" class="sub">{{ row.sub }}</span>
            </div>
            <div
              v-else-if="row.kind === 'note'"
              class="note"
              :class="[row.tone ?? 'plain', { shown: row.at <= click, fresh: isNew(row.at) }]"
              :style="{ left: `${CENTRE[row.lane]}%` }"
            >{{ row.text }}</div>
            <div v-else class="wait" :class="{ shown: row.at <= click, fresh: isNew(row.at) }">{{ row.text }}</div>
          </div>
        </template>

        <div
          v-for="(w, i) in s.windows"
          :key="`w${i}`"
          class="window"
          :class="windowState(w)"
          :style="{ gridRow: windowRows(w), gridColumn: w.col + 2 }"
        >
          <span class="w-label">{{ w.label }}</span>
          <span v-if="w.close && click >= w.close.at && w.close.text" class="w-out">{{ w.close.text }}</span>
        </div>
      </div>
      <p class="say" v-html="s.say[click]" />
    </div>

    <aside class="side">
    <section class="history">
      <h3>Event History</h3>
      <p class="what">What Temporal records, oldest first</p>
      <ol>
        <li v-for="(e, i) in events" :key="i" :class="[e.tone ?? 'plain', { ghost: e.ghost, fresh: isNew(e.at) }]">
          <span class="ev">{{ e.name }}</span>
          <span v-if="e.detail" class="detail">{{ e.detail }}</span>
        </li>
        <li v-if="!events.length" class="empty">No events yet</li>
      </ol>
    </section>
    </aside>
  </div>
</template>

<style scoped>
.ts {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 15.5rem;
  gap: 1rem;
  /* A fixed height so the rows share it (1fr each) instead of growing the
     slide past its footer; the caption stays in view below them. */
  height: 23.5rem;
  min-height: 0;
  font-size: 0.72rem;
}
.diagram {
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.opts {
  align-self: flex-start;
  margin-bottom: 0.45rem;
  font-size: 0.68rem !important;
}
.grid {
  flex: 1 1 auto;
  min-height: 0;
  display: grid;
  grid-template-columns: 3.4rem 5.2rem 5.2rem minmax(0, 1fr);
  column-gap: 0.3rem;
}
.head {
  padding-bottom: 0.35rem;
  border-bottom: 1px solid var(--temporal-border);
  color: var(--temporal-text-muted);
}
.time {
  grid-column: 1;
  align-self: center;
  font-variant-numeric: tabular-nums;
  color: var(--temporal-text-muted);
  visibility: hidden;
}
.time.head {
  visibility: visible;
}
.time.shown {
  visibility: visible;
}
.lanes {
  grid-column: 4;
  position: relative;
  min-height: 1.2rem;
}
.lanes.head {
  height: 1.3rem;
}
.lane-title {
  position: absolute;
  transform: translateX(-50%);
  white-space: nowrap;
  font-weight: 500;
  color: var(--temporal-text-strong);
  letter-spacing: 0.02em;
}
.lifeline {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 0;
  border-left: 1px dashed rgba(255, 255, 255, 0.14);
}

/* A message: a line between two lifelines with a head at the receiving end. */
.msg {
  position: absolute;
  top: 50%;
  height: 0;
  border-top: 2px solid var(--temporal-lavender);
  visibility: hidden;
}
.msg::after {
  content: "";
  position: absolute;
  right: -1px;
  top: -6px;
  border-style: solid;
  border-width: 5px 0 5px 8px;
  border-color: transparent transparent transparent var(--temporal-lavender);
}
.msg.rtl::after {
  right: auto;
  left: -1px;
  border-width: 5px 8px 5px 0;
  border-color: transparent var(--temporal-lavender) transparent transparent;
}
.msg .label {
  position: absolute;
  left: 50%;
  bottom: 3px;
  transform: translateX(-50%);
  white-space: nowrap;
  font-family: "Noto Sans Mono", ui-monospace, monospace;
  font-size: 0.66rem;
  color: var(--temporal-text-strong);
}
.msg .sub {
  position: absolute;
  left: 50%;
  top: 4px;
  transform: translateX(-50%);
  white-space: nowrap;
  font-size: 0.6rem;
  font-style: italic;
  color: var(--temporal-text-muted);
}
.msg.ok { border-top-color: var(--temporal-green); }
.msg.ok::after { border-left-color: var(--temporal-green); }
.msg.ok.rtl::after { border-left-color: transparent; border-right-color: var(--temporal-green); }
.msg.fail { border-top-color: #ff7a6b; border-top-style: dashed; }
.msg.fail::after { border-left-color: #ff7a6b; }
.msg.fail.rtl::after { border-left-color: transparent; border-right-color: #ff7a6b; }

.note {
  position: absolute;
  top: 50%;
  transform: translate(-50%, -50%);
  white-space: nowrap;
  padding: 0.12rem 0.5rem;
  border-radius: 0.3rem;
  border: 1px solid var(--temporal-border-hi);
  background: var(--temporal-bg-elev);
  color: var(--temporal-text-strong);
  visibility: hidden;
}
.note.fail { border-color: #ff7a6b; color: #ffb3aa; }
.note.ok { border-color: var(--temporal-green); }
.wait {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-style: italic;
  color: var(--temporal-text-muted);
  visibility: hidden;
}
.shown { visibility: visible; }
.fresh { animation: pop 380ms ease-out; }
@keyframes pop {
  from { opacity: 0; transform: translateY(4px); }
}
.note.fresh { animation: pop-note 380ms ease-out; }
@keyframes pop-note {
  from { opacity: 0; transform: translate(-50%, -30%); }
}

/* Timeout windows: a bar in a gutter column spanning the rows it covers. */
.window {
  position: relative;
  margin: 0.25rem 0;
  border-radius: 0.35rem;
  border: 1px solid var(--temporal-lavender);
  background: rgba(154, 128, 255, 0.12);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 0.25rem 0.3rem;
  line-height: 1.15;
}
.window.hidden { visibility: hidden; }
.window.open { animation: breathe 1.6s ease-in-out infinite; }
@keyframes breathe {
  50% { background: rgba(154, 128, 255, 0.22); }
}
.window.closed.fail { border-color: #ff7a6b; background: rgba(255, 122, 107, 0.14); }
.window.closed.ok { border-color: var(--temporal-green); background: rgba(89, 253, 160, 0.1); }
.window.closed.plain { border-style: dashed; opacity: 0.6; }
.w-label { font-weight: 500; color: var(--temporal-text-strong); font-size: 0.64rem; }
.w-out { font-size: 0.62rem; color: var(--temporal-text); }

.say {
  margin: 0.6rem 0 0 !important;
  min-height: 2.6em;
  font-size: 0.9rem !important;
  color: var(--temporal-text-strong);
}

.side {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  min-height: 0;
}
/* Event History, as the Temporal UI lists it: newest at the bottom. */
.history {
  flex: 1 1 auto;
  border: 1px solid var(--temporal-border);
  border-radius: 0.45rem;
  background: var(--temporal-bg-elev);
  padding: 0.6rem 0.7rem;
  min-height: 0;
  overflow: hidden;
}
.history .what {
  margin: 0 0 0.5rem !important;
  font-size: 0.6rem !important;
  color: var(--temporal-text-muted);
}
.history h3 {
  margin: 0 0 0.15rem !important;
  font-size: 0.66rem !important;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--temporal-green);
}
.history ol {
  list-style: none;
  margin: 0 !important;
  padding: 0 !important;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}
.history li {
  padding: 0.25rem 0.4rem 0.25rem 0.5rem !important;
  border-left: 3px solid var(--temporal-border-hi);
  border-radius: 0.2rem;
  background: rgba(255, 255, 255, 0.03);
  display: flex;
  flex-direction: column;
  font-size: 0.66rem !important;
  margin: 0 !important;
}
.history li::before { display: none; }
.history li.ok { border-left-color: var(--temporal-green); }
.history li.fail { border-left-color: #ff7a6b; }
.history li.fresh { background: rgba(154, 128, 255, 0.16); animation: pop 380ms ease-out; }
.history li.ghost {
  border-left-style: dashed;
  background: none;
  font-style: italic;
  color: var(--temporal-text-muted);
}
.history li.empty { border: none; background: none; color: var(--temporal-text-muted); }
.ev { font-family: "Noto Sans Mono", ui-monospace, monospace; color: var(--temporal-text-strong); }
.ghost .ev { font-family: inherit; color: var(--temporal-text-muted); }
.detail { font-family: "Noto Sans Mono", ui-monospace, monospace; color: var(--temporal-text-muted); font-size: 0.6rem; }
</style>
