<template>
  <div :class="['roadmap-node', 'step-node', `step-type-${data.nodeType}`]">
    <div class="step-header">
      <div class="step-title-row">
        <span v-if="data.code" class="rm-code">{{ data.code }}</span>
        <span class="step-title">{{ truncate(data.title, 50) }}</span>
      </div>
      <span v-if="data.status" :class="['rm-status', `rm-status-${statusSlug(data.status)}`]">{{ data.status }}</span>
    </div>
    <p v-if="data.description" class="step-desc">{{ truncate(data.description, 100) }}</p>
    <div class="step-footer">
      <span class="step-type-badge">{{ data.nodeType }}</span>
    </div>
    <Handle type="target" :position="targetPosition" />
    <Handle type="source" :position="sourcePosition" />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

defineProps<{
  data: {
    code: string
    title: string
    description: string
    nodeType: string
    status: string
  }
  sourcePosition?: Position
  targetPosition?: Position
}>()

function truncate(text: string, len: number): string {
  if (!text) return ''
  return text.length > len ? text.slice(0, len) + '...' : text
}

function statusSlug(s: string): string {
  return s.replace(/[^a-z0-9]/gi, '-').toLowerCase()
}
</script>

<style scoped>
.step-node {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--space-4) var(--space-5);
  min-width: 280px;
  max-width: 380px;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
.step-node:hover {
  border-color: var(--color-border-strong);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

/* Type-based left accent */
.step-type-milestone { border-left: 3px solid rgba(168, 130, 255, 0.6); }
.step-type-decision  { border-left: 3px solid rgba(240, 184, 73, 0.6); }
.step-type-info      { border-left: 3px solid rgba(108, 197, 224, 0.6); }
.step-type-group     { border-left: 3px solid rgba(160, 160, 160, 0.4); }
.step-type-step      { border-left: 3px solid rgba(107, 203, 119, 0.6); }

.step-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
  margin-bottom: var(--space-1);
}
.step-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}
.step-title {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-text);
  line-height: 1.3;
}
.step-desc {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  line-height: 1.4;
  margin-bottom: var(--space-2);
}
.step-footer {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
.step-type-badge {
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  line-height: 1;
}
.rm-code {
  font-size: 0.625rem;
  font-weight: 700;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.1rem 0.3rem;
  border-radius: 3px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
.rm-status {
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-weight: 600;
  padding: 0.15rem 0.4rem;
  border-radius: 3px;
  line-height: 1;
  white-space: nowrap;
  flex-shrink: 0;
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
}
.rm-status-completed,
.rm-status-mastered,
.rm-status-done {
  background: rgba(107, 203, 119, 0.15);
  color: rgba(107, 203, 119, 1);
}
.rm-status-in-progress,
.rm-status-learning,
.rm-status-active {
  background: rgba(108, 197, 224, 0.15);
  color: rgba(108, 197, 224, 1);
}
.rm-status-not-started,
.rm-status-pending {
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
}
.rm-status-skipped,
.rm-status-blocked {
  background: rgba(239, 107, 107, 0.12);
  color: rgba(239, 107, 107, 1);
}
</style>
