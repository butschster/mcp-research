<template>
  <div class="graph-page">
    <!-- Sidebar -->
    <aside :class="['graph-sidebar', { collapsed: sidebarCollapsed }]">
      <button class="sidebar-toggle" @click="sidebarCollapsed = !sidebarCollapsed" :title="sidebarCollapsed ? 'Open panel' : 'Close panel'">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path v-if="sidebarCollapsed" d="m9 18 6-6-6-6"/>
          <path v-else d="m15 18-6-6 6-6"/>
        </svg>
      </button>

      <template v-if="!sidebarCollapsed">
        <div class="sidebar-header">
          <NuxtLink :to="`/research/${researchSlug}`" class="sidebar-back">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
            Back
          </NuxtLink>
          <span class="sidebar-title">Knowledge Graph</span>
        </div>

        <!-- Stats -->
        <div class="sidebar-section">
          <div class="sidebar-stats">
            <span>{{ filteredNodeCount }} nodes</span>
            <span>{{ filteredEdgeCount }} edges</span>
          </div>
        </div>

        <!-- Nodes -->
        <div class="sidebar-section">
          <div class="sidebar-section-title">Nodes</div>
          <label
            v-for="nt in nodeTypeFilters"
            :key="nt.key"
            class="sidebar-check"
          >
            <input
              type="checkbox"
              :checked="visibleNodeTypes.has(nt.key)"
              @change="toggleNodeType(nt.key)"
            />
            <span class="check-dot" :style="{ background: nt.color }"></span>
            <span class="check-label">{{ nt.label }}</span>
            <span class="check-count">{{ nodeCountByType[nt.key] || 0 }}</span>
          </label>
        </div>

        <!-- Edges -->
        <div class="sidebar-section">
          <div class="sidebar-section-title">Edges</div>
          <label
            v-for="et in edgeTypeFilters"
            :key="et.key"
            class="sidebar-check"
          >
            <input
              type="checkbox"
              :checked="visibleEdgeTypes.has(et.key)"
              @change="toggleEdgeType(et.key)"
            />
            <span class="check-label">{{ et.label }}</span>
            <span class="check-count">{{ edgeCountByType[et.key] || 0 }}</span>
          </label>
        </div>

        <!-- Filters -->
        <div class="sidebar-section">
          <div class="sidebar-section-title">Filters</div>
          <label class="sidebar-check">
            <input type="checkbox" v-model="hideOrphans" />
            <span class="check-label">Hide orphan nodes</span>
          </label>
          <label class="sidebar-check">
            <input type="checkbox" v-model="showArrows" />
            <span class="check-label">Show edge direction</span>
          </label>
        </div>

        <!-- Focus depth -->
        <div class="sidebar-section">
          <div class="sidebar-section-title">
            Focus depth
            <span class="depth-value">{{ focusDepth }}</span>
          </div>
          <input
            type="range"
            min="1"
            max="5"
            v-model.number="focusDepth"
            class="depth-slider"
          />
          <div class="depth-labels">
            <span>1</span>
            <span>2</span>
            <span>3</span>
            <span>4</span>
            <span>5</span>
          </div>
          <div class="sidebar-hint">Right-click a node to focus</div>
          <button
            v-if="focusedNodeId"
            class="btn-clear-focus"
            @click="clearFocus"
          >Clear focus</button>
        </div>
      </template>
    </aside>

    <!-- Main area -->
    <div class="graph-main">
      <!-- Loading -->
      <div v-if="loading" class="graph-loading">
        <div class="spinner"></div>
        Loading graph...
      </div>

      <!-- Canvas -->
      <canvas
        v-show="!loading"
        ref="canvasRef"
        class="graph-canvas"
        @mousedown="onMouseDown"
        @mousemove="onMouseMove"
        @mouseup="onMouseUp"
        @mouseleave="onMouseUp"
        @wheel="onWheel"
        @dblclick="onDblClick"
        @contextmenu.prevent="onContextMenu"
      ></canvas>

      <!-- Tooltip -->
      <div v-if="tooltip" class="graph-tooltip" :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }">
        <span class="tooltip-type" :style="{ color: getNodeColor(tooltip.type) }">{{ tooltip.type }}</span>
        <span class="tooltip-code">{{ tooltip.code }}</span>
        <span class="tooltip-label">{{ tooltip.label }}</span>
        <span v-if="tooltip.status" class="tooltip-status">{{ tooltip.status }}</span>
        <span v-if="tooltip.connections !== undefined" class="tooltip-connections">{{ tooltip.connections }} connections</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import * as d3Force from 'd3-force'

