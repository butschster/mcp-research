<template>
  <EmptyState
    v-if="excluded"
    title="Not part of this link"
    description="The person who shared this research didn't include roadmaps. Ask them if you need them."
  >
    <NuxtLink class="btn btn-primary" :to="researchPath(slug)">Back to the research</NuxtLink>
  </EmptyState>

  <div v-else class="roadmap-page">
    <div class="roadmap-toolbar">
      <div class="toolbar-left">
        <NuxtLink :to="roadmapsPath(slug)" class="btn btn-sm toolbar-back">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
          Back
        </NuxtLink>
        <span v-if="roadmap" class="toolbar-title">{{ roadmap.title }}</span>
        <span v-if="roadmap?.code" class="toolbar-code">{{ roadmap.code }}</span>
      </div>
      <div class="toolbar-right">
        <RoadmapViewToggle v-model="view" />
        <div v-if="progress.total > 0" class="toolbar-progress">
          <span class="progress-text">{{ progress.completed }}/{{ progress.total }}</span>
        </div>
        <template v-if="view === 'graph'">
        <button
          :class="['btn btn-sm', { active: layoutDirection === 'LR' }]"
          title="Left to right"
          @click="setLayoutDirection('LR')"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
        </button>
        <button
          :class="['btn btn-sm', { active: layoutDirection === 'TB' }]"
          title="Top to bottom"
          @click="setLayoutDirection('TB')"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14"/><path d="m19 12-7 7-7-7"/></svg>
        </button>
        <button class="btn btn-sm" title="Fit view" @click="fitAll">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/></svg>
        </button>
        </template>
      </div>
    </div>

    <div v-if="loading" class="roadmap-loading">
      <div class="skeleton-card" style="width: 200px; height: 80px;"></div>
      <p class="card-meta mt-4">Loading roadmap…</p>
    </div>

    <div v-else-if="error" class="roadmap-loading">
      <p class="card-meta">This roadmap could not be loaded.</p>
      <button class="btn btn-sm mt-4" @click="refresh()">Try again</button>
    </div>

    <div v-else class="roadmap-canvas">
      <RoadmapStagesBoard
        v-if="view === 'stages'"
        :stages="roadmap?.stages ?? []"
        :nodes="rawNodes"
        :edges="rawEdges"
        @node-click="openNode"
        @switch-graph="view = 'graph'"
      />
      <RoadmapTimeline
        v-else-if="view === 'timeline'"
        :nodes="rawNodes"
        :edges="rawEdges"
        @node-click="openNode"
        @switch-graph="view = 'graph'"
      />
      <VueFlow
        v-else
        :nodes="nodes"
        :edges="edges"
        :node-types="nodeTypes"
        :default-viewport="{ x: 0, y: 0, zoom: 0.85 }"
        :min-zoom="0.15"
        :max-zoom="2"
        :fit-view-on-init="true"
        :nodes-draggable="false"
        :nodes-connectable="false"
        :edges-updatable="false"
        :pan-on-drag="true"
        :zoom-on-scroll="true"
        class="roadmap-flow"
        @node-click="onNodeClick"
        @pane-click="selectedNode = null"
      >
        <MiniMap :node-color="minimapNodeColor" mask-color="rgba(12, 18, 32, 0.7)" position="bottom-right" />
        <Controls position="bottom-left" />
      </VueFlow>
    </div>

    <!-- The popover gates its status chips on canWrite, which the share lock
         holds false. Navigation out of it goes through the path helpers, so a
         node pointing at an entry lands inside the shared view. -->
    <RoadmapNodePopover
      :node="selectedNode"
      :statuses="[...(roadmap?.statuses ?? [])]"
      :stages="[...(roadmap?.stages ?? [])]"
      @navigate="onNavigate"
      @close="selectedNode = null"
    />
  </div>
</template>

<script setup lang="ts">
import { VueFlow, useVueFlow } from '@vue-flow/core'
import { MiniMap } from '@vue-flow/minimap'
import { Controls } from '@vue-flow/controls'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/minimap/dist/style.css'
import '@vue-flow/controls/dist/style.css'

import RoadmapRootNode from '~/components/roadmap/RoadmapRootNode.vue'
import RoadmapStepNode from '~/components/roadmap/RoadmapStepNode.vue'
import RoadmapRefNode from '~/components/roadmap/RoadmapRefNode.vue'
import RoadmapNodePopover from '~/components/roadmap/RoadmapNodePopover.vue'
import RoadmapViewToggle, { type RoadmapView } from '~/components/roadmap/RoadmapViewToggle.vue'
import RoadmapStagesBoard from '~/components/roadmap/RoadmapStagesBoard.vue'
import RoadmapTimeline from '~/components/roadmap/RoadmapTimeline.vue'

const route = useRoute()
const roadmapId = route.params.roadmapId as string

const { researchId, researchCode, include, slug } = useShare()

// Without this the page rendered "This roadmap could not be loaded" and a Try
// again button that can never succeed, wrapped in a full toolbar for a graph
// that will never exist.
const excluded = computed(() => !include.value.roadmaps)

