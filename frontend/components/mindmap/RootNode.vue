<template>
  <div class="mindmap-node root-node">
    <div class="root-header">
      <span class="root-name">{{ data.name }}</span>
      <StatusBadge :status="data.status" />
    </div>
    <p v-if="data.goal" class="root-goal">{{ truncate(data.goal, 80) }}</p>
    <div class="root-stats">
      <span>{{ data.sectionCount }} sections</span>
      <span>{{ data.entryCount }} entries</span>
      <span>{{ data.questionCount }} questions</span>
      <span>{{ data.taskCount }} tasks</span>
    </div>
    <Handle id="right" type="source" :position="Position.Right" />
    <Handle id="left" type="source" :position="Position.Left" />
  </div>
</template>

<script setup lang="ts">
import { truncate } from '~/utils/truncate'
import { Handle, Position } from '@vue-flow/core'

defineProps<{
  data: {
    name: string
    goal: string
    status: string
    sectionCount: number
    entryCount: number
    questionCount: number
    taskCount: number
  }
  sourcePosition?: Position
}>()

</script>

<style scoped>
.root-node {
  background: var(--color-surface);
  border: 2px solid var(--color-primary);
  border-radius: var(--radius-lg);
  padding: var(--space-5) var(--space-6);
  min-width: 340px;
  max-width: 400px;
  box-shadow: var(--shadow-glow);
}
.root-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  margin-bottom: var(--space-2);
}
.root-name {
  font-size: var(--type-lg);
  font-weight: var(--weight-bold);
  letter-spacing: -0.02em;
  color: var(--color-text);
}
.root-goal {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  line-height: 1.4;
  margin-bottom: var(--space-3);
}
.root-stats {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.root-stats span {
  background: var(--color-surface-hover);
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
  font-variant-numeric: tabular-nums;
}
</style>