const route = useRoute()
const researchSlug = route.params.id as string

const {
  nodes, edges, loading, error,
  fetchGraph,
  visibleEdgeTypes, visibleNodeTypes,
  toggleEdgeType, toggleNodeType,
} = useResearchGraph(researchSlug)

const canvasRef = ref<HTMLCanvasElement | null>(null)
const sidebarCollapsed = ref(false)
const hideOrphans = ref(false)
const showArrows = ref(false)
const focusDepth = ref(1)

const nodeTypeFilters = [
  { key: 'entry', label: 'Entries', color: '#6cc5e0' },
  { key: 'section', label: 'Sections', color: '#a78bfa' },
  { key: 'session', label: 'Sessions', color: '#f0b849' },
  { key: 'question', label: 'Questions', color: '#fbbf24' },
  { key: 'task', label: 'Tasks', color: '#ef6b6b' },
]

const edgeTypeFilters = [
  { key: 'crossref', label: 'Cross-references' },
  { key: 'tag', label: 'Shared tags' },
  { key: 'section', label: 'Section hierarchy' },
  { key: 'session', label: 'Session links' },
]

// Computed stats
const nodeCountByType = computed(() => {
  const counts: Record<string, number> = {}
  for (const n of nodes.value) {
    counts[n.type] = (counts[n.type] || 0) + 1
  }
  return counts
})

const edgeCountByType = computed(() => {
  const counts: Record<string, number> = {}
  for (const e of edges.value) {
    counts[e.type] = (counts[e.type] || 0) + 1
  }
  return counts
})

const filteredNodeCount = ref(0)
const filteredEdgeCount = ref(0)

function getNodeColor(type: string): string {
  const colors: Record<string, string> = {
    entry: '#6cc5e0',
    section: '#a78bfa',
    session: '#f0b849',
    question: '#fbbf24',
    task: '#ef6b6b',
  }
  return colors[type] || '#888'
}

const BASE_RADIUS: Record<string, number> = {
  section: 8,
  entry: 6,
  session: 7,
  question: 5,
  task: 6,
}

// Connection counts per node, recalculated when simulation rebuilds
let connectionCounts: Map<string, number> = new Map()

function recalcConnectionCounts() {
  const counts = new Map<string, number>()
  for (const e of simEdges) {
    const srcId = typeof e.source === 'string' ? e.source : e.source?.id
    const tgtId = typeof e.target === 'string' ? e.target : e.target?.id
    if (srcId) counts.set(srcId, (counts.get(srcId) || 0) + 1)
    if (tgtId) counts.set(tgtId, (counts.get(tgtId) || 0) + 1)
  }
  connectionCounts = counts
}

function getNodeRadius(type: string, nodeId?: string): number {
  const base = BASE_RADIUS[type] || 5
  if (!nodeId) return base
  const conn = connectionCounts.get(nodeId) || 0
  // Scale: base + sqrt(connections) * 1.5, capped at base * 4
  return Math.min(base * 4, base + Math.sqrt(conn) * 1.5)
}

function getEdgeColor(type: string): string {
  const colors: Record<string, string> = {
    crossref: 'rgba(167,139,250,0.4)',
    tag: 'rgba(108,197,224,0.25)',
    section: 'rgba(167,139,250,0.15)',
    session: 'rgba(240,184,73,0.25)',
  }
  return colors[type] || 'rgba(255,255,255,0.1)'
}

// --- Simulation state ---
let simulation: d3Force.Simulation<any, any> | null = null
let simNodes: any[] = []
let simEdges: any[] = []
let transform = { x: 0, y: 0, k: 1 }
let dragNode: any = null
let isPanning = false
let panStart = { x: 0, y: 0, tx: 0, ty: 0 }
let animFrame = 0

const tooltip = ref<{ x: number; y: number; label: string; code: string; type: string; status: string; connections?: number } | null>(null)

