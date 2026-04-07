import type { Meta, StoryObj } from '@storybook/vue3'
import GroupLabelNode from './GroupLabelNode.vue'

const meta: Meta<typeof GroupLabelNode> = {
  title: 'Mindmap/GroupLabelNode',
  component: GroupLabelNode,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 320px;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof GroupLabelNode>

export const Questions: Story = {
  args: {
    data: {
      label: 'Questions',
      count: 12,
      icon: 'question',
    },
  },
}

export const Tasks: Story = {
  args: {
    data: {
      label: 'Tasks',
      count: 7,
      icon: 'task',
    },
  },
}

export const ZeroCount: Story = {
  args: {
    data: {
      label: 'Questions',
      count: 0,
      icon: 'question',
    },
  },
}

export const LargeCount: Story = {
  args: {
    data: {
      label: 'Tasks',
      count: 128,
      icon: 'task',
    },
  },
}
