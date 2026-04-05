<template>
  <div v-if="active" class="activity-indicator">
    <span class="activity-dot"></span>
    <span class="activity-label">{{ label }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  active: boolean
  label?: string
}>()
</script>

<style scoped>
.activity-indicator {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-xs);
  color: var(--color-primary);
  padding: var(--space-1) var(--space-3);
  border-radius: 999px;
  background: rgba(56,189,248,0.08);
  border: 1px solid rgba(56,189,248,0.2);
  line-height: 1;
}
.activity-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-primary);
}
.activity-label { font-weight: 500; }
</style>