// --- Focus mode ---
const focusedNodeId = ref<string | null>(null)
const focusedConnected = ref<Set<string>>(new Set())
const focusedEdges = ref<Set<string>>(new Set())

// Build adjacency list from current simEdges
function buildAdjacency(): Map<string, Array<{ neighborId: string; edgeIdx: number }>> {
  const adj = new Map<string, Array<{ neighborId: string; edgeIdx: number }>>()
  for (let i = 0; i < simEdges.length; i++) {
    const e = simEdges[i]
    const srcId = typeof e.source === 'string' ? e.source : e.source?.id
    const tgtId = typeof e.target === 'string' ? e.target : e.target?.id
    if (!srcId || !tgtId) continue
    if (!adj.has(srcId)) adj.set(srcId, [])
    if (!adj.has(tgtId)) adj.set(tgtId, [])
    adj.get(srcId)!.push({ neighborId: tgtId, edgeIdx: i })
    adj.get(tgtId)!.push({ neighborId: srcId, edgeIdx: i })
  }
  return adj
}

function updateFocusSets() {
  if (!focusedNodeId.value) {
    focusedConnected.value = new Set()
    focusedEdges.value = new Set()
    return
  }

  const adj = buildAdjacency()
  const depth = focusDepth.value
  const visited = new Set<string>([focusedNodeId.value])
  const edgeSet = new Set<string>()
  let frontier = [focusedNodeId.value]

  for (let d = 0; d < depth; d++) {
    const nextFrontier: string[] = []
    for (const nodeId of frontier) {
      const neighbors = adj.get(nodeId) ?? []
      for (const { neighborId, edgeIdx } of neighbors) {
        edgeSet.add(String(edgeIdx))
        if (!visited.has(neighborId)) {
          visited.add(neighborId)
          nextFrontier.push(neighborId)
        }
      }
    }
    frontier = nextFrontier
  }

  focusedConnected.value = visited
  focusedEdges.value = edgeSet
}

function clearFocus() {
  focusedNodeId.value = null
  updateFocusSets()
  render()
}

// --- Build simulation ---
function buildSimulation() {
  const filteredNodes = nodes.value.filter(n => visibleNodeTypes.value.has(n.type))
  let nodeIds = new Set(filteredNodes.map(n => n.id))

  let filteredEdges = edges.value.filter(e => {
    const srcId = typeof e.source === 'string' ? e.source : e.source.id
    const tgtId = typeof e.target === 'string' ? e.target : e.target.id
    return visibleEdgeTypes.value.has(e.type) && nodeIds.has(srcId) && nodeIds.has(tgtId)
  })

  // Determine connected nodes for orphan filtering
  let connectedNodeIds: Set<string> | null = null
  if (hideOrphans.value) {
    connectedNodeIds = new Set<string>()
    for (const e of filteredEdges) {
      const srcId = typeof e.source === 'string' ? e.source : e.source.id
      const tgtId = typeof e.target === 'string' ? e.target : e.target.id
      connectedNodeIds.add(srcId)
      connectedNodeIds.add(tgtId)
    }
  }

  // Copy nodes (filter orphans if needed)
  simNodes = filteredNodes
    .filter(n => !connectedNodeIds || connectedNodeIds.has(n.id))
    .map(n => ({ ...n }))

  // Re-filter node IDs after orphan removal
  nodeIds = new Set(simNodes.map(n => n.id))
  filteredEdges = filteredEdges.filter(e => {
    const srcId = typeof e.source === 'string' ? e.source : e.source.id
    const tgtId = typeof e.target === 'string' ? e.target : e.target.id
    return nodeIds.has(srcId) && nodeIds.has(tgtId)
  })

  simEdges = filteredEdges.map(e => ({
    source: typeof e.source === 'string' ? e.source : e.source.id,
    target: typeof e.target === 'string' ? e.target : e.target.id,
    type: e.type,
    label: e.label,
  }))

  filteredNodeCount.value = simNodes.length
  filteredEdgeCount.value = simEdges.length

  recalcConnectionCounts()

  if (simulation) simulation.stop()

  const nodeCount = simNodes.length
  const linkDist = Math.max(150, 80 + nodeCount * 2)
  const chargeStrength = -Math.max(400, 200 + nodeCount * 8)

  simulation = d3Force.forceSimulation(simNodes)
    .force('link', d3Force.forceLink(simEdges).id((d: any) => d.id).distance(linkDist).strength(0.15))
    .force('charge', d3Force.forceManyBody().strength(chargeStrength).distanceMax(1200))
    .force('center', d3Force.forceCenter(0, 0).strength(0.03))
    .force('collision', d3Force.forceCollide().radius((d: any) => getNodeRadius(d.type, d.id) + 30).strength(0.7))
    .force('x', d3Force.forceX(0).strength(0.02))
    .force('y', d3Force.forceY(0).strength(0.02))
    .alphaDecay(0.015)
    .velocityDecay(0.3)
    .on('tick', render)

  // Update focus sets with new edges
  updateFocusSets()

  const canvas = canvasRef.value
  if (canvas) {
    transform = { x: canvas.width / 2, y: canvas.height / 2, k: 1 }
  }
}

