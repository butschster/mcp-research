<template>
  <div class="rm-axis">
    <!-- Quarter band row: a label spanning the months of each quarter -->
    <div class="rm-axis-quarters">
      <div
        v-for="band in quarterBands"
        :key="band.label + band.start"
        class="rm-axis-quarter"
        :style="{ gridColumn: `span ${band.span}` }"
      >
        {{ band.label }}
      </div>
    </div>
    <!-- Month cells -->
    <div class="rm-axis-months" :style="gridStyle">
      <div v-for="m in months" :key="m.key" class="rm-axis-month">
        <span class="rm-axis-month-label">{{ m.label }}</span>
        <span class="rm-axis-month-count" v-if="m.nodes.length">{{ m.nodes.length }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TimelineMonth } from '~/utils/roadmap'

const props = defineProps<{ months: TimelineMonth[]; cellWidth: number }>()

const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${props.months.length}, ${props.cellWidth}px)`,
}))

// Group contiguous months into quarter bands so the label ("Q1 2026") spans its
// months. A quarterLabel is set only on the first month of each quarter.
const quarterBands = computed(() => {
  const bands: { label: string; span: number; start: number }[] = []
  props.months.forEach((m, i) => {
    if (m.quarterLabel || bands.length === 0) {
      bands.push({ label: m.quarterLabel || '', span: 1, start: i })
    } else {
      const last = bands[bands.length - 1]
      if (last) last.span++
    }
  })
  return bands
})
</script>

<style scoped>
.rm-axis {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--color-bg);
}
.rm-axis-quarters {
  display: grid;
  grid-auto-flow: column;
  border-bottom: 1px solid var(--color-border);
}
.rm-axis-quarter {
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-faint);
  padding: 0.15rem var(--space-2);
  border-right: 1px solid var(--color-border);
  white-space: nowrap;
}
.rm-axis-months {
  display: grid;
}
.rm-axis-month {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: var(--space-2);
  border-right: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
}
.rm-axis-month-label {
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-text-muted);
}
.rm-axis-month-count {
  font-size: 0.5625rem;
  font-variant-numeric: tabular-nums;
  color: var(--color-text-faint);
  background: var(--color-surface-hover);
  border-radius: var(--radius-xs);
  padding: 0.05rem 0.3rem;
}
</style>
