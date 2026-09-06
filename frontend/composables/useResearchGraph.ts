export interface GraphNode {
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

export interface GraphEdge {
  source: string | GraphNode
  target: string | GraphNode
  type: string // crossref, tag, section, session
  label?: string
}

interface GraphData {
  nodes: GraphNode[]
  edges: GraphEdge[]
  /** The node types this caller could have received. See the graph route. */
  available_node_types?: string[]
}

/**
 * Fetches a JSON body for a path relative to the API root — `/researches/R1/graph`.
 *
 * Injected rather than sniffed from `useShare`: that is module state, and a
 * composable that silently read it would behave differently depending on a page
 * you cannot see from its file. The owner's page passes nothing and gets the
 * authenticated fetch; the shared page passes `shareFetch`, which carries the
 * link and never the owner's token.
 */
export type GraphFetcher = <T>(path: string) => Promise<T>

export const ALL_GRAPH_NODE_TYPES = ['entry', 'section', 'session', 'question', 'task']

export function useResearchGraph(researchId: string, options: { fetcher?: GraphFetcher } = {}) {
  const nodes = ref<GraphNode[]>([])
  const edges = ref<GraphEdge[]>([])
  const loading = ref(true)
  const error = ref<string | null>(null)
  /** The HTTP status of the last failure, when there was one. */
  const errorStatus = ref<number | null>(null)
  /**
   * What the server says this caller may see. Everything, until the first
   * answer says otherwise — a share link's answer lists only the types its
   * include flags cover, and the sidebar builds its rows from this rather
   * than repeating the server's mapping.
   */
  const availableNodeTypes = ref<string[]>([...ALL_GRAPH_NODE_TYPES])

  // Filter toggles
  const visibleEdgeTypes = ref<Set<string>>(new Set(['crossref', 'section']))
  const visibleNodeTypes = ref<Set<string>>(new Set(['entry', 'section', 'question', 'task']))

  async function fetchGraph() {
    loading.value = true
    error.value = null
    errorStatus.value = null
    try {
      let fetcher = options.fetcher
      if (!fetcher) {
        const config = useRuntimeConfig()
        const base = config.public.apiBase || ''
        const { authFetch } = useAuth()
        fetcher = <T>(path: string) => authFetch<T>(`${base}/api${path}`)
      }

      const res = await fetcher<GraphData>(`/researches/${researchId}/graph`)
      nodes.value = res.nodes ?? []
      edges.value = res.edges ?? []
      if (res.available_node_types?.length) availableNodeTypes.value = res.available_node_types
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to load graph'
      errorStatus.value = e?.response?.status ?? e?.statusCode ?? null
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

  /** Re-check every type the caller may see — the way out of "nothing matches these filters". */
  function showAllNodeTypes() {
    visibleNodeTypes.value = new Set(availableNodeTypes.value)
  }

  return {
    nodes,
    edges,
    loading,
    error,
    errorStatus,
    availableNodeTypes,
    fetchGraph,
    visibleEdgeTypes,
    visibleNodeTypes,
    toggleEdgeType,
    toggleNodeType,
    showAllNodeTypes,
  }
}
