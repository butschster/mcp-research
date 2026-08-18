import type { RoadmapCardData } from '~/components/roadmap/RoadmapNodeCard.vue'

// Shared shapes for the board and timeline, which read the raw API roadmap
// rather than the Vue Flow graph. Kept loose (the API node has more fields than
// any one view needs) but explicit about what these views consume.
export interface RawRoadmapNode {
  id: string
  code: string
  title: string
  description: string
  node_type: string
  status: string
  stage?: string
  node_date?: string
  ref_type?: string
  ref_id?: string
  ref_data?: RoadmapCardData['refData']
}

export interface RawRoadmapEdge {
  id: string
  source_node_id: string
  target_node_id: string
  label: string
  edge_type: string
}

// nodeCardData maps a raw API node to the camelCase shape RoadmapNodeCard wants —
// the same shape the graph's buildGraph already produces, so all three views
// feed the one card component identically.
export function nodeCardData(n: RawRoadmapNode): RoadmapCardData {
  return {
    code: n.code,
    title: n.title,
    description: n.description,
    nodeType: n.node_type,
    status: n.status,
    refType: n.ref_type,
    refId: n.ref_id,
    refData: n.ref_data,
  }
}

// depsByNode returns, for each node id, the predecessor node codes it depends on
// (edges pointing INTO it). Board and timeline render these as chips instead of
// drawing edges across independently-scrolling columns. Derived purely from the
// data already on the payload — no server field needed.
export function depsByNode(
  nodes: readonly RawRoadmapNode[],
  edges: readonly RawRoadmapEdge[],
): Map<string, { code: string }[]> {
  const codeById = new Map(nodes.map(n => [n.id, n.code]))
  const out = new Map<string, { code: string }[]>()
  for (const e of edges) {
    const code = codeById.get(e.source_node_id)
    if (!code) continue
    const list = out.get(e.target_node_id) ?? []
    list.push({ code })
    out.set(e.target_node_id, list)
  }
  return out
}

// relatedIds returns the set of node ids connected to the given node by any edge
// (predecessor or successor), for the hover/focus highlight.
export function relatedIds(nodeId: string, edges: readonly RawRoadmapEdge[]): Set<string> {
  const set = new Set<string>()
  for (const e of edges) {
    if (e.source_node_id === nodeId) set.add(e.target_node_id)
    if (e.target_node_id === nodeId) set.add(e.source_node_id)
  }
  return set
}

// --- Stages bucketing ---

export interface StageColumn {
  name: string
  nodes: RawRoadmapNode[]
  leftover: boolean
}

// bucketByStage assigns each node to a declared stage column, collecting anything
// with an empty or unknown stage into a trailing "leftover" column that only
// appears when it has content. Declared stages always render (even empty), the
// same contract statuses[] already follows.
export function bucketByStage(nodes: readonly RawRoadmapNode[], stages: readonly string[]): StageColumn[] {
  const known = new Set(stages)
  const cols: StageColumn[] = stages.map(name => ({ name, nodes: [], leftover: false }))
  const byName = new Map(cols.map(c => [c.name, c]))
  const leftover: RawRoadmapNode[] = []
  for (const n of nodes) {
    const stage = n.stage ?? ''
    if (stage !== '' && known.has(stage)) {
      byName.get(stage)!.nodes.push(n)
    } else {
      leftover.push(n)
    }
  }
  if (leftover.length) {
    cols.push({ name: 'Unassigned', nodes: leftover, leftover: true })
  }
  return cols
}

// --- Timeline month axis ---

export interface TimelineMonth {
  key: string // 'YYYY-MM'
  label: string // 'Jan'
  quarterLabel: string // 'Q1 2026' on the first month of a quarter, else ''
  nodes: RawRoadmapNode[]
}

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

function monthKey(iso: string): string | null {
  // Expect 'YYYY-MM-DD'; take the year-month. Reject anything malformed.
  const m = /^(\d{4})-(\d{2})-\d{2}$/.exec(iso)
  return m ? `${m[1]}-${m[2]}` : null
}

// buildMonthAxis returns the contiguous run of months from the earliest to the
// latest dated node (inclusive), each carrying the nodes that fall in it, plus
// the undated nodes set aside. A gap month in the middle still renders (an empty
// cell) so the axis reads as continuous time, not a list of populated months.
export function buildMonthAxis(nodes: readonly RawRoadmapNode[]): {
  months: TimelineMonth[]
  undated: RawRoadmapNode[]
} {
  const undated: RawRoadmapNode[] = []
  const dated: { key: string; node: RawRoadmapNode }[] = []
  for (const n of nodes) {
    const key = n.node_date ? monthKey(n.node_date) : null
    if (key) dated.push({ key, node: n })
    else undated.push(n)
  }
  if (dated.length === 0) return { months: [], undated }

  const keys = dated.map(d => d.key).sort()
  const [firstY, firstM] = keys[0]!.split('-').map(Number) as [number, number]
  const [lastY, lastM] = keys[keys.length - 1]!.split('-').map(Number) as [number, number]

  const months: TimelineMonth[] = []
  const byKey = new Map<string, TimelineMonth>()
  let y = firstY
  let m = firstM
  // Bounded walk: at most a few hundred months even for absurd inputs.
  for (let guard = 0; guard < 1200; guard++) {
    const key = `${y}-${String(m).padStart(2, '0')}`
    // The very first month always carries its quarter caption, even mid-quarter,
    // so the axis never opens with an unlabelled band.
    const quarterStart = m === 1 || m === 4 || m === 7 || m === 10 || months.length === 0
    const tm: TimelineMonth = {
      key,
      label: MONTHS[m - 1] ?? '',
      quarterLabel: quarterStart ? `Q${Math.floor((m - 1) / 3) + 1} ${y}` : '',
      nodes: [],
    }
    months.push(tm)
    byKey.set(key, tm)
    if (y === lastY && m === lastM) break
    m++
    if (m > 12) {
      m = 1
      y++
    }
  }
  for (const d of dated) byKey.get(d.key)?.nodes.push(d.node)
  return { months, undated }
}
