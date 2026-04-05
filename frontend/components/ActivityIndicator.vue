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
  gap: 0.375rem;
  font-size: 0.75rem;
  color: var(--color-text-muted);
  padding: 0.25rem 0.625rem;
  border-radius: 999px;
  background: rgba(56,189,248,0.08);
  border: 1px solid rgba(56,189,248,0.2);
}
.activity-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-primary);
  animation: activity-pulse 1.5s ease-in-out infinite;
}
@keyframes activity-pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50%       { opacity: 0.4; transform: scale(0.8); }
}
.activity-label { color: var(--color-primary); font-weight: 500; }
</style>
