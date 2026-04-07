import type { Meta, StoryObj } from '@storybook/vue3'
import CreateTaskModal from './CreateTaskModal.vue'

const meta: Meta<typeof CreateTaskModal> = {
  title: 'Tasks/CreateTaskModal',
  component: CreateTaskModal,
  tags: ['autodocs'],
  argTypes: {
    visible: { control: 'boolean' },
  },
}
export default meta
type Story = StoryObj<typeof CreateTaskModal>

export const EmptyForm: Story = {
  args: {
    visible: true,
  },
}

export const Hidden: Story = {
  args: {
    visible: false,
  },
}