// --- Render ---
function render() {
  const canvas = canvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const w = canvas.width
  const h = canvas.height

  ctx.clearRect(0, 0, w, h)
  ctx.save()
  ctx.translate(transform.x, transform.y)
  ctx.scale(transform.k, transform.k)

  const hasFocus = focusedNodeId.value !== null

  // Draw edges
  const arrows = showArrows.value
  for (let i = 0; i < simEdges.length; i++) {
    const edge = simEdges[i]
    const src = edge.source
    const tgt = edge.target
    if (!src || !tgt || src.x == null || tgt.x == null) continue

    const isHighlighted = hasFocus && focusedEdges.value.has(String(i))
    const isDimmed = hasFocus && !isHighlighted

    let strokeColor: string
    let lineWidth: number

    if (isHighlighted) {
      strokeColor = getNodeColor(edge.type === 'crossref' ? 'section' : 'entry')
      lineWidth = 2
      ctx.globalAlpha = 0.9
    } else {
      strokeColor = getEdgeColor(edge.type)
      lineWidth = edge.type === 'crossref' ? 1.2 : 0.6
      ctx.globalAlpha = isDimmed ? 0.06 : 1
    }

    // Shorten line to stop at target node edge
    const dx = tgt.x - src.x
    const dy = tgt.y - src.y
    const dist = Math.sqrt(dx * dx + dy * dy)
    if (dist < 1) { ctx.globalAlpha = 1; continue }
    const ux = dx / dist
    const uy = dy / dist
    const tgtR = getNodeRadius(tgt.type, tgt.id) + 3
    const endX = tgt.x - ux * tgtR
    const endY = tgt.y - uy * tgtR

    ctx.beginPath()
    ctx.moveTo(src.x, src.y)
    ctx.lineTo(endX, endY)
    ctx.strokeStyle = strokeColor
    ctx.lineWidth = lineWidth
    ctx.stroke()

    // Arrowhead
    if (arrows) {
      const arrowLen = Math.max(6, lineWidth * 4)
      const arrowAngle = 0.45
      ctx.beginPath()
      ctx.moveTo(endX, endY)
      ctx.lineTo(
        endX - arrowLen * Math.cos(Math.atan2(uy, ux) - arrowAngle),
        endY - arrowLen * Math.sin(Math.atan2(uy, ux) - arrowAngle),
      )
      ctx.moveTo(endX, endY)
      ctx.lineTo(
        endX - arrowLen * Math.cos(Math.atan2(uy, ux) + arrowAngle),
        endY - arrowLen * Math.sin(Math.atan2(uy, ux) + arrowAngle),
      )
      ctx.strokeStyle = strokeColor
      ctx.lineWidth = Math.max(1, lineWidth * 0.8)
      ctx.stroke()
    }

    ctx.globalAlpha = 1
  }

  // Draw nodes
  for (const node of simNodes) {
    if (node.x == null) continue
    const r = getNodeRadius(node.type, node.id)
    const color = getNodeColor(node.type)

    const isConnected = !hasFocus || focusedConnected.value.has(node.id)
    const isFocused = node.id === focusedNodeId.value
    const dimAlpha = isConnected ? 1 : 0.08

    ctx.globalAlpha = dimAlpha

    // Glow for focused node
    if (isFocused) {
      ctx.beginPath()
      ctx.arc(node.x, node.y, r + 8, 0, Math.PI * 2)
      ctx.fillStyle = color.replace(')', ',0.25)').replace('rgb', 'rgba')
      ctx.fill()
    }

    ctx.beginPath()
    ctx.arc(node.x, node.y, r + 2, 0, Math.PI * 2)
    ctx.fillStyle = color.replace(')', ',0.15)').replace('rgb', 'rgba')
    ctx.fill()

    // Circle
    ctx.beginPath()
    ctx.arc(node.x, node.y, isFocused ? r + 1 : r, 0, Math.PI * 2)
    ctx.fillStyle = color
    ctx.fill()

    // Ring for focused node
    if (isFocused) {
      ctx.beginPath()
      ctx.arc(node.x, node.y, r + 4, 0, Math.PI * 2)
      ctx.strokeStyle = color
      ctx.lineWidth = 1.5
      ctx.stroke()
    }

    // Label
    const label = node.label || node.code || node.id
    const maxLen = 30
    const displayLabel = label.length > maxLen ? label.slice(0, maxLen) + '...' : label
    ctx.font = `${Math.max(10, 11 / Math.max(transform.k, 0.5))}px system-ui, -apple-system, sans-serif`
    ctx.fillStyle = isConnected ? 'rgba(255,255,255,0.85)' : 'rgba(255,255,255,0.06)'
    ctx.textAlign = 'left'
    ctx.textBaseline = 'middle'
    ctx.fillText(displayLabel, node.x + r + 5, node.y)

    ctx.globalAlpha = 1
  }

  ctx.restore()
}

