import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import GraphCanvas from './GraphCanvas.vue'
import GraphSidebar from './GraphSidebar.vue'
import type { GraphNode } from '~/composables/useResearchGraph'
import { GRAPH_EDGE_TYPE_FILTERS, GRAPH_NODE_TYPE_FILTERS } from '~/composables/useGraphFilters'
import {
  mockGraphEdges,
  mockGraphNodeCountByType,
  mockGraphNodes,
  mockLargeGraph,
  mockLongLabelGraphNode,
} from '../../__mocks__/graph'

/**
 * The force-directed canvas: a d3 simulation painted into a `<canvas>`, with
 * pan, zoom, drag, hover, focus mode and an auto-fit the first time the layout
 * settles.
 *
 * It knows nothing about where the data came from or where a node leads. The
 * owner's graph page and the shared one both render this component; they differ
 * in the fetch and in the destinations, and both of those stay in the pages.
 * That is why every story here passes plain arrays and listens for `open-node`
 * rather than navigating.
 *
 * **These stories are alive.** The simulation runs, so a canvas takes about a
 * second to spread out and then fits itself to the panel. Give it that second
 * before judging a frame. Filters rebuild the simulation; the viewport is kept
 * across a rebuild on purpose — a reader who has zoomed into a cluster to trace
 * a reference should not lose their place because somebody else's edit arrived.
 *
 * **It is mouse-only, and that is not an oversight to fix here.** Drag to pan,
 * wheel to zoom, double-click to open, right-click to focus. A keyboard, a
 * screen reader or a phone gets a labelled black rectangle out of it, which is
 * why `GraphNodeList` exists beside it as a peer rather than a fallback.
 */
const meta: Meta<typeof GraphCanvas> = {
  title: 'Graph/GraphCanvas',
  component: GraphCanvas,
  tags: ['autodocs'],
  decorators: [
    () => ({
      // `.graph-main` is `flex: 1`, so it needs a flex parent with a height or
      // the canvas measures zero and paints nothing.
      template: `
        <div style="height: 520px; display: flex; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: var(--radius); overflow: hidden;">
          <story />
        </div>
      `,
    }),
  ],
  argTypes: {
    focusDepth: { control: { type: 'range', min: 1, max: 5 } },
    loading: { control: 'boolean' },
    hideOrphans: { control: 'boolean' },
    showArrows: { control: 'boolean' },
    ariaLabel: { control: 'text' },
    // Hyphenated because the emit is: Storybook derives the arg name from the
    // event, so `open-node` is `onOpen-node` and `onOpenNode` would be a prop
    // this component does not have.
    'onOpen-node': { action: 'open-node' },
    'onUpdate:focusedNodeId': { action: 'update:focusedNodeId' },
    onCounts: { action: 'counts' },
  },
  args: {
    nodes: mockGraphNodes,
    edges: mockGraphEdges,
    loading: false,
    visibleNodeTypes: new Set(['entry', 'section', 'question', 'task']),
    visibleEdgeTypes: new Set(['crossref', 'section']),
    hideOrphans: false,
    showArrows: false,
    focusDepth: 1,
    focusedNodeId: null,
  },
}
export default meta
type Story = StoryObj<typeof GraphCanvas>

/**
 * Twelve nodes, as the owner's page opens: documents, sections, questions and
 * tasks, joined by cross-references and the section hierarchy. Shared-tag edges
 * are off, because on a real project they outnumber everything else and turn the
 * picture into a mesh.
 *
 * Node radius grows with the number of connections, so the busiest document is
 * the biggest circle without anything having to say so.
 */
export const Default: Story = {}

/**
 * Waiting for the first payload. The canvas is `v-show`n away — it has no size
 * to measure until it appears, which is why the first build waits for the frame
 * after loading turns false rather than running on mount.
 */
export const Loading: Story = {
  args: { loading: true },
}

/**
 * Every edge type on. This is what the shared-tag edges do to a twelve-node
 * project, and the argument for leaving them off by default on a project with
 * four hundred documents.
 */
export const AllEdgeTypes: Story = {
  args: {
    visibleNodeTypes: new Set(['entry', 'section', 'session', 'question', 'task']),
    visibleEdgeTypes: new Set(['crossref', 'tag', 'section', 'session']),
  },
}

