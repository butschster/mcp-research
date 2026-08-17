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
      <span :class="['step-type-badge', `badge-${data.nodeType}`]">
        <svg v-if="data.nodeType === 'step'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="m8 12 3 3 5-5"/></svg>
        <svg v-else-if="data.nodeType === 'milestone'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z"/><line x1="4" x2="4" y1="22" y2="15"/></svg>
        <svg v-else-if="data.nodeType === 'decision'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M16 3h5v5"/><path d="M8 3H3v5"/><path d="M12 22v-8.3a4 4 0 0 0-1.172-2.872L3 3"/><path d="m15 9 6-6"/></svg>
        <svg v-else-if="data.nodeType === 'info'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
        <svg v-else-if="data.nodeType === 'group'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><rect width="7" height="7" x="3" y="3" rx="1"/><rect width="7" height="7" x="14" y="3" rx="1"/><rect width="7" height="7" x="14" y="14" rx="1"/><rect width="7" height="7" x="3" y="14" rx="1"/></svg>
        {{ data.nodeType }}
      </span>
    </div>
    <Handle type="target" :position="targetPosition" />
    <Handle type="source" :position="sourcePosition" />
  </div>
</template>

<script setup lang="ts">
import { truncate } from '~/utils/truncate'
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
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
.step-node:hover {
  border-color: var(--color-border-strong);
  box-shadow: var(--shadow-1);
}

/* Subtle background tint per type */
.step-type-step      { background: rgba(107, 203, 119, 0.04); }
.step-type-milestone { background: rgba(168, 130, 255, 0.06); }
.step-type-decision  { background: rgba(240, 184, 73, 0.05); }
.step-type-info      { background: rgba(108, 197, 224, 0.04); }
.step-type-group     { background: rgba(160, 160, 160, 0.04); }

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
  font-weight: var(--weight-semibold);
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

/* Type badge with icon and color */
.step-type-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
  line-height: 1;
  font-weight: var(--weight-semibold);
}
.badge-step {
  background: rgba(107, 203, 119, 0.12);
  color: rgba(107, 203, 119, 1);
}
.badge-milestone {
  background: rgba(168, 130, 255, 0.12);
  color: rgba(168, 130, 255, 1);
}
.badge-decision {
  background: rgba(240, 184, 73, 0.12);
  color: rgba(240, 184, 73, 1);
}
.badge-info {
  background: rgba(108, 197, 224, 0.12);
  color: rgba(108, 197, 224, 1);
}
.badge-group {
  background: rgba(160, 160, 160, 0.12);
  color: rgba(160, 160, 160, 1);
}

.rm-code {
  font-size: 0.625rem;
  font-weight: var(--weight-bold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.1rem 0.3rem;
  border-radius: var(--radius-xs);
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
.rm-status {
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-weight: var(--weight-semibold);
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
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
