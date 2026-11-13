<!--
  ABOUTME: One code block, with a bubble beside the highlighted line explaining it.
  ABOUTME: Ports the PPT pattern where an arrow points at a line and a caption explains it.

  The shared code-walkthrough layout, plus an optional `subheading` (a one-line
  definition under the heading).

  `title` is reserved by Slidev for slide metadata and never reaches a layout as
  a prop, so this uses `heading` — same workaround as code-stack and the theme's
  exercise layout.

  Frontmatter usage:
    ---
    layout: code-walkthrough
    heading: Go Workflow anatomy
    subheading: One line defining the concept. HTML, like notes
    notes:
      - Everything the Workflow needs comes from this one package
      - "<strong>Context first</strong>: every Workflow takes a workflow.Context"
      - ExecuteActivity schedules the Activity. Not a direct function call
    ---

    ::code::

    ```go {all|2|4|10}
    // workflows.go
    import "go.temporal.io/sdk/workflow"
    ...
    ```

  One `notes` entry per highlight step, in order. The code block's `{all|…}`
  burns click 0 showing the whole snippet unannotated, so note 1 lands on
  click 1 alongside the first highlight and the two stay in step. Drop the
  leading `all` and set `notesStart: 0`.

  The bubble tracks the highlight. It sits in a column to the right of the code
  and is positioned vertically on the centre of whatever lines are lit, with a
  tail pointing back at them, so the eye never travels between the code and its
  explanation. `bubbleWidth` sets that column (default 32%) and `minCodeSize`
  the floor the code may shrink to in what is left.

  Notes are rendered as HTML, so `<strong>` and `<code>` work inside them.
  Markdown does not — a YAML scalar never reaches the markdown parser. For a
  note that needs more than a line, leave `notes` off and put your own
  `::default::` content in instead; it renders under the code and you own the
  click wiring.

  Why the layout emits the clicks rather than the markdown doing it: a caption
  has to occupy the *same* click as its highlight, which means an absolute
  range (`v-click="[2, 3]"`). `<v-switch>` cannot express that — its `#N` slots
  are numbered relative to whatever consumed clicks before them, so a preceding
  code block silently shifts every caption past its highlight.

  The code pane measures itself after render and steps its font size down until
  it fits, so a snippet never scrolls or clips. If it shrinks past legibility,
  cut lines or split the slide — that is a content problem, not a layout one.
-->
<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from "vue";
import TemporalFooter from "../theme/components/TemporalFooter.vue";

const props = withDefaults(
  defineProps<{
    heading?: string;
    subheading?: string;
    notes?: string[];
    notesStart?: number;
    bubbleWidth?: string;
    minCodeSize?: number;
  }>(),
  {
    notesStart: 1,
    bubbleWidth: "32%",
    minCodeSize: 0.5,
  },
);

// Each note owns one click: [at, at + 1] shows it for that click and hides it
// again on the next. The last note is the exception — it gets a bare `at` and
// stays up. A closing range there would register a hide at `at + 1`, which
// Slidev counts as another click on the slide, so the presenter would click
// once to blank the annotation and again to actually move on.
const noteRanges = computed(() => {
  const notes = props.notes ?? [];
  return notes.map((html, i) => {
    const at = props.notesStart + i;
    const isLast = i === notes.length - 1;
    return { html, at: isLast ? at : ([at, at + 1] as [number, number]) };
  });
});

const MAX_CODE_SIZE = 1.25; // matches the theme's default, in rem
const STEP = 0.05;

const body = ref<HTMLElement | null>(null);
const codePane = ref<HTMLElement | null>(null);
const codeSize = ref(MAX_CODE_SIZE);

// Vertical centre of the lit lines, in px from the top of the body. Null until
// measured, and while no line is lit — on click 0 the whole snippet carries the
// `highlighted` class, which is not something to point a bubble at.
const bubbleTop = ref<number | null>(null);

let fitObserver: ResizeObserver | undefined;
let contentObserver: MutationObserver | undefined;
let highlightObserver: MutationObserver | undefined;
let fitScheduled = false;
let aimScheduled = false;
let fitting = false;

function overflows(el: HTMLElement) {
  if (el.scrollHeight > el.clientHeight + 1) return true;
  // Width has to be measured on the <pre>, not the pane: Slidev gives code
  // blocks their own overflow-x, so a long line scrolls inside the pre and
  // never widens its container.
  const pre = el.querySelector("pre");
  return pre ? pre.scrollWidth > pre.clientWidth + 1 : false;
}

