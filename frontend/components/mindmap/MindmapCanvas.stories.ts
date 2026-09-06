import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import { Position, type Edge, type Node } from '@vue-flow/core'
import MindmapCanvas from './MindmapCanvas.vue'
import { withoutShare } from '../../__mocks__/share'

/**
 * The mind map's canvas: Vue Flow with this product's node types, the
 * cross-reference hover, the minimap and the controls, painted in the product's
 * palette rather than the library's.
 *
 * The owner's page and the shared page both render it. What differs — where the
 * data comes from, which toolbar sits above it, whether it takes over the screen
 * — stays in the pages, which is why these stories hand it finished `nodes` and
 * `edges` and listen for `node-click` instead of navigating anywhere.
 *
 * **Positions here are written by hand.** In the product they come out of dagre,
 * laid out left-to-right with the root in the middle: sections and documents to
 * the right, sessions and tasks to the left. The fixture below keeps that shape
 * so the handle sides — `sourcePosition` / `targetPosition` — are the ones a
 * real payload produces. A node whose handles point the wrong way still renders,
 * and its edges enter and leave from behind the card.
 *
 * **The hover is the part worth clicking.** An edge whose id starts with
 * `xref-` is a `[[E1]]` written inside another document; hovering it names both
 * ends, thickens that edge, fades the other references and outlines the two
 * cards. Nothing else in the map explains what a dashed violet line means.
 */
const meta: Meta<typeof MindmapCanvas> = {
  title: 'Mindmap/MindmapCanvas',
  component: MindmapCanvas,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
  // Entry, question and answer nodes navigate through the path helpers, which
  // read module share state; start every story outside a link.
  decorators: [
    withoutShare(),
    () => ({
      template: `
        <div style="height: 600px; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: var(--radius); overflow: hidden;">
          <story />
        </div>
      `,
    }),
  ],
  argTypes: {
    draggable: { control: 'boolean' },
    // Hyphenated because the emit is — see GraphCanvas.stories.ts.
    'onNode-click': { action: 'node-click' },
  },
}
export default meta
type Story = StoryObj<typeof MindmapCanvas>

// --- Fixture -----------------------------------------------------------------
// One project, R7. Right of the root: two sections and four documents. Left of
// it: a session with its questions and an answer, and the task list.

const right = { sourcePosition: Position.Right, targetPosition: Position.Left }
const left = { sourcePosition: Position.Left, targetPosition: Position.Right }

const rootNode: Node = {
  id: 'root',
  type: 'root',
  position: { x: -180, y: -80 },
  data: {
    name: 'Pricing benchmark, Q3',
    goal: 'Work out where our seat pricing sits against the eight vendors procurement keeps naming, and what to change before the October renewal round.',
    status: 'active',
    sectionCount: 2,
    entryCount: 4,
    questionCount: 2,
    taskCount: 2,
  },
  ...right,
}

