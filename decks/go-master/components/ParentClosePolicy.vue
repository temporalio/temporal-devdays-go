<!--
  ABOUTME: One Parent Close Policy as a two-lane timeline: the parent starts a child,
  ABOUTME: the parent closes, and the child's lane shows what that policy does to it.

    <ParentClosePolicy policy="TERMINATE" />
    <ParentClosePolicy policy="ABANDON" />
    <ParentClosePolicy policy="REQUEST_CANCEL" />
-->
<script setup lang="ts">
import { computed } from 'vue'

type Policy = 'TERMINATE' | 'ABANDON' | 'REQUEST_CANCEL'
const props = defineProps<{ policy: Policy }>()

const COPY: Record<Policy, { tag?: string; caption: string }> = {
  TERMINATE: { tag: 'default', caption: 'The child stops the moment the parent closes. No cleanup runs.' },
  ABANDON: { caption: 'The child keeps running on its own, to completion.' },
  REQUEST_CANCEL: { caption: 'The child is asked to cancel, runs its cleanup, then ends.' },
}
const copy = computed(() => COPY[props.policy])

// Shared geometry: the parent closes at CLOSE on both lanes.
const P = 38
const C = 98
const START = 64
const SPAWN = 104
const CLOSE = 178
</script>

<template>
  <figure class="pcp" :class="policy.toLowerCase()">
    <figcaption>
      <code>{{ policy }}</code>
      <span v-if="copy.tag" class="tag">{{ copy.tag }}</span>
    </figcaption>

    <svg viewBox="0 0 320 136" role="img" :aria-label="`${policy}: ${copy.caption}`">
      <defs>
        <marker :id="`pcp-head-${policy}`" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
          <path d="M0,0 L10,5 L0,10 z" fill="context-stroke" />
        </marker>
      </defs>

      <text x="6" :y="P + 4" class="lane">Parent</text>
      <text x="6" :y="C + 4" class="lane">Child</text>

      <!-- Parent: runs, then closes. -->
      <line :x1="START" :y1="P" :x2="CLOSE" :y2="P" class="parent" />
      <circle :cx="START" :cy="P" r="4" class="parent-dot" />
      <rect :x="CLOSE - 4" :y="P - 9" width="8" height="18" rx="1.5" class="close" />
      <text :x="CLOSE" :y="P - 15" class="note" text-anchor="middle">parent closes</text>

      <!-- The parent starts the child. -->
      <line :x1="SPAWN" :y1="P + 4" :x2="SPAWN" :y2="C - 8" class="spawn" :marker-end="`url(#pcp-head-${policy})`" />
      <circle :cx="SPAWN" :cy="C" r="4" class="child-dot" />

      <template v-if="policy === 'TERMINATE'">
        <line :x1="SPAWN" :y1="C" :x2="CLOSE" :y2="C" class="child" />
        <line :x1="CLOSE" :y1="P + 9" :x2="CLOSE" :y2="C - 12" class="kill" />
        <g class="x" :transform="`translate(${CLOSE} ${C})`">
          <line x1="-7" y1="-7" x2="7" y2="7" />
          <line x1="-7" y1="7" x2="7" y2="-7" />
        </g>
        <text :x="CLOSE" :y="C + 24" class="note bad" text-anchor="middle">terminated</text>
      </template>

      <template v-else-if="policy === 'ABANDON'">
        <line :x1="SPAWN" :y1="C" :x2="308" :y2="C" class="child" :marker-end="`url(#pcp-head-${policy})`" />
        <text x="250" :y="C + 24" class="note good" text-anchor="middle">keeps running</text>
      </template>

      <template v-else>
        <line :x1="SPAWN" :y1="C" :x2="CLOSE" :y2="C" class="child" />
        <line :x1="CLOSE" :y1="P + 9" :x2="CLOSE" :y2="C - 8" class="cancel" :marker-end="`url(#pcp-head-${policy})`" />
        <text :x="CLOSE + 6" :y="(P + C) / 2 + 4" class="note warn">cancel request</text>
        <line :x1="CLOSE" :y1="C" :x2="248" :y2="C" class="cleanup" />
        <circle cx="248" :cy="C" r="5" class="end" />
        <text :x="213" :y="C + 24" class="note warn" text-anchor="middle">cleanup, then cancelled</text>
      </template>
    </svg>

    <p>{{ copy.caption }}</p>
  </figure>
</template>

<style scoped>
.pcp {
  margin: 0; padding: 0.7rem 0.8rem 0.8rem; border-radius: 0.6rem;
  border: 1px solid var(--temporal-border); background: var(--temporal-bg-elev);
  display: flex; flex-direction: column; gap: 0.4rem; min-width: 0;
}
figcaption { display: flex; align-items: center; gap: 0.5rem; }
figcaption code { font-size: 0.95rem !important; color: var(--temporal-text-strong); }
.tag {
  font-size: 0.6rem; text-transform: uppercase; letter-spacing: 0.08em; padding: 0.05rem 0.45rem;
  border-radius: 999px; border: 1px solid var(--temporal-border-hi); color: var(--temporal-text-muted);
}
svg { width: 100%; height: auto; overflow: visible; }
p { margin: 0; font-size: 0.78rem; line-height: 1.4; color: var(--temporal-text); }

.lane { font-size: 11px; fill: var(--temporal-text-muted); font-weight: 600; }
.note { font-size: 10.5px; fill: var(--temporal-text-muted); }
.note.bad { fill: #f9a8d4; }
.note.good { fill: var(--temporal-green); }
.note.warn { fill: var(--temporal-yellow); }

line { stroke-width: 3; stroke-linecap: round; }
.parent { stroke: var(--temporal-lavender); }
.parent-dot { fill: var(--temporal-lavender); }
.close { fill: var(--temporal-text-strong); }
.spawn { stroke: var(--temporal-lavender); stroke-width: 1.5; stroke-dasharray: 3 3; }
.child { stroke: var(--temporal-cyan); }
.child-dot { fill: var(--temporal-cyan); }
.kill { stroke: #f472b6; stroke-width: 1.5; stroke-dasharray: 3 3; }
.x line { stroke: #f472b6; stroke-width: 3; }
.cancel { stroke: var(--temporal-yellow); stroke-width: 1.5; stroke-dasharray: 3 3; }
.cleanup { stroke: var(--temporal-yellow); }
.end { fill: none; stroke: var(--temporal-yellow); stroke-width: 2.5; }
</style>
