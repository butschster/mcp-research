<template>
  <div class="mindmap-node section-node" @click="$emit('toggle', $event)">
    <div class="section-header">
      <div class="section-title-row">
        <span v-if="data.code" class="mm-code">{{ data.code }}</span>
        <span class="section-name">{{ data.name }}</span>
      </div>
      <StatusBadge :status="data.status" />
    </div>
    <p v-if="data.description" class="section-desc">{{ truncate(data.description, 60) }}</p>
    <span class="section-count">{{ data.entryCount }} entries</span>
    <Handle type="target" :position="targetPosition" />
    <Handle type="source" :position="sourcePosition" />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

defineProps<{
  data: {
    code: string
    name: string
    description: string
    status: string
    entryCount: number
  }
  sourcePosition?: Position
  targetPosition?: Position
}>()

defineEmits<{ toggle: [e: MouseEvent] }>()

function truncate(text: string, len: number): string {
  return text.length > len ? text.slice(0, len) + '...' : text
}
</script>

<style scoped>
.section-node {
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius);
  padding: var(--space-4) var(--space-5);
  min-width: 280px;
  max-width: 360px;
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
.section-node:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-glow);
}
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  margin-bottom: var(--space-1);
}
.section-title-row { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }
.mm-code {
  font-size: 0.625rem; font-weight: var(--weight-bold); color: var(--color-primary);
  background: var(--color-primary-muted); padding: 0.1rem 0.3rem;
  border-radius: var(--radius-xs); font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0; line-height: 1;
}
.section-name {
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
  letter-spacing: -0.01em;
}
.section-desc {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  line-height: 1.4;
  margin-bottom: var(--space-2);
}
.section-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}
</style>