const documentNodes: Node[] = [
  {
    id: 'section-1',
    type: 'section',
    position: { x: 380, y: -260 },
    data: { code: 'S1', name: 'Market landscape', description: 'Who we actually meet in deals, and what they charge', status: 'active', entryCount: 2 },
    ...right,
  },
  {
    id: 'section-2',
    type: 'section',
    position: { x: 380, y: 100 },
    data: { code: 'S2', name: 'Pricing models', description: 'Seat, usage and the hybrids in between', status: 'completed', entryCount: 2 },
    ...right,
  },
  {
    id: 'entry-1',
    type: 'entry',
    position: { x: 820, y: -340 },
    data: {
      title: 'Seat-tier pricing across eight competitors',
      description: 'Published list prices as of March, normalised to a 50-seat annual contract. See [[E3]] for the Northlight sheet this was checked against.',
      status: 'completed',
      tags: ['pricing', 'competitors'],
      createdAt: new Date(Date.now() - 6 * 86_400_000).toISOString(),
      researchSlug: 'R7',
      entrySlug: 'E1',
    },
    ...right,
  },
  {
    id: 'entry-2',
    type: 'entry',
    position: { x: 820, y: -180 },
    data: {
      title: 'Usage-based pricing: where it breaks down',
      description: 'Two of the eight moved to usage in 2025 and both kept a seat floor. [[E1]] has the numbers.',
      status: 'completed',
      tags: ['pricing'],
      createdAt: new Date(Date.now() - 5 * 86_400_000).toISOString(),
      researchSlug: 'R7',
      entrySlug: 'E2',
    },
    ...right,
  },
  {
    id: 'entry-3',
    type: 'entry',
    position: { x: 820, y: 40 },
    data: {
      title: 'Northlight published price list, March',
      description: 'Captured from their pricing page before the April revision.',
      status: 'draft',
      tags: ['competitors'],
      createdAt: new Date(Date.now() - 3 * 86_400_000).toISOString(),
      researchSlug: 'R7',
      entrySlug: 'E3',
    },
    ...right,
  },
  {
    id: 'entry-4',
    type: 'entry',
    position: { x: 820, y: 200 },
    data: {
      title: 'Enterprise discount ladders',
      description: 'What the ladder looks like above 250 seats, and where it stops being a ladder.',
      status: 'active',
      tags: ['pricing', 'enterprise', 'renewals'],
      createdAt: new Date(Date.now() - 86_400_000).toISOString(),
      researchSlug: 'R7',
      entrySlug: 'E4',
    },
    ...right,
  },
]

const documentEdges: Edge[] = [
  { id: 'root-section-1', source: 'root', sourceHandle: 'right', target: 'section-1', type: 'smoothstep', style: { stroke: 'var(--color-primary)', strokeWidth: 2 }, animated: true },
  { id: 'root-section-2', source: 'root', sourceHandle: 'right', target: 'section-2', type: 'smoothstep', style: { stroke: 'var(--color-primary)', strokeWidth: 2 } },
  { id: 'section-1-entry-1', source: 'section-1', target: 'entry-1', type: 'smoothstep', style: { stroke: 'rgba(var(--color-primary-rgb), 0.25)', strokeWidth: 1.5 } },
  { id: 'section-1-entry-2', source: 'section-1', target: 'entry-2', type: 'smoothstep', style: { stroke: 'rgba(var(--color-primary-rgb), 0.25)', strokeWidth: 1.5 } },
  { id: 'section-2-entry-3', source: 'section-2', target: 'entry-3', type: 'smoothstep', style: { stroke: 'rgba(var(--color-primary-rgb), 0.25)', strokeWidth: 1.5 } },
  { id: 'section-2-entry-4', source: 'section-2', target: 'entry-4', type: 'smoothstep', style: { stroke: 'rgba(var(--color-primary-rgb), 0.25)', strokeWidth: 1.5 } },
]

/** The dashed violet ones. Only edges whose id starts with `xref-` respond to
 *  hover, which is what makes the tooltip mean "cross-reference". */
const crossrefEdges: Edge[] = [
  { id: 'xref-e2-e1', source: 'entry-2', target: 'entry-1', type: 'smoothstep', style: { stroke: 'rgba(var(--hue-5-rgb), 0.35)', strokeWidth: 1, strokeDasharray: '4 4' } },
  { id: 'xref-e4-e3', source: 'entry-4', target: 'entry-3', type: 'smoothstep', style: { stroke: 'rgba(var(--hue-5-rgb), 0.35)', strokeWidth: 1, strokeDasharray: '4 4' } },
]