/**
 * Arrowheads on. Direction matters for a cross-reference — `[[E1]]` written
 * inside E2 is not the same fact as the reverse — and it is unreadable without
 * them, so the toggle exists rather than the arrows being permanent clutter.
 */
export const WithArrows: Story = {
  args: {
    showArrows: true,
    visibleEdgeTypes: new Set(['crossref', 'section', 'session']),
    visibleNodeTypes: new Set(['entry', 'section', 'session', 'question', 'task']),
  },
}

/**
 * `hideOrphans`. E6 — a stray note nothing links to and which links to nothing —
 * leaves the picture. Orphans are the reason a graph of a young project looks
 * like a scattering of dots, and hiding them is how a reader gets to the part
 * that is actually connected.
 */
export const OrphansHidden: Story = {
  args: { hideOrphans: true },
}

/**
 * Focus mode, one hop. Right-clicking a node sets `focusedNodeId`; everything
 * outside its neighbourhood drops to 8% opacity rather than disappearing, so the
 * reader keeps the shape of the whole while reading one part of it.
 *
 * The prop is the source of truth and the component only asks for it to change,
 * so the page can clear the focus — closing a panel, following a link — without
 * reaching inside the canvas.
 */
export const Focused: Story = {
  args: {
    focusedNodeId: 'ent_1',
    focusDepth: 1,
    visibleEdgeTypes: new Set(['crossref', 'section', 'tag']),
  },
}

/** The same focus, three hops out. Depth is the difference between "what does
 *  this document touch" and "what neighbourhood is it in". */
export const FocusedThreeHops: Story = {
  args: {
    focusedNodeId: 'ent_1',
    focusDepth: 3,
    visibleEdgeTypes: new Set(['crossref', 'section', 'tag']),
  },
}

/**
 * What a share link that withholds sessions and tasks looks like.
 *
 * The visitor's payload has no session, question or task rows in it at all — the
 * server leaves them out — so this is `visibleNodeTypes` narrowed rather than a
 * canvas hiding something it holds. The canvas cannot tell the difference and
 * must not: it draws what it is given.
 */
export const DocumentsAndSectionsOnly: Story = {
  args: {
    nodes: mockGraphNodes.filter((n) => n.type === 'entry' || n.type === 'section'),
    edges: mockGraphEdges.filter((e) => e.type === 'section' || e.type === 'crossref'),
    visibleNodeTypes: new Set(['entry', 'section']),
  },
}

/**
 * Every node type switched off. The canvas paints an empty field and emits
 * `counts(0, 0)`; the page turns that into "Nothing matches these filters" with
 * a way back. A blank canvas with no explanation is the state this pair of
 * components exists to avoid, and the canvas alone cannot fix it.
 */
export const NothingVisible: Story = {
  args: { visibleNodeTypes: new Set<string>() },
}

/** A project with nothing in it yet. No nodes, no simulation, no error — the
 *  page decides whether that deserves an empty state. */
export const EmptyGraph: Story = {
  args: { nodes: [], edges: [] },
}

/**
 * 126 nodes and 246 edges. Watch the first second: the layout spreads, then
 * fits itself once, and after that the viewport is the reader's.
 *
 * Above 600 nodes the pages start narrowed to documents and sections and say so
 * in a notice, because the simulation freezes a phone. That decision lives in
 * the pages; the canvas draws whatever it is handed.
 */
export const LargeProject: Story = {
  args: {
    nodes: mockLargeGraph.nodes,
    edges: mockLargeGraph.edges,
    visibleNodeTypes: new Set(['entry', 'section']),
  },
}

/**
 * A 100-character label. The canvas truncates at 30 characters and a hover
 * tooltip carries the whole thing; the list view wraps it instead. Both are
 * right for their medium — a canvas label cannot reflow around its neighbours.
 */
export const LongLabel: Story = {
  args: {
    nodes: [...mockGraphNodes, mockLongLabelGraphNode],
    edges: [...mockGraphEdges, { source: 'ent_1', target: 'ent_long', type: 'crossref' }],
  },
}

