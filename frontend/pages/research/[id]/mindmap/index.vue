<template>
  <div class="mindmap-page">
    <!-- Toolbar -->
    <div class="mindmap-toolbar">
      <div class="toolbar-left">
        <NuxtLink :to="`/research/${researchSlug}`" class="btn btn-sm toolbar-back">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
          Back
        </NuxtLink>
        <span v-if="researchName" class="toolbar-title">{{ researchName }}</span>
      </div>
      <div class="toolbar-right">
        <!-- Filter chips -->
        <button
          v-for="group in filterGroups"
          :key="group.key"
          :class="['btn btn-sm filter-chip', { active: visibleGroups.has(group.key) }]"
          @click="toggleGroup(group.key)"
        >{{ group.label }}</button>

        <button
          :class="['btn btn-sm filter-chip crossref-chip', { active: showCrossrefs }]"
          @click="toggleCrossrefs"
        >
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
          Crossrefs
        </button>

        <span class="toolbar-sep"></span>

        <!-- Layout toggle -->
        <button
          :class="['btn btn-sm', { active: layoutDirection === 'LR' }]"
          @click="setLayoutDirection('LR')"
          title="Left to right"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
        </button>
        <button
          :class="['btn btn-sm', { active: layoutDirection === 'TB' }]"
          @click="setLayoutDirection('TB')"
          title="Top to bottom"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14"/><path d="m19 12-7 7-7-7"/></svg>
        </button>

        <span class="toolbar-sep"></span>

        <!-- Expand/collapse -->
        <button class="btn btn-sm" @click="expandAll">Expand all</button>
        <button class="btn btn-sm" @click="collapseAll">Collapse</button>

        <!-- Fit view -->
        <button class="btn btn-sm" @click="fitAll" title="Fit view">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/></svg>
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="mindmap-loading">
      <div class="skeleton-card" style="width: 200px; height: 80px;"></div>
      <p class="card-meta mt-4">Loading mindmap...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="mindmap-loading">
      <p class="card-meta">{{ error }}</p>
      <button class="btn btn-sm mt-4" @click="refresh">Retry</button>
    </div>

    <!-- Canvas -->
    <div v-else class="mindmap-canvas">
      <VueFlow
        ref="vueFlowRef"
        :nodes="nodes"
        :edges="edges"
        :node-types="nodeTypes"
        :default-viewport="{ x: 0, y: 0, zoom: 0.85 }"
        :min-zoom="0.15"
        :max-zoom="2"
        :fit-view-on-init="true"
        :nodes-draggable="true"
        :nodes-connectable="false"
        :edges-updatable="false"
        :pan-on-drag="true"
        :zoom-on-scroll="true"
        class="mindmap-flow"
        @node-click="onNodeClick"
        @edge-mouse-enter="onEdgeEnter"
        @edge-mouse-leave="onEdgeLeave"
      >
        <!-- Crossref tooltip -->
        <div v-if="hoveredEdge" class="xref-tooltip" :style="tooltipStyle">
          <span class="xref-from">{{ hoveredEdge.sourceLabel }}</span>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
          <span class="xref-to">{{ hoveredEdge.targetLabel }}</span>
        </div>
        <MiniMap
          :node-color="minimapNodeColor"
          :mask-color="'rgba(12, 18, 32, 0.7)'"
          position="bottom-right"
        />
        <Controls position="bottom-left" />
      </VueFlow>
    </div>
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

import RootNode from '~/components/mindmap/RootNode.vue'
import SectionNode from '~/components/mindmap/SectionNode.vue'
import EntryNode from '~/components/mindmap/EntryNode.vue'
import GroupLabelNode from '~/components/mindmap/GroupLabelNode.vue'
import QuestionNode from '~/components/mindmap/QuestionNode.vue'
import TaskNode from '~/components/mindmap/TaskNode.vue'
import AnswerNode from '~/components/mindmap/AnswerNode.vue'

const route = useRoute()
const id = route.params.id as string

