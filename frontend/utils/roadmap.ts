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
  node_end_date?: string
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

// --- Timeline axis: month / quarter / year, with range bars ---

export type TimeUnit = 'month' | 'quarter' | 'year'

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

function parseISO(iso: string | undefined): { y: number; m: number } | null {
  if (!iso) return null
  const mm = /^(\d{4})-(\d{2})-\d{2}$/.exec(iso)
  return mm ? { y: Number(mm[1]), m: Number(mm[2]) } : null
}

function unitIndex(y: number, m: number, unit: TimeUnit): number {
  if (unit === 'year') return y
  if (unit === 'quarter') return y * 4 + Math.floor((m - 1) / 3)
  return y * 12 + (m - 1)
}

export interface TimeCell {
  key: string
  label: string
  band: string
  count: number
}

function cellMeta(index: number, unit: TimeUnit): { key: string; label: string; band: string } {
  if (unit === 'year') return { key: `${index}`, label: `${index}`, band: '' }
  if (unit === 'quarter') {
    const y = Math.floor(index / 4)
    const q = index % 4
    return { key: `${y}-Q${q + 1}`, label: `Q${q + 1}`, band: `${y}` }
  }
  const y = Math.floor(index / 12)
  const m = (index % 12) + 1
  return { key: `${y}-${String(m).padStart(2, '0')}`, label: MONTHS[m - 1] ?? '', band: `Q${Math.floor((m - 1) / 3) + 1} ${y}` }
}

export interface TimelineBar {
  node: RawRoadmapNode
  startIdx: number
  endIdx: number
  lane: number
  invalid: boolean
}

export interface TimelinePoint {
  node: RawRoadmapNode
  idx: number
  invalid: boolean
}

export interface TimeAxis {
  units: TimeCell[]
  first: number
  bars: TimelineBar[]
  points: TimelinePoint[]
  undated: RawRoadmapNode[]
  laneCount: number
}

// autoUnit picks the opening zoom from the dated span in months, so a multi-year
// plan does not open as dozens of month columns. Mobile biases one step coarser.
export function autoUnit(nodes: readonly RawRoadmapNode[], narrow = false): TimeUnit {
  let min = Infinity
  let max = -Infinity
  for (const n of nodes) {
    const s = parseISO(n.node_date)
    if (!s) continue
    const si = unitIndex(s.y, s.m, 'month')
    const e = parseISO(n.node_end_date)
    const ei = e ? unitIndex(e.y, e.m, 'month') : si
    min = Math.min(min, si)
    max = Math.max(max, Math.max(si, ei))
  }
  if (min === Infinity) return narrow ? 'quarter' : 'month'
  const months = max - min + 1
  let unit: TimeUnit = months <= 12 ? 'month' : months <= 36 ? 'quarter' : 'year'
  if (narrow) unit = unit === 'month' ? 'quarter' : 'year'
  return unit
}

// buildTimeAxis buckets nodes into cells of the chosen unit and splits them into
// range bars (start + valid end, not a milestone) and points (everything else,
// undated set aside). Gap cells still render so time reads continuously; bars are
// greedily laned so overlaps do not stack on top of each other.
export function buildTimeAxis(nodes: readonly RawRoadmapNode[], unit: TimeUnit): TimeAxis {
  const undated: RawRoadmapNode[] = []
  const rawBars: { node: RawRoadmapNode; startIdx: number; endIdx: number }[] = []
  const points: TimelinePoint[] = []

  for (const n of nodes) {
    const s = parseISO(n.node_date)
    if (!s) {
      undated.push(n)
      continue
    }
    const startIdx = unitIndex(s.y, s.m, unit)
    const e = parseISO(n.node_end_date)
    const isMilestone = n.node_type === 'milestone'
    if (e && !isMilestone) {
      const endIdx = unitIndex(e.y, e.m, unit)
      const startsBefore = s.y < e.y || (s.y === e.y && s.m <= e.m)
      if (startsBefore) {
        rawBars.push({ node: n, startIdx, endIdx: Math.max(endIdx, startIdx) })
        continue
      }
      points.push({ node: n, idx: startIdx, invalid: true })
      continue
    }
    points.push({ node: n, idx: startIdx, invalid: false })
  }

  if (rawBars.length === 0 && points.length === 0) {
    return { units: [], first: 0, bars: [], points: [], undated, laneCount: 0 }
  }

  let min = Infinity
  let max = -Infinity
  for (const b of rawBars) {
    min = Math.min(min, b.startIdx)
    max = Math.max(max, b.endIdx)
  }
  for (const p of points) {
    min = Math.min(min, p.idx)
    max = Math.max(max, p.idx)
  }

  const units: TimeCell[] = []
  for (let i = min; i <= max; i++) units.push({ ...cellMeta(i, unit), count: 0 })
  const countAt = (idx: number) => {
    const cell = units[idx - min]
    if (cell) cell.count++
  }
  for (const b of rawBars) countAt(b.startIdx)
  for (const p of points) countAt(p.idx)

  const sorted = [...rawBars].sort((a, b) => a.startIdx - b.startIdx || a.endIdx - b.endIdx)
  const laneEnds: number[] = []
  const bars: TimelineBar[] = sorted.map(b => {
    let lane = laneEnds.findIndex(end => end < b.startIdx)
    if (lane === -1) {
      lane = laneEnds.length
      laneEnds.push(b.endIdx)
    } else {
      laneEnds[lane] = b.endIdx
    }
    return { node: b.node, startIdx: b.startIdx, endIdx: b.endIdx, lane, invalid: false }
  })

  return { units, first: min, bars, points, undated, laneCount: laneEnds.length }
}