async function fit() {
  const el = codePane.value;
  if (!el || fitting) return;
  fitting = true;
  try {
    codeSize.value = MAX_CODE_SIZE;
    await nextTick();
    while (codeSize.value > props.minCodeSize && overflows(el)) {
      codeSize.value = Math.round((codeSize.value - STEP) * 100) / 100;
      await nextTick();
    }
  } finally {
    fitting = false;
  }
  aim();
}

// Point the bubble at the middle of the currently lit lines.
function aim() {
  const pre = codePane.value?.querySelector("pre");
  const frame = body.value;
  if (!pre || !frame) return;

  const lines = Array.from(pre.querySelectorAll<HTMLElement>(".line"));
  const lit = lines.filter((l) => l.classList.contains("highlighted"));

  // Nothing lit, or everything lit (the `all` step): no single line to aim at.
  if (!lit.length || lit.length === lines.length) {
    bubbleTop.value = null;
    return;
  }

  // Slidev renders the slide at a fixed canvas width and CSS-scales it to the
  // viewport, so getBoundingClientRect() returns screen pixels while the `top`
  // we write back is interpreted in unscaled canvas pixels. Measuring the frame
  // both ways recovers the factor between them; skip it and the bubble drifts
  // further down the further down the snippet the highlight is.
  const frameRect = frame.getBoundingClientRect();
  const scale = frame.offsetHeight ? frameRect.height / frame.offsetHeight : 1;

  const first = lit[0].getBoundingClientRect();
  const last = lit[lit.length - 1].getBoundingClientRect();
  const centre = (first.top + last.bottom) / 2 - frameRect.top;

  bubbleTop.value = centre / (scale || 1);
}

// Coalesce bursts of observer callbacks into one pass on the next frame.
function scheduleFit() {
  if (fitScheduled) return;
  fitScheduled = true;
  requestAnimationFrame(() => {
    fitScheduled = false;
    fit();
  });
}

function scheduleAim() {
  if (aimScheduled) return;
  aimScheduled = true;
  requestAnimationFrame(() => {
    aimScheduled = false;
    aim();
  });
}

onMounted(() => {
  fit();

  // The mono webfont almost never lands before the first fit. Its metrics
  // differ from the fallback's, so a snippet that measured as fitting can be
  // a few pixels too tall once the real font swaps in — and nothing below
  // notices, because a font swap changes no boxes the observers watch.
  if (typeof document !== "undefined" && document.fonts?.ready) {
    document.fonts.ready.then(scheduleFit).catch(() => {});
  }

  const el = codePane.value;
  if (!el) return;
  const pre = el.querySelector("pre");

  if (typeof ResizeObserver !== "undefined") {
    fitObserver = new ResizeObserver(scheduleFit);
    fitObserver.observe(el);
    // The pane's box is pinned by the grid, so it never resizes. The <pre>
    // inside it does, and that is the thing whose height decides the fit.
    if (pre) fitObserver.observe(pre);
  }

  if (typeof MutationObserver !== "undefined") {
    // A ResizeObserver alone is not enough. Stepping through the highlight
    // swaps the pane's contents without changing its box, so nothing resizes
    // and the old font size sticks while the new content clips.
    contentObserver = new MutationObserver(scheduleFit);
    contentObserver.observe(el, {
      childList: true,
      subtree: true,
      characterData: true,
    });

    // Moving between clicks only rewrites the `class` on each .line, which the
    // content observer above deliberately ignores. Watch that separately so the
    // bubble follows the highlight. Scoped to the <pre> and filtered to class,
    // so our own font-size writes on the pane cannot retrigger it.
    if (pre) {
      highlightObserver = new MutationObserver(scheduleAim);
      highlightObserver.observe(pre, {
        attributes: true,
        attributeFilter: ["class"],
        subtree: true,
      });
    }
  }
});

onBeforeUnmount(() => {
  fitObserver?.disconnect();
  contentObserver?.disconnect();
  highlightObserver?.disconnect();
});

watch(() => [props.bubbleWidth, props.minCodeSize], scheduleFit);
</script>

<template>
  <div class="slidev-layout code-walkthrough bg-grid">
    <h2 v-if="heading" class="heading">{{ heading }}</h2>
    <p v-if="subheading" class="subheading" v-html="subheading" />
    <div
      ref="body"
      class="cw-body"
      :style="{ '--cw-bubble-width': bubbleWidth }"
    >
      <div
        ref="codePane"
        class="code-pane"
        :style="{ '--cw-code-size': `${codeSize}rem` }"
      >
        <slot name="code" />
      </div>
      <div
        class="bubble-layer"
        :class="{ 'is-placed': bubbleTop !== null }"
        :style="{ '--cw-bubble-top': `${bubbleTop ?? 0}px` }"
      >
        <!-- The position lives on the layer, not on the bubbles. A `:class` or
             `:style` binding on the same element as `v-click` gets re-patched by
             Vue whenever its value changes, which wipes the classes the
             directive sets imperatively and leaves every bubble permanently
             hidden. Bind the state to the parent and let CSS inherit it. -->
        <div
          v-for="(note, i) in noteRanges"
          :key="i"
          v-click="note.at"
          class="bubble"
          v-html="note.html"
        />
      </div>
    </div>
    <div class="cw-extra"><slot /></div>
    <TemporalFooter />
  </div>
