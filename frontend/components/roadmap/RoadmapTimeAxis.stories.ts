import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapTimeAxis from './RoadmapTimeAxis.vue'
import { buildMonthAxis, type TimelineMonth } from '~/utils/roadmap'
import { NODES } from './roadmap.fixtures'

// The sticky month header for the timeline. Groups contiguous months into
// quarter bands (the label spans its months) and shows a per-month node count.
const meta: Meta<typeof RoadmapTimeAxis> = {
  title: 'Roadmap/TimeAxis',
  component: RoadmapTimeAxis,
  tags: ['autodocs'],
  argTypes: {
    cellWidth: { control: { type: 'range', min: 80, max: 320, step: 10 } },
  },
  decorators: [
    () => ({ template: '<div style="overflow-x: auto;"><story /></div>' }),
  ],
}
export default meta
type Story = StoryObj<typeof RoadmapTimeAxis>

// Feb–May 2026: the Q1 band (Feb, Mar) then the Q2 band starting at Apr.
const months: TimelineMonth[] = [
  { key: '2026-02', label: 'Feb', quarterLabel: '', nodes: [] },
  { key: '2026-03', label: 'Mar', quarterLabel: '', nodes: [] },
  { key: '2026-04', label: 'Apr', quarterLabel: 'Q2 2026', nodes: [] },
  { key: '2026-05', label: 'May', quarterLabel: '', nodes: [] },
]

export const QuarterBoundary: Story = {
  name: 'A few months (quarter boundary)',
  args: { months, cellWidth: 160 },
}

export const FromRoadmap: Story = {
  name: 'Derived from the fixture roadmap',
  args: { months: buildMonthAxis(NODES).months, cellWidth: 160 },
}
