import type { RawRoadmapNode, RawRoadmapEdge } from '~/utils/roadmap'

// Shared fixtures for the board / timeline / column stories. These mirror the
// raw API roadmap payload (snake_case, ref_data on the node) that the three
// non-graph views consume, so the catalogue exercises the same nodeCardData /
// bucketByStage / buildMonthAxis path the product does.
//
// Not a story file: the `components/**/*.stories.ts` glob skips it, and the
// board/column/timeline stories import from here to avoid restating a roadmap
// in four places.

export const STAGES = ['Discovery', 'Design', 'Build', 'Launch']

export const NODES: RawRoadmapNode[] = [
  {
    id: 'n1',
    code: 'N1',
    title: 'Landscape review',
    description: 'Survey the existing tools and where they fall short.',
    node_type: 'step',
    status: 'completed',
    stage: 'Discovery',
    node_date: '2026-01-15',
    ref_type: 'entry',
    ref_id: 'entry-uuid-1',
    ref_data: {
      title: 'Competitive Landscape Review',
      status: 'completed',
      code: 'E3',
      section_name: 'Discovery',
      content: 'Service mesh gives transparent routing and observability without touching application code, which is where the incumbents stop.',
    },
  },
  {
    id: 'n2',
    code: 'N2',
    title: 'User interviews',
    description: 'Talk to eight practitioners about their workflow.',
    node_type: 'step',
    status: 'in_progress',
    stage: 'Discovery',
    node_date: '2026-02-05',
    ref_type: 'session',
    ref_id: 'session-uuid-1',
    ref_data: {
      title: 'Deep-dive User Interviews',
      status: 'active',
      code: 'SS2',
      total_questions: 12,
      answered_questions: 8,
    },
  },
  {
    id: 'n3',
    code: 'N3',
    title: 'Pick a state layer',
    description: 'Decide between Pinia, composables, or a bespoke store.',
    node_type: 'decision',
    status: 'pending',
    stage: 'Design',
  },
  {
    id: 'n4',
    code: 'N4',
    title: 'Component system',
    description: 'Design the shared card, badges, and layout primitives.',
    node_type: 'step',
    status: 'in_progress',
    stage: 'Design',
    node_date: '2026-02-20',
  },
  {
    id: 'n5',
    code: 'N5',
    title: 'Wire the API',
    description: 'Connect the frontend to the write endpoints.',
    node_type: 'step',
    status: 'pending',
    stage: 'Build',
    node_date: '2026-03-10',
    ref_type: 'task',
    ref_id: 'task-uuid-1',
    ref_data: {
      title: 'Configure GitHub Actions CI/CD pipeline',
      status: 'in_progress',
      code: 'T5',
      priority: 'high',
      result: 'Build, test and deploy stages green. Average build time 3m 20s, deploys to staging on merge.',
    },
  },
  {
    id: 'n6',
    code: 'N6',
    title: 'Public beta',
    description: 'First cohort of external users.',
    node_type: 'milestone',
    status: 'pending',
    stage: 'Launch',
    node_date: '2026-04-01',
  },
]

export const EDGES: RawRoadmapEdge[] = [
  { id: 'e1', source_node_id: 'n1', target_node_id: 'n2', label: '', edge_type: 'depends' },
  { id: 'e2', source_node_id: 'n2', target_node_id: 'n3', label: '', edge_type: 'depends' },
  { id: 'e3', source_node_id: 'n3', target_node_id: 'n4', label: '', edge_type: 'depends' },
  { id: 'e4', source_node_id: 'n4', target_node_id: 'n5', label: '', edge_type: 'depends' },
  { id: 'e5', source_node_id: 'n5', target_node_id: 'n6', label: '', edge_type: 'depends' },
]

// Every node with an empty / unknown stage — drives the board's "all unassigned"
// banner and the leftover column.
export const UNASSIGNED_NODES: RawRoadmapNode[] = NODES.map(n => ({ ...n, stage: '' }))

// Nothing dated — drives the timeline EmptyState + expanded tray.
export const UNDATED_NODES: RawRoadmapNode[] = NODES.map(n => ({ ...n, node_date: undefined }))
