import type { Meta, StoryObj } from '@storybook/vue3'
import GraphSidebar from './GraphSidebar.vue'

const nodeTypes = [
  { key: 'entry', label: 'Documents', color: 'var(--color-primary)' },
  { key: 'section', label: 'Sections', color: 'var(--hue-5)' },
  { key: 'session', label: 'Sessions', color: 'var(--color-warning)' },
  { key: 'question', label: 'Questions', color: 'var(--hue-6)' },
  { key: 'task', label: 'Tasks', color: 'var(--color-error)' },
]

const edgeTypes = [
  { key: 'crossref', label: 'Cross-references' },
  { key: 'tag', label: 'Shared tags' },
  { key: 'section', label: 'Section hierarchy' },
  { key: 'session', label: 'Session links' },
]

const meta: Meta<typeof GraphSidebar> = {
  title: 'Graph/GraphSidebar',
  component: GraphSidebar,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="height: 700px; background: var(--color-bg); display: flex;"><story /></div>',
    }),
  ],
  argTypes: {
    focusDepth: { control: { type: 'range', min: 1, max: 5 } },
  },
}
export default meta
type Story = StoryObj<typeof GraphSidebar>

export const Default: Story = {
  args: {
    collapsed: false,
    nodeTypes,
    edgeTypes,
    visibleNodeTypes: new Set(['entry', 'section', 'question', 'task']),
    visibleEdgeTypes: new Set(['crossref', 'section']),
    nodeCountByType: { entry: 42, section: 5, session: 3, question: 18, task: 7 },
    edgeCountByType: { crossref: 23, tag: 56, section: 42, session: 21 },
    nodeCount: 72,
    edgeCount: 65,
    hideOrphans: false,
    showArrows: false,
    focusDepth: 1,
    hasFocus: false,
  },
}

export const WithFocus: Story = {
  args: {
    ...Default.args,
    hasFocus: true,
    focusDepth: 2,
  },
}

export const AllFiltersOn: Story = {
  args: {
    ...Default.args,
    visibleNodeTypes: new Set(['entry', 'section', 'session', 'question', 'task']),
    visibleEdgeTypes: new Set(['crossref', 'tag', 'section', 'session']),
    hideOrphans: true,
    showArrows: true,
    hasFocus: true,
    focusDepth: 3,
  },
}

export const MinimalGraph: Story = {
  args: {
    ...Default.args,
    nodeCountByType: { entry: 3, section: 1, session: 0, question: 0, task: 0 },
    edgeCountByType: { crossref: 1, tag: 0, section: 3, session: 0 },
    nodeCount: 4,
    edgeCount: 4,
    visibleNodeTypes: new Set(['entry', 'section']),
    visibleEdgeTypes: new Set(['crossref', 'section']),
  },
}

export const LargeGraph: Story = {
  args: {
    ...Default.args,
    nodeCountByType: { entry: 156, section: 12, session: 8, question: 64, task: 23 },
    edgeCountByType: { crossref: 89, tag: 234, section: 156, session: 72 },
    nodeCount: 263,
    edgeCount: 551,
    hideOrphans: true,
  },
}

export const Collapsed: Story = {
  args: {
    ...Default.args,
    collapsed: true,
  },
}

export const DeepFocus: Story = {
  args: {
    ...Default.args,
    hasFocus: true,
    focusDepth: 5,
    showArrows: true,
  },
}

export const NoEdges: Story = {
  args: {
    ...Default.args,
    visibleEdgeTypes: new Set<string>(),
    edgeCountByType: { crossref: 0, tag: 0, section: 0, session: 0 },
    nodeCount: 42,
    edgeCount: 0,
  },
}
