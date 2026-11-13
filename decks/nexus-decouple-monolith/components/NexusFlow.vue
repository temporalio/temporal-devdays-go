<!--
  One payment (TXN-A from the lab's starter), followed click by click from the
  Payments Workflow, over the Nexus Endpoint, through the Compliance Namespace and
  back. Driven by the slide's $clicks: step 0 is the Workflow start, step 10 is the
  PaymentResult. The packet animates along the edge between consecutive steps and
  snaps when you jump.
-->
<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'

const props = defineProps<{ step: number }>()

type Pt = [number, number]

const W = 240
const H = 50

const nodes = {
  // title: the call as it appears in the lab code. sdk: the Go SDK call that makes it a
  // Workflow, Activity or Nexus piece. file: where it lives.
  start: { x: 160, y: 76, title: 'PaymentProcessingWorkflow()', sdk: 'w.RegisterWorkflow', file: 'payments/workflows.go' },
  validate: { x: 160, y: 148, title: 'ValidatePayment()', sdk: 'workflow.ExecuteActivity', file: 'payments/activities.go' },
  nexusOp: { x: 160, y: 220, title: 'ExecuteOperation()', sdk: 'workflow.NewNexusClient', file: 'payments/workflows.go' },
  execute: { x: 160, y: 292, title: 'ExecutePayment()', sdk: 'workflow.ExecuteActivity', file: 'payments/activities.go' },
  handler: { x: 780, y: 96, title: 'CheckComplianceOperation', sdk: 'temporalnexus.NewWorkflowRunOperation', file: 'compliance/handler.go' },
  cwf: { x: 780, y: 180, title: 'ComplianceWorkflow()', sdk: 'w.RegisterWorkflow', file: 'compliance/workflows.go' },
  activity: { x: 780, y: 264, title: 'CheckCompliance()', sdk: 'workflow.ExecuteActivity', file: 'compliance/activities.go' },
} as const
type NodeKey = keyof typeof nodes

// The packet parks on the border where it entered a node, so it never covers the text.
const stops: Pt[] = [
  [160, 51], [160, 123], [160, 195], [370, 220], [660, 96], [780, 155],
  [780, 239], [780, 205], [255, 195], [160, 267], [40, 76],
]
// The route into each step, along the drawn edges.
const routes: Pt[][] = [
  [],
  [[160, 51], [160, 123]],
  [[160, 123], [160, 195]],
  [[160, 195], [280, 195], [280, 220], [370, 220]],
  [[370, 220], [605, 220], [605, 96], [660, 96]],
  [[660, 96], [660, 121], [780, 121], [780, 155]],
  [[780, 155], [780, 239]],
  [[780, 239], [780, 205]],
  [[780, 205], [660, 205], [660, 182], [255, 182], [255, 195]],
  [[255, 195], [280, 195], [280, 245], [160, 245], [160, 267]],
  [[160, 267], [40, 267], [40, 292], [25, 292], [25, 76], [40, 76]],
]

const edges = [
  { from: 1, kind: 'pay', d: 'M160 101 V123' },
  { from: 2, kind: 'pay', d: 'M160 173 V195' },
  { from: 3, kind: 'nexus', d: 'M280 220 H370' },
  { from: 4, kind: 'nexus', d: 'M570 220 H605 V96 H660' },
  { from: 5, kind: 'comp', d: 'M780 121 V155' },
  { from: 6, kind: 'comp', d: 'M780 205 V239' },
  { from: 8, kind: 'callback', d: 'M660 182 H255 V195' },
  { from: 9, kind: 'pay', d: 'M160 245 V267' },
  { from: 10, kind: 'result', d: 'M40 292 H25 V76 H40' },
]

// Which node is "current" at each step, for the highlight ring.
const current: (NodeKey | 'endpoint')[] = [
  'start', 'validate', 'nexusOp', 'endpoint', 'handler', 'cwf',
  'activity', 'cwf', 'nexusOp', 'execute', 'start',
]

const captions = [
  { where: 'Payments', title: 'A client starts the payment Workflow',
    data: `PaymentRequest{ TransactionID: "TXN-A", Amount: 250, Currency: "USD", US → US }` },
  { where: 'Payments', title: 'ValidatePayment runs on the Payments Worker',
    data: `→ true` },
  { where: 'Payments', title: 'The Workflow calls ExecuteOperation(CheckComplianceOperation) on its Nexus client',
    data: `ComplianceRequest{ TransactionID: "TXN-A", Amount: 250, US → US, Description }` },
  { where: 'Nexus', title: 'The Endpoint resolves the name, not the caller',
    data: `compliance-endpoint → compliance-namespace / compliance-risk` },
  { where: 'Compliance', title: 'The handler starts ComplianceWorkflow and returns an operation token',
    data: `StartWorkflowOptions{ ID: "compliance-TXN-A" } → Operation started (async)` },
  { where: 'Compliance', title: 'ComplianceWorkflow runs in the Compliance Namespace',
    data: `ComplianceRequest{ TransactionID: "TXN-A", ... }` },
  { where: 'Compliance', title: 'The CheckCompliance Activity scores the risk',
    data: `→ RiskLevel: "LOW"` },
  { where: 'Compliance', title: 'ComplianceWorkflow completes',
    data: `ComplianceResult{ Approved: true, RiskLevel: "LOW", Explanation: "Routine … No regulatory concerns." }` },
  { where: 'Nexus', title: 'A completion callback delivers the result to the caller',
    data: `NexusOperationCompleted → ExecuteOperation(...).Get() fills in the ComplianceResult` },
  { where: 'Payments', title: 'Approved, so ExecutePayment runs',
    data: `→ ConfirmationNumber` },
  { where: 'Payments', title: 'The payment Workflow completes',
    data: `PaymentResult{ Success: true, Status: "COMPLETED", RiskLevel: "LOW", ConfirmationNumber }` },
]

const step = computed(() => Math.max(0, Math.min(stops.length - 1, props.step ?? 0)))
const caption = computed(() => captions[step.value])

// What the packet is carrying on the hop INTO each step. Shown only while it moves;
// at rest the caption bar carries the detail and the chip would cover node text.
const carrying = [
  'PaymentRequest', 'PaymentRequest', 'valid: true', 'ComplianceRequest', 'ComplianceRequest',
  'ComplianceRequest', 'ComplianceRequest', 'RiskLevel: LOW', 'ComplianceResult',
  'PaymentRequest', 'PaymentResult',
]

// Packet animation. Forward one step: walk the route. Back one step: walk it reversed.
// Anything else (jumps, first render, export): snap.
const pos = ref<Pt>(stops[step.value])
const label = ref(carrying[step.value])
const trail = ref('')
let raf = 0

const chipW = computed(() => label.value.length * 6.6 + 16)
// Keep the label inside the canvas: flip it left of the dot near the right edge.
const chipX = computed(() => (pos.value[0] + 14 + chipW.value > 935 ? -14 - chipW.value : 14))

function walk(points: Pt[], ms: number) {
  cancelAnimationFrame(raf)
  trail.value = 'M' + points.map((p) => p.join(' ')).join(' L')
  const segs = points.slice(1).map((p, i) => Math.hypot(p[0] - points[i][0], p[1] - points[i][1]))
  const total = segs.reduce((a, b) => a + b, 0) || 1
  const t0 = performance.now()
  const tick = (now: number) => {
    const u = Math.min(1, (now - t0) / ms)
    const e = u < 0.5 ? 2 * u * u : 1 - (-2 * u + 2) ** 2 / 2
    let d = e * total
    let i = 0
    while (i < segs.length - 1 && d > segs[i]) d -= segs[i++]
    const f = segs[i] ? d / segs[i] : 1
    const [a, b] = [points[i], points[i + 1]]
    pos.value = [a[0] + (b[0] - a[0]) * f, a[1] + (b[1] - a[1]) * f]
    if (u < 1) raf = requestAnimationFrame(tick)
    else trail.value = ''
  }
  raf = requestAnimationFrame(tick)
}

watch(step, (next, prev) => {
  label.value = carrying[next === prev - 1 ? prev : next]
  if (next === prev + 1) walk(routes[next], 1300 + routes[next].length * 150)
  else if (next === prev - 1) walk([...routes[prev]].reverse(), 700)
  else {
    cancelAnimationFrame(raf)
    trail.value = ''
    pos.value = stops[next]
  }
})
onBeforeUnmount(() => cancelAnimationFrame(raf))

const visited = (k: NodeKey | 'endpoint') => current.slice(0, step.value + 1).includes(k)

// The three things a Nexus request cannot be sent without. Each tag sits on the top
// edge of the box it applies to and lights up on the step where that piece is used.
// Kept short and left-aligned so it clears the packet's stops at the box centres.
const required = [
  { x: 36, y: 195, at: 2, text: 'service + endpoint' },
  { x: 366, y: 194, at: 3, text: '→ compliance-risk' },
  { x: 656, y: 66, at: 4, text: 'polls compliance-risk' },
].map((r) => ({ ...r, w: r.text.length * 5.3 + 26 }))
</script>

<template>
  <div class="flow">
    <svg viewBox="0 0 940 350" class="diagram" role="img"
      aria-label="A payment flowing from the Payments Workflow, over a Nexus Endpoint, to the Compliance Workflow and back">
      <!-- Namespaces -->
      <rect x="10" y="8" width="300" height="334" rx="14" class="ns ns-pay" />
      <text x="26" y="26" class="team pay">PAYMENTS TEAM</text>
      <text x="26" y="40" class="ns-label">payments-namespace · queue: <tspan class="pay">payments-processing</tspan></text>
      <rect x="630" y="8" width="300" height="334" rx="14" class="ns ns-comp" />
      <text x="646" y="26" class="team comp">COMPLIANCE TEAM</text>
      <text x="646" y="40" class="ns-label">compliance-namespace · queue: <tspan class="comp">compliance-risk</tspan></text>

      <!-- Edges: dim until the packet has crossed them -->
      <path v-for="(e, i) in edges" :key="i" :d="e.d"
        :class="['edge', `edge-${e.kind}`, { lit: step >= e.from }]" />
      <text x="458" y="175" class="edge-label" :class="{ lit: step >= 8 }">completion callback</text>
      <text x="595" y="150" class="edge-label" :class="{ lit: step >= 4 }"
        transform="rotate(-90 595 150)" text-anchor="middle">start request</text>

      <!-- Endpoint -->
      <g :class="['node', 'endpoint', { visited: visited('endpoint'), current: current[step] === 'endpoint' }]">
        <rect x="370" y="194" width="200" height="52" rx="10" />
        <text x="470" y="216" class="n-sub" text-anchor="middle">Nexus Endpoint</text>
        <text x="470" y="234" class="n-title" text-anchor="middle">compliance-endpoint</text>
      </g>

      <!-- Workflow and Activity nodes -->
      <g v-for="(n, k) in nodes" :key="k"
        :class="['node', k === 'handler' || k === 'cwf' || k === 'activity' ? 'comp' : 'pay',
          { visited: visited(k), current: current[step] === k }]">
        <rect :x="n.x - W / 2" :y="n.y - H / 2" :width="W" :height="H" rx="9" />
        <text :x="n.x" :y="n.y - 7" class="n-title" text-anchor="middle">{{ n.title }}</text>
        <text :x="n.x" :y="n.y + 7" class="n-sdk" text-anchor="middle">{{ n.sdk }}</text>
        <text :x="n.x" :y="n.y + 19" class="n-file" text-anchor="middle">{{ n.file }}</text>
      </g>

      <!-- What a Nexus request needs: a key tag on each box it applies to -->
      <g v-for="r in required" :key="r.text" :class="['req', { active: step === r.at }]"
        :transform="`translate(${r.x} ${r.y - 9})`">
        <rect :width="r.w" height="18" rx="9" class="req-pill" />
        <ph-key-fill x="6" y="3.5" width="11" height="11" />
        <text x="20" y="12.5" class="req-text">{{ r.text }}</text>
      </g>

      <!-- The payment -->
      <path v-if="trail" :d="trail" class="trail" />
      <g :transform="`translate(${pos[0]} ${pos[1]})`" class="packet">
        <circle r="10" class="halo" />
        <circle r="5" class="core" />
        <g :transform="`translate(${chipX} -24)`" :class="['chip', { moving: trail }]">
          <rect :width="chipW" height="18" rx="9" />
          <text :x="chipW / 2" y="12.5" text-anchor="middle">{{ label }}</text>
        </g>
      </g>
    </svg>

    <div class="caption">
      <span :class="['where', caption.where.toLowerCase()]">{{ caption.where }}</span>
      <div class="text">
        <p class="title">{{ caption.title }}</p>
        <code class="data">{{ caption.data }}</code>
      </div>
      <span class="count">{{ step }} / {{ stops.length - 1 }}</span>
    </div>
  </div>
