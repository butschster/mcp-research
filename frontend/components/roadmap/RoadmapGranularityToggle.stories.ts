import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import RoadmapGranularityToggle from './RoadmapGranularityToggle.vue'
import type { TimeUnit } from '~/utils/roadmap'

// Segmented control that switches the timeline axis unit between Month, Quarter
// and Year. Not a tablist — it toggles the axis zoom via v-model, the same
// pressed-button control as RoadmapViewToggle.
const meta: Meta<typeof RoadmapGranularityToggle> = {
  title: 'Roadmap/GranularityToggle',
  component: RoadmapGranularityToggle,
  tags: ['autodocs'],
  argTypes: {
    modelValue: { control: 'select', options: ['month', 'quarter', 'year'] },
  },
}
export default meta
type Story = StoryObj<typeof RoadmapGranularityToggle>

export const Month: Story = { args: { modelValue: 'month' } }
export const Quarter: Story = { args: { modelValue: 'quarter' } }
export const Year: Story = { args: { modelValue: 'year' } }

export const Interactive: Story = {
  name: 'Interactive (v-model)',
  render: () => ({
    components: { RoadmapGranularityToggle },
    setup() {
      const unit = ref<TimeUnit>('quarter')
      return { unit }
    },
    template: `
      <div style="display: flex; flex-direction: column; gap: 1rem; align-items: flex-start;">
        <RoadmapGranularityToggle v-model="unit" />
        <p style="font-size: 0.8125rem; color: var(--color-text-muted);">Selected: <strong>{{ unit }}</strong></p>
      </div>
    `,
  }),
}
