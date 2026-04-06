import dagre from '@dagrejs/dagre'
import { type Node, type Edge, Position } from '@vue-flow/core'

interface ResearchData {
  research: {
    id: string
    name: string
    goal: string
    status: string
    tags: string[]
  }
  sections: Array<{
    id: string
    name: string
    display_name: string
    description: string
    status: string
    entries_count: number
    position: number
  }>
  active_session: any
}

interface EntryData {
  id: string
  title: string
  description: string
  status: string
  tags: string[]
  created_at: string
  section_id: string
}

interface TaskData {
  id: string
  title: string
  result: string
  status: string
  priority: string
}

interface SessionData {
  session: {
    id: string
    title: string
    status: string
  }
  questions: Record<string, Array<{
    id: string
    text: string
    status: string
    answer: string
    area: string
    priority: string
  }>>
  progress: {
    total: number
    answered: number
  }
}

const NODE_SIZES: Record<string, { width: number; height: number }> = {
  root: { width: 380, height: 150 },
  section: { width: 340, height: 110 },
  entry: { width: 400, height: 100 },
  'group-label': { width: 200, height: 60 },
  question: { width: 400, height: 100 },
  task: { width: 400, height: 100 },
}

export function useResearchMindmap(researchId: string) {
  const nodes = ref<Node[]>([])
  const edges = ref<Edge[]>([])
  const loading = ref(true)
  const error = ref<string | null>(null)
  const collapsedIds = ref<Set<string>>(new Set())
  const layoutDirection = ref<'LR' | 'TB'>('LR')
  const visibleGroups = ref<Set<string>>(new Set(['entries', 'questions', 'tasks']))
  const showCrossrefs = ref(false)

  async function fetchAllData() {
    const config = useRuntimeConfig()
    const base = config.public.apiBase || ''
    const { authFetch } = useAuth()

    const researchRes = await authFetch<{ data: ResearchData }>(`${base}/api/researches/${researchId}`)
    const research = researchRes.data.research
    const sections = researchRes.data.sections ?? []

    const tasksRes = await authFetch<{ data: TaskData[] }>(`${base}/api/researches/${researchId}/tasks`)
    const tasks = tasksRes.data ?? []

    const sessionsRes = await authFetch<{ data: any[] }>(`${base}/api/researches/${researchId}/sessions`)
    const sessions = sessionsRes.data ?? []

    // Fetch entries for all sections in parallel
    const entriesBySection: Record<string, EntryData[]> = {}
    await Promise.all(
      sections.map(async (sec) => {
        const res = await authFetch<{ data: EntryData[] }>(
          `${base}/api/researches/${researchId}/sections/${sec.id}/entries`
        )
        entriesBySection[sec.id] = res.data ?? []
      })
    )

    // Fetch questions from all sessions
    const allQuestions: Array<{ id: string; text: string; status: string; answer: string; sessionTitle: string }> = []
    await Promise.all(
      sessions.map(async (sess: any) => {
        try {
          const res = await authFetch<{ data: SessionData }>(`${base}/api/sessions/${sess.id}`)
          const questions = res.data?.questions ?? {}
          for (const statusGroup of Object.values(questions)) {
            for (const q of statusGroup) {
              allQuestions.push({
                id: q.id,
                code: q.code,
                text: q.text,
                status: q.status,
                answer: q.answer,
                sessionTitle: sess.title,
              })
            }
          }
        } catch {
          // session might not be accessible
        }
      })
    )

    // Fetch cross-references
    const crossrefsRes = await authFetch<{ data: any[] }>(`${base}/api/researches/${researchId}/crossrefs`)
    const crossrefs = crossrefsRes.data ?? []

    return { research, sections, entriesBySection, tasks, allQuestions, crossrefs }
  }

  function buildGraph(data: Awaited<ReturnType<typeof fetchAllData>>) {
    const { research, sections, entriesBySection, tasks, allQuestions } = data
    const rawNodes: Node[] = []
    const rawEdges: Edge[] = []

    const totalEntries = Object.values(entriesBySection).reduce((sum, e) => sum + e.length, 0)

    // Root node
    rawNodes.push({
      id: 'root',
      type: 'root',
      position: { x: 0, y: 0 },
      data: {
        name: research.name,
        goal: research.goal,
        status: research.status,
        sectionCount: sections.length,
        entryCount: totalEntries,
        questionCount: allQuestions.length,
        taskCount: tasks.length,
      },
    })

    // Section nodes + entry nodes
    if (visibleGroups.value.has('entries')) {
      for (const sec of sections) {
        const secNodeId = `section-${sec.id}`
        rawNodes.push({
          id: secNodeId,
          type: 'section',
          position: { x: 0, y: 0 },
          data: {
            code: sec.code,
            name: sec.display_name || sec.name,
            description: sec.description,
            status: sec.status,
            entryCount: entriesBySection[sec.id]?.length ?? 0,
          },
        })
        rawEdges.push({
          id: `root-${secNodeId}`,
          source: 'root',
          sourceHandle: 'right',
          target: secNodeId,
          type: 'smoothstep',
          style: { stroke: '#6cc5e0', strokeWidth: 2 },
          animated: sec.status === 'active',
        })

        if (!collapsedIds.value.has(secNodeId)) {
          const entries = entriesBySection[sec.id] ?? []
          for (const entry of entries) {
            const entryNodeId = `entry-${entry.id}`
            rawNodes.push({
              id: entryNodeId,
              type: 'entry',
              position: { x: 0, y: 0 },
              data: {
                title: entry.title,
                description: entry.description,
                status: entry.status,
                tags: entry.tags ?? [],
                createdAt: entry.created_at,
                researchSlug: research.code || researchId,
                entrySlug: entry.code || entry.id,
              },
            })
            rawEdges.push({
              id: `${secNodeId}-${entryNodeId}`,
              source: secNodeId,
              target: entryNodeId,
              type: 'smoothstep',
              style: { stroke: 'rgba(108,197,224,0.25)', strokeWidth: 1.5 },
            })
          }
        }
      }
    }

    // Questions group
    if (visibleGroups.value.has('questions') && allQuestions.length > 0) {
      const questionsGroupId = 'group-questions'
      rawNodes.push({
        id: questionsGroupId,
        type: 'group-label',
        position: { x: 0, y: 0 },
        data: { label: 'Questions', count: allQuestions.length, icon: 'question' },
      })
      rawEdges.push({
        id: `root-${questionsGroupId}`,
        source: 'root',
        sourceHandle: 'left',
        target: questionsGroupId,
        type: 'smoothstep',
        style: { stroke: 'rgba(240,184,73,0.5)', strokeWidth: 2 },
      })

      if (!collapsedIds.value.has(questionsGroupId)) {
        for (const q of allQuestions) {
          const qNodeId = `question-${q.id}`
          rawNodes.push({
            id: qNodeId,
            type: 'question',
            position: { x: 0, y: 0 },
            data: {
              code: q.code,
              text: q.text,
              status: q.status,
              answer: q.answer,
              sessionTitle: q.sessionTitle,
            },
          })
          rawEdges.push({
            id: `${questionsGroupId}-${qNodeId}`,
            source: questionsGroupId,
            target: qNodeId,
            type: 'smoothstep',
            style: { stroke: 'rgba(240,184,73,0.25)', strokeWidth: 1.5 },
          })
        }
      }
    }

    // Tasks group
    if (visibleGroups.value.has('tasks') && tasks.length > 0) {
      const tasksGroupId = 'group-tasks'
      rawNodes.push({
        id: tasksGroupId,
        type: 'group-label',
        position: { x: 0, y: 0 },
        data: { label: 'Tasks', count: tasks.length, icon: 'task' },
      })
      rawEdges.push({
        id: `root-${tasksGroupId}`,
        source: 'root',
        sourceHandle: 'left',
        target: tasksGroupId,
        type: 'smoothstep',
        style: { stroke: 'rgba(239,107,107,0.5)', strokeWidth: 2 },
      })

      if (!collapsedIds.value.has(tasksGroupId)) {
        for (const t of tasks) {
          const tNodeId = `task-${t.id}`
          rawNodes.push({
            id: tNodeId,
            type: 'task',
            position: { x: 0, y: 0 },
            data: {
              code: t.code,
              title: t.title,
              result: t.result,
              status: t.status,
              priority: t.priority,
            },
          })
          rawEdges.push({
            id: `${tasksGroupId}-${tNodeId}`,
            source: tasksGroupId,
            target: tNodeId,
            type: 'smoothstep',
            style: { stroke: 'rgba(239,107,107,0.25)', strokeWidth: 1.5 },
          })
        }
      }
    }

    // Cross-reference edges are built separately (not included in dagre layout)
    const crossrefEdges: Edge[] = []
    const nodeIds = new Set(rawNodes.map(n => n.id))
    for (const ref of data.crossrefs) {
      if (!ref.resolved || !ref.target_entry_id) continue
      // Map source_type + source_id to node ID
      const sourceId = ref.source_type === 'entry'
        ? `entry-${ref.source_id}`
        : ref.source_type === 'question'
          ? `question-${ref.source_id}`
          : `task-${ref.source_id}`
      const targetId = `entry-${ref.target_entry_id}`
      if (nodeIds.has(sourceId) && nodeIds.has(targetId)) {
        crossrefEdges.push({
          id: `xref-${ref.source_id}-${ref.target_entry_id}`,
          source: sourceId,
          target: targetId,
          type: 'smoothstep',
          animated: false,
          style: { stroke: 'rgba(167,139,250,0.35)', strokeWidth: 1, strokeDasharray: '4 4' },
        })
      }
    }

    return { rawNodes, rawEdges, crossrefEdges }
  }

  // IDs that belong on the left side (questions + tasks)
  const LEFT_PREFIXES = ['group-questions', 'question-', 'group-tasks', 'task-']

  function isLeftNode(id: string): boolean {
    return LEFT_PREFIXES.some(p => id.startsWith(p))
  }

  function applyLayout(rawNodes: Node[], rawEdges: Edge[]) {
    const isHorizontal = layoutDirection.value === 'LR'

    // Split nodes into left (questions/tasks) and right (sections/entries)
    const rootNode = rawNodes.find(n => n.id === 'root')!
    const leftNodes = rawNodes.filter(n => isLeftNode(n.id))
    const rightNodes = rawNodes.filter(n => n.id !== 'root' && !isLeftNode(n.id))
    const leftEdges = rawEdges.filter(e => isLeftNode(e.target))
    const rightEdges = rawEdges.filter(e => !isLeftNode(e.target))

    const rootSize = NODE_SIZES.root

    // Layout right side (sections/entries) — normal LR
    const rightPositions = new Map<string, { x: number; y: number }>()
    if (rightNodes.length) {
      const g = new dagre.graphlib.Graph()
      g.setDefaultEdgeLabel(() => ({}))
      g.setGraph({ rankdir: isHorizontal ? 'LR' : 'TB', nodesep: 60, ranksep: 120 })
      g.setNode('root', { width: rootSize.width, height: rootSize.height })
      for (const node of rightNodes) {
        const size = NODE_SIZES[node.type ?? 'entry'] ?? NODE_SIZES.entry
        g.setNode(node.id, { width: size.width, height: size.height })
      }
      for (const edge of rightEdges) {
        g.setEdge(edge.source, edge.target)
      }
      dagre.layout(g)
      const rootPos = g.node('root')
      for (const node of rightNodes) {
        const pos = g.node(node.id)
        const size = NODE_SIZES[node.type ?? 'entry'] ?? NODE_SIZES.entry
        rightPositions.set(node.id, {
          x: pos.x - rootPos.x,
          y: pos.y - rootPos.y,
        })
      }
    }

    // Layout left side (questions/tasks) — mirrored
    const leftPositions = new Map<string, { x: number; y: number }>()
    if (leftNodes.length) {
      const g = new dagre.graphlib.Graph()
      g.setDefaultEdgeLabel(() => ({}))
      g.setGraph({ rankdir: isHorizontal ? 'LR' : 'TB', nodesep: 60, ranksep: 120 })
      g.setNode('root', { width: rootSize.width, height: rootSize.height })
      for (const node of leftNodes) {
        const size = NODE_SIZES[node.type ?? 'entry'] ?? NODE_SIZES.entry
        g.setNode(node.id, { width: size.width, height: size.height })
      }
      for (const edge of leftEdges) {
        g.setEdge(edge.source, edge.target)
      }
      dagre.layout(g)
      const rootPos = g.node('root')
      for (const node of leftNodes) {
        const pos = g.node(node.id)
        // Mirror: negate X (horizontal) or Y (vertical) relative to root
        if (isHorizontal) {
          leftPositions.set(node.id, { x: -(pos.x - rootPos.x), y: pos.y - rootPos.y })
        } else {
          leftPositions.set(node.id, { x: pos.x - rootPos.x, y: -(pos.y - rootPos.y) })
        }
      }
    }

    // Assemble final positions, centered on root at (0, 0)
    const layoutNodes = rawNodes.map((node) => {
      const size = NODE_SIZES[node.type ?? 'entry'] ?? NODE_SIZES.entry
      let x = 0, y = 0

      if (node.id === 'root') {
        x = -size.width / 2
        y = -size.height / 2
      } else if (leftPositions.has(node.id)) {
        const pos = leftPositions.get(node.id)!
        x = pos.x - size.width / 2
        y = pos.y - size.height / 2
      } else if (rightPositions.has(node.id)) {
        const pos = rightPositions.get(node.id)!
        x = pos.x - size.width / 2
        y = pos.y - size.height / 2
      }

      // Determine handle positions based on side
      let srcPos: Position, tgtPos: Position
      if (isLeftNode(node.id)) {
        srcPos = isHorizontal ? Position.Left : Position.Top
        tgtPos = isHorizontal ? Position.Right : Position.Bottom
      } else {
        srcPos = isHorizontal ? Position.Right : Position.Bottom
        tgtPos = isHorizontal ? Position.Left : Position.Top
      }

      // Root needs handles on both sides
      if (node.id === 'root') {
        srcPos = isHorizontal ? Position.Right : Position.Bottom
        tgtPos = isHorizontal ? Position.Left : Position.Top
      }

      return {
        ...node,
        position: { x, y },
        sourcePosition: srcPos,
        targetPosition: tgtPos,
      }
    })

    return layoutNodes
  }

  let cachedData: Awaited<ReturnType<typeof fetchAllData>> | null = null

  async function refresh() {
    loading.value = true
    error.value = null
    try {
      cachedData = await fetchAllData()
      rebuildGraph()
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to load data'
    } finally {
      loading.value = false
    }
  }

  function rebuildGraph() {
    if (!cachedData) return
    const { rawNodes, rawEdges, crossrefEdges } = buildGraph(cachedData)
    nodes.value = applyLayout(rawNodes, rawEdges)
    edges.value = showCrossrefs.value ? [...rawEdges, ...crossrefEdges] : rawEdges
  }

  function toggleCollapse(id: string) {
    const s = new Set(collapsedIds.value)
    if (s.has(id)) s.delete(id)
    else s.add(id)
    collapsedIds.value = s
    rebuildGraph()
  }

  function expandAll() {
    collapsedIds.value = new Set()
    rebuildGraph()
  }

  function collapseAll() {
    if (!cachedData) return
    const ids = new Set<string>()
    for (const sec of cachedData.sections) ids.add(`section-${sec.id}`)
    ids.add('group-questions')
    ids.add('group-tasks')
    collapsedIds.value = ids
    rebuildGraph()
  }

  function setLayoutDirection(dir: 'LR' | 'TB') {
    layoutDirection.value = dir
    rebuildGraph()
  }

  function toggleGroup(group: string) {
    const s = new Set(visibleGroups.value)
    if (s.has(group)) s.delete(group)
    else s.add(group)
    visibleGroups.value = s
    rebuildGraph()
  }

  function toggleCrossrefs() {
    showCrossrefs.value = !showCrossrefs.value
    rebuildGraph()
  }

  return {
    nodes,
    edges,
    loading,
    error,
    refresh,
    collapsedIds,
    toggleCollapse,
    expandAll,
    collapseAll,
    layoutDirection,
    setLayoutDirection,
    visibleGroups,
    toggleGroup,
    showCrossrefs,
    toggleCrossrefs,
  }
}