</template>

<style scoped>
.flow {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}
.diagram {
  width: 100%;
  height: auto;
  max-height: 330px;
  overflow: visible;
}

.ns { fill: rgba(199, 215, 254, 0.03); stroke-width: 1; stroke-dasharray: 4 4; }
.ns-pay { stroke: rgba(89, 253, 160, 0.4); }
.ns-comp { stroke: rgba(0, 224, 224, 0.4); }
.team { font: 700 11px 'Inter', sans-serif; letter-spacing: 0.14em; }
.team.pay, .ns-label .pay { fill: var(--temporal-green); }
.team.comp, .ns-label .comp { fill: var(--temporal-cyan); }
.ns-label { font: 400 9.5px 'Noto Sans Mono', ui-monospace, monospace; fill: var(--temporal-text-muted); }

.edge {
  fill: none;
  stroke: var(--temporal-border-hi);
  stroke-width: 1.5;
  transition: stroke 0.4s, opacity 0.4s;
  opacity: 0.5;
}
.edge.lit { opacity: 1; stroke-width: 2; }
.edge-pay.lit { stroke: var(--temporal-green); }
.edge-comp.lit { stroke: var(--temporal-cyan); }
.edge-nexus.lit { stroke: var(--temporal-lavender); }
.edge-callback { stroke-dasharray: 5 5; }
.edge-callback.lit { stroke: var(--temporal-lavender); }
.edge-result { stroke-dasharray: 5 5; }
.edge-result.lit { stroke: var(--temporal-green); }