const nodeTypes = {
  root: markRaw(RootNode),
  section: markRaw(SectionNode),
  entry: markRaw(EntryNode),
  'group-label': markRaw(GroupLabelNode),
  question: markRaw(QuestionNode),
  answer: markRaw(AnswerNode),
  task: markRaw(TaskNode),
}

const {
  nodes,
  edges,
  loading,
  error,
  refresh,
  toggleCollapse,
  expandAll,
  collapseAll,
  layoutDirection,
  setLayoutDirection,
  visibleGroups,
  toggleGroup,
  showCrossrefs,
  toggleCrossrefs,
} = useResearchMindmap(id)

const filterGroups = [
  { key: 'entries', label: 'Entries' },
  { key: 'questions', label: 'Sessions' },
  { key: 'tasks', label: 'Tasks' },
]

// Research name for toolbar
const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? '')
const researchSlug = computed(() => researchData.value?.data?.research?.code || id)

// Vue Flow instance
const vueFlowRef = ref()
const { fitView } = useVueFlow()

function fitAll() {
  fitView({ padding: 0.15, duration: 300 })
}

function onNodeClick({ node }: { node: any }) {
  if (node.type === 'section' || node.type === 'group-label') {
    toggleCollapse(node.id)
    nextTick(() => fitView({ padding: 0.15, duration: 300 }))
  }
}

// Crossref edge hover
const hoveredEdge = ref<{ sourceLabel: string; targetLabel: string } | null>(null)
const tooltipPos = ref({ x: 0, y: 0 })
const tooltipStyle = computed(() => ({
  left: `${tooltipPos.value.x}px`,
  top: `${tooltipPos.value.y}px`,
}))

const { getNodes, getEdges, setEdges, setNodes } = useVueFlow()

function onEdgeEnter({ edge, event }: { edge: any; event: MouseEvent }) {
  if (!edge.id.startsWith('xref-')) return

  // Find source and target node labels
  const sourceNode = getNodes.value.find((n: any) => n.id === edge.source)
  const targetNode = getNodes.value.find((n: any) => n.id === edge.target)
  const sourceLabel = sourceNode?.data?.entrySlug
    ? `${sourceNode.data.entrySlug} ${sourceNode.data.title}`
    : sourceNode?.data?.title ?? edge.source
  const targetLabel = targetNode?.data?.entrySlug
    ? `${targetNode.data.entrySlug} ${targetNode.data.title}`
    : targetNode?.data?.title ?? edge.target

  hoveredEdge.value = { sourceLabel, targetLabel }
  tooltipPos.value = { x: event.clientX + 12, y: event.clientY - 30 }

  // Highlight the edge and connected nodes
  setEdges(getEdges.value.map((e: any) => ({
    ...e,
    style: e.id === edge.id
      ? { ...e.style, stroke: '#a78bfa', strokeWidth: 2.5, strokeDasharray: '4 4' }
      : e.id.startsWith('xref-')
        ? { ...e.style, stroke: 'rgba(167,139,250,0.12)' }
        : e.style,
  })))

  setNodes(getNodes.value.map((n: any) => ({
    ...n,
    class: n.id === edge.source || n.id === edge.target ? 'xref-highlight' : '',
  })))
}

function onEdgeLeave({ edge }: { edge: any }) {
  if (!edge.id.startsWith('xref-')) return
  hoveredEdge.value = null

  // Reset styles
  setEdges(getEdges.value.map((e: any) => ({
    ...e,
    style: e.id.startsWith('xref-')
      ? { stroke: 'rgba(167,139,250,0.35)', strokeWidth: 1, strokeDasharray: '4 4' }
      : e.style,
  })))

  setNodes(getNodes.value.map((n: any) => ({
    ...n,
    class: '',
  })))
}

