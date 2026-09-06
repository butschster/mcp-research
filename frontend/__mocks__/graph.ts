/**
 * Nodes and edges for the knowledge graph, in the shape
 * `GET /api/researches/{id}/graph` returns them.
 *
 * One project, twelve nodes, seen twice: the canvas draws them with a force
 * simulation and the list groups them by type. Both stories read this file, so
 * a node that has a page in one view has the same page in the other — the two
 * views disagreeing about where a node leads is the failure the list exists to
 * prevent, not to demonstrate.
 *
 * Deliberate shapes in here:
 *   - `E6` touches no edge at all, which is what `hideOrphans` removes.
 *   - `SS1`, `Q1`, `Q2` and `T1` are the types a share link may withhold; a
 *     visitor's payload simply has no rows of those types, so a story can
 *     model one by narrowing `visibleNodeTypes`.
 *   - Codes are the product's own (`E3`, `SS1`, `Q2`), because the tooltip and
 *     the list both print them next to the label.
 */
import type { GraphNode, GraphEdge } from '~/composables/useResearchGraph'

export const mockGraphNodes: GraphNode[] = [
  { id: 'sec_1', code: 'S1', label: 'Market landscape', type: 'section', status: 'active' },
  { id: 'sec_2', code: 'S2', label: 'Pricing models', type: 'section', status: 'active' },

  { id: 'ent_1', code: 'E1', label: 'Seat-tier pricing across eight competitors', type: 'entry', status: 'completed', tags: ['pricing', 'competitors'] },
  { id: 'ent_2', code: 'E2', label: 'Usage-based pricing: where it breaks down', type: 'entry', status: 'completed', tags: ['pricing'] },
  { id: 'ent_3', code: 'E3', label: 'Northlight published price list, March', type: 'entry', status: 'draft', tags: ['competitors'] },
  { id: 'ent_4', code: 'E4', label: 'Enterprise discount ladders', type: 'entry', status: 'active', tags: ['pricing', 'enterprise'] },
  { id: 'ent_5', code: 'E5', label: 'What procurement actually asks for', type: 'entry', status: 'active', tags: ['enterprise'] },
  // No edge touches this one. It is what `hideOrphans` takes away.
  { id: 'ent_6', code: 'E6', label: 'Stray note on currency rounding', type: 'entry', status: 'draft', tags: [] },

  { id: 'ses_1', code: 'SS1', label: 'Kickoff interview with the sales lead', type: 'session', status: 'active' },
  { id: 'qst_1', code: 'Q1', label: 'Which competitor do deals most often stall against?', type: 'question', status: 'answered' },
  { id: 'qst_2', code: 'Q2', label: 'Where does the seat model stop making sense?', type: 'question', status: 'pending' },

  { id: 'tsk_1', code: 'T1', label: 'Recheck the February seat numbers', type: 'task', status: 'in_progress' },
]

export const mockGraphEdges: GraphEdge[] = [
  // Section hierarchy — the spine of the layout.
  { source: 'sec_1', target: 'ent_1', type: 'section' },
  { source: 'sec_1', target: 'ent_2', type: 'section' },
  { source: 'sec_1', target: 'ent_3', type: 'section' },
  { source: 'sec_2', target: 'ent_4', type: 'section' },
  { source: 'sec_2', target: 'ent_5', type: 'section' },

  // `[[E1]]` written inside another document.
  { source: 'ent_2', target: 'ent_1', type: 'crossref', label: '[[E1]]' },
  { source: 'ent_4', target: 'ent_3', type: 'crossref', label: '[[E3]]' },
  { source: 'ent_5', target: 'ent_1', type: 'crossref', label: '[[E1]]' },

  // Shared tags — off by default, and the reason `Shared tags` has the biggest
  // count in the sidebar on a real project.
  { source: 'ent_1', target: 'ent_4', type: 'tag', label: 'pricing' },
  { source: 'ent_2', target: 'ent_4', type: 'tag', label: 'pricing' },
  { source: 'ent_4', target: 'ent_5', type: 'tag', label: 'enterprise' },
  { source: 'ent_1', target: 'ent_3', type: 'tag', label: 'competitors' },

  // The interview and what came out of it.
  { source: 'ses_1', target: 'qst_1', type: 'session' },
  { source: 'ses_1', target: 'qst_2', type: 'session' },
  { source: 'qst_1', target: 'ent_1', type: 'session' },
  { source: 'ses_1', target: 'tsk_1', type: 'session' },
]

/** Every node type the API can send, in the order the sidebar lists them. */
export const mockGraphNodeCountByType: Record<string, number> = {
  entry: 6,
  section: 2,
  session: 1,
  question: 2,
  task: 1,
}

/**
 * A project big enough that the force layout matters: 120 documents across six
 * sections, cross-referenced in a chain. This is where the auto-fit on first
 * settle earns its keep — at zoom 1 the reader's first sight of this is a
 * corner of it.
 */
export const mockLargeGraph: { nodes: GraphNode[]; edges: GraphEdge[] } = (() => {
  const nodes: GraphNode[] = []
  const edges: GraphEdge[] = []
  const sectionNames = ['Market landscape', 'Pricing models', 'Procurement', 'Packaging', 'Renewals', 'Churn']
  const statuses = ['draft', 'active', 'completed']

  sectionNames.forEach((name, s) => {
    nodes.push({ id: `sec_${s}`, code: `S${s + 1}`, label: name, type: 'section', status: 'active' })
  })
  for (let i = 0; i < 120; i++) {
    const s = i % sectionNames.length
    nodes.push({
      id: `ent_${i}`,
      code: `E${i + 1}`,
      label: `${sectionNames[s]} — finding ${i + 1}`,
      type: 'entry',
      status: statuses[i % statuses.length]!,
      tags: ['pricing'],
    })
    edges.push({ source: `sec_${s}`, target: `ent_${i}`, type: 'section' })
    // A chain of references, so most nodes have a neighbour and a handful do not.
    if (i > 3 && i % 3 !== 0) edges.push({ source: `ent_${i}`, target: `ent_${i - 3}`, type: 'crossref' })
  }
  return { nodes, edges }
})()

/**
 * The one node whose label is longer than the canvas will draw.
 *
 * The canvas truncates at 30 characters; the list wraps and never truncates.
 * Both are deliberate and the difference only shows with a row like this.
 */
export const mockLongLabelGraphNode: GraphNode = {
  id: 'ent_long',
  code: 'E97',
  label:
    'Пересчёт февральских данных по посадочным местам после исправления тарифной сетки корпоративного сегмента',
  type: 'entry',
  status: 'active',
  tags: ['pricing'],
}
