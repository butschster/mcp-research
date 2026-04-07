import type { Meta, StoryObj } from '@storybook/vue3'
import TaskDetailModal from './TaskDetailModal.vue'
import { mockTask, mockTaskHigh, mockTaskCompleted } from '../../__mocks__/task'

const meta: Meta<typeof TaskDetailModal> = {
  title: 'Tasks/TaskDetailModal',
  component: TaskDetailModal,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof TaskDetailModal>

export const FullTask: Story = {
  args: {
    task: {
      ...mockTaskCompleted,
      description: 'Check all components for line count and identify candidates for decomposition.\n\n- Components over 200 lines\n- Deeply nested templates\n- Mixed concerns',
    },
    researchSlug: 'R1',
  },
}

export const EmptyDescription: Story = {
  args: {
    task: {
      ...mockTask,
      description: null,
      result: null,
    },
    researchSlug: 'R1',
  },
}

export const HighPriorityTask: Story = {
  args: {
    task: mockTaskHigh,
    researchSlug: 'R1',
  },
}

export const Closed: Story = {
  args: {
    task: null,
    researchSlug: 'R1',
  },
}
