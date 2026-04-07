import type { Meta, StoryObj } from '@storybook/vue3'
import ShortCode from './ShortCode.vue'

const meta: Meta<typeof ShortCode> = {
  title: 'Base/ShortCode',
  component: ShortCode,
  tags: ['autodocs'],
  argTypes: {
    code: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof ShortCode>

export const Research: Story = { args: { code: 'R1' } }
export const Entry: Story = { args: { code: 'E15' } }
export const Session: Story = { args: { code: 'SS3' } }
export const Task: Story = { args: { code: 'T42' } }
export const Question: Story = { args: { code: 'Q7' } }

export const AllVariants: Story = {
  render: () => ({
    components: { ShortCode },
    template: `
      <div style="display: flex; gap: 0.75rem; align-items: center; flex-wrap: wrap;">
        <ShortCode v-for="c in codes" :key="c" :code="c" />
      </div>
    `,
    setup() {
      return { codes: ['R1', 'R12', 'E15', 'E3', 'SS3', 'SS10', 'T42', 'T1', 'Q7', 'Q23'] }
    },
  }),
}
