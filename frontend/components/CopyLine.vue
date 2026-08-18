<template>
  <div class="copy-line">
    <p v-if="label" class="copy-label">{{ label }}</p>
    <div class="copy-box">
      <code class="copy-text">{{ text }}</code>
      <button type="button" class="btn btn-sm copy-btn" @click="copy">{{ copied ? 'Copied' : 'Copy' }}</button>
    </div>
    <!-- The three hand-rolled copy buttons already in this product change their
         label and tell a screen reader nothing. One live region, said once. -->
    <span class="sr-only" role="status">{{ announcement }}</span>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ text: string; label?: string }>()

const copied = ref(false)
const announcement = ref('')
let timer: ReturnType<typeof setTimeout> | undefined

async function copy() {
  try {
    await navigator.clipboard.writeText(props.text)
    copied.value = true
    announcement.value = 'Copied to the clipboard'
  } catch {
    // A refused clipboard is not an error worth a toast: the text is on screen
    // and selectable, which is the fallback anyway.
    announcement.value = 'Could not copy — select the text and copy it yourself'
  }
  clearTimeout(timer)
  timer = setTimeout(() => {
    copied.value = false
    announcement.value = ''
  }, 2000)
}

onBeforeUnmount(() => clearTimeout(timer))
</script>

<style scoped>
.copy-label {
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: var(--space-2);
}
.copy-box {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  background: var(--color-surface-hover);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}
.copy-text {
  flex: 1;
  font-size: var(--type-xs);
  color: var(--color-text);
  /* A command with a long slug in it must wrap rather than push the card. */
  overflow-wrap: anywhere;
}
.copy-btn { flex-shrink: 0; }
</style>
