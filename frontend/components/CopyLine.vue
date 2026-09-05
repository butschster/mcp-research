<template>
  <div class="copy-line">
    <p v-if="label" class="copy-label">{{ label }}</p>
    <div class="copy-box">
      <code class="copy-text">{{ text }}</code>
      <button type="button" class="btn btn-sm copy-btn" @click="copy">{{ copied ? 'Copied' : 'Copy' }}</button>
    </div>
    <!-- The copy buttons this product grew by hand change their label and tell
         a screen reader nothing. One live region, said once. -->
    <span class="sr-only" role="status">{{ announcement }}</span>
  </div>
</template>

<script setup lang="ts">
import { useCopyToClipboard } from '~/composables/useCopyToClipboard'

const props = defineProps<{ text: string; label?: string }>()

const { copied, announcement, copy: writeToClipboard } = useCopyToClipboard()
const copy = () => writeToClipboard(props.text)
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
