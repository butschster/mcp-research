import type { Meta, StoryObj } from '@storybook/vue3'
import KanbanCard from './KanbanCard.vue'
import { mockTask, mockTaskHigh } from '../../__mocks__/task'

const meta: Meta<typeof KanbanCard> = {
  title: 'Tasks/KanbanCard',
  component: KanbanCard,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 280px"><story /></div>',
    }),
  ],
  argTypes: {
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof KanbanCard>

export const Default: Story = {
  args: {
    task: mockTask,
    researchSlug: 'R1',
  },
}

export const HighPriority: Story = {
  args: {
    task: mockTaskHigh,
    researchSlug: 'R1',
  },
}

export const LongTitle: Story = {
  args: {
    task: {
      ...mockTask,
      code: 'T99',
      title: 'Investigate the long-term implications of migrating from Options API to Composition API across all enterprise-level Vue components in the monorepo',
    },
    researchSlug: 'R1',
  },
}
