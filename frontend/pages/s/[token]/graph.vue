<template>
  <div class="share-graph-page">
    <div class="view-toolbar">
      <div class="toolbar-left">
        <NuxtLink :to="researchPath(slug)" class="btn btn-sm toolbar-back">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
          Back to project
        </NuxtLink>
        <span class="toolbar-title">Knowledge graph</span>
        <span v-if="researchCode" class="toolbar-code">{{ researchCode }}</span>
      </div>
      <div class="toolbar-right">
        <button v-if="view === 'graph'" type="button" class="btn btn-sm" title="Fit view" aria-label="Fit view" @click="canvas?.fit()">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/></svg>
        </button>
        <SegmentedToggle
          v-model="view"
          :options="[{ value: 'graph', label: 'Graph' }, { value: 'list', label: 'List' }]"
          label="Graph view"
        />
      </div>
    </div>

    <p v-if="largeNotice" class="card-meta large-notice" role="status">
      Large project — showing documents and sections. Add more from the panel.
      <button type="button" class="link-btn" @click="largeNotice = false">Dismiss</button>
    </p>

    <!-- Inside the shell rather than a fixed takeover: a takeover would cover
         the banner, which is the only thing telling a visitor where they are
         and, after this change, where the theme toggle lives. The shared
         roadmap page set the precedent and the numbers. -->
    <div class="view-panel" :class="{ 'view-panel--list': view === 'list' }">
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
          <span class="sidebar-title">Filters</span>
        </template>
      </GraphSidebar>

      <!-- Errors first, then the two empties, then the views. A 404 is not an
           error here: it means the whole link has gone, and the shell owns
           that screen. -->
      <div v-if="!loading && error" class="view-state">
        <EmptyState
          title="Couldn't draw the graph"
          description="The server didn't answer. The link is fine — try again in a moment."
        >
          <button class="btn btn-primary" @click="fetchGraph()">Try again</button>
        </EmptyState>
      </div>

      <div v-else-if="!loading && isEmptyProject" class="view-state">
        <EmptyState
          title="Nothing to connect yet"
          description="This project has one document and no cross-references between documents, so there is no graph to draw. It fills in as the project grows — you can leave this page open."
        >
          <NuxtLink class="btn btn-primary" :to="researchPath(slug)">Back to project</NuxtLink>
        </EmptyState>
      </div>

      <div v-else-if="!loading && nothingVisible" class="view-state">
        <EmptyState
          title="Nothing matches these filters"
          description="Every node type is switched off. Turn one back on in the panel."
        >
          <button class="btn btn-primary" @click="showEverything">Show all node types</button>
        </EmptyState>
      </div>

      <GraphNodeList
        v-else-if="view === 'list' && !loading"
        :nodes="nodes"
        :edges="edges"
        :node-types="nodeTypeFilters"
        :visible-node-types="visibleNodeTypes"
        :href-for="hrefFor"
      />

      <GraphCanvas
        v-else
        ref="canvas"
        v-model:focused-node-id="focusedNodeId"
        :nodes="nodes"
        :edges="edges"
        :loading="loading"
        :visible-node-types="visibleNodeTypes"
        :visible-edge-types="visibleEdgeTypes"
        :hide-orphans="hideOrphans"
        :show-arrows="showArrows"
        :focus-depth="focusDepth"
        :aria-label="`Knowledge graph: ${filteredNodeCount} nodes, ${filteredEdgeCount} edges. Switch to List view to read them.`"
        @open-node="openNode"
        @counts="onCounts"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * The shared knowledge graph.
 *
 * Same canvas as the owner's, fed through `shareFetch` so the owner opening
 * their own link sees what a visitor sees, and every node leads only to a page
 * that exists inside the link. The types a link withholds never reach this
 * page — the server leaves them out of the payload and out of
 * `available_node_types` — so the sidebar has no row for them, rather than a
 * disabled one that would say "there is something here you cannot see".
 */
import { GRAPH_NODE_TYPE_FILTERS, GRAPH_EDGE_TYPE_FILTERS } from '~/composables/useGraphFilters'
import type { GraphNode } from '~/composables/useResearchGraph'

const { shareFetch, researchId, researchCode, include, slug, markGone } = useShare()

const {
  nodes, edges, loading, error, errorStatus, availableNodeTypes,
  fetchGraph,
  visibleEdgeTypes, visibleNodeTypes,
  toggleEdgeType, toggleNodeType, showAllNodeTypes,
} = useResearchGraph(researchId.value, { fetcher: (path) => shareFetch(path) })

