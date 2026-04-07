import type { Meta, StoryObj } from '@storybook/vue3'
import PrintButton from './PrintButton.vue'

const meta: Meta<typeof PrintButton> = {
  title: 'Base/PrintButton',
  component: PrintButton,
  tags: ['autodocs'],
}
export default meta
type Story = StoryObj<typeof PrintButton>

export const Default: Story = {}
