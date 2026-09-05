<template>
  <div class="empty-state">
    <div v-if="icon" class="empty-icon">{{ icon }}</div>
    <p class="empty-title">{{ title }}</p>
    <p v-if="description" class="empty-desc card-meta">{{ description }}</p>
    <div v-if="command" class="empty-command">
      <code class="command-text">{{ command }}</code>
      <button class="copy-btn" :class="{ copied }" @click="copy">
        {{ copied ? '&#x2713; Copied' : 'Copy' }}
      </button>
    </div>
    <!-- Visible, not only announced. Outside a secure context the clipboard API
         is absent — an ordinary way to run this product on a LAN — and a button
         that does nothing teaches a sighted reader nothing. -->
    <p v-if="failed" class="empty-copy-failed">{{ announcement }}</p>
    <span class="sr-only" role="status">{{ announcement }}</span>
    <!-- What the reader should do next. An empty state that names no action is
         a dead end; an error state without one is worse. -->
    <div v-if="$slots.default" class="empty-actions">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useCopyToClipboard } from '~/composables/useCopyToClipboard'

const props = defineProps<{
  icon?: string
  title: string
  description?: string
  command?: string
}>()

const { copied, failed, announcement, copy: writeToClipboard } = useCopyToClipboard()

function copy() {
  if (!props.command) return
  return writeToClipboard(props.command)
}
</script>

<style scoped>
.empty-copy-failed { margin-top: var(--space-2); font-size: var(--type-xs); color: var(--color-text-muted); }
.empty-icon  { font-size: 2.5rem; margin-bottom: var(--space-5); line-height: 1; opacity: 0.7; }
.empty-title { font-size: var(--type-lg); font-weight: var(--weight-semibold); margin-bottom: var(--space-2); color: var(--color-text); line-height: var(--line-tight); letter-spacing: -0.01em; overflow-wrap: anywhere; }
.empty-desc  { font-size: var(--type-sm); overflow-wrap: anywhere; max-width: 40ch; margin-left: auto; margin-right: auto; }
.empty-actions { margin-top: var(--space-4); display: flex; gap: var(--space-2); justify-content: center; }
</style>
