import type { Meta, StoryObj } from '@storybook/vue3'
import KanbanBoard from './KanbanBoard.vue'
import { mockTask, mockTaskHigh, mockTaskCompleted, mockTaskFailed } from '../../__mocks__/task'

const columns = [
  { status: 'pending', label: 'Pending' },
  { status: 'in_progress', label: 'In Progress' },
  { status: 'completed', label: 'Completed' },
  { status: 'failed', label: 'Failed' },
]

const meta: Meta<typeof KanbanBoard> = {
  title: 'Tasks/KanbanBoard',
  component: KanbanBoard,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof KanbanBoard>

export const AllColumnsPopulated: Story = {
  args: {
    columns,
    tasks: [
      mockTask,
      { ...mockTask, id: 'task_005', code: 'T5', title: 'Write unit tests for composables', priority: 'low' },
      mockTaskHigh,
      mockTaskCompleted,
      mockTaskFailed,
    ],
    researchSlug: 'R1',
  },
}

export const SomeEmpty: Story = {
  args: {
    columns,
    tasks: [
      mockTask,
      mockTaskCompleted,
    ],
    researchSlug: 'R1',
  },
}

export const AllEmpty: Story = {
  args: {
    columns,
    tasks: [],
    researchSlug: 'R1',
  },
}
