<template>
  <div class="rm-timeline-wrap">
    <!-- Nothing dated: an empty state that still lists every node in the tray -->
    <EmptyState
      v-if="!axis.units.length"
      icon="&#x1F4C5;"
      title="No dates on this roadmap"
      :description="emptyDescription"
    >
      <button class="btn btn-sm" @click="emit('switch-graph')">Switch to Graph</button>
    </EmptyState>

    <template v-else>
      <!-- Local toolbar: zoom + the undated jump -->
      <div class="rm-tl-toolbar">
        <RoadmapGranularityToggle :model-value="unit" @update:model-value="onZoom" />
        <span v-if="axis.undated.length" class="rm-tl-undated-note">No date ({{ axis.undated.length }})</span>
      </div>

      <div class="rm-tl-scroll">
        <RoadmapTimeAxis :units="axis.units" :cell-width="cellWidth" />

        <!-- Ranges band: greedy-laned bars spanning their cells -->
        <div v-if="axis.bars.length" class="rm-tl-ranges" :style="rangesStyle">
          <div
            v-for="b in axis.bars"
            :key="b.node.id"
            class="rm-tl-bar-slot"
            :style="barStyle(b)"
          >
            <RoadmapTimelineBar
              :data="nodeCardData(b.node)"
              :deps="deps.get(b.node.id)"
              :start-label="cellLabel(b.startIdx)"
              :end-label="cellLabel(b.endIdx)"
              :highlighted="isHighlighted(b.node.id)"
              :dimmed="isDimmed(b.node.id)"
              @click="emit('node-click', b.node.id)"
              @hover="onHover(b.node.id, $event)"
            />
          </div>
        </div>

        <!-- Points row: one cell per unit, points stacked within -->
        <div v-if="axis.points.length" class="rm-tl-points" :style="pointsStyle">
          <div
            v-for="(cell, i) in pointCells"
            :key="i"
            class="rm-tl-cell"
            :style="{ gridColumn: i + 1 }"
          >
            <template v-for="p in cell" :key="p.node.id">
              <!-- Milestone → diamond marker -->
              <button
                v-if="p.node.node_type === 'milestone'"
                type="button"
                class="rm-milestone"
                :class="{ 'is-dimmed': isDimmed(p.node.id) }"
                @click="emit('node-click', p.node.id)"
                @mouseenter="onHover(p.node.id, true)"
                @mouseleave="onHover(p.node.id, false)"
                @focus="onHover(p.node.id, true)"
                @blur="onHover(p.node.id, false)"
              >
                <span class="rm-milestone-diamond">&#x25C6;</span>
                <span class="rm-milestone-label">
                  <span v-if="p.node.code" class="rm-milestone-code">{{ p.node.code }}</span>
                  {{ p.node.title }}
                </span>
              </button>
              <!-- Regular point → compact card -->
              <button
                v-else
                type="button"
                class="rm-card-btn"
                @click="emit('node-click', p.node.id)"
                @mouseenter="onHover(p.node.id, true)"
                @mouseleave="onHover(p.node.id, false)"
                @focus="onHover(p.node.id, true)"
                @blur="onHover(p.node.id, false)"
              >
                <span v-if="p.invalid" class="rm-tl-warn" title="End date is before start date">&#x26A0;</span>
                <RoadmapNodeCard
                  :data="nodeCardData(p.node)"
                  :deps="deps.get(p.node.id)"
                  :highlighted="isHighlighted(p.node.id)"
                  :dimmed="isDimmed(p.node.id)"
                  compact
                />
              </button>
            </template>
          </div>
        </div>
      </div>

      <!-- Undated tray -->
      <details v-if="axis.undated.length" class="rm-tray">
        <summary class="rm-tray-head">No date ({{ axis.undated.length }})</summary>
        <div class="rm-tray-body">
          <button
            v-for="n in axis.undated"
            :key="n.id"
            type="button"
            class="rm-card-btn rm-tray-card"
            @click="emit('node-click', n.id)"
          >
            <RoadmapNodeCard :data="nodeCardData(n)" :deps="deps.get(n.id)" compact />
          </button>
        </div>
      </details>
    </template>
  </div>
</template>

<script setup lang="ts">
import {
  buildTimeAxis, autoUnit, depsByNode, relatedIds, nodeCardData,
  type RawRoadmapNode, type RawRoadmapEdge, type TimeUnit, type TimelineBar,
} from '~/utils/roadmap'

const props = defineProps<{
  nodes: readonly RawRoadmapNode[]
  edges: readonly RawRoadmapEdge[]
}>()

const emit = defineEmits<{ 'node-click': [string]; 'switch-graph': [] }>()

// Zoom: initialised once from the dated span (so a multi-year plan opens fitted),
// then local — a manual choice survives a data refresh within the session.
const unit = ref<TimeUnit>('month')
// Only a real click sets this — a programmatic auto-fit must not trip it, or the
// span-based default would freeze on whatever the first (often empty) data gave.
let unitTouched = false
const isNarrow = ref(false)

function onZoom(u: TimeUnit) {
  unitTouched = true
  unit.value = u
}

