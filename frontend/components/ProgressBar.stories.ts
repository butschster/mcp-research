import type { Meta, StoryObj } from '@storybook/vue3'
import ProgressBar from './ProgressBar.vue'

const meta: Meta<typeof ProgressBar> = {
  title: 'Base/ProgressBar',
  component: ProgressBar,
  tags: ['autodocs'],
  argTypes: {
    value: { control: { type: 'range', min: 0, max: 100 } },
    total: { control: { type: 'number', min: 1 } },
    showLabel: { control: 'boolean' },
  },
}
export default meta
type Story = StoryObj<typeof ProgressBar>

export const Empty: Story = { args: { value: 0, total: 100, showLabel: true } }
export const Quarter: Story = { args: { value: 25, total: 100, showLabel: true } }
export const Half: Story = { args: { value: 50, total: 100, showLabel: true } }
export const ThreeQuarters: Story = { args: { value: 75, total: 100, showLabel: true } }
export const Complete: Story = { args: { value: 100, total: 100, showLabel: true } }

export const WithoutLabel: Story = { args: { value: 60, total: 100, showLabel: false } }

export const AllStates: Story = {
  render: () => ({
    components: { ProgressBar },
    template: `
      <div style="display: flex; flex-direction: column; gap: 1.5rem; max-width: 400px;">
        <div v-for="v in values" :key="v">
          <div style="font-size: 0.75rem; color: var(--color-text-muted); margin-bottom: 0.25rem;">{{ v }}%</div>
          <ProgressBar :value="v" :total="100" showLabel />
        </div>
      </div>
    `,
    setup() {
      return { values: [0, 10, 25, 50, 75, 100] }
    },
  }),
}
