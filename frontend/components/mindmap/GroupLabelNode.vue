<template>
  <div :class="['mindmap-node group-label-node', `group-${data.icon}`]" @click="$emit('toggle', $event)">
    <div class="group-header">
      <svg v-if="data.icon === 'question'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/></svg>
      <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
      <span class="group-name">{{ data.label }}</span>
      <span class="group-count">{{ data.count }}</span>
    </div>
    <Handle type="target" :position="targetPosition" />
    <Handle type="source" :position="sourcePosition" />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

defineProps<{
  data: {
    label: string
    count: number
    icon: string
  }
  sourcePosition?: Position
  targetPosition?: Position
}>()

defineEmits<{ toggle: [e: MouseEvent] }>()
</script>

<style scoped>
.group-label-node {
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius);
  padding: var(--space-3) var(--space-4);
  cursor: pointer;
  transition: border-color var(--transition-fast);
}
.group-question { border-color: rgba(240, 184, 73, 0.3); }
.group-task { border-color: rgba(239, 107, 107, 0.3); }
.group-label-node:hover { border-color: var(--color-primary); }
.group-header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
.group-name {
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
}
.group-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
  font-variant-numeric: tabular-nums;
}
</style>