watch(() => props.nodes, (ns) => {
  if (!unitTouched) unit.value = autoUnit(ns, isNarrow.value)
}, { immediate: true })

onMounted(() => {
  isNarrow.value = window.matchMedia('(max-width: 768px)').matches
  if (!unitTouched) unit.value = autoUnit(props.nodes, isNarrow.value)
})

const CELL_WIDTH: Record<TimeUnit, number> = { month: 200, quarter: 200, year: 240 }
const cellWidth = computed(() => CELL_WIDTH[unit.value])

const axis = computed(() => buildTimeAxis(props.nodes, unit.value))
const deps = computed(() => depsByNode(props.nodes, props.edges))

const emptyDescription = computed(() =>
  props.nodes.length
    ? 'Add a date to nodes to place them on a timeline, or switch to Graph. All nodes are listed below.'
    : 'This roadmap has no nodes yet. Add some, or switch to Graph.')

function cellLabel(idx: number): string {
  return axis.value.units[idx - axis.value.first]?.label ?? ''
}

// Points bucketed per cell, in axis order, so each cell stacks its points.
const pointCells = computed(() => {
  const cells: { node: RawRoadmapNode; invalid: boolean }[][] = axis.value.units.map(() => [])
  for (const p of axis.value.points) {
    const col = cells[p.idx - axis.value.first]
    if (col) col.push({ node: p.node, invalid: p.invalid })
  }
  return cells
})

const rangesStyle = computed(() => ({
  gridTemplateColumns: `repeat(${axis.value.units.length}, ${cellWidth.value}px)`,
  gridTemplateRows: `repeat(${axis.value.laneCount}, var(--space-8))`,
}))
const pointsStyle = computed(() => ({
  gridTemplateColumns: `repeat(${axis.value.units.length}, ${cellWidth.value}px)`,
}))
function barStyle(b: TimelineBar) {
  return {
    gridColumn: `${b.startIdx - axis.value.first + 1} / span ${b.endIdx - b.startIdx + 1}`,
    gridRow: `${b.lane + 1}`,
  }
}

// Hover/focus highlight — same mechanics as the stages board, now on the timeline.
const hoveredId = ref<string | null>(null)
const highlightIds = computed<Set<string> | null>(() => {
  if (!hoveredId.value) return null
  const set = relatedIds(hoveredId.value, props.edges)
  set.add(hoveredId.value)
  return set
})
function onHover(id: string, on: boolean) {
  hoveredId.value = on ? id : (hoveredId.value === id ? null : hoveredId.value)
}
const isHighlighted = (id: string) => !!highlightIds.value?.has(id)
const isDimmed = (id: string) => !!highlightIds.value && !highlightIds.value.has(id)
</script>

<style scoped>
.rm-timeline-wrap {
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
}
.rm-tl-toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-5);
  border-bottom: 1px solid var(--color-border);
}
.rm-tl-undated-note {
  margin-left: auto;
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.rm-tl-scroll {
  flex: 1;
  min-height: 0;
  overflow: auto;
}
.rm-tl-ranges {
  display: grid;
  grid-auto-rows: var(--space-8);
  border-bottom: 1px solid var(--color-border);
  padding-top: var(--space-1);
}
.rm-tl-bar-slot {
  padding: var(--space-1) 2px;
  min-width: 0;
}
.rm-tl-points {
  display: grid;
  align-items: start;
}
.rm-tl-cell {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding: var(--space-3);
  border-right: 1px solid var(--color-border);
  min-height: var(--space-8);
}
.rm-card-btn {
  appearance: none;
  background: none;
  border: none;
  padding: 0;
  margin: 0;
  cursor: pointer;
  display: block;
  width: 100%;
  border-radius: var(--radius);
  position: relative;
}
.rm-card-btn:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
.rm-tl-warn {
  position: absolute;
  top: -6px;
  right: -6px;
  z-index: 1;
  font-size: 0.75rem;
  line-height: 1;
}

.rm-milestone {
  appearance: none;
  background: none;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-1) 0;
  text-align: left;
  transition: opacity var(--transition-fast);
}
.rm-milestone.is-dimmed { opacity: 0.5; }
.rm-milestone:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
  border-radius: var(--radius-xs);
}
.rm-milestone-diamond {
  color: rgba(168, 130, 255, 1);
  font-size: var(--type-base);
  line-height: 1;
  flex-shrink: 0;
}
.rm-milestone-label {
  font-size: var(--type-xs);
  color: var(--color-text);
  display: flex;
  align-items: center;
  gap: 0.3rem;
  overflow-wrap: anywhere;
}
.rm-milestone-code {
  font-size: 0.5625rem;
  font-weight: var(--weight-bold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.05rem 0.25rem;
  border-radius: var(--radius-xs);
  font-family: 'JetBrains Mono', monospace;
}

.rm-tray {
  flex-shrink: 0;
  border-top: 1px solid var(--color-border);
}
.rm-tray-head {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  padding: var(--space-2) var(--space-5);
  cursor: pointer;
  user-select: none;
}
.rm-tray-body {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-5) var(--space-5);
  overflow-y: auto;
  max-height: 40%;
}
.rm-tray-card {
  width: 300px;
}
</style>
