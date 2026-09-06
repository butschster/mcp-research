import type { Meta, StoryObj } from '@storybook/vue3'
import GraphNodeList from './GraphNodeList.vue'
import type { GraphNode } from '~/composables/useResearchGraph'
import { GRAPH_NODE_TYPE_FILTERS } from '~/composables/useGraphFilters'
import { entryPath, sessionPath, tasksPath } from '~/composables/useResearchPaths'
import {
  mockGraphEdges,
  mockGraphNodes,
  mockLargeGraph,
  mockLongLabelGraphNode,
} from '../../__mocks__/graph'
import { withShare, withoutShare } from '../../__mocks__/share'

/**
 * The knowledge graph as a list: the same nodes, grouped by type, each with the
 * number of edges that touch it.
 *
 * **This is not a fallback for a canvas that failed.** The canvas is mouse-only
 * — pan by drag, open by double-click, focus by right-click — so a visitor with
 * a keyboard, a screen reader or a phone gets a labelled black rectangle out of
 * it. The list is the peer that can be read and operated, on the same screen,
 * one switch away, and it is the default on a phone.
 *
 * It draws exactly the groups the sidebar draws, in the same order and the same
 * colours, from the same `GRAPH_NODE_TYPE_FILTERS`. A group with nothing in it
 * still appears and says so — the reader has just ticked that filter on, and a
 * heading that vanishes answers a different question from "there are none".
 *
 * Where a node leads is `hrefFor`, supplied by the page. A node with no page of
 * its own comes back with `''` and renders as inert muted text rather than a
 * link to nowhere: a section has no page, a question has none under a share, and
 * a session has none when the link withholds sessions.
 */
const meta: Meta<typeof GraphNodeList> = {
  title: 'Graph/GraphNodeList',
  component: GraphNodeList,
  tags: ['autodocs'],
  // The path helpers below read module share state, so every story starts
  // outside a link rather than from whatever the last one left behind.
  decorators: [withoutShare(), () => ({ template: '<div style="max-width: 720px;"><story /></div>' })],
  argTypes: {
    nodes: { control: 'object' },
    edges: { control: 'object' },
    nodeTypes: { control: 'object' },
  },
  args: {
    nodes: mockGraphNodes,
    edges: mockGraphEdges,
    nodeTypes: GRAPH_NODE_TYPE_FILTERS,
    visibleNodeTypes: new Set(['entry', 'section', 'session', 'question', 'task']),
    hrefFor: ownerHref,
  },
}
export default meta
type Story = StoryObj<typeof GraphNodeList>

/**
 * Where a node leads for a signed-in owner: a document to its page, a session to
 * its page, a task to the board with that card open. A section and a question
 * come back empty — the first has no page anywhere, the second is reached
 * through its session.
 */
function ownerHref(node: GraphNode): string {
  const code = node.code || node.id
  if (node.type === 'entry') return entryPath('R7', code)
  if (node.type === 'session') return sessionPath('R7', code)
  if (node.type === 'task') return tasksPath('R7')
  return ''
}

/** Nothing leads anywhere. This is not a story about a broken page — it is what
 *  a page passes while it is still deciding, and the list has to stay readable
 *  with every row inert. */
const noHref = () => ''

/**
 * All five types, as the owner's list opens. Connection counts come from the
 * edges the page holds, before any filtering — "how connected is this document"
 * is a fact about the project, not about the current filters.
 *
 * E1 is the busiest document at four connections; E6 sits in Documents with
 * none, which is what an orphan looks like when it cannot be hidden.
 */
export const MixedTypes: Story = {}

/**
 * Sessions and tasks filtered off. The groups disappear entirely rather than
 * appearing empty — an unticked filter is the reader's own doing and needs no
 * report.
 */
export const DocumentsAndSectionsOnly: Story = {
  args: { visibleNodeTypes: new Set(['entry', 'section']) },
}

/**
 * A group with nothing in it: Tasks is ticked on and this project has none.
 *
 * "None in this project." rather than an absent heading, because the two states
 * are different answers — the reader has just asked to see tasks, and silence
 * would read as the filter not working.
 */
export const EmptyGroup: Story = {
  args: {
    nodes: mockGraphNodes.filter((n) => n.type !== 'task' && n.type !== 'question'),
    visibleNodeTypes: new Set(['entry', 'section', 'session', 'question', 'task']),
  },
}

/** Every group empty — a project with a graph route and nothing in it yet. Five
 *  headings and five "None in this project." lines, which is honest and still
 *  not a screen anybody should be looking at; the page shows an empty state
 *  instead. */
export const EverythingEmpty: Story = {
  args: { nodes: [], edges: [] },
}

/**
 * Nodes with no destination. Sections and questions are always inert; here
 * `hrefFor` returns `''` for everything, so the whole list is muted text with no
 * underline on hover and no pointer.
 *
 * Inertness is carried by the absence of affordance, not by a message. A row
 * that says "not available" invites the question of who could see it.
 */
export const InertNodes: Story = {
  args: { hrefFor: noHref },
}

/**
 * Under a share link that withholds tasks: sessions lead to the session page
 * inside `/s/{token}/`, and the task node — which a real payload would not even
 * carry — has nowhere to go.
 *
 * Every href in this story starts with the token. A hand-written `/research/…`
 * here would drop an anonymous visitor on the login wall, which is the whole
 * reason the path helpers exist.
 */
export const WithinAShare: Story = {
  decorators: [withShare()],
  args: {
    hrefFor: (node: GraphNode) => {
      const code = node.code || node.id
      if (node.type === 'entry') return entryPath('R7', code)
      if (node.type === 'session') return sessionPath('R7', code)
      return ''
    },
  },
}

/**
 * A 100-character label. It wraps and pushes the row taller; the connection
 * count stays on its own line at the right. The canvas truncates the same label
 * at 30 characters — it has no way to reflow — and this is the view that can
 * afford to show all of it.
 */
export const LongLabel: Story = {
  args: {
    nodes: [...mockGraphNodes, mockLongLabelGraphNode],
    edges: [...mockGraphEdges, { source: 'ent_1', target: 'ent_long', type: 'crossref' }],
  },
}

/** 126 nodes. The list scrolls and nothing paginates; the group counts are how a
 *  reader decides where to start. */
export const LargeProject: Story = {
  args: {
    nodes: mockLargeGraph.nodes,
    edges: mockLargeGraph.edges,
    visibleNodeTypes: new Set(['entry', 'section']),
  },
}

/** A node with no code — every node the API sends has one, but the label falls
 *  back through code to id, and the badge is simply absent rather than empty. */
export const WithoutCodes: Story = {
  args: {
    nodes: mockGraphNodes.map((n) => ({ ...n, code: '' })),
    visibleNodeTypes: new Set(['entry', 'section']),
  },
}

/** The count's wording at its two boundaries: "1 connection", not "1
 *  connections", and "0 connections" for an orphan rather than a blank. E1 has
 *  exactly one edge here and E6 has none. */
export const ConnectionWording: Story = {
  args: {
    nodes: [mockGraphNodes[2]!, mockGraphNodes[7]!],
    edges: [{ source: 'ent_1', target: 'ent_2', type: 'crossref' }],
    visibleNodeTypes: new Set(['entry']),
  },
}

/** ≤480px: the row breaks to two lines, the label above and the state and count
 *  below. This is the view a phone gets by default, so the target height on the
 *  link is deliberate rather than incidental. */
export const Mobile: Story = {
  parameters: { viewport: { defaultViewport: 'mobile' } },
}
