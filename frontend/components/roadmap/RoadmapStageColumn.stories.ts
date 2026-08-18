import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapStageColumn from './RoadmapStageColumn.vue'
import { depsByNode, type RawRoadmapNode } from '~/utils/roadmap'
import { NODES, EDGES } from './roadmap.fixtures'

const deps = depsByNode(NODES, EDGES)
const discovery = NODES.filter(n => n.stage === 'Discovery')

// One column of the stages board. Renders its nodes as RoadmapNodeCards with
// dependency chips, an empty message when it has none, and a leftover variant
// for the trailing Unassigned column.
const meta: Meta<typeof RoadmapStageColumn> = {
  title: 'Roadmap/StageColumn',
  component: RoadmapStageColumn,
  tags: ['autodocs'],
  decorators: [
    () => ({ template: '<div style="height: 520px; display: flex;"><story /></div>' }),
  ],
}
export default meta
type Story = StoryObj<typeof RoadmapStageColumn>

export const Populated: Story = {
  args: { name: 'Discovery', nodes: discovery, deps, highlightIds: null },
}

export const Empty: Story = {
  name: 'Empty ("Nothing here yet")',
  args: { name: 'Launch', nodes: [], deps, highlightIds: null },
}

export const Leftover: Story = {
  name: 'Leftover / Unassigned',
  args: {
    name: 'Unassigned',
    nodes: NODES.map(n => ({ ...n, stage: '' })) as RawRoadmapNode[],
    deps,
    leftover: true,
    highlightIds: null,
  },
}

export const LongName: Story = {
  name: 'Long stage name (truncates)',
  args: {
    name: 'Discovery, validation and stakeholder alignment workshop',
    nodes: discovery,
    deps,
    highlightIds: null,
  },
}

export const Highlighted: Story = {
  name: 'Highlight / dim',
  args: {
    name: 'Discovery',
    nodes: discovery,
    deps,
    // Highlight the first Discovery node; the other is dimmed.
    highlightIds: new Set([discovery[0]!.id]),
  },
}