// A 404 under the prefix is the link, not the graph: revoked or expired since
// the page opened. Hand it to the shell, which owns the one dead screen.
watch(errorStatus, (status) => { if (status === 404) markGone() })

const nodeTypeFilters = computed(() =>
  GRAPH_NODE_TYPE_FILTERS.filter(nt => availableNodeTypes.value.includes(nt.key)),
)
const edgeTypeFilters = GRAPH_EDGE_TYPE_FILTERS

// The phone is where a canvas is least useful and a list most; and a 240px
// panel inside a 375px viewport leaves 100px of graph.
const narrow = typeof window !== 'undefined' && window.matchMedia('(max-width: 768px)').matches
const view = ref<'graph' | 'list'>(narrow ? 'list' : 'graph')
const sidebarCollapsed = ref(narrow)
const hideOrphans = ref(false)
const showArrows = ref(false)
const focusDepth = ref(1)
const focusedNodeId = ref<string | null>(null)
const largeNotice = ref(false)
const canvas = ref<{ fit: () => void } | null>(null)

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

// One document and nothing to connect it to is not a graph; distinct from
// "you switched everything off", which is the reader's own doing.
const isEmptyProject = computed(() =>
  nodes.value.filter(n => n.type !== 'section').length <= 1
  && !edges.value.some(e => e.type === 'crossref'),
)
const nothingVisible = computed(() =>
  nodes.value.length > 0 && !nodes.value.some(n => visibleNodeTypes.value.has(n.type)),
)

function showEverything() {
  showAllNodeTypes()
  // The button names a panel the reader cannot see while it is collapsed.
  sidebarCollapsed.value = false
}

/**
 * Where a node leads. The same rules `linkRefs` follows, and it must stay
 * that way or the two surfaces disagree about what the link exposes: a
 * document always; a session only when the link includes them (otherwise the
 * node is not here to be clicked); a task to the board, likewise; a question
 * nowhere, because there is no shared question page; a section nowhere.
 */
function hrefFor(node: GraphNode): string {
  const code = node.code || node.id
  if (node.type === 'entry') return entryPath(slug.value, code)
  if (node.type === 'session' && include.value.sessions) return sessionPath(slug.value, code)
  if (node.type === 'task' && include.value.tasks) return tasksPath(slug.value)
  return ''
}

function openNode(node: GraphNode) {
  const href = hrefFor(node)
  if (href) navigateTo(href)
}

async function load() {
  await fetchGraph()
  // Above this the simulation freezes a phone; start narrow and say so.
  if (nodes.value.length > 600) {
    visibleNodeTypes.value = new Set(['entry', 'section'])
    hideOrphans.value = true
    largeNotice.value = true
    return
  }
  // A visitor starts with everything the link carries switched on. The
  // owner's default leaves sessions off — noise on a busy workspace — but a
  // link was made to show something, and a row unticked by default is a row
  // most people never tick.
  visibleNodeTypes.value = new Set(availableNodeTypes.value)
}

onMounted(() => { void load() })

useResearchRealtime(
  () => slug.value,
  () => void fetchGraph(),
  { onResync: () => void fetchGraph(), researchId: () => researchId.value },
)

useHead({ title: () => (researchCode.value ? `${researchCode.value} — graph` : 'Knowledge graph') })
</script>

<style scoped>
.share-graph-page { display: flex; flex-direction: column; }
.view-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
  margin-bottom: var(--space-4);
}
.toolbar-left, .toolbar-right { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }
.toolbar-title { font-size: var(--type-sm); font-weight: var(--weight-semibold); overflow-wrap: anywhere; }
.toolbar-code { font-family: 'JetBrains Mono', monospace; font-size: var(--type-xs); color: var(--color-text-muted); }
.large-notice { margin: 0 0 var(--space-3); display: flex; gap: var(--space-3); flex-wrap: wrap; }

/* The shared roadmap page's panel, verbatim, so the shared canvases are one
   family: a definite height, a border, and its own scrolling. */
.view-panel {
  display: flex;
  height: 70vh;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--color-bg-deep);
}
.view-panel--list { background: var(--color-surface); }
.view-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
  min-width: 0;
}
.sidebar-title {
  font-size: 14px;
  font-weight: var(--weight-semibold);
  color: var(--color-text);
}

@media (max-width: 768px) {
  .view-panel { height: 60vh; }
}
</style>
