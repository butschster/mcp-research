<template>
  <div class="rm-board-wrap">
    <!-- No nodes at all: the board would otherwise be a blank strip. Mirror the
         timeline's empty state so the two board modes behave the same. -->
    <EmptyState
      v-if="!nodes.length"
      icon="&#x1F5C2;"
      title="This roadmap has no nodes yet"
      description="Add nodes to organise them into stage columns, or switch to Graph."
    >
      <button class="btn btn-sm" @click="emit('switch-graph')">Switch to Graph</button>
    </EmptyState>

    <template v-else>
      <p v-if="allUnassigned" class="rm-board-banner">
        <span>Stages are set by the assistant — nothing is placed in a stage yet.</span>
        <button class="rm-board-link" @click="emit('switch-graph')">Switch to Graph</button>
      </p>
      <div class="rm-board" role="list">
      <RoadmapStageColumn
        v-for="col in columns"
        :key="col.name"
        :name="col.name"
        :nodes="col.nodes"
        :deps="deps"
        :leftover="col.leftover"
        :highlight-ids="highlightIds"
        role="listitem"
        @node-click="emit('node-click', $event)"
        @hover="onHover"
      />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { bucketByStage, depsByNode, relatedIds, type RawRoadmapNode, type RawRoadmapEdge } from '~/utils/roadmap'

const props = defineProps<{
  stages: readonly string[]
  nodes: readonly RawRoadmapNode[]
  edges: readonly RawRoadmapEdge[]
}>()

const emit = defineEmits<{ 'node-click': [string]; 'switch-graph': [] }>()

const columns = computed(() => bucketByStage(props.nodes, props.stages))
const deps = computed(() => depsByNode(props.nodes, props.edges))

// Everything landed in the leftover column: the stages exist (or don't) but no
// node is attached to one. The board still renders; a banner names the state.
const allUnassigned = computed(() =>
  props.nodes.length > 0 &&
  columns.value.every(c => c.leftover || c.nodes.length === 0),
)

// Hover/focus a card → highlight it and its graph neighbours, dim the rest.
const hoveredId = ref<string | null>(null)
const highlightIds = computed<Set<string> | null>(() => {
  if (!hoveredId.value) return null
  const set = relatedIds(hoveredId.value, props.edges)
  set.add(hoveredId.value)
  return set
})
function onHover(id: string | null) {
  hoveredId.value = id
}
</script>

<style scoped>
.rm-board-wrap {
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
}
.rm-board-banner {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  padding: var(--space-2) var(--space-5);
  border-bottom: 1px solid var(--color-border);
}
.rm-board-link {
  appearance: none;
  background: none;
  border: none;
  padding: 0;
  color: var(--color-primary);
  font: inherit;
  font-size: var(--type-xs);
  cursor: pointer;
  text-decoration: underline;
}
.rm-board-link:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; }
.rm-board {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: var(--space-4);
  padding: var(--space-5);
  overflow-x: auto;
  overflow-y: hidden;
  align-items: stretch;
  scroll-snap-type: x proximity;
}
</style>