.edge-label {
  font: 400 10.5px 'Inter', sans-serif;
  fill: var(--temporal-text-muted);
  opacity: 0.6;
  transition: fill 0.4s, opacity 0.4s;
}
.edge-label.lit { fill: var(--temporal-lavender); opacity: 1; }

.node rect {
  fill: var(--temporal-bg-elev);
  stroke: var(--temporal-border-hi);
  stroke-width: 1;
  transition: stroke 0.4s, fill 0.4s;
}
.node .n-title { font: 500 12.5px 'Noto Sans Mono', ui-monospace, monospace; fill: var(--temporal-text-muted); transition: fill 0.4s; }
.node .n-sub { font: 400 10px 'Noto Sans Mono', ui-monospace, monospace; fill: var(--temporal-text-muted); }
.node.visited .n-title { fill: var(--temporal-text-strong); }
.node .n-sdk { font: 400 9.5px 'Noto Sans Mono', ui-monospace, monospace; fill: var(--temporal-lavender); opacity: 0.85; }
.node .n-file { font: italic 400 9.5px 'Noto Sans Mono', ui-monospace, monospace; fill: var(--temporal-text-muted); }
.node.pay.visited rect { stroke: rgba(89, 253, 160, 0.5); }
.node.comp.visited rect { stroke: rgba(0, 224, 224, 0.5); }
.node.endpoint rect { stroke: rgba(195, 160, 255, 0.45); }
.node.endpoint .n-sub { fill: var(--temporal-lavender); }
.node.endpoint.visited rect { fill: rgba(93, 36, 189, 0.25); }
.node.current rect { stroke-width: 2; }
.node.pay.current rect { stroke: var(--temporal-green); fill: rgba(89, 253, 160, 0.08); }
.node.comp.current rect { stroke: var(--temporal-cyan); fill: rgba(0, 224, 224, 0.08); }
.node.endpoint.current rect { stroke: var(--temporal-lavender); fill: rgba(93, 36, 189, 0.4); }