// --- Interaction helpers ---
function screenToWorld(sx: number, sy: number) {
  return {
    x: (sx - transform.x) / transform.k,
    y: (sy - transform.y) / transform.k,
  }
}

function findNodeAt(wx: number, wy: number): any | null {
  for (let i = simNodes.length - 1; i >= 0; i--) {
    const n = simNodes[i]
    const r = getNodeRadius(n.type, n.id) + 4
    const dx = n.x - wx
    const dy = n.y - wy
    if (dx * dx + dy * dy < r * r) return n
  }
  return null
}

function countNodeConnections(nodeId: string): number {
  let count = 0
  for (const e of simEdges) {
    const srcId = typeof e.source === 'string' ? e.source : e.source?.id
    const tgtId = typeof e.target === 'string' ? e.target : e.target?.id
    if (srcId === nodeId || tgtId === nodeId) count++
  }
  return count
}

function onMouseDown(e: MouseEvent) {
  const canvas = canvasRef.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  const sx = (e.clientX - rect.left) * (canvas.width / rect.width)
  const sy = (e.clientY - rect.top) * (canvas.height / rect.height)
  const { x: wx, y: wy } = screenToWorld(sx, sy)

  const node = findNodeAt(wx, wy)
  if (node) {
    dragNode = node
    dragNode.fx = dragNode.x
    dragNode.fy = dragNode.y
    simulation?.alphaTarget(0.3).restart()
  } else {
    if (focusedNodeId.value) {
      clearFocus()
    }
    isPanning = true
    panStart = { x: e.clientX, y: e.clientY, tx: transform.x, ty: transform.y }
  }
}

function onMouseMove(e: MouseEvent) {
  const canvas = canvasRef.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  const sx = (e.clientX - rect.left) * (canvas.width / rect.width)
  const sy = (e.clientY - rect.top) * (canvas.height / rect.height)

  if (dragNode) {
    const { x: wx, y: wy } = screenToWorld(sx, sy)
    dragNode.fx = wx
    dragNode.fy = wy
    tooltip.value = null
    return
  }

  if (isPanning) {
    const dx = e.clientX - panStart.x
    const dy = e.clientY - panStart.y
    transform.x = panStart.tx + dx
    transform.y = panStart.ty + dy
    render()
    tooltip.value = null
    return
  }

  // Hover tooltip
  const { x: wx, y: wy } = screenToWorld(sx, sy)
  const node = findNodeAt(wx, wy)
  if (node) {
    tooltip.value = {
      x: e.clientX + 12,
      y: e.clientY + 12,
      label: node.label || '',
      code: node.code || '',
      type: node.type,
      status: node.status || '',
      connections: countNodeConnections(node.id),
    }
    canvas.style.cursor = 'grab'
  } else {
    tooltip.value = null
    canvas.style.cursor = 'default'
  }
}

