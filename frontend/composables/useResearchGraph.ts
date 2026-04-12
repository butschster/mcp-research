interface GraphNode {
  id: string
  code: string
  label: string
  type: string // entry, section, session, question, task
  status: string
  tags?: string[]
  group?: string
  // d3-force simulation fields
  x?: number
  y?: number
  vx?: number
  vy?: number
  fx?: number | null
  fy?: number | null
}

interface GraphEdge {
  source: string | GraphNode
  target: string | GraphNode
  type: string // crossref, tag, section, session
  label?: string
}

interface GraphData {
  nodes: GraphNode[]
  edges: GraphEdge[]
}

export function useResearchGraph(researchId: string) {
  const nodes = ref<GraphNode[]>([])
  const edges = ref<GraphEdge[]>([])
  const loading = ref(true)
  const error = ref<string | null>(null)

  // Filter toggles
  const visibleEdgeTypes = ref<Set<string>>(new Set(['crossref', 'section']))
  const visibleNodeTypes = ref<Set<string>>(new Set(['entry', 'section', 'question', 'task']))

  async function fetchGraph() {
    loading.value = true
    error.value = null
    try {
      const config = useRuntimeConfig()
      const base = config.public.apiBase || ''
      const { authFetch } = useAuth()

      const res = await authFetch<GraphData>(`${base}/api/researches/${researchId}/graph`)
      nodes.value = res.nodes ?? []
      edges.value = res.edges ?? []
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to load graph'
    } finally {
      loading.value = false
    }
  }

  function toggleEdgeType(type: string) {
    const s = new Set(visibleEdgeTypes.value)
    if (s.has(type)) s.delete(type)
    else s.add(type)
    visibleEdgeTypes.value = s
  }

  function toggleNodeType(type: string) {
    const s = new Set(visibleNodeTypes.value)
    if (s.has(type)) s.delete(type)
    else s.add(type)
    visibleNodeTypes.value = s
  }

  return {
    nodes,
    edges,
    loading,
    error,
    fetchGraph,
    visibleEdgeTypes,
    visibleNodeTypes,
    toggleEdgeType,
    toggleNodeType,
  }
}
