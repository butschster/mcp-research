<template>
  <div class="roadmap-page">
    <!-- Toolbar -->
    <div class="roadmap-toolbar">
      <div class="toolbar-left">
        <NuxtLink :to="`/research/${researchSlug}/roadmaps`" class="btn btn-sm toolbar-back">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
          Back
        </NuxtLink>
        <span v-if="roadmap" class="toolbar-title">{{ roadmap.title }}</span>
        <span v-if="roadmap?.code" class="toolbar-code">{{ roadmap.code }}</span>
      </div>
      <div class="toolbar-right">
        <!-- Progress -->
        <div v-if="progress.total > 0" class="toolbar-progress">
          <span class="progress-text">{{ progress.completed }}/{{ progress.total }}</span>
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: progress.percent + '%' }"></div>
          </div>
        </div>

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

        <!-- Auto layout -->
        <button class="btn btn-sm" @click="onAutoLayout" title="Auto layout">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
          Auto layout
        </button>

        <!-- Fit view -->
        <button class="btn btn-sm" @click="fitAll" title="Fit view">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/></svg>
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="roadmap-loading">
      <div class="skeleton-card" style="width: 200px; height: 80px;"></div>
      <p class="card-meta mt-4">Loading roadmap...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="roadmap-loading">
      <p class="card-meta">{{ error }}</p>
      <button class="btn btn-sm mt-4" @click="refresh">Retry</button>
    </div>

    <!-- Canvas -->
    <div v-else class="roadmap-canvas">
      <VueFlow
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
        class="roadmap-flow"
        @node-click="onNodeClick"
        @node-drag-stop="onNodeDragStop"
      >
        <MiniMap
          :node-color="minimapNodeColor"
          :mask-color="'rgba(12, 18, 32, 0.7)'"
          position="bottom-right"
        />
        <Controls position="bottom-left" />
      </VueFlow>
    </div>

    <!-- Node popover -->
    <RoadmapNodePopover
      v-if="selectedNode"
      :node="selectedNode"
      :statuses="roadmap?.statuses ?? []"
      :position="popoverPosition"
      @update-status="onUpdateStatus"
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
import RoadmapNodePopover from '~/components/roadmap/RoadmapNodePopover.vue'

const route = useRoute()
const researchId = route.params.id as string
const roadmapId = route.params.roadmapId as string

// Resolve research slug for back link
const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${researchId}`)
const researchSlug = computed(() => researchData.value?.data?.research?.code || researchId)

const nodeTypes = {
  'roadmap-root': markRaw(RoadmapRootNode),
  'roadmap-step': markRaw(RoadmapStepNode),
}

const {
  roadmap,
  nodes,
  edges,
  loading,
  error,
  progress,
  refresh,
  updateNodeStatus,
  updateNodePosition,
  autoLayout,
  layoutDirection,
  setLayoutDirection,
  shouldSuppressRefresh,
} = useRoadmap(roadmapId)

// Vue Flow instance
const { fitView } = useVueFlow()

function fitAll() {
  fitView({ padding: 0.15, duration: 300 })
}

// Node click → popover
const selectedNode = ref<{
  id: string
  title: string
  description: string
  nodeType: string
  status: string
} | null>(null)
const popoverPosition = ref({ x: 0, y: 0 })

function onNodeClick({ node, event }: { node: any; event: MouseEvent }) {
  // Don't show popover for root node
  if (node.type === 'roadmap-root') return

  selectedNode.value = {
    id: node.id,
    title: node.data.title,
    description: node.data.description || '',
    nodeType: node.data.nodeType || 'step',
    status: node.data.status || '',
  }
  popoverPosition.value = { x: event.clientX + 12, y: event.clientY - 20 }
}

async function onUpdateStatus(nodeId: string, status: string) {
  await updateNodeStatus(nodeId, status)
  selectedNode.value = null
}

function onNodeDragStop({ node }: { node: any }) {
  updateNodePosition(node.id, node.position.x, node.position.y)
}

async function onAutoLayout() {
  await autoLayout()
  nextTick(() => fitView({ padding: 0.15, duration: 300 }))
}

function minimapNodeColor(node: any): string {
  if (node.type === 'roadmap-root') return '#6cc5e0'
  return '#7f8ea3'
}

// Close popover on click outside
function onDocumentClick() {
  selectedNode.value = null
}
onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  refresh()
})
onUnmounted(() => {
  document.removeEventListener('click', onDocumentClick)
})

// Real-time updates (skip refresh if we just made a local change)
useRealtimeUpdates(async (event) => {
  if (event.entity === 'roadmap' && !shouldSuppressRefresh()) {
    await refresh()
    nextTick(() => fitView({ padding: 0.15, duration: 300 }))
  }
})
</script>

<style scoped>
.roadmap-page {
  position: fixed;
  inset: 0;
  display: flex;
  flex-direction: column;
  background: var(--color-bg);
  z-index: var(--z-overlay);
}

.roadmap-toolbar {
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
  gap: var(--space-3);
}
.toolbar-back { gap: var(--space-1); }
.toolbar-title {
  font-size: var(--type-sm);
  font-weight: 600;
  color: var(--color-text);
  letter-spacing: -0.01em;
}
.toolbar-code {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
}
.toolbar-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
.toolbar-sep {
  width: 1px;
  height: 20px;
  background: var(--color-border-strong);
  margin: 0 var(--space-1);
}

.toolbar-progress {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
.progress-text {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
.progress-bar {
  width: 100px;
  height: 4px;
  background: var(--color-surface-hover);
  border-radius: 2px;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  background: rgba(107, 203, 119, 0.8);
  border-radius: 2px;
  transition: width 0.3s ease;
}

.roadmap-canvas {
  flex: 1;
  min-height: 0;
}
.roadmap-flow {
  width: 100%;
  height: 100%;
}
.roadmap-loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

/* Vue Flow overrides */
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
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
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

/* Responsive */
@media (max-width: 768px) {
  .roadmap-toolbar {
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
