<template>
  <div class="rm-timeline-wrap">
    <!-- Nothing dated: an empty state that still lists every node in the tray -->
    <EmptyState
      v-if="months.length === 0"
      icon="&#x1F4C5;"
      title="No dates on this roadmap"
      description="Add a date to nodes to place them on a timeline, or switch to Graph. All nodes are listed below."
    >
      <button class="btn btn-sm" @click="emit('switch-graph')">Switch to Graph</button>
    </EmptyState>

    <div v-else class="rm-timeline-scroll">
      <RoadmapTimeAxis :months="months" :cell-width="cellWidth" />
      <div class="rm-timeline-lane" :style="laneStyle">
        <div v-for="m in months" :key="m.key" class="rm-timeline-cell">
          <template v-for="n in m.nodes" :key="n.id">
            <!-- Milestone: a diamond marker on the axis, not a full card -->
            <button
              v-if="n.node_type === 'milestone'"
              type="button"
              class="rm-milestone"
              @click="emit('node-click', n.id)"
            >
              <span class="rm-milestone-diamond">&#x25C6;</span>
              <span class="rm-milestone-label">
                <span v-if="n.code" class="rm-milestone-code">{{ n.code }}</span>
                {{ n.title }}
              </span>
            </button>
            <!-- Regular node: full card -->
            <button
              v-else
              type="button"
              class="rm-card-btn"
              @click="emit('node-click', n.id)"
            >
              <RoadmapNodeCard :data="nodeCardData(n)" :deps="deps.get(n.id)" compact />
            </button>
          </template>
        </div>
      </div>
    </div>

    <!-- Undated tray: expanded by default when the axis is empty, else collapsible -->
    <details v-if="undated.length" class="rm-tray" :open="months.length === 0">
      <summary class="rm-tray-head">No date ({{ undated.length }})</summary>
      <div class="rm-tray-body">
        <button
          v-for="n in undated"
          :key="n.id"
          type="button"
          class="rm-card-btn rm-tray-card"
          @click="emit('node-click', n.id)"
        >
          <RoadmapNodeCard :data="nodeCardData(n)" :deps="deps.get(n.id)" compact />
        </button>
      </div>
    </details>
  </div>
</template>

<script setup lang="ts">
import { buildMonthAxis, depsByNode, nodeCardData, type RawRoadmapNode, type RawRoadmapEdge } from '~/utils/roadmap'

const props = defineProps<{
  nodes: readonly RawRoadmapNode[]
  edges: readonly RawRoadmapEdge[]
}>()

const emit = defineEmits<{ 'node-click': [string]; 'switch-graph': [] }>()

const cellWidth = 200

const axis = computed(() => buildMonthAxis(props.nodes))
const months = computed(() => axis.value.months)
const undated = computed(() => axis.value.undated)
const deps = computed(() => depsByNode(props.nodes, props.edges))

const laneStyle = computed(() => ({
  gridTemplateColumns: `repeat(${months.value.length}, ${cellWidth}px)`,
}))
</script>

<style scoped>
.rm-timeline-wrap {
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
}
.rm-timeline-scroll {
  flex: 1;
  min-height: 0;
  overflow: auto;
}
.rm-timeline-lane {
  display: grid;
  align-items: start;
}
.rm-timeline-cell {
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
}
.rm-card-btn:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
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
}
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
