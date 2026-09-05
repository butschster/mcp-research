<template>
  <div :class="['roadmap-edge-label', `edge-type-${edgeType}`]">
    <span class="edge-label-text">{{ label }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  edgeType: string
}>()
</script>

<style scoped>
.roadmap-edge-label {
  display: inline-flex;
  align-items: center;
  padding: 0.15rem 0.45rem;
  border-radius: var(--radius-xs);
  font-size: 0.5625rem;
  font-weight: var(--weight-medium);
  letter-spacing: 0.02em;
  line-height: 1;
  white-space: nowrap;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
}
.edge-type-success {
  border-color: rgba(var(--color-success-rgb), 0.3);
  color: rgba(var(--color-success-rgb), 1);
}
.edge-type-warning {
  border-color: rgba(var(--color-warning-rgb), 0.3);
  color: rgba(var(--color-warning-rgb), 1);
}
.edge-type-optional {
  border-color: var(--color-border);
  color: var(--color-text-muted);
  opacity: 0.7;
}
</style>