function onMouseUp() {
  if (dragNode) {
    dragNode.fx = null
    dragNode.fy = null
    dragNode = null
    simulation?.alphaTarget(0)
  }
  isPanning = false
}

function onWheel(e: WheelEvent) {
  e.preventDefault()
  const canvas = canvasRef.value
  if (!canvas) return

  const rect = canvas.getBoundingClientRect()
  const mx = (e.clientX - rect.left) * (canvas.width / rect.width)
  const my = (e.clientY - rect.top) * (canvas.height / rect.height)

  const factor = e.deltaY < 0 ? 1.1 : 0.9
  const newK = Math.max(0.1, Math.min(5, transform.k * factor))

  transform.x = mx - (mx - transform.x) * (newK / transform.k)
  transform.y = my - (my - transform.y) * (newK / transform.k)
  transform.k = newK

  render()
}

function onDblClick(e: MouseEvent) {
  const canvas = canvasRef.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  const sx = (e.clientX - rect.left) * (canvas.width / rect.width)
  const sy = (e.clientY - rect.top) * (canvas.height / rect.height)
  const { x: wx, y: wy } = screenToWorld(sx, sy)

  const node = findNodeAt(wx, wy)
  if (node) {
    if (node.type === 'entry') {
      navigateTo(`/research/${researchSlug}/entry/${node.code || node.id}`)
    } else if (node.type === 'session') {
      navigateTo(`/research/${researchSlug}/session/${node.code || node.id}`)
    } else if (node.type === 'task') {
      navigateTo(`/research/${researchSlug}/tasks`)
    }
  }
}

function onContextMenu(e: MouseEvent) {
  const canvas = canvasRef.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  const sx = (e.clientX - rect.left) * (canvas.width / rect.width)
  const sy = (e.clientY - rect.top) * (canvas.height / rect.height)
  const { x: wx, y: wy } = screenToWorld(sx, sy)

  const node = findNodeAt(wx, wy)
  if (node) {
    if (focusedNodeId.value === node.id) {
      focusedNodeId.value = null
    } else {
      focusedNodeId.value = node.id
    }
    updateFocusSets()
    render()
  }
}

function resizeCanvas() {
  const canvas = canvasRef.value
  if (!canvas) return
  const dpr = window.devicePixelRatio || 1
  const rect = canvas.getBoundingClientRect()
  canvas.width = rect.width * dpr
  canvas.height = rect.height * dpr

  if (simNodes.length === 0) {
    transform = { x: canvas.width / 2, y: canvas.height / 2, k: 1 }
  }
  render()
}

// Watch filters to rebuild simulation
watch([visibleEdgeTypes, visibleNodeTypes, hideOrphans], () => {
  buildSimulation()
}, { deep: true })

// Watch depth to update focus highlight
watch(focusDepth, () => {
  if (focusedNodeId.value) {
    updateFocusSets()
    render()
  }
})

watch(showArrows, () => render())

onMounted(async () => {
  await fetchGraph()
  await nextTick()
  resizeCanvas()
  buildSimulation()
  window.addEventListener('resize', resizeCanvas)
})

onUnmounted(() => {
  if (simulation) simulation.stop()
  window.removeEventListener('resize', resizeCanvas)
  if (animFrame) cancelAnimationFrame(animFrame)
})

useRealtimeUpdates((event: any) => {
  if (event.research_id === researchSlug || event.entity_id === researchSlug) {
    fetchGraph().then(() => buildSimulation())
  }
})
</script>

<style scoped>
.graph-page {
  position: fixed;
  inset: 0;
  display: flex;
  background: #111;
  z-index: 100;
}

/* --- Sidebar --- */
.graph-sidebar {
  width: 240px;
  flex-shrink: 0;
  background: rgba(0,0,0,0.5);
  border-right: 1px solid rgba(255,255,255,0.06);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  transition: width 0.2s;
  position: relative;
}

.graph-sidebar.collapsed {
  width: 36px;
  overflow: hidden;
}

.sidebar-toggle {
  position: absolute;
  top: 10px;
  right: 8px;
  background: none;
  border: none;
  color: rgba(255,255,255,0.4);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  z-index: 2;
}
.sidebar-toggle:hover {
  color: rgba(255,255,255,0.8);
  background: rgba(255,255,255,0.06);
}

