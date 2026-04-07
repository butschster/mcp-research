import dagre from '@dagrejs/dagre'
import { type Node, type Edge, Position } from '@vue-flow/core'

interface RoadmapNodeData {
  id: string
  code: string
  title: string
  description: string
  node_type: string
  status: string
  position_x: number
  position_y: number
  parent_id: string
}

interface RoadmapEdgeData {
  id: string
  source_node_id: string
  target_node_id: string
  label: string
  edge_type: string
}

interface RoadmapData {
  id: string
  code: string
  research_id: string
  title: string
  description: string
  statuses: string[]
  status: string
  nodes: RoadmapNodeData[]
  edges: RoadmapEdgeData[]
}

const NODE_SIZES: Record<string, { width: number; height: number }> = {
  'roadmap-root': { width: 400, height: 160 },
  'roadmap-step': { width: 320, height: 120 },
}

const EDGE_STYLES: Record<string, Record<string, any>> = {
  default: { stroke: 'rgba(140,150,170,0.5)', strokeWidth: 2 },
  success: { stroke: 'rgba(107,203,119,0.6)', strokeWidth: 2 },
  warning: { stroke: 'rgba(240,184,73,0.6)', strokeWidth: 2 },
  optional: { stroke: 'rgba(140,150,170,0.35)', strokeWidth: 1.5, strokeDasharray: '4 4' },
}

