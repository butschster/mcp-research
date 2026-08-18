<template>
  <div class="rm-axis">
    <!-- Coarser-unit band row (Q for months, year for quarters, hidden at year) -->
    <div v-if="hasBands" class="rm-axis-bands" :style="gridStyle">
      <div
        v-for="band in bands"
        :key="band.label + band.start"
        class="rm-axis-band"
        :style="{ gridColumn: `${band.start + 1} / span ${band.span}` }"
      >
        {{ band.label }}
      </div>
    </div>
    <!-- Unit cells (month / quarter / year) -->
    <div class="rm-axis-cells" :style="gridStyle">
      <div v-for="c in units" :key="c.key" class="rm-axis-cell">
        <span class="rm-axis-cell-label">{{ c.label }}</span>
        <span v-if="c.count" class="rm-axis-cell-count">{{ c.count }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TimeCell } from '~/utils/roadmap'

const props = defineProps<{ units: TimeCell[]; cellWidth: number }>()

const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${props.units.length}, ${props.cellWidth}px)`,
}))

const hasBands = computed(() => props.units.some(c => c.band !== ''))

// Group contiguous cells that share a band caption ("Q1 2026", "2026") so the
// label spans its cells. Year zoom has no band and this is skipped.
const bands = computed(() => {
  const out: { label: string; span: number; start: number }[] = []
  props.units.forEach((c, i) => {
    const last = out[out.length - 1]
    if (last && last.label === c.band) {
      last.span++
    } else {
      out.push({ label: c.band, span: 1, start: i })
    }
  })
  return out
})
</script>

<style scoped>
.rm-axis {
  position: sticky;
  top: 0;
  /* Above the point cards AND their absolutely-positioned warn badge (z-index 1),
     so nothing bleeds over the sticky header while the points row scrolls. */
  z-index: 2;
  background: var(--color-bg);
}
.rm-axis-bands {
  display: grid;
  border-bottom: 1px solid var(--color-border);
}
.rm-axis-band {
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-faint);
  padding: 0.15rem var(--space-2);
  border-right: 1px solid var(--color-border);
  white-space: nowrap;
}
.rm-axis-cells {
  display: grid;
}
.rm-axis-cell {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: var(--space-2);
  border-right: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
}
.rm-axis-cell-label {
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-text-muted);
}
.rm-axis-cell-count {
  font-size: 0.5625rem;
  font-variant-numeric: tabular-nums;
  color: var(--color-text-faint);
  background: var(--color-surface-hover);
  border-radius: var(--radius-xs);
  padding: 0.05rem 0.3rem;
}
</style>
