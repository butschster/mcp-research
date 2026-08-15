<template>
  <div v-if="active" class="activity-indicator" aria-hidden="true">
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
  padding: 0.2rem 0.5rem;
  border-radius: var(--radius-xs);
  background: var(--color-primary-muted);
  border: 1px solid rgba(108, 197, 224, 0.12);
  line-height: 1;
  animation: indicator-in 0.2s ease both;
}
@keyframes indicator-in {
  from { opacity: 0; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1); }
}
.activity-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-primary);
  animation: pulse-dot 1.5s ease-in-out infinite;
}
@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.8); }
}
.activity-label { font-weight: var(--weight-medium); }
</style>
