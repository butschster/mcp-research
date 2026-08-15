<template>
  <div class="progress-wrap">
    <div class="progress-bar" :title="`${pct}%`">
      <div class="progress-bar-fill" :class="fillClass" :style="{ width: pct + '%' }"></div>
    </div>
    <span v-if="showLabel" class="progress-label">{{ pct }}%</span>
  </div>
  <slot />
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  value: number
  total: number
  showLabel?: boolean
}>(), {
  showLabel: false,
})

const pct = computed(() => props.total > 0 ? Math.round((props.value / props.total) * 100) : 0)

const fillClass = computed(() => {
  if (pct.value === 100) return 'fill-complete'
  if (pct.value >= 70) return 'fill-good'
  if (pct.value >= 30) return 'fill-mid'
  return 'fill-low'
})
</script>

<style scoped>
.progress-wrap { display: flex; align-items: center; gap: var(--space-2); }
.progress-bar {
  flex: 1;
  height: 4px;
  background: var(--color-surface-hover);
  border-radius: var(--radius-hair);
  overflow: hidden;
}
.progress-bar-fill {
  height: 100%;
  border-radius: var(--radius-hair);
  transition: width 0.3s;
}
.fill-complete { background: var(--color-success); }
.fill-good     { background: var(--color-info); }
.fill-mid      { background: var(--color-warning); }
.fill-low      { background: var(--color-error); }
.progress-label {
  font-size: var(--type-xs);
  font-weight: var(--weight-medium);
  color: var(--color-text-muted);
  min-width: 2.25rem;
  text-align: right;
  font-variant-numeric: tabular-nums;
}
</style>
