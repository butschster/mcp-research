/**
 * The filter rows both graph pages draw. One list, because the shared page
 * subtracts from it (the types a link withholds) and the owner's page draws
 * all of it — two hand-copied arrays would drift on the first colour change.
 */
export interface GraphNodeTypeFilter {
  key: string
  label: string
  color: string
}

export interface GraphEdgeTypeFilter {
  key: string
  label: string
}

export const GRAPH_NODE_TYPE_FILTERS: GraphNodeTypeFilter[] = [
  { key: 'entry', label: 'Documents', color: 'var(--color-primary)' },
  { key: 'section', label: 'Sections', color: 'var(--hue-5)' },
  { key: 'session', label: 'Sessions', color: 'var(--color-warning)' },
  { key: 'question', label: 'Questions', color: 'var(--hue-6)' },
  { key: 'task', label: 'Tasks', color: 'var(--color-error)' },
]

export const GRAPH_EDGE_TYPE_FILTERS: GraphEdgeTypeFilter[] = [
  { key: 'crossref', label: 'Cross-references' },
  { key: 'tag', label: 'Shared tags' },
  { key: 'section', label: 'Section hierarchy' },
  { key: 'session', label: 'Session links' },
]
