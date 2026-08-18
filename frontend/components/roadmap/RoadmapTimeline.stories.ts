import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapTimeline from './RoadmapTimeline.vue'
import type { RawRoadmapNode } from '~/utils/roadmap'
import { NODES, EDGES, UNDATED_NODES } from './roadmap.fixtures'

// The timeline places dated nodes on a month axis (buildMonthAxis), renders
// milestones as diamond markers rather than cards, and sets undated nodes aside
// in a tray. With nothing dated it shows an EmptyState and lists every node.
const meta: Meta<typeof RoadmapTimeline> = {
  title: 'Roadmap/Timeline',
  component: RoadmapTimeline,
  tags: ['autodocs'],
  decorators: [
    () => ({ template: '<div style="height: 560px; border: 1px solid var(--color-border); border-radius: var(--radius);"><story /></div>' }),
  ],
}
export default meta
type Story = StoryObj<typeof RoadmapTimeline>

export const Dated: Story = {
  name: 'Dated nodes across months (with milestone)',
  args: { nodes: NODES, edges: EDGES },
}

export const SomeUndated: Story = {
  name: 'Some undated (tray collapsed)',
  args: {
    // n3 (decision) already has no date; drop the date on one more so the tray
    // has content but the axis still renders — the tray stays collapsed.
    nodes: NODES.map(n => (n.id === 'n4' ? { ...n, node_date: undefined } : n)) as RawRoadmapNode[],
    edges: EDGES,
  },
}

export const NothingDated: Story = {
  name: 'Nothing dated (EmptyState + tray)',
  args: { nodes: UNDATED_NODES, edges: EDGES },
}