/**
 * The whole thing wired to a sidebar, as a page assembles it.
 *
 * Right-click a node to focus it and the sidebar's Clear appears; double-click
 * one and the line underneath says which node the page would have opened. The
 * counts under the filters come from the canvas's `counts` emit, not from a
 * second copy of the filtering — two places deciding what is visible is how they
 * come to disagree.
 */
export const Interactive: Story = {
  render: () => ({
    components: { GraphCanvas, GraphSidebar },
    setup() {
      const visibleNodeTypes = ref(new Set(['entry', 'section', 'question', 'task']))
      const visibleEdgeTypes = ref(new Set(['crossref', 'section']))
      const hideOrphans = ref(false)
      const showArrows = ref(false)
      const focusDepth = ref(1)
      const focusedNodeId = ref<string | null>(null)
      const collapsed = ref(false)
      const nodeCount = ref(0)
      const edgeCount = ref(0)
      const opened = ref('')
      const canvas = ref<{ fit: () => void } | null>(null)

      const edgeCountByType: Record<string, number> = {}
      for (const e of mockGraphEdges) edgeCountByType[e.type] = (edgeCountByType[e.type] || 0) + 1

      function toggle(set: typeof visibleNodeTypes, key: string) {
        const next = new Set(set.value)
        if (next.has(key)) next.delete(key)
        else next.add(key)
        set.value = next
      }

      return {
        nodes: mockGraphNodes,
        edges: mockGraphEdges,
        nodeTypes: GRAPH_NODE_TYPE_FILTERS,
        edgeTypes: GRAPH_EDGE_TYPE_FILTERS,
        nodeCountByType: mockGraphNodeCountByType,
        edgeCountByType,
        visibleNodeTypes,
        visibleEdgeTypes,
        hideOrphans,
        showArrows,
        focusDepth,
        focusedNodeId,
        collapsed,
        nodeCount,
        edgeCount,
        opened,
        canvas,
        toggleNodeType: (k: string) => toggle(visibleNodeTypes, k),
        toggleEdgeType: (k: string) => toggle(visibleEdgeTypes, k),
        onCounts: (n: number, e: number) => {
          nodeCount.value = n
          edgeCount.value = e
        },
        onOpenNode: (node: GraphNode) => {
          opened.value = `${node.code} — ${node.label}`
        },
      }
    },
    template: `
      <div style="flex: 1; display: flex; flex-direction: column; min-width: 0;">
        <div style="flex: 1; display: flex; min-height: 0;">
          <GraphSidebar
            :collapsed="collapsed"
            :node-types="nodeTypes"
            :edge-types="edgeTypes"
            :visible-node-types="visibleNodeTypes"
            :visible-edge-types="visibleEdgeTypes"
            :node-count-by-type="nodeCountByType"
            :edge-count-by-type="edgeCountByType"
            :node-count="nodeCount"
            :edge-count="edgeCount"
            :hide-orphans="hideOrphans"
            :show-arrows="showArrows"
            :focus-depth="focusDepth"
            :has-focus="!!focusedNodeId"
            @update:collapsed="collapsed = $event"
            @toggle-node-type="toggleNodeType"
            @toggle-edge-type="toggleEdgeType"
            @update:hide-orphans="hideOrphans = $event"
            @update:show-arrows="showArrows = $event"
            @update:focus-depth="focusDepth = $event"
            @clear-focus="focusedNodeId = null"
          />
          <GraphCanvas
            ref="canvas"
            v-model:focused-node-id="focusedNodeId"
            :nodes="nodes"
            :edges="edges"
            :loading="false"
            :visible-node-types="visibleNodeTypes"
            :visible-edge-types="visibleEdgeTypes"
            :hide-orphans="hideOrphans"
            :show-arrows="showArrows"
            :focus-depth="focusDepth"
            @counts="onCounts"
            @open-node="onOpenNode"
          />
        </div>
        <div style="display: flex; align-items: center; gap: var(--space-3); padding: var(--space-2) var(--space-3); border-top: 1px solid var(--color-border); font-size: var(--type-xs); color: var(--color-text-muted); flex: none;">
          <button class="btn btn-sm" @click="canvas?.fit()">Fit view</button>
          <span>Showing {{ nodeCount }} nodes, {{ edgeCount }} edges</span>
          <span>Double-clicked: <code>{{ opened || '—' }}</code></span>
        </div>
      </div>
    `,
  }),
}
