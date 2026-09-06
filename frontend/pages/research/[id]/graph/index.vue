<template>
  <div class="graph-page">
    <GraphSidebar
      :collapsed="sidebarCollapsed"
      :node-types="nodeTypeFilters"
      :edge-types="edgeTypeFilters"
      :visible-node-types="visibleNodeTypes"
      :visible-edge-types="visibleEdgeTypes"
      :node-count-by-type="nodeCountByType"
      :edge-count-by-type="edgeCountByType"
      :node-count="filteredNodeCount"
      :edge-count="filteredEdgeCount"
      :hide-orphans="hideOrphans"
      :show-arrows="showArrows"
      :focus-depth="focusDepth"
      :has-focus="!!focusedNodeId"
      @update:collapsed="sidebarCollapsed = $event"
      @toggle-node-type="toggleNodeType"
      @toggle-edge-type="toggleEdgeType"
      @update:hide-orphans="hideOrphans = $event"
      @update:show-arrows="showArrows = $event"
      @update:focus-depth="focusDepth = $event"
      @clear-focus="focusedNodeId = null"
    >
      <template #back>
        <NuxtLink :to="researchPath(researchSlug)" class="sidebar-back">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
          Back
        </NuxtLink>
        <span class="sidebar-title">Knowledge Graph</span>
      </template>
    </GraphSidebar>

    <GraphCanvas
      v-model:focused-node-id="focusedNodeId"
      :nodes="nodes"
      :edges="edges"
      :loading="loading"
      :visible-node-types="visibleNodeTypes"
      :visible-edge-types="visibleEdgeTypes"
      :hide-orphans="hideOrphans"
      :show-arrows="showArrows"
      :focus-depth="focusDepth"
      :aria-label="`Knowledge graph: ${filteredNodeCount} nodes, ${filteredEdgeCount} edges.`"
      @open-node="openNode"
      @counts="onCounts"
    />
  </div>
</template>

<script setup lang="ts">
import { GRAPH_NODE_TYPE_FILTERS, GRAPH_EDGE_TYPE_FILTERS } from '~/composables/useGraphFilters'
import type { GraphNode } from '~/composables/useResearchGraph'

const route = useRoute()
const researchSlug = route.params.id as string

const {
  nodes, edges, loading,
  fetchGraph,
  visibleEdgeTypes, visibleNodeTypes,
  toggleEdgeType, toggleNodeType,
} = useResearchGraph(researchSlug)

const sidebarCollapsed = ref(false)
const hideOrphans = ref(false)
const showArrows = ref(false)
const focusDepth = ref(1)
const focusedNodeId = ref<string | null>(null)

const nodeTypeFilters = GRAPH_NODE_TYPE_FILTERS
const edgeTypeFilters = GRAPH_EDGE_TYPE_FILTERS

const nodeCountByType = computed(() => {
  const counts: Record<string, number> = {}
  for (const n of nodes.value) counts[n.type] = (counts[n.type] || 0) + 1
  return counts
})

const edgeCountByType = computed(() => {
  const counts: Record<string, number> = {}
  for (const e of edges.value) counts[e.type] = (counts[e.type] || 0) + 1
  return counts
})

const filteredNodeCount = ref(0)
const filteredEdgeCount = ref(0)
function onCounts(n: number, e: number) {
  filteredNodeCount.value = n
  filteredEdgeCount.value = e
}

// Through the path helpers, so the same component under a share link lands
// inside the link rather than on the login wall.
function openNode(node: GraphNode) {
  const code = node.code || node.id
  if (node.type === 'entry') navigateTo(entryPath(researchSlug, code))
  else if (node.type === 'session') navigateTo(sessionPath(researchSlug, code))
  else if (node.type === 'task') navigateTo(tasksPath(researchSlug))
}

onMounted(() => { void fetchGraph() })

// `researchSlug` here is the raw route param — a short code from a link, or a
// UUID from a pasted address — and either matches. No `researchId` is passed
// because this page loads a graph, not the research, so it has no second
// identity to offer.
useResearchRealtime(() => researchSlug, () => void fetchGraph(), { onResync: () => void fetchGraph() })
</script>

<style scoped>
.graph-page {
  position: fixed;
  inset: 0;
  display: flex;
  background: var(--color-bg-deep);
  z-index: 100;
}

/* Styles for back link in sidebar slot */
.sidebar-back {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--color-text-muted);
  text-decoration: none;
  font-size: 12px;
}
.sidebar-back:hover { color: var(--color-text); }

.sidebar-title {
  font-size: 14px;
  font-weight: var(--weight-semibold);
  color: var(--color-text);
}
</style>