.req { color: var(--temporal-magenta); }
.req-pill {
  fill: var(--temporal-bg);
  stroke: var(--temporal-magenta);
  stroke-width: 1.25;
  transition: fill 0.3s;
}
.req-text { font: 600 9.5px 'Inter', sans-serif; fill: var(--temporal-magenta); }
.req.active { color: var(--temporal-bg); }
.req.active .req-pill { fill: var(--temporal-magenta); }
.req.active .req-text { fill: var(--temporal-bg); }
.req.active { filter: drop-shadow(0 0 6px rgba(255, 107, 255, 0.8)); }


.trail {
  fill: none;
  stroke: var(--temporal-yellow);
  stroke-width: 2.5;
  stroke-dasharray: 6 6;
  opacity: 0.8;
  animation: march 0.6s linear infinite;
}
@keyframes march { to { stroke-dashoffset: -12; } }
.chip { opacity: 0; transition: opacity 0.35s; }
.chip.moving { opacity: 1; transition: opacity 0.15s; }
.chip rect { fill: var(--temporal-yellow); }
.chip text { font: 600 10.5px 'Noto Sans Mono', ui-monospace, monospace; fill: var(--temporal-bg); }

.packet .core { fill: var(--temporal-yellow); }
.packet .halo { fill: rgba(254, 203, 47, 0.25); animation: pulse 1.4s ease-in-out infinite; transform-box: fill-box; transform-origin: center; }
@keyframes pulse {
  0%, 100% { transform: scale(0.8); opacity: 0.9; }
  50% { transform: scale(1.35); opacity: 0.3; }
}

.caption {
  display: flex;
  align-items: center;
  gap: 0.9rem;
  padding: 0.55rem 0.9rem;
  border: 1px solid var(--temporal-border);
  border-radius: 10px;
  background: var(--temporal-bg-overlay);
  min-height: 3.6rem;
}
.where {
  flex: none;
  width: 6.2rem;
  text-align: center;
  font: 600 0.62rem 'Inter', sans-serif;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  padding: 0.3rem 0;
  border-radius: 999px;
  border: 1px solid currentColor;
}
.where.payments { color: var(--temporal-green); }
.where.nexus { color: var(--temporal-lavender); }
.where.compliance { color: var(--temporal-cyan); }
.text { flex: 1 1 auto; min-width: 0; }
.title { margin: 0; font-size: 0.95rem; color: var(--temporal-text-strong); line-height: 1.3; }
.data {
  display: block;
  margin-top: 0.15rem;
  font-size: 0.68rem;
  color: var(--temporal-yellow);
  background: none;
  border: none;
  padding: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.count { flex: none; font: 400 0.72rem 'Noto Sans Mono', monospace; color: var(--temporal-text-muted); }
</style>
