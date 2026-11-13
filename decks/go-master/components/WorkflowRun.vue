<!--
  ABOUTME: One Activity call inside a Workflow run, a click per step: the Command, the Events,
  ABOUTME: the Tasks and the Worker's actions, each concept in its own colour, on three lanes.

  Lanes: the Worker, then the Temporal Service's Event History and Task Queue.
  Step N appears on click N; click 0 shows the empty lanes. Set `clicks: 9`.
  ActivityTaskStarted is drawn at step 7, beside ActivityTaskCompleted, because
  the Service writes it when the attempt finishes, not when the Task is taken.
-->
<script setup lang="ts">
import { computed } from 'vue'
import { useSlideContext } from '@slidev/client'

type Kind = 'worker' | 'command' | 'event' | 'task'
type Pill = { icon: string; text: string; mono?: boolean; dim?: boolean }
// `dx` nudges a step right of its slot when its pill is too wide for the gap.
type Step = { kind: Kind; lane: number; pills: Pill[]; say: string; dx?: number }

const LANES = [
  { icon: '⚙️', label: 'Worker' },
  { icon: '📜', label: 'Event History' },
  { icon: '📥', label: 'Task Queue' },
]
const LANE_Y = [62, 196, 300]
const COMMAND_Y = 129

const STEPS: Step[] = [
  { kind: 'worker', lane: 0, pills: [{ icon: '▶️', text: 'ExecuteActivity()', mono: true }], say: 'Your Workflow calls workflow.ExecuteActivity with ChargeCard. Nothing runs yet: the call is a request' },
  { kind: 'command', lane: -1, pills: [{ icon: '📨', text: 'ScheduleActivityTask', mono: true }], say: 'The call becomes a Command, sent when the Workflow Task ends' },
  { kind: 'event', lane: 1, pills: [{ icon: '📝', text: 'ActivityTaskScheduled', mono: true }], say: 'The Service writes it down' },
  { kind: 'task', lane: 2, pills: [{ icon: '🎫', text: 'Activity Task' }], say: 'An Activity Task waits on the queue: the Activity Type, ChargeCard, and an Activity ID' },
  { kind: 'worker', lane: 0, pills: [{ icon: '🙋', text: 'poll' }], say: 'A Worker polls and takes it' },
  { kind: 'worker', lane: 0, dx: 30, pills: [{ icon: '💳', text: 'ChargeCard runs', mono: true }], say: 'The Worker matches ChargeCard to a function it registered at startup, and runs it' },
  {
    kind: 'event',
    lane: 1,
    pills: [
      { icon: '📝', text: 'ActivityTaskStarted', mono: true, dim: true },
      { icon: '✅', text: 'ActivityTaskCompleted', mono: true },
    ],
    say: 'The result is recorded. Started is written now too',
  },
  { kind: 'task', lane: 2, pills: [{ icon: '🎫', text: 'Workflow Task' }], say: 'A new Workflow Task wakes the Workflow' },
  { kind: 'worker', lane: 0, pills: [{ icon: '↩️', text: 'Get() returns' }], say: 'A Worker takes the Workflow Task: the Future’s Get returns the result, and the next line runs' },
]

const LEGEND: { kind: Kind; label: string }[] = [
  { kind: 'worker', label: 'Worker action' },
  { kind: 'command', label: 'Command' },
  { kind: 'event', label: 'Event' },
  { kind: 'task', label: 'Task' },
]

const X0 = 190
const DX = 86
const PILL_H = 26

function width(p: Pill): number {
  return Math.round(p.text.length * (p.mono ? 7.1 : 6.9) + 40)
}
function centre(i: number): { x: number; y: number } {
  const s = STEPS[i]
  return { x: X0 + i * DX + (s.dx ?? 0), y: s.lane < 0 ? COMMAND_Y : LANE_Y[s.lane] }
}
function pillY(i: number, j: number): number {
  const n = STEPS[i].pills.length
  return centre(i).y + (j - (n - 1) / 2) * (PILL_H + 6)
}

const { $clicks } = useSlideContext()
const shown = computed(() => Math.min($clicks.value, STEPS.length))
const say = computed(() => (shown.value ? STEPS[shown.value - 1].say : 'One Activity call, from your code to the next line'))
</script>

