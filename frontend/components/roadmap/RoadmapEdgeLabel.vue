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
  border-radius: 3px;
  font-size: 0.5625rem;
  font-weight: 500;
  letter-spacing: 0.02em;
  line-height: 1;
  white-space: nowrap;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
}
.edge-type-success {
  border-color: rgba(107, 203, 119, 0.3);
  color: rgba(107, 203, 119, 1);
}
.edge-type-warning {
  border-color: rgba(240, 184, 73, 0.3);
  color: rgba(240, 184, 73, 1);
}
.edge-type-optional {
  border-color: var(--color-border);
  color: var(--color-text-muted);
  opacity: 0.7;
}
</style>
