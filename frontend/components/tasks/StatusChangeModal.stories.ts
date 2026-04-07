import type { Meta, StoryObj } from '@storybook/vue3'
import StatusChangeModal from './StatusChangeModal.vue'
import { mockTask } from '../../__mocks__/task'

const meta: Meta<typeof StatusChangeModal> = {
  title: 'Tasks/StatusChangeModal',
  component: StatusChangeModal,
  tags: ['autodocs'],
  argTypes: {
    visible: { control: 'boolean' },
    targetStatus: { control: 'select', options: ['pending', 'in_progress', 'completed', 'failed'] },
    statusLabel: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof StatusChangeModal>

export const MoveToInProgress: Story = {
  args: {
    visible: true,
    task: mockTask,
    targetStatus: 'in_progress',
    statusLabel: 'In Progress',
  },
}

export const MoveToCompleted: Story = {
  args: {
    visible: true,
    task: mockTask,
    targetStatus: 'completed',
    statusLabel: 'Completed',
  },
}

export const MoveToFailed: Story = {
  args: {
    visible: true,
    task: mockTask,
    targetStatus: 'failed',
    statusLabel: 'Failed',
  },
}

export const MoveToPending: Story = {
  args: {
    visible: true,
    task: mockTask,
    targetStatus: 'pending',
    statusLabel: 'Pending',
  },
}

export const Hidden: Story = {
  args: {
    visible: false,
    task: mockTask,
    targetStatus: 'in_progress',
    statusLabel: 'In Progress',
  },
}
