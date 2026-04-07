<template>
  <div class="roadmap-node roadmap-root-node">
    <div class="rm-root-header">
      <div class="rm-root-title-row">
        <span v-if="data.code" class="rm-root-code">{{ data.code }}</span>
        <span class="rm-root-name">{{ data.title }}</span>
      </div>
      <span :class="['badge', `badge-${data.status}`]">{{ data.status }}</span>
    </div>
    <p v-if="data.description" class="rm-root-desc">{{ truncate(data.description, 100) }}</p>
    <div v-if="data.statuses?.length" class="rm-root-statuses">
      <span v-for="s in data.statuses" :key="s" class="rm-root-status-chip">{{ s }}</span>
    </div>
    <div class="rm-root-stats">
      <span>{{ data.nodeCount }} nodes</span>
      <span>{{ data.edgeCount }} edges</span>
    </div>
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
    status: string
    statuses: string[]
    nodeCount: number
    edgeCount: number
  }
  sourcePosition?: Position
}>()

function truncate(text: string, len: number): string {
  if (!text) return ''
  return text.length > len ? text.slice(0, len) + '...' : text
}
</script>

<style scoped>
.roadmap-root-node {
  background: var(--color-surface);
  border: 2px solid var(--color-primary);
  border-radius: var(--radius-lg);
  padding: var(--space-5) var(--space-6);
  min-width: 320px;
  max-width: 400px;
  box-shadow: 0 4px 24px rgba(108, 197, 224, 0.15);
}
.rm-root-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
  margin-bottom: var(--space-2);
}
.rm-root-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}
.rm-root-code {
  font-size: var(--type-xs);
  font-weight: 700;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.2rem 0.45rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
.rm-root-name {
  font-size: var(--type-lg);
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--color-text);
}
.rm-root-desc {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  line-height: 1.45;
  margin-bottom: var(--space-3);
}
.rm-root-statuses {
  display: flex;
  gap: var(--space-1);
  flex-wrap: wrap;
  margin-bottom: var(--space-3);
}
.rm-root-status-chip {
  font-size: 0.625rem;
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
  line-height: 1;
}
.rm-root-stats {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.rm-root-stats span {
  background: var(--color-surface-hover);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-variant-numeric: tabular-nums;
}
</style>
