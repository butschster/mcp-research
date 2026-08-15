import dagre from '@dagrejs/dagre'
import { type Node, type Edge, Position } from '@vue-flow/core'

interface RoadmapNodeRefData {
  title?: string
  status?: string
  code?: string
  description?: string
  research_id?: string
  section_name?: string
  content?: string
  priority?: string
  result?: string
  total_questions?: number
  answered_questions?: number
  section_count?: number
  entry_count?: number
}

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
  ref_type?: string
  ref_id?: string
  metadata?: string
  ref_data?: RoadmapNodeRefData
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
  'roadmap-ref': { width: 340, height: 140 },
}

const EDGE_STYLES: Record<string, Record<string, any>> = {
  default: { stroke: 'rgba(140,150,170,0.5)', strokeWidth: 2 },
  success: { stroke: 'rgba(107,203,119,0.6)', strokeWidth: 2 },
  warning: { stroke: 'rgba(240,184,73,0.6)', strokeWidth: 2 },
  optional: { stroke: 'rgba(140,150,170,0.35)', strokeWidth: 1.5, strokeDasharray: '4 4' },
}

export function useRoadmap(researchId: string, roadmapId: string) {
  const config = useRuntimeConfig()
  const base = config.public.apiBase || ''
  const { authFetch } = useAuth()

  const roadmap = ref<RoadmapData | null>(null)
  const nodes = ref<Node[]>([])
  const edges = ref<Edge[]>([])
  const loading = ref(true)
  const error = ref<string | null>(null)
  const layoutDirection = ref<'LR' | 'TB'>('TB')

  // Debounced position saves
  const pendingPositions = new Map<string, { x: number; y: number }>()
  let positionTimer: ReturnType<typeof setTimeout> | null = null
  // A drag is a gesture, not a write, and a refetch landing in the middle of one
  // yanks the node out from under the pointer. Our *own* writes no longer need
  // suppressing — the event names the tab that caused it — but a remote change
  // arriving mid-drag still has to wait for the hand to come off the mouse.
  let interactingUntil = 0

  // A share visitor reads the same roadmap over the public prefix, with no
  // Authorization header. The writes below are unreachable for them — the read-
  // only lock turns off every control that calls one, and the server refuses a
  // share on a write route — so only the read needs a second path.
  async function fetchRoadmap() {
    if (shareActive()) {
      const { shareFetch } = useShare()
      const res = await shareFetch<{ data: RoadmapData }>(`/researches/${researchId}/roadmaps/${roadmapId}`)
      return res.data
    }
    const res = await authFetch<{ data: RoadmapData }>(`${base}/api/researches/${researchId}/roadmaps/${roadmapId}`)
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
      const hasRef = !isRoot && n.ref_type && n.ref_id

      let nodeType: string
      if (isRoot) nodeType = 'roadmap-root'
      else if (hasRef) nodeType = 'roadmap-ref'
      else nodeType = 'roadmap-step'

      let nodeData: Record<string, any>
      if (isRoot) {
        nodeData = {
          code: data.code,
          title: data.title,
          description: data.description,
          status: data.status,
          statuses: data.statuses,
          nodeCount: data.nodes.length,
          edgeCount: data.edges.length,
        }
      } else {
        nodeData = {
          code: n.code,
          title: n.title,
          description: n.description,
          nodeType: n.node_type,
          status: n.status,
          refType: n.ref_type,
          refId: n.ref_id,
          metadata: n.metadata,
          refData: n.ref_data,
        }
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

  // Update node data in-place without replacing the array (preserves Vue Flow viewport)
  function patchGraph() {
    if (!roadmap.value) return
    const result = buildGraph(roadmap.value)

    // Patch existing nodes in-place
    const newById = new Map(result.nodes.map(n => [n.id, n]))
    for (const existing of nodes.value) {
      const fresh = newById.get(existing.id)
      if (fresh) {
        existing.data = fresh.data
        // Don't touch position — keep current drag state
      }
    }

    // Add new nodes, remove deleted ones
    const existingIds = new Set(nodes.value.map(n => n.id))
    const freshIds = new Set(result.nodes.map(n => n.id))
    for (const n of result.nodes) {
      if (!existingIds.has(n.id)) nodes.value.push(n)
    }
    nodes.value = nodes.value.filter(n => freshIds.has(n.id))

    // Edges can be fully replaced (they don't affect viewport)
    edges.value = result.edges
  }

  async function refresh(soft = false) {
    if (!soft) loading.value = true
    error.value = null
    try {
      roadmap.value = await fetchRoadmap()
      if (soft) {
        patchGraph()
      } else {
        rebuildGraph()
      }
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to load roadmap'
    } finally {
      loading.value = false
    }
  }

  async function updateNodeStatus(nodeId: string, newStatus: string) {
    try {
      await authFetch(`${base}/api/roadmap-nodes/${nodeId}`, {
        method: 'PUT',
        body: { status: newStatus },
      })
      // Update local state immediately
      if (roadmap.value) {
        const node = roadmap.value.nodes.find(n => n.id === nodeId)
        if (node) node.status = newStatus
        patchGraph()
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
    interactingUntil = Date.now() + 2000
    if (positionTimer) clearTimeout(positionTimer)
    positionTimer = setTimeout(flushPositions, 500)
  }

  async function flushPositions() {
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

  const ENTITY_COMPLETED = new Set(['completed', 'done', 'archived', 'answered', 'final'])

  const progress = computed(() => {
    if (!roadmap.value) return { total: 0, completed: 0, percent: 0 }
    const statuses = roadmap.value.statuses
    const completedStatus = statuses.length > 0 ? statuses[statuses.length - 1] : null
    const total = roadmap.value.nodes.length
    let completed = 0
    for (const n of roadmap.value.nodes) {
      if (n.ref_type && n.ref_data?.status) {
        // Ref node: use entity status
        if (ENTITY_COMPLETED.has(n.ref_data.status)) completed++
      } else if (completedStatus && n.status === completedStatus) {
        // Plain node: use roadmap status
        completed++
      }
    }
    return { total, completed, percent: total > 0 ? Math.round((completed / total) * 100) : 0 }
  })

  /** True while a drag is settling, when a repaint would fight the pointer. */
  function isInteracting(): boolean {
    return Date.now() < interactingUntil
  }

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
    isInteracting,
  }
}
