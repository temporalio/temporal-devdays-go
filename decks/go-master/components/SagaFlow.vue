<!--
  ABOUTME: Embeds public/diagrams/saga-flow.html and drives its stepper from the slide's clicks.
  ABOUTME: Click N shows tour position N; arrow keys pressed inside the iframe still move the deck.

  Frontmatter usage (clicks must equal the diagram's TOUR length minus one):
    ---
    clicks: 15
    ---

    <SagaFlow />

  `panes="without"` (or "with") shows one half of the story; set `clicks` to
  that pane's step count.
-->
<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useSlideContext } from '@slidev/client'

const props = withDefaults(defineProps<{ height?: string; panes?: string }>(), { height: '420px' })

const { $clicks, $nav } = useSlideContext()
const frame = ref<HTMLIFrameElement>()
const src = `${import.meta.env.BASE_URL}diagrams/saga-flow.html?embed=1&theme=dark${props.panes ? `&panes=${props.panes}` : ''}`

function send() {
  frame.value?.contentWindow?.postMessage({ type: 'saga-flow:goto', index: $clicks.value }, '*')
}

function onMessage(ev: MessageEvent) {
  if (ev.source !== frame.value?.contentWindow || ev.data?.type !== 'saga-flow:key') return
  const key = ev.data.key
  if (key === 'ArrowRight' || key === 'ArrowDown' || key === ' ' || key === 'PageDown') $nav.value.next()
  else $nav.value.prev()
}

watch($clicks, send)
onMounted(() => window.addEventListener('message', onMessage))
onBeforeUnmount(() => window.removeEventListener('message', onMessage))
</script>

<template>
  <iframe
    ref="frame"
    :src="src"
    :style="{ height: props.height }"
    class="saga-flow"
    title="Saga order fulfillment, step by step"
    @load="send"
  />
</template>

<style scoped>
.saga-flow {
  width: 100%;
  border: 0;
  background: transparent;
  color-scheme: dark;
}
</style>
