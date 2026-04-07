<template>
  <div class="card roadmap-card" @click="$emit('click')">
    <div class="rm-card-header">
      <div class="rm-card-title-row">
        <span v-if="roadmap.code" class="rm-card-code">{{ roadmap.code }}</span>
        <h3 class="rm-card-title">{{ roadmap.title }}</h3>
      </div>
      <span :class="['badge', `badge-${roadmap.status}`]">{{ roadmap.status }}</span>
    </div>
    <p v-if="roadmap.description" class="rm-card-desc">{{ roadmap.description }}</p>
    <div v-if="roadmap.statuses?.length" class="rm-card-statuses">
      <span v-for="s in roadmap.statuses" :key="s" class="rm-card-status-chip">{{ s }}</span>
    </div>
    <div class="rm-card-meta">
      <span v-if="roadmap.nodeCount !== undefined" class="rm-card-count">{{ roadmap.nodeCount }} nodes</span>
      <span v-if="roadmap.edgeCount !== undefined" class="rm-card-count">{{ roadmap.edgeCount }} edges</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  roadmap: {
    code: string
    title: string
    description: string
    status: string
    statuses: string[]
    nodeCount?: number
    edgeCount?: number
  }
}>()

defineEmits<{ click: [] }>()
</script>

<style scoped>
.roadmap-card {
  display: flex;
  flex-direction: column;
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
.roadmap-card:hover {
  border-color: var(--color-border-strong);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}
.rm-card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}
.rm-card-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}
.rm-card-code {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
.rm-card-title {
  font-size: var(--type-sm);
  font-weight: 600;
  color: var(--color-text);
  margin: 0;
  line-height: 1.3;
}
.rm-card-desc {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  line-height: 1.5;
  margin-top: var(--space-2);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.rm-card-statuses {
  display: flex;
  gap: var(--space-1);
  flex-wrap: wrap;
  margin-top: var(--space-3);
}
.rm-card-status-chip {
  font-size: 0.625rem;
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
  line-height: 1;
}
.rm-card-meta {
  display: flex;
  gap: var(--space-3);
  margin-top: var(--space-3);
  padding-top: var(--space-3);
  border-top: 1px solid var(--color-border);
}
.rm-card-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}
</style>
