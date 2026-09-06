<template>
  <VueFlow
    :nodes="nodes"
    :edges="edges"
    :node-types="nodeTypes"
    :default-viewport="{ x: 0, y: 0, zoom: 0.85 }"
    :min-zoom="0.15"
    :max-zoom="2"
    :fit-view-on-init="true"
    :nodes-draggable="draggable"
    :nodes-connectable="false"
    :edges-updatable="false"
    :pan-on-drag="true"
    :zoom-on-scroll="true"
    class="mindmap-flow"
    @node-click="emit('node-click', $event)"
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
      :mask-color="'var(--color-nav)'"
      position="bottom-right"
    />
    <Controls position="bottom-left" />
  </VueFlow>
</template>

<script setup lang="ts">
/**
 * The mind map's canvas: Vue Flow with the product's node types, the
 * cross-reference hover, the minimap and the controls, painted in the
 * product's palette rather than the library's.
 *
 * The owner's page and the shared page both render it. What differs — where
 * the data comes from, which toolbar sits above it, whether it takes over the
 * screen — stays in the pages.
 */
import { VueFlow, useVueFlow, type Node, type Edge, type NodeTypesObject, type EdgeMouseEvent } from '@vue-flow/core'
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

withDefaults(defineProps<{
  nodes: Node[]
  edges: Edge[]
  /** Whether a reader may drag nodes around. Local to the browser either way. */
  draggable?: boolean
}>(), { draggable: true })

const emit = defineEmits<{ 'node-click': [payload: { node: any }] }>()

// Cast once: Vue Flow types node components loosely, and every SFC here has
// its own props type that the library's `NodeTypesObject` does not model.
const nodeTypes = {
  root: markRaw(RootNode),
  section: markRaw(SectionNode),
  entry: markRaw(EntryNode),
  'group-label': markRaw(GroupLabelNode),
  question: markRaw(QuestionNode),
  answer: markRaw(AnswerNode),
  task: markRaw(TaskNode),
} as unknown as NodeTypesObject

const { fitView, getNodes, getEdges, setEdges, setNodes } = useVueFlow()

function fitAll() {
  fitView({ padding: 0.15, duration: 300 })
}

// Crossref edge hover
const hoveredEdge = ref<{ sourceLabel: string; targetLabel: string } | null>(null)
const tooltipPos = ref({ x: 0, y: 0 })
const tooltipStyle = computed(() => ({
  left: `${tooltipPos.value.x}px`,
  top: `${tooltipPos.value.y}px`,
}))

function onEdgeEnter({ edge, event }: EdgeMouseEvent) {
  if (!edge.id.startsWith('xref-')) return
  const pointer = event as MouseEvent

  const sourceNode = getNodes.value.find((n: any) => n.id === edge.source)
  const targetNode = getNodes.value.find((n: any) => n.id === edge.target)
  const sourceLabel = sourceNode?.data?.entrySlug
    ? `${sourceNode.data.entrySlug} ${sourceNode.data.title}`
    : sourceNode?.data?.title ?? edge.source
  const targetLabel = targetNode?.data?.entrySlug
    ? `${targetNode.data.entrySlug} ${targetNode.data.title}`
    : targetNode?.data?.title ?? edge.target

  hoveredEdge.value = { sourceLabel, targetLabel }
  tooltipPos.value = { x: pointer.clientX + 12, y: pointer.clientY - 30 }

  setEdges(getEdges.value.map((e: any) => ({
    ...e,
    style: e.id === edge.id
      ? { ...e.style, stroke: 'var(--hue-5)', strokeWidth: 2.5, strokeDasharray: '4 4' }
      : e.id.startsWith('xref-')
        ? { ...e.style, stroke: 'rgba(var(--hue-5-rgb), 0.12)' }
        : e.style,
  })))

  setNodes(getNodes.value.map((n: any) => ({
    ...n,
    class: n.id === edge.source || n.id === edge.target ? 'xref-highlight' : '',
  })))
}

function onEdgeLeave({ edge }: EdgeMouseEvent) {
  if (!edge.id.startsWith('xref-')) return
  hoveredEdge.value = null

  setEdges(getEdges.value.map((e: any) => ({
    ...e,
    style: e.id.startsWith('xref-')
      ? { stroke: 'rgba(var(--hue-5-rgb), 0.35)', strokeWidth: 1, strokeDasharray: '4 4' }
      : e.style,
  })))

  setNodes(getNodes.value.map((n: any) => ({
    ...n,
    class: '',
  })))
}

function minimapNodeColor(node: any): string {
  switch (node.type) {
    case 'root': return 'var(--color-primary)'
    case 'section': return 'var(--color-info)'
    case 'entry': return 'var(--color-text-muted)'
    case 'group-label': return 'var(--color-warning)'
    case 'question': return 'var(--color-warning)'
    case 'answer': return 'var(--color-success)'
    case 'task': return 'var(--color-error)'
    default: return 'var(--color-text-muted)'
  }
}

defineExpose({ fitAll })
</script>

<style scoped>
.mindmap-flow {
  width: 100%;
  height: 100%;
}

/* Crossref tooltip */
.xref-tooltip {
  position: fixed;
  z-index: var(--z-toast);
  background: var(--color-surface);
  border: 1px solid rgba(var(--hue-5-rgb), 0.3);
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
.xref-tooltip svg { color: rgba(var(--hue-5-rgb), 0.6); flex-shrink: 0; }
.xref-from, .xref-to {
  font-weight: var(--weight-semibold);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Highlighted nodes on crossref hover */
.mindmap-flow :deep(.xref-highlight) {
  outline: 2px solid rgba(var(--hue-5-rgb), 0.6);
  outline-offset: 2px;
  border-radius: var(--radius);
}

/* Vue Flow renders its chrome in the library's own light theme unless told
   otherwise. One copy of the overrides, here, for both pages. */
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
</style>