const sessionAndTaskNodes: Node[] = [
  {
    id: 'group-sessions',
    type: 'group-label',
    position: { x: -620, y: -180 },
    data: { label: 'Sessions', count: 1, icon: 'question' },
    ...left,
  },
  {
    id: 'session-node-1',
    type: 'group-label',
    position: { x: -980, y: -180 },
    data: { label: 'Kickoff interview with the sales lead', count: 2, icon: 'question', status: 'active' },
    ...left,
  },
  {
    id: 'question-1',
    type: 'question',
    position: { x: -1400, y: -280 },
    data: {
      id: 'qst_1',
      code: 'Q1',
      text: 'Which competitor do deals most often stall against, and at what seat count?',
      status: 'answered',
      answer: 'Northlight, almost always above 200 seats — see [[E3]] for their published ladder.',
      sessionId: 'SS1',
      sessionTitle: 'Kickoff interview with the sales lead',
      researchSlug: 'R7',
    },
    ...left,
  },
  {
    id: 'answer-qst_1',
    type: 'answer',
    position: { x: -1820, y: -280 },
    data: {
      answer: 'Northlight, almost always above 200 seats — see [[E3]] for their published ladder.',
      questionCode: 'Q1',
      sessionId: 'SS1',
      researchSlug: 'R7',
    },
    ...left,
  },
  {
    id: 'question-2',
    type: 'question',
    position: { x: -1400, y: -80 },
    data: {
      id: 'qst_2',
      code: 'Q2',
      text: 'Where does the seat model stop making sense for a customer?',
      status: 'pending',
      answer: '',
      sessionId: 'SS1',
      sessionTitle: 'Kickoff interview with the sales lead',
      researchSlug: 'R7',
    },
    ...left,
  },
  {
    id: 'group-tasks',
    type: 'group-label',
    position: { x: -620, y: 140 },
    data: { label: 'Tasks', count: 2, icon: 'task' },
    ...left,
  },
  {
    id: 'task-1',
    type: 'task',
    position: { x: -980, y: 80 },
    data: { code: 'T1', title: 'Recheck the February seat numbers', result: 'Two rows were double-counted; corrected in [[E1]].', status: 'completed', priority: 'high' },
    ...left,
  },
  {
    id: 'task-2',
    type: 'task',
    position: { x: -980, y: 240 },
    data: { code: 'T2', title: 'Ask procurement for the Northlight quote', result: '', status: 'in_progress', priority: 'medium' },
    ...left,
  },
]

const sessionAndTaskEdges: Edge[] = [
  { id: 'root-group-sessions', source: 'root', sourceHandle: 'left', target: 'group-sessions', type: 'smoothstep', style: { stroke: 'rgba(var(--color-warning-rgb), 0.5)', strokeWidth: 2 } },
  { id: 'group-sessions-session-node-1', source: 'group-sessions', target: 'session-node-1', type: 'smoothstep', style: { stroke: 'rgba(var(--color-warning-rgb), 0.35)', strokeWidth: 1.5 } },
  { id: 'session-node-1-question-1', source: 'session-node-1', target: 'question-1', type: 'smoothstep', style: { stroke: 'rgba(var(--color-warning-rgb), 0.25)', strokeWidth: 1.5 } },
  { id: 'session-node-1-question-2', source: 'session-node-1', target: 'question-2', type: 'smoothstep', style: { stroke: 'rgba(var(--color-warning-rgb), 0.25)', strokeWidth: 1.5 } },
  { id: 'question-1-answer-qst_1', source: 'question-1', target: 'answer-qst_1', type: 'smoothstep', style: { stroke: 'rgba(var(--color-success-rgb), 0.35)', strokeWidth: 1.5 } },
  { id: 'root-group-tasks', source: 'root', sourceHandle: 'left', target: 'group-tasks', type: 'smoothstep', style: { stroke: 'rgba(var(--color-error-rgb), 0.5)', strokeWidth: 2 } },
  { id: 'group-tasks-task-1', source: 'group-tasks', target: 'task-1', type: 'smoothstep', style: { stroke: 'rgba(var(--color-error-rgb), 0.25)', strokeWidth: 1.5 } },
  { id: 'group-tasks-task-2', source: 'group-tasks', target: 'task-2', type: 'smoothstep', style: { stroke: 'rgba(var(--color-error-rgb), 0.25)', strokeWidth: 1.5 } },
]

const allNodes = [rootNode, ...documentNodes, ...sessionAndTaskNodes]
const allEdges = [...documentEdges, ...sessionAndTaskEdges, ...crossrefEdges]

