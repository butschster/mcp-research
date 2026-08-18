import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapTimeline from './RoadmapTimeline.vue'
import type { RawRoadmapNode } from '~/utils/roadmap'
import { NODES, EDGES, MIXED_NODES, MIXED_EDGES, UNDATED_NODES } from './roadmap.fixtures'

// The timeline places dated nodes on a unit axis (month / quarter / year, chosen
// by autoUnit and switchable from the zoom toolbar), draws ranged nodes (both
// node_date and node_end_date) as laned Gantt bars, renders milestones as
// markers rather than cards, and sets undated nodes aside in a tray. With nothing
// dated it shows an EmptyState and lists every node.
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

export const RangedBars: Story = {
  name: 'Ranged bars + points + milestone',
  // N4/N5 carry node_end_date → Gantt bars (N4 overlaps N1's span, so they lane
  // separately); N1/N2 are points, N6 a milestone, N3 undated. The dated span
  // sets autoUnit's opening zoom, and the toolbar can switch it.
  args: { nodes: MIXED_NODES, edges: MIXED_EDGES },
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
