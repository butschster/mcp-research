import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapTimeAxis from './RoadmapTimeAxis.vue'
import type { TimeCell } from '~/utils/roadmap'

// The sticky axis header for the timeline. It renders a row of unit cells
// (month / quarter / year) with an optional per-cell node count, and groups
// contiguous cells that share a `band` caption into a spanning band row above.
// Year zoom carries no band ("") so the band row is skipped entirely.
const meta: Meta<typeof RoadmapTimeAxis> = {
  title: 'Roadmap/TimeAxis',
  component: RoadmapTimeAxis,
  tags: ['autodocs'],
  argTypes: {
    cellWidth: { control: { type: 'range', min: 60, max: 320, step: 10 } },
  },
  decorators: [
    () => ({ template: '<div style="overflow-x: auto;"><story /></div>' }),
  ],
}
export default meta
type Story = StoryObj<typeof RoadmapTimeAxis>

// Month zoom: Jan–Apr 2026. The Q1 band (Jan, Feb, Mar) spans three cells, then
// Q2 begins at Apr. Counts sit on the months that carry nodes.
const monthUnits: TimeCell[] = [
  { key: '2026-01', label: 'Jan', band: 'Q1 2026', count: 1 },
  { key: '2026-02', label: 'Feb', band: 'Q1 2026', count: 2 },
  { key: '2026-03', label: 'Mar', band: 'Q1 2026', count: 0 },
  { key: '2026-04', label: 'Apr', band: 'Q2 2026', count: 1 },
]

export const MonthZoom: Story = {
  name: 'Month zoom (quarter bands)',
  args: { units: monthUnits, cellWidth: 140 },
}

// Quarter zoom: Q1–Q4 2026, all sharing the "2026" band.
const quarterUnits: TimeCell[] = [
  { key: '2026-Q1', label: 'Q1', band: '2026', count: 3 },
  { key: '2026-Q2', label: 'Q2', band: '2026', count: 1 },
  { key: '2026-Q3', label: 'Q3', band: '2026', count: 0 },
  { key: '2026-Q4', label: 'Q4', band: '2026', count: 2 },
]

export const QuarterZoom: Story = {
  name: 'Quarter zoom (year band)',
  args: { units: quarterUnits, cellWidth: 120 },
}

// Year zoom: 2026–2028, band "" for each cell — the band row does not render.
const yearUnits: TimeCell[] = [
  { key: '2026', label: '2026', band: '', count: 4 },
  { key: '2027', label: '2027', band: '', count: 2 },
  { key: '2028', label: '2028', band: '', count: 1 },
]

export const YearZoom: Story = {
  name: 'Year zoom (no band row)',
  args: { units: yearUnits, cellWidth: 120 },
}
