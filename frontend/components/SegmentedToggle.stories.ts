import type { Meta, StoryObj } from '@storybook/vue3'
import SegmentedToggle from './SegmentedToggle.vue'

/** The generic pressed-button segmented control the roadmap view and zoom
 *  toggles wrap. Switches a value among a few options — not a TabBar. */
const meta: Meta<typeof SegmentedToggle> = {
  title: 'Primitives/SegmentedToggle',
  component: SegmentedToggle,
}
export default meta
type Story = StoryObj<typeof SegmentedToggle>

const options = [
  { value: 'month', label: 'Month' },
  { value: 'quarter', label: 'Quarter' },
  { value: 'year', label: 'Year' },
]

export const Default: Story = {
  args: { modelValue: 'quarter', options, label: 'Timeline zoom' },
}

export const Interactive: Story = {
  render: (args) => ({
    components: { SegmentedToggle },
    setup: () => {
      const value = ref('month')
      return { args, value, options }
    },
    template: '<SegmentedToggle :model-value="value" :options="options" aria-label="Demo" @update:model-value="v => value = v" />',
  }),
}

export const TwoOptions: Story = {
  args: {
    modelValue: 'a',
    options: [{ value: 'a', label: 'First' }, { value: 'b', label: 'Second' }],
    label: 'Two',
  },
}