// --- Stories -----------------------------------------------------------------

/**
 * The whole map: the project at the centre, its documents to the right, the
 * interview and the task list to the left.
 *
 * It fits itself on init, so what you see is the whole graph rather than the
 * corner of it the default viewport would land on. Hover one of the dashed
 * violet edges between two documents.
 */
export const Default: Story = {
  args: { nodes: allNodes, edges: allEdges, draggable: true },
}

/**
 * Root, sections and documents only — what a share link that withholds sessions
 * and tasks carries.
 *
 * There is no "hidden" state to render: the payload simply has no question or
 * task nodes in it, so the left half of the map does not exist. That is the
 * point of gating on the server rather than in the canvas.
 */
export const DocumentsOnly: Story = {
  args: {
    nodes: [rootNode, ...documentNodes],
    edges: [...documentEdges, ...crossrefEdges],
    draggable: true,
  },
}

/**
 * A project with nothing in it yet: one card, floating.
 *
 * The map is technically correct and useless, which is why the pages check the
 * item count and show an empty state instead of this. Kept as a story because
 * the pages' check is the only thing standing between a new user and it.
 */
export const JustTheRoot: Story = {
  args: { nodes: [rootNode], edges: [], draggable: true },
}

/**
 * `draggable: false`. Panning and zooming still work; the cards are pinned.
 *
 * Dragging never persists — the layout comes back from dagre on the next load
 * either way — so this is about whether a reader can accidentally scramble a
 * picture they are trying to read, not about permissions.
 */
export const NotDraggable: Story = {
  args: { nodes: allNodes, edges: allEdges, draggable: false },
}

/**
 * A document node with more tags than fit and a title that has to wrap, next to
 * an answer long enough to be truncated. Both are the ordinary case on a real
 * project rather than an edge one.
 */
export const OverloadedCards: Story = {
  args: {
    draggable: true,
    nodes: [
      rootNode,
      documentNodes[0]!,
      {
        ...documentNodes[1]!,
        id: 'entry-2',
        data: {
          ...(documentNodes[1]!.data as Record<string, unknown>),
          title: 'Пересчёт февральских данных по посадочным местам после исправления тарифной сетки корпоративного сегмента',
          tags: ['pricing', 'competitors', 'enterprise', 'renewals', 'churn', 'procurement'],
        },
      },
    ],
    edges: [documentEdges[0]!, documentEdges[2]!, documentEdges[3]!],
  },
}

/**
 * Wired to a page's two jobs: reacting to a click, and re-fitting on demand
 * through the exposed `fitAll()`.
 *
 * Click any card and the line underneath names it. In the product that click is
 * a navigation — to the document, or to the session that asked the question —
 * and the destination is decided by the page, not here.
 */
export const Interactive: Story = {
  render: () => ({
    components: { MindmapCanvas },
    setup() {
      const canvas = ref<{ fitAll: () => void } | null>(null)
      const clicked = ref('')
      return {
        canvas,
        clicked,
        nodes: allNodes,
        edges: allEdges,
        onNodeClick: ({ node }: { node: any }) => {
          clicked.value = `${node.type} · ${node.data?.title || node.data?.name || node.data?.label || node.id}`
        },
      }
    },
    template: `
      <div style="height: 100%; display: flex; flex-direction: column;">
        <div style="display: flex; align-items: center; gap: var(--space-3); padding: var(--space-2) var(--space-3); border-bottom: 1px solid var(--color-border); font-size: var(--type-xs); color: var(--color-text-muted); flex: none;">
          <button class="btn btn-sm" @click="canvas?.fitAll()">Fit view</button>
          <span>Clicked: <code>{{ clicked || '—' }}</code></span>
        </div>
        <div style="flex: 1; min-height: 0;">
          <MindmapCanvas ref="canvas" :nodes="nodes" :edges="edges" @node-click="onNodeClick" />
        </div>
      </div>
    `,
  }),
}