const nodeTypes = {
  'roadmap-root': markRaw(RoadmapRootNode),
  'roadmap-step': markRaw(RoadmapStepNode),
  'roadmap-ref': markRaw(RoadmapRefNode),
}

const { roadmap, nodes, edges, loading, error, progress, refresh, layoutDirection, setLayoutDirection } =
  useRoadmap(researchId.value, roadmapId)

const { fitView } = useVueFlow()
function fitAll() { fitView({ padding: 0.15, duration: 300 }) }

const view = ref<RoadmapView>('graph')
let viewInitialised = false
watch(roadmap, (rm) => {
  if (rm && !viewInitialised) {
    view.value = (rm.view as RoadmapView) || 'graph'
    viewInitialised = true
  }
}, { immediate: true })
const rawNodes = computed(() => roadmap.value?.nodes ?? [])
const rawEdges = computed(() => roadmap.value?.edges ?? [])

const selectedNode = ref<any | null>(null)

// Open a node by id from any view, reading the raw node (stage/date included).
function openNode(nodeId: string) {
  const n = roadmap.value?.nodes.find(x => x.id === nodeId)
  if (!n) return
  selectedNode.value = {
    id: n.id, title: n.title, description: n.description || '',
    nodeType: n.node_type || 'step', status: n.status || '',
    refType: n.ref_type, refId: n.ref_id, refData: n.ref_data,
    stage: n.stage || '', node_date: n.node_date || '', node_end_date: n.node_end_date || '',
  }
}

function onNodeClick({ node }: { node: any }) {
  if (node.type === 'roadmap-root') return
  openNode(node.id)
}

/**
 * A node pointing at something else in the research.
 *
 * Only the destinations that exist inside this link are offered. A reference to
 * a question has no page here, and one to another research has no page at all —
 * both close the popover and go nowhere rather than opening a tab on the login
 * screen.
 */
function onNavigate(node: any) {
  const { refType, refId } = node
  selectedNode.value = null
  if (!refType || !refId) return
  const refResearch = node.refData?.research_id
  if (refResearch && refResearch !== researchId.value) return

  const include = shareInclude()
  let path = ''
  if (refType === 'entry') path = entryPath(slug.value, refId)
  else if (refType === 'session' && include.sessions) path = sessionPath(slug.value, refId)
  else if (refType === 'task' && include.tasks) path = tasksPath(slug.value)
  else if (refType === 'research') path = researchPath(slug.value)
  if (path) navigateTo(path)
}

function minimapNodeColor(node: any): string {
  if (node.type === 'roadmap-root') return '#6cc5e0'
  if (node.type === 'roadmap-ref') return '#a882ff'
  return '#7f8ea3'
}

onMounted(() => { if (!excluded.value) refresh() })

useResearchRealtime(
  () => slug.value,
  (event) => { if (event.entity === 'roadmap' || event.entity === 'roadmap_node') void refresh(true) },
  { onResync: () => void refresh(true), researchId: () => researchId.value },
)
</script>

<style scoped>
.roadmap-page { display: flex; flex-direction: column; min-height: 70vh; }
.roadmap-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
  margin-bottom: var(--space-4);
}
.toolbar-left, .toolbar-right { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }
.toolbar-title { font-weight: var(--weight-semibold); overflow-wrap: anywhere; }
.toolbar-code { font-family: 'JetBrains Mono', monospace; font-size: var(--type-xs); color: var(--color-text-muted); }
.toolbar-progress { font-size: var(--type-xs); color: var(--color-text-muted); }
/* A definite height, not min-height: the stages board and timeline fill their
   ancestor and scroll internally, which needs the ancestor to resolve to a real
   height (the owner page gets that from position:fixed; here it is explicit). */
.roadmap-canvas { flex: 1; height: 70vh; border: 1px solid var(--color-border); border-radius: var(--radius); overflow: hidden; }
.roadmap-flow { height: 100%; min-height: 60vh; }
.roadmap-loading { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 40vh; }

/* The Vue Flow chrome renders in the library's own light theme unless it is
   told otherwise. The owner's roadmap page carries these; the shared one
   rendered a white minimap and white controls on the dark app until it did too.
   The duplication is the argument for a RoadmapCanvas component — noted, and
   larger than this branch. */
.roadmap-flow :deep(.vue-flow__edge-path) {
  stroke-linecap: round;
}
.roadmap-flow :deep(.vue-flow__background) {
  background: var(--color-bg);
}
.roadmap-flow :deep(.vue-flow__minimap) {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}
.roadmap-flow :deep(.vue-flow__controls) {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-1);
}
.roadmap-flow :deep(.vue-flow__controls-button) {
  background: var(--color-surface);
  border-color: var(--color-border);
  color: var(--color-text-muted);
  fill: var(--color-text-muted);
}
.roadmap-flow :deep(.vue-flow__controls-button:hover) {
  background: var(--color-surface-hover);
}

</style>