function minimapNodeColor(node: any): string {
  switch (node.type) {
    case 'root': return '#6cc5e0'
    case 'section': return '#6b9df0'
    case 'entry': return '#7f8ea3'
    case 'group-label': return '#f0b849'
    case 'question': return '#f0b849'
    case 'answer': return '#34d399'
    case 'task': return '#ef6b6b'
    default: return '#7f8ea3'
  }
}

// Initial load
onMounted(() => {
  refresh()
})

// Real-time updates
//
// The refit is deliberately not repeated. Fitting the whole map on every event
// discards the reader's pan and zoom, and an agent at work emits one every few
// seconds — so the page fought whoever was reading it, precisely when it was
// most worth reading.
async function reloadMindmap() {
  await refresh()
}

useResearchRealtime(() => id, reloadMindmap, {
  researchId: () => researchData.value?.data?.research?.id,
  onResync: reloadMindmap,
})
</script>

<style scoped>
.mindmap-page {
  position: fixed;
  inset: 0;
  display: flex;
  flex-direction: column;
  background: var(--color-bg);
  z-index: var(--z-overlay);
}

.mindmap-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3) var(--space-5);
  background: rgba(21, 29, 46, 0.9);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--color-border);
  gap: var(--space-4);
  flex-shrink: 0;
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}
.toolbar-back { gap: var(--space-1); }
.toolbar-title {
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
  letter-spacing: -0.01em;
}
.toolbar-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}
.toolbar-sep {
  width: 1px;
  height: 20px;
  background: var(--color-border-strong);
  margin: 0 var(--space-1);
}

.filter-chip {
  color: var(--color-text-muted);
  border-color: var(--color-border);
}
.filter-chip.active {
  color: var(--color-primary);
  border-color: rgba(108, 197, 224, 0.3);
  background: var(--color-primary-muted);
}
.crossref-chip.active {
  color: var(--hue-5);
  border-color: rgba(167, 139, 250, 0.3);
  background: rgba(167, 139, 250, 0.1);
}

.mindmap-canvas {
  flex: 1;
  min-height: 0;
}

.mindmap-flow {
  width: 100%;
  height: 100%;
}

.mindmap-loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

/* Crossref tooltip */
.xref-tooltip {
  position: fixed;
  z-index: 1000;
  background: var(--color-surface);
  border: 1px solid rgba(167, 139, 250, 0.3);
  border-radius: var(--radius-sm);
  padding: var(--space-2) var(--space-3);
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-xs);
  color: var(--color-text);
  box-shadow: var(--shadow-2);
  pointer-events: none;
  white-space: nowrap;
}
.xref-tooltip svg { color: rgba(167, 139, 250, 0.6); flex-shrink: 0; }
.xref-from, .xref-to {
  font-weight: var(--weight-semibold);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Highlighted nodes on crossref hover */
.mindmap-flow :deep(.xref-highlight) {
  outline: 2px solid rgba(167, 139, 250, 0.6);
  outline-offset: 2px;
  border-radius: var(--radius);
}

/* Vue Flow theme overrides */
.mindmap-flow :deep(.vue-flow__edge-path) {
  stroke-linecap: round;
}
.mindmap-flow :deep(.vue-flow__background) {
  background: var(--color-bg);
}
.mindmap-flow :deep(.vue-flow__minimap) {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}
.mindmap-flow :deep(.vue-flow__controls) {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-1);
}
.mindmap-flow :deep(.vue-flow__controls-button) {
  background: var(--color-surface);
  border-color: var(--color-border);
  color: var(--color-text-muted);
  fill: var(--color-text-muted);
}
.mindmap-flow :deep(.vue-flow__controls-button:hover) {
  background: var(--color-surface-hover);
}

/* Responsive */
@media (max-width: 768px) {
  .mindmap-toolbar {
    flex-wrap: wrap;
    padding: var(--space-2) var(--space-3);
    gap: var(--space-2);
  }
  .toolbar-right {
    flex-wrap: wrap;
    gap: var(--space-1);
  }
  .toolbar-title { display: none; }
  .toolbar-sep { display: none; }
}
</style>
