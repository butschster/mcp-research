import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import RoadmapViewToggle from './RoadmapViewToggle.vue'

// Segmented control that switches the roadmap body between Graph, Stages and
// Timeline. Not a tablist — it toggles a rendered body via v-model.
const meta: Meta<typeof RoadmapViewToggle> = {
  title: 'Roadmap/ViewToggle',
  component: RoadmapViewToggle,
  tags: ['autodocs'],
  argTypes: {
    modelValue: { control: 'select', options: ['graph', 'stages', 'timeline'] },
  },
}
export default meta
type Story = StoryObj<typeof RoadmapViewToggle>

export const Graph: Story = { args: { modelValue: 'graph' } }
export const Stages: Story = { args: { modelValue: 'stages' } }
export const Timeline: Story = { args: { modelValue: 'timeline' } }

export const Interactive: Story = {
  name: 'Interactive (v-model)',
  render: () => ({
    components: { RoadmapViewToggle },
    setup() {
      const view = ref<'graph' | 'stages' | 'timeline'>('stages')
      return { view }
    },
    template: `
      <div style="display: flex; flex-direction: column; gap: 1rem; align-items: flex-start;">
        <RoadmapViewToggle v-model="view" />
        <p style="font-size: 0.8125rem; color: var(--color-text-muted);">Selected: <strong>{{ view }}</strong></p>
      </div>
    `,
  }),
}
