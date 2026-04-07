import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapEdgeLabel from './RoadmapEdgeLabel.vue'

const meta: Meta<typeof RoadmapEdgeLabel> = {
  title: 'Roadmap/EdgeLabel',
  component: RoadmapEdgeLabel,
  tags: ['autodocs'],
}
export default meta
type Story = StoryObj<typeof RoadmapEdgeLabel>

export const Default: Story = {
  args: {
    label: 'next',
    edgeType: 'default',
  },
}

export const Success: Story = {
  args: {
    label: 'if yes',
    edgeType: 'success',
  },
}

export const Warning: Story = {
  args: {
    label: 'if failed',
    edgeType: 'warning',
  },
}

export const Optional: Story = {
  args: {
    label: 'alternative',
    edgeType: 'optional',
  },
}

export const AllTypes: Story = {
  render: () => ({
    components: { RoadmapEdgeLabel },
    template: `
      <div style="display: flex; gap: 0.75rem; align-items: center; flex-wrap: wrap;">
        <RoadmapEdgeLabel v-for="e in edges" :key="e.label" :label="e.label" :edge-type="e.edgeType" />
      </div>
    `,
    setup() {
      const edges = [
        { label: 'next', edgeType: 'default' },
        { label: 'complete', edgeType: 'success' },
        { label: 'if blocked', edgeType: 'warning' },
        { label: 'skip', edgeType: 'optional' },
        { label: 'prerequisite', edgeType: 'default' },
      ]
      return { edges }
    },
  }),
}