<template>
  <figure class="run">
    <svg viewBox="0 0 980 350" role="img" aria-label="One Activity call inside a Workflow run, step by step">
      <defs>
        <marker id="wr-head" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto">
          <path d="M0,0 L10,5 L0,10 z" class="head" />
        </marker>
      </defs>

      <!-- The Temporal Service holds the history and the queue. -->
      <rect x="8" y="160" width="964" height="182" rx="12" class="service" />
      <text x="958" y="180" text-anchor="end" class="service-label">🏛️ Temporal Service</text>

      <g v-for="(lane, l) in LANES" :key="lane.label">
        <line x1="150" :y1="LANE_Y[l]" x2="964" :y2="LANE_Y[l]" class="lane" />
        <text x="22" :y="LANE_Y[l] + 6" class="lane-label">
          <tspan class="lane-icon">{{ lane.icon }}</tspan> {{ lane.label }}
        </text>
      </g>

      <!-- The path through the steps so far. -->
      <line
        v-for="i in Math.max(shown - 1, 0)"
        :key="`seg-${i}`"
        :x1="centre(i - 1).x"
        :y1="centre(i - 1).y"
        :x2="centre(i).x"
        :y2="centre(i).y"
        class="path"
        marker-end="url(#wr-head)"
      />

      <g v-for="(step, i) in STEPS" :key="i" class="step" :class="[step.kind, { on: i < shown }]">
        <g v-for="(pill, j) in step.pills" :key="j" :class="{ dim: pill.dim }">
          <rect
            :x="centre(i).x - width(pill) / 2"
            :y="pillY(i, j) - PILL_H / 2"
            :width="width(pill)"
            :height="PILL_H"
            rx="13"
            class="pill"
          />
          <text :x="centre(i).x" :y="pillY(i, j) + 4.5" text-anchor="middle" class="pill-text" :class="{ mono: pill.mono }">
            {{ pill.icon }} {{ pill.text }}
          </text>
        </g>
        <circle :cx="centre(i).x" :cy="pillY(i, 0) - PILL_H / 2 - 9" r="8" class="badge" />
        <text :x="centre(i).x" :y="pillY(i, 0) - PILL_H / 2 - 5.5" text-anchor="middle" class="badge-text">{{ i + 1 }}</text>
      </g>
    </svg>

    <figcaption>
      <span class="say">{{ say }}</span>
      <span class="legend">
        <span v-for="item in LEGEND" :key="item.kind" class="key" :class="item.kind"><i />{{ item.label }}</span>
      </span>
    </figcaption>
  </figure>
</template>

<style scoped>
.run { margin: 0.2rem 0 0; display: flex; flex-direction: column; gap: 0.5rem; }
svg { width: 100%; height: auto; overflow: visible; }

.worker { --c: var(--temporal-cyan); }
.command { --c: var(--temporal-lavender); }
.event { --c: var(--temporal-green); }
.task { --c: var(--temporal-yellow); }

.service { fill: rgba(195, 160, 255, 0.05); stroke: var(--temporal-border-hi); stroke-dasharray: 4 4; }
.service-label { font-size: 11px; fill: var(--temporal-text-muted); letter-spacing: 0.06em; text-transform: uppercase; }
.lane { stroke: var(--temporal-border); stroke-width: 1.5; }
.lane-label { font-size: 13px; font-weight: 600; fill: var(--temporal-text-strong); }
.lane-icon { font-size: 15px; }

.path { stroke: var(--temporal-text-muted); stroke-width: 1.5; stroke-dasharray: 4 4; }
.head { fill: var(--temporal-text-muted); }

.step { opacity: 0; transition: opacity 0.35s ease; }
.step.on { opacity: 1; }
.pill { fill: color-mix(in srgb, var(--c) 16%, var(--temporal-bg-elev)); stroke: var(--c); stroke-width: 1.5; }
.pill-text { font-size: 12px; fill: var(--temporal-text-strong); }
.pill-text.mono { font-family: var(--slidev-code-font-family, ui-monospace, monospace); font-size: 11.5px; }
.dim { opacity: 0.55; }
.badge { fill: var(--c); }
.badge-text { font-size: 10.5px; font-weight: 700; fill: var(--temporal-bg); }

figcaption { display: flex; justify-content: space-between; align-items: center; gap: 1rem; }
.say { font-size: 1.05rem; color: var(--temporal-text-strong); min-height: 1.5em; }
.legend { display: flex; gap: 0.9rem; flex-shrink: 0; }
.key { display: inline-flex; align-items: center; gap: 0.35rem; font-size: 0.72rem; color: var(--temporal-text-muted); }
.key i { width: 0.65rem; height: 0.65rem; border-radius: 999px; background: var(--c); }
</style>