export function useRoadmap(roadmapId: string) {
  const roadmap = ref<RoadmapData | null>(null)
  const nodes = ref<Node[]>([])
  const edges = ref<Edge[]>([])
  const loading = ref(true)
  const error = ref<string | null>(null)
  const layoutDirection = ref<'LR' | 'TB'>('TB')

  // Debounced position saves
  const pendingPositions = new Map<string, { x: number; y: number }>()
  let positionTimer: ReturnType<typeof setTimeout> | null = null

  async function fetchRoadmap() {
    const config = useRuntimeConfig()
    const base = config.public.apiBase || ''
    const { authFetch } = useAuth()

    const res = await authFetch<{ data: RoadmapData }>(`${base}/api/roadmaps/${roadmapId}`)
    return res.data
  }

  function findRootNodeId(data: RoadmapData): string | null {
    const targets = new Set(data.edges.map(e => e.target_node_id))
    const root = data.nodes.find(n => !targets.has(n.id))
    return root?.id ?? data.nodes[0]?.id ?? null
  }

  function hasStoredPositions(data: RoadmapData): boolean {
    return data.nodes.some(n => n.position_x !== 0 || n.position_y !== 0)
  }

  function buildGraph(data: RoadmapData) {
    const rootId = findRootNodeId(data)
    const useStored = hasStoredPositions(data)

    const rawNodes: Node[] = data.nodes.map(n => {
      const isRoot = n.id === rootId
      const nodeType = isRoot ? 'roadmap-root' : 'roadmap-step'

      const nodeData: Record<string, any> = isRoot
        ? {
            code: data.code,
            title: data.title,
            description: data.description,
            status: data.status,
            statuses: data.statuses,
            nodeCount: data.nodes.length,
            edgeCount: data.edges.length,
          }
        : {
            code: n.code,
            title: n.title,
            description: n.description,
            nodeType: n.node_type,
            status: n.status,
          }

      return {
        id: n.id,
        type: nodeType,
        position: useStored ? { x: n.position_x, y: n.position_y } : { x: 0, y: 0 },
        data: nodeData,
      }
    })

    const rawEdges: Edge[] = data.edges.map(e => ({
      id: e.id,
      source: e.source_node_id,
      target: e.target_node_id,
      type: 'smoothstep',
      label: e.label || undefined,
      style: EDGE_STYLES[e.edge_type] ?? EDGE_STYLES.default,
      labelStyle: { fill: 'var(--color-text-muted)', fontSize: '0.625rem' },
      labelBgStyle: { fill: 'var(--color-surface)', fillOpacity: 0.9 },
      labelBgPadding: [4, 2] as [number, number],
      labelBgBorderRadius: 3,
    }))

    if (!useStored) {
      return applyDagreLayout(rawNodes, rawEdges)
    }

    // Use stored positions — just set handle positions
    const isHorizontal = layoutDirection.value === 'LR'
    for (const node of rawNodes) {
      node.sourcePosition = isHorizontal ? Position.Right : Position.Bottom
      node.targetPosition = isHorizontal ? Position.Left : Position.Top
    }

    return { nodes: rawNodes, edges: rawEdges }
  }

  function applyDagreLayout(rawNodes: Node[], rawEdges: Edge[]) {
    const isHorizontal = layoutDirection.value === 'LR'
    const g = new dagre.graphlib.Graph()
    g.setDefaultEdgeLabel(() => ({}))
    g.setGraph({ rankdir: isHorizontal ? 'LR' : 'TB', nodesep: 80, ranksep: 140 })

    for (const node of rawNodes) {
      const size = NODE_SIZES[node.type ?? 'roadmap-step'] ?? NODE_SIZES['roadmap-step']
      g.setNode(node.id, { width: size.width, height: size.height })
    }
    for (const edge of rawEdges) {
      g.setEdge(edge.source, edge.target)
    }

    dagre.layout(g)

    const layoutNodes = rawNodes.map(node => {
      const pos = g.node(node.id)
      const size = NODE_SIZES[node.type ?? 'roadmap-step'] ?? NODE_SIZES['roadmap-step']
      return {
        ...node,
        position: { x: pos.x - size.width / 2, y: pos.y - size.height / 2 },
        sourcePosition: isHorizontal ? Position.Right : Position.Bottom,
        targetPosition: isHorizontal ? Position.Left : Position.Top,
      }
    })

    return { nodes: layoutNodes, edges: rawEdges }
  }

  function rebuildGraph() {
    if (!roadmap.value) return
    const result = buildGraph(roadmap.value)
    nodes.value = result.nodes
    edges.value = result.edges
  }

  async function refresh() {
    loading.value = true
    error.value = null
    try {
      roadmap.value = await fetchRoadmap()
      rebuildGraph()
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to load roadmap'
    } finally {
      loading.value = false
    }
  }

  async function updateNodeStatus(nodeId: string, newStatus: string) {
    const config = useRuntimeConfig()
    const base = config.public.apiBase || ''
    const { authFetch } = useAuth()

    try {
      await authFetch(`${base}/api/roadmap-nodes/${nodeId}`, {
        method: 'PUT',
        body: { status: newStatus },
      })
      // Update local state immediately
      if (roadmap.value) {
        const node = roadmap.value.nodes.find(n => n.id === nodeId)
        if (node) node.status = newStatus
        rebuildGraph()
      }
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to update node status'
    }
  }

  function updateNodePosition(nodeId: string, x: number, y: number) {
    pendingPositions.set(nodeId, { x, y })
    // Update local roadmap data immediately (so rebuild uses new positions)
    if (roadmap.value) {
      const node = roadmap.value.nodes.find(n => n.id === nodeId)
      if (node) {
        node.position_x = x
        node.position_y = y
      }
    }
    if (positionTimer) clearTimeout(positionTimer)
    positionTimer = setTimeout(flushPositions, 500)
  }

  async function flushPositions() {
    const config = useRuntimeConfig()
    const base = config.public.apiBase || ''
    const { authFetch } = useAuth()

    const entries = Array.from(pendingPositions.entries())
    pendingPositions.clear()

    await Promise.all(
      entries.map(([nodeId, pos]) =>
        authFetch(`${base}/api/roadmap-nodes/${nodeId}`, {
          method: 'PUT',
          body: { position_x: pos.x, position_y: pos.y },
        }).catch(() => {}) // silently ignore position save errors
      )
    )
  }

  async function autoLayout() {
    if (!roadmap.value) return
    // Force dagre layout regardless of stored positions
    const rawNodes: Node[] = roadmap.value.nodes.map(n => ({
      id: n.id,
      type: n.id === findRootNodeId(roadmap.value!) ? 'roadmap-root' : 'roadmap-step',
      position: { x: 0, y: 0 },
      data: {},
    }))
    const rawEdges: Edge[] = roadmap.value.edges.map(e => ({
      id: e.id,
      source: e.source_node_id,
      target: e.target_node_id,
    }))

    const result = applyDagreLayout(rawNodes, rawEdges)

    // Save all new positions
    const config = useRuntimeConfig()
    const base = config.public.apiBase || ''
    const { authFetch } = useAuth()

    await Promise.all(
      result.nodes.map(n =>
        authFetch(`${base}/api/roadmap-nodes/${n.id}`, {
          method: 'PUT',
          body: { position_x: n.position.x, position_y: n.position.y },
        }).catch(() => {})
      )
    )

    // Refresh to get consistent state
    await refresh()
  }

  function setLayoutDirection(dir: 'LR' | 'TB') {
    layoutDirection.value = dir
    rebuildGraph()
  }

  const progress = computed(() => {
    if (!roadmap.value) return { total: 0, completed: 0, percent: 0 }
    const statuses = roadmap.value.statuses
    // The last status in the list is considered "completed"
    const completedStatus = statuses.length > 0 ? statuses[statuses.length - 1] : null
    const total = roadmap.value.nodes.length
    const completed = completedStatus
      ? roadmap.value.nodes.filter(n => n.status === completedStatus).length
      : 0
    return { total, completed, percent: total > 0 ? Math.round((completed / total) * 100) : 0 }
  })

  return {
    roadmap: readonly(roadmap),
    nodes,
    edges,
    loading,
    error,
    progress,
    refresh,
    updateNodeStatus,
    updateNodePosition,
    autoLayout,
    layoutDirection: readonly(layoutDirection),
    setLayoutDirection,
  }
}
