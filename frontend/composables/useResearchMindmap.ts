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
  answer: { width: 360, height: 80 },
  task: { width: 400, height: 100 },
}

/**
 * Fetches a JSON body for a path relative to the API root. Injected for the
 * same reason `useResearchGraph` takes one: the owner's page wants the
 * authenticated fetch, the shared page wants `shareFetch`, and neither should
 * be guessed from module state.
 */
export type MindmapFetcher = <T>(path: string) => Promise<T>

export interface MindmapOptions {
  fetcher?: MindmapFetcher
  /**
   * Which optional parts to ask for. A share link that excludes sessions
   * answers 404 to the sessions list; not asking is quieter than asking and
   * catching, and it keeps a withheld part from ever appearing as a request in
   * the visitor's network tab.
   */
  parts?: { sessions?: boolean; tasks?: boolean }
}

export function useResearchMindmap(researchId: string, options: MindmapOptions = {}) {
  const nodes = ref<Node[]>([])
  const edges = ref<Edge[]>([])
  const loading = ref(true)
  const error = ref<string | null>(null)
  const collapsedIds = ref<Set<string>>(new Set())
  const layoutDirection = ref<'LR' | 'TB'>('LR')
  const visibleGroups = ref<Set<string>>(new Set(['entries', 'questions', 'tasks']))
  const showCrossrefs = ref(false)

  function resolveFetcher(): MindmapFetcher {
    if (options.fetcher) return options.fetcher
    const config = useRuntimeConfig()
    const base = config.public.apiBase || ''
    const { authFetch } = useAuth()
    return <T>(path: string) => authFetch<T>(`${base}/api${path}`)
  }

  async function fetchAllData() {
    const fetch = resolveFetcher()
    const wantSessions = options.parts?.sessions ?? true
    const wantTasks = options.parts?.tasks ?? true

    // The research and its sections are the map; without them there is nothing
    // to draw and the failure is the page's. Everything after this is a part
    // that may legitimately be absent — a link that excludes tasks answers 404
    // to the tasks list, and that is the link working, not an error.
    const researchRes = await fetch<{ data: ResearchData }>(`/researches/${researchId}`)
    const research = researchRes.data.research
    const sections = researchRes.data.sections ?? []

    const tasks = wantTasks
      ? await fetch<{ data: TaskData[] }>(`/researches/${researchId}/tasks`).then(r => r.data ?? []).catch(() => [] as TaskData[])
      : []

    const sessions = wantSessions
      ? await fetch<{ data: any[] }>(`/researches/${researchId}/sessions`).then(r => r.data ?? []).catch(() => [] as any[])
      : []

    // Fetch entries for all sections in parallel. A section whose list fails
    // draws empty rather than taking the map down.
    const entriesBySection: Record<string, EntryData[]> = {}
    await Promise.all(
      sections.map(async (sec) => {
        try {
          const res = await fetch<{ data: EntryData[] }>(`/researches/${researchId}/sections/${sec.id}/entries`)
          entriesBySection[sec.id] = res.data ?? []
        } catch {
          entriesBySection[sec.id] = []
        }
      })
    )

    // Fetch questions from all sessions, grouped by session
    const sessionGroups: Array<{
      id: string; code: string; title: string; focus: string; status: string
      questions: Array<{ id: string; code: string; text: string; status: string; answer: string }>
    }> = []
    await Promise.all(
      sessions.map(async (sess: any) => {
        try {
          const res = await fetch<{ data: SessionData }>(`/researches/${researchId}/sessions/${sess.id}`)
          const questions: Array<{ id: string; code: string; text: string; status: string; answer: string }> = []
          for (const statusGroup of Object.values(res.data?.questions ?? {})) {
            for (const q of statusGroup) {
              questions.push({ id: q.id, code: q.code, text: q.text, status: q.status, answer: q.answer })
            }
          }
          sessionGroups.push({
            id: sess.id,
            code: sess.code || sess.id,
            title: sess.title,
            focus: sess.focus || '',
            status: sess.status,
            questions,
          })
        } catch {
          // session might not be accessible
        }
      })
    )

    // Fetch cross-references
    const crossrefs = await fetch<{ data: any[] }>(`/researches/${researchId}/crossrefs`).then(r => r.data ?? []).catch(() => [] as any[])

    return { research, sections, entriesBySection, tasks, sessionGroups, crossrefs }
  }

  function buildGraph(data: Awaited<ReturnType<typeof fetchAllData>>) {
    const { research, sections, entriesBySection, tasks, sessionGroups } = data
    const rawNodes: Node[] = []
    const rawEdges: Edge[] = []

    const totalEntries = Object.values(entriesBySection).reduce((sum, e) => sum + e.length, 0)
    const totalQuestions = sessionGroups.reduce((sum, s) => sum + s.questions.length, 0)

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
        questionCount: totalQuestions,
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
          style: { stroke: 'var(--color-primary)', strokeWidth: 2 },
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
              style: { stroke: 'rgba(var(--color-primary-rgb), 0.25)', strokeWidth: 1.5 },
            })
          }
        }
      }
    }

    // Sessions with questions (grouped by session)
    if (visibleGroups.value.has('questions') && sessionGroups.length > 0) {
      const sessionsGroupId = 'group-sessions'
      rawNodes.push({
        id: sessionsGroupId,
        type: 'group-label',
        position: { x: 0, y: 0 },
        data: { label: 'Sessions', count: sessionGroups.length, icon: 'question' },
      })
      rawEdges.push({
        id: `root-${sessionsGroupId}`,
        source: 'root',
        sourceHandle: 'left',
        target: sessionsGroupId,
        type: 'smoothstep',
        style: { stroke: 'rgba(var(--color-warning-rgb), 0.5)', strokeWidth: 2 },
      })

      if (!collapsedIds.value.has(sessionsGroupId)) {
        for (const sess of sessionGroups) {
          const sessNodeId = `session-node-${sess.id}`
          rawNodes.push({
            id: sessNodeId,
            type: 'group-label',
            position: { x: 0, y: 0 },
            data: {
              label: sess.title,
              count: sess.questions.length,
              icon: 'question',
              status: sess.status,
            },
          })
          rawEdges.push({
            id: `${sessionsGroupId}-${sessNodeId}`,
            source: sessionsGroupId,
            target: sessNodeId,
            type: 'smoothstep',
            style: { stroke: 'rgba(var(--color-warning-rgb), 0.35)', strokeWidth: 1.5 },
          })

          if (!collapsedIds.value.has(sessNodeId)) {
            for (const q of sess.questions) {
              const qNodeId = `question-${q.id}`
              rawNodes.push({
                id: qNodeId,
                type: 'question',
                position: { x: 0, y: 0 },
                data: {
                  id: q.id,
                  code: q.code,
                  text: q.text,
                  status: q.status,
                  answer: q.answer,
                  sessionId: sess.code,
                  sessionTitle: sess.title,
                  researchSlug: research.code || researchId,
                },
              })
              rawEdges.push({
                id: `${sessNodeId}-${qNodeId}`,
                source: sessNodeId,
                target: qNodeId,
                type: 'smoothstep',
                style: { stroke: 'rgba(var(--color-warning-rgb), 0.25)', strokeWidth: 1.5 },
              })

              // Add answer node for answered questions
              if (q.answer) {
                const answerNodeId = `answer-${q.id}`
                rawNodes.push({
                  id: answerNodeId,
                  type: 'answer',
                  position: { x: 0, y: 0 },
                  data: {
                    answer: q.answer,
                    questionCode: q.code || q.id,
                    sessionId: sess.code,
                    researchSlug: research.code || researchId,
                  },
                })
                rawEdges.push({
                  id: `${qNodeId}-${answerNodeId}`,
                  source: qNodeId,
                  target: answerNodeId,
                  type: 'smoothstep',
                  style: { stroke: 'rgba(var(--color-success-rgb), 0.35)', strokeWidth: 1.5 },
                })
              }
            }
          }
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
        style: { stroke: 'rgba(var(--color-error-rgb), 0.5)', strokeWidth: 2 },
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
            style: { stroke: 'rgba(var(--color-error-rgb), 0.25)', strokeWidth: 1.5 },
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
          style: { stroke: 'rgba(var(--hue-5-rgb), 0.35)', strokeWidth: 1, strokeDasharray: '4 4' },
        })
      }
    }

    return { rawNodes, rawEdges, crossrefEdges }
  }

  // IDs that belong on the left side (questions + tasks)
  const LEFT_PREFIXES = ['group-sessions', 'session-node-', 'question-', 'answer-', 'group-tasks', 'task-']

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

  /** The HTTP status of the last failure, when there was one. */
  const errorStatus = ref<number | null>(null)
  /** How many things the map holds besides the root — for the empty state. */
  const itemCount = ref(0)

  async function refresh() {
    loading.value = true
    error.value = null
    errorStatus.value = null
    try {
      cachedData = await fetchAllData()
      const d = cachedData
      itemCount.value = Object.values(d.entriesBySection).reduce((n, e) => n + e.length, 0)
        + d.sessionGroups.length + d.tasks.length
      rebuildGraph()
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to load data'
      errorStatus.value = e?.response?.status ?? e?.statusCode ?? null
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
    ids.add('group-sessions')
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
    errorStatus,
    itemCount,
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