</template>

<style scoped>
.code-walkthrough {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.subheading {
  margin: -0.35rem 0 0.8rem;
  font-size: 1.05rem;
  line-height: 1.4;
  color: var(--temporal-text);
}
.subheading :deep(strong) {
  color: var(--temporal-green);
  font-weight: 500;
}
.heading {
  margin: 0 0 0.7rem;
  font-size: 1.8rem;
  font-weight: 200;
  color: var(--temporal-text-strong);
}
.cw-body {
  flex: 1 1 auto;
  min-height: 0;
  display: grid;
  /* Code left, bubbles right. The bubble column is reserved whether or not a
     bubble is showing, so the code never reflows between clicks. */
  grid-template-columns: minmax(0, 1fr) var(--cw-bubble-width);
  gap: 1.1rem;
  overflow: hidden;
}
.code-pane {
  min-height: 0;
  overflow: hidden;
  /* Centre a short snippet rather than stranding it under the heading. */
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.code-pane {
  /* Slidev sizes code from these two vars, and the highlighted-line
     decoration reads the line height from them too. Setting font-size on the
     <pre> alone shrinks the glyphs but leaves each line's box at its original
     height, so a snippet keeps overflowing however small the text gets. */
  --slidev-code-font-size: var(--cw-code-size);
  --slidev-code-line-height: calc(var(--cw-code-size) * 1.35);
}
.code-pane :deep(pre.shiki),
.code-pane :deep(pre code),
.code-pane :deep(pre .line) {
  font-size: var(--cw-code-size);
  line-height: calc(var(--cw-code-size) * 1.35);
}
.code-pane :deep(pre) {
  margin: 0;
}

.bubble-layer {
  position: relative;
  min-height: 0;
}
.bubble {
  position: absolute;
  left: 0;
  right: 0;
  /* Centre of the lit lines, measured by the script onto the layer, then
     pulled up by half the bubble's own height so it straddles them. */
  top: var(--cw-bubble-top);
  transform: translateY(-50%);
  padding: 0.55rem 0.8rem;
  border: 1px solid var(--temporal-green);
  border-radius: 0.5rem;
  background: var(--temporal-bg-elev);
  box-shadow: 0 10px 30px -12px rgb(0 0 0 / 70%);
  font-size: 0.92rem;
  line-height: 1.4;
  color: var(--temporal-text-strong);
  /* Moving between clicks slides the bubble to its new line rather than
     teleporting, which makes the connection to the code easier to follow. */
  transition:
    top 220ms ease,
    opacity 160ms ease;
}
/* Until the script has measured, park it at the top rather than at 0 with a
   -50% transform, which would hang it off the slide. */
.bubble-layer:not(.is-placed) .bubble {
  top: 1.2rem;
  transform: none;
}
/* The tail, pointing back at the code. Two stacked triangles fake a 1px
   border: the outer one is the border colour, the inner one sits a pixel
   inside it in the fill colour. */
.bubble::before,
.bubble::after {
  content: "";
  position: absolute;
  top: 50%;
  width: 0;
  height: 0;
  border-style: solid;
  transform: translateY(-50%);
}
.bubble::before {
  left: -0.5rem;
  border-width: 0.44rem 0.5rem 0.44rem 0;
  border-color: transparent var(--temporal-green) transparent transparent;
}
.bubble::after {
  left: -0.42rem;
  border-width: 0.38rem 0.44rem 0.38rem 0;
  border-color: transparent var(--temporal-bg-elev) transparent transparent;
}
.bubble :deep(p) {
  margin: 0;
}
.bubble :deep(code) {
  font-size: 0.9em;
}
.bubble :deep(strong) {
  color: var(--temporal-green);
  font-weight: 500;
}

/* Anything passed as default slot content, under the code. Collapses to
   nothing when the slot is empty. */
.cw-extra:empty {
  display: none;
}
.cw-extra {
  flex: 0 0 auto;
  margin-top: 0.5rem;
  font-size: 0.95rem;
}
</style>
