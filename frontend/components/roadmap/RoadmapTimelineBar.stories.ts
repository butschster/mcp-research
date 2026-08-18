import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapTimelineBar from './RoadmapTimelineBar.vue'
import type { RoadmapCardData } from './RoadmapNodeCard.vue'

// The compact Gantt bar the timeline lays across several axis cells. The parent
// applies the grid span (grid-column / grid-row); the bar just fills 100% of
// whatever box it is given. Every story wraps it in a fixed-width div so it has
// something to fill and its clip/ellipsis behaviour is visible.
const meta: Meta<typeof RoadmapTimelineBar> = {
  title: 'Roadmap/TimelineBar',
  component: RoadmapTimelineBar,
  tags: ['autodocs'],
  argTypes: {
    highlighted: { control: 'boolean' },
    dimmed: { control: 'boolean' },
  },
  decorators: [
    () => ({ template: '<div style="width: 400px; height: 28px;"><story /></div>' }),
  ],
}
export default meta
type Story = StoryObj<typeof RoadmapTimelineBar>

const step: RoadmapCardData = { code: 'N4', title: 'Component system', nodeType: 'step', status: 'in_progress' }

export const Step: Story = {
  name: 'Plain step bar',
  args: { data: step, startLabel: 'Feb 2026', endLabel: 'Mar 2026' },
}

export const LongTitle: Story = {
  name: 'Long title (clips with ellipsis)',
  args: {
    data: {
      code: 'N5',
      title: 'Wire the API to every write endpoint and reconcile the optimistic cache on failure',
      nodeType: 'step',
      status: 'pending',
    },
    startLabel: 'Mar 2026',
    endLabel: 'May 2026',
  },
}

export const WithDeps: Story = {
  name: 'With dependency chips',
  args: {
    data: step,
    startLabel: 'Feb 2026',
    endLabel: 'Mar 2026',
    deps: [{ code: 'N2' }, { code: 'N3' }],
  },
}

export const Highlighted: Story = {
  args: { data: step, startLabel: 'Feb 2026', endLabel: 'Mar 2026', highlighted: true },
}

export const Dimmed: Story = {
  args: { data: step, startLabel: 'Feb 2026', endLabel: 'Mar 2026', dimmed: true },
}

// The left-accent tint keys off nodeType, mirroring the card's tints. This shows
// every type a bar can carry side by side.
export const AllTypes: Story = {
  name: 'All node types',
  render: () => ({
    components: { RoadmapTimelineBar },
    setup() {
      const bars: { data: RoadmapCardData; start: string; end: string }[] = [
        { data: { code: 'N1', title: 'Landscape review', nodeType: 'step' }, start: 'Jan 2026', end: 'Feb 2026' },
        { data: { code: 'N3', title: 'Pick a state layer', nodeType: 'decision' }, start: 'Feb 2026', end: 'Mar 2026' },
        { data: { code: 'N7', title: 'TypeScript recommended', nodeType: 'info' }, start: 'Mar 2026', end: 'Apr 2026' },
        { data: { code: 'N6', title: 'Public beta', nodeType: 'milestone' }, start: 'Apr 2026', end: 'Apr 2026' },
        { data: { code: 'N8', title: 'Foundations', nodeType: 'group' }, start: 'Jan 2026', end: 'May 2026' },
      ]
      return { bars }
    },
    template: `
      <div style="display: flex; flex-direction: column; gap: 6px; width: 400px;">
        <div v-for="b in bars" :key="b.data.code" style="height: 28px;">
          <RoadmapTimelineBar :data="b.data" :start-label="b.start" :end-label="b.end" />
        </div>
      </div>
    `,
  }),
}