.sidebar-header {
  padding: 12px 14px 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.sidebar-back {
  display: flex;
  align-items: center;
  gap: 4px;
  color: rgba(255,255,255,0.45);
  text-decoration: none;
  font-size: 12px;
}
.sidebar-back:hover { color: rgba(255,255,255,0.8); }

.sidebar-title {
  font-size: 14px;
  font-weight: 600;
  color: rgba(255,255,255,0.85);
}

.sidebar-section {
  padding: 10px 14px;
  border-top: 1px solid rgba(255,255,255,0.04);
}

.sidebar-section-title {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: rgba(255,255,255,0.35);
  margin-bottom: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sidebar-stats {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: rgba(255,255,255,0.45);
}

.sidebar-check {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  cursor: pointer;
  font-size: 12px;
  color: rgba(255,255,255,0.7);
}

.sidebar-check input[type="checkbox"] {
  appearance: none;
  width: 14px;
  height: 14px;
  border: 1.5px solid rgba(255,255,255,0.2);
  border-radius: 3px;
  background: transparent;
  cursor: pointer;
  position: relative;
  flex-shrink: 0;
}

.sidebar-check input[type="checkbox"]:checked {
  background: rgba(255,255,255,0.15);
  border-color: rgba(255,255,255,0.4);
}

.sidebar-check input[type="checkbox"]:checked::after {
  content: '';
  position: absolute;
  top: 1px;
  left: 4px;
  width: 4px;
  height: 7px;
  border: solid rgba(255,255,255,0.9);
  border-width: 0 1.5px 1.5px 0;
  transform: rotate(45deg);
}

.check-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.check-label {
  flex: 1;
}

.check-count {
  font-size: 11px;
  color: rgba(255,255,255,0.25);
  font-variant-numeric: tabular-nums;
}

.depth-slider {
  width: 100%;
  appearance: none;
  height: 4px;
  background: rgba(255,255,255,0.1);
  border-radius: 2px;
  outline: none;
  margin: 4px 0 2px;
}

.depth-slider::-webkit-slider-thumb {
  appearance: none;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: #a78bfa;
  cursor: pointer;
  border: 2px solid #1a1a2e;
}

.depth-slider::-moz-range-thumb {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: #a78bfa;
  cursor: pointer;
  border: 2px solid #1a1a2e;
}

.depth-labels {
  display: flex;
  justify-content: space-between;
  font-size: 9px;
  color: rgba(255,255,255,0.2);
  padding: 0 2px;
}

.depth-value {
  color: #a78bfa;
  font-weight: 700;
  font-size: 12px;
}

.sidebar-hint {
  font-size: 10px;
  color: rgba(255,255,255,0.2);
  margin-top: 6px;
  font-style: italic;
}

.btn-clear-focus {
  display: block;
  width: 100%;
  margin-top: 8px;
  padding: 5px;
  background: rgba(167,139,250,0.1);
  border: 1px solid rgba(167,139,250,0.25);
  border-radius: 5px;
  color: #a78bfa;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-clear-focus:hover {
  background: rgba(167,139,250,0.2);
}

/* --- Main area --- */
.graph-main {
  flex: 1;
  position: relative;
  display: flex;
}

.graph-canvas {
  flex: 1;
  width: 100%;
  cursor: default;
}

.graph-loading {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: rgba(255,255,255,0.5);
  font-size: 14px;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255,255,255,0.15);
  border-top-color: rgba(255,255,255,0.6);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.graph-tooltip {
  position: fixed;
  z-index: 100;
  background: rgba(20,20,20,0.95);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 8px;
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  pointer-events: none;
  max-width: 300px;
  backdrop-filter: blur(8px);
}

.tooltip-type {
  font-size: 10px;
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.tooltip-code {
  font-size: 11px;
  color: rgba(255,255,255,0.4);
  font-family: monospace;
}

.tooltip-label {
  font-size: 13px;
  color: rgba(255,255,255,0.9);
  line-height: 1.3;
}

.tooltip-status {
  font-size: 11px;
  color: rgba(255,255,255,0.4);
}

.tooltip-connections {
  font-size: 10px;
  color: rgba(255,255,255,0.3);
}
</style>
