<template>
  <div class="mindmap-page">
    <!-- Toolbar -->
    <div class="mindmap-toolbar">
      <div class="toolbar-left">
        <NuxtLink :to="`/research/${id}`" class="btn btn-sm toolbar-back">
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
      >
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

const route = useRoute()
const id = route.params.id as string

const nodeTypes = {
  root: markRaw(RootNode),
  section: markRaw(SectionNode),
  entry: markRaw(EntryNode),
  'group-label': markRaw(GroupLabelNode),
  question: markRaw(QuestionNode),
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
} = useResearchMindmap(id)

const filterGroups = [
  { key: 'entries', label: 'Entries' },
  { key: 'questions', label: 'Questions' },
  { key: 'tasks', label: 'Tasks' },
]

// Research name for toolbar
const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? '')

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

function minimapNodeColor(node: any): string {
  switch (node.type) {
    case 'root': return '#6cc5e0'
    case 'section': return '#6b9df0'
    case 'entry': return '#7f8ea3'
    case 'group-label': return '#f0b849'
    case 'question': return '#f0b849'
    case 'task': return '#ef6b6b'
    default: return '#7f8ea3'
  }
}

// Initial load
onMounted(() => {
  refresh()
})

// Real-time updates
useRealtimeUpdates(async (event) => {
  if (event.research_id && event.research_id !== id) return
  await refresh()
  nextTick(() => fitView({ padding: 0.15, duration: 300 }))
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
  font-weight: 600;
  color: var(--color-text);
  letter-spacing: -0.01em;
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

.filter-chip {
  color: var(--color-text-muted);
  border-color: var(--color-border);
}
.filter-chip.active {
  color: var(--color-primary);
  border-color: rgba(108, 197, 224, 0.3);
  background: var(--color-primary-muted);
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
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
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
</style>
