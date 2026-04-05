<template>
  <div class="mindmap-node task-node">
    <div class="t-header">
      <span class="t-title">{{ truncate(data.title, 60) }}</span>
      <StatusBadge :status="data.status" />
    </div>
    <div class="t-meta">
      <StatusBadge v-if="data.priority" :status="data.priority" />
      <span v-if="data.result" class="t-result">{{ truncate(data.result, 50) }}</span>
    </div>
    <Handle type="target" :position="targetPosition" />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

defineProps<{
  data: {
    title: string
    result: string
    status: string
    priority: string
  }
  targetPosition?: Position
}>()

function truncate(text: string, len: number): string {
  if (!text) return ''
  return text.length > len ? text.slice(0, len) + '...' : text
}
</script>

<style scoped>
.task-node {
  background: var(--color-surface);
  border: 1px solid rgba(239, 107, 107, 0.2);
  border-radius: var(--radius);
  padding: var(--space-4) var(--space-5);
  min-width: 340px;
  max-width: 420px;
}
.t-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
  margin-bottom: var(--space-1);
}
.t-title {
  font-size: var(--type-xs);
  font-weight: 500;
  color: var(--color-text);
  line-height: 1.35;
}
.t-meta {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
.t-result {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  line-height: 1.3;
}
</style>
