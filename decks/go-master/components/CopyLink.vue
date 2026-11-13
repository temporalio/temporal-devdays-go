<!--
  ABOUTME: A link with a copy-to-clipboard button beside it, for the lab URL on
  ABOUTME: the exercise slides.

    <CopyLink href="https://t.mp/ddnov122026" />

  The label defaults to the href without its scheme. Clicks stop here, so
  pressing the button copies instead of advancing the slide. Copying uses the
  Clipboard API, with a hidden-textarea fallback for when it is unavailable
  (an insecure origin, or a browser that refuses).
-->
<script setup lang="ts">
import { computed, ref } from 'vue'

const props = defineProps<{ href: string; label?: string }>()
const text = computed(() => props.label ?? props.href.replace(/^https?:\/\//, ''))
const copied = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined

function fallbackCopy(value: string): void {
  const area = document.createElement('textarea')
  area.value = value
  area.setAttribute('readonly', '')
  area.style.position = 'fixed'
  area.style.opacity = '0'
  document.body.appendChild(area)
  area.select()
  document.execCommand('copy')
  area.remove()
}

async function copy(): Promise<void> {
  try {
    await navigator.clipboard.writeText(props.href)
  } catch {
    fallbackCopy(props.href)
  }
  copied.value = true
  clearTimeout(timer)
  timer = setTimeout(() => (copied.value = false), 1500)
}
</script>

<template>
  <span class="copy-link" @click.stop>
    <a :href="href" target="_blank" rel="noopener">{{ text }}</a>
    <button type="button" :class="{ copied }" :aria-label="copied ? 'Copied' : `Copy ${text}`" @click.stop="copy">
      {{ copied ? 'Copied' : 'Copy' }}
    </button>
  </span>
</template>

<style scoped>
.copy-link {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
}
.copy-link a {
  font-weight: 700;
}
button {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  padding: 0.25rem 0.6rem;
  border-radius: 999px;
  border: 1px solid var(--temporal-border-hi);
  color: var(--temporal-text-strong);
  background: rgb(154 128 255 / 15%);
  cursor: pointer;
}
button:hover {
  border-color: var(--temporal-lavender);
}
button.copied {
  color: var(--temporal-green);
  border-color: var(--temporal-green);
  background: rgba(89, 253, 160, 0.12);
}
</style>
