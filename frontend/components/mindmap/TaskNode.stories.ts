import type { Meta, StoryObj } from '@storybook/vue3'
import TaskNode from './TaskNode.vue'

const meta: Meta<typeof TaskNode> = {
  title: 'Mindmap/TaskNode',
  component: TaskNode,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 440px;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof TaskNode>

export const Pending: Story = {
  args: {
    data: {
      code: 'T1',
      title: 'Review CORS configuration for production endpoints',
      result: '',
      status: 'pending',
      priority: 'high',
    },
  },
}

export const InProgress: Story = {
  args: {
    data: {
      code: 'T3',
      title: 'Benchmark SQLite WAL checkpoint intervals',
      result: '',
      status: 'in_progress',
      priority: 'medium',
    },
  },
}

export const Completed: Story = {
  args: {
    data: {
      code: 'T2',
      title: 'Document OAuth2 PKCE flow for external clients',
      result: 'Added sequence diagrams and example curl commands to the API docs section',
      status: 'completed',
      priority: 'high',
    },
  },
}

export const LowPriority: Story = {
  args: {
    data: {
      code: 'T7',
      title: 'Add optional request logging middleware',
      result: '',
      status: 'pending',
      priority: 'low',
    },
  },
}

export const Failed: Story = {
  args: {
    data: {
      code: 'T5',
      title: 'Migrate to pgx driver for PostgreSQL support',
      result: 'Blocked by dependency on SQLite-specific features in the migration layer',
      status: 'failed',
      priority: 'high',
    },
  },
}

export const NoPriority: Story = {
  args: {
    data: {
      code: 'T4',
      title: 'Clean up unused API handler stubs',
      result: '',
      status: 'pending',
      priority: '',
    },
  },
}

export const LongTitle: Story = {
  args: {
    data: {
      code: 'T9',
      title: 'Investigate and document all edge cases in the cross-reference resolution logic when references span multiple research projects',
      result: '',
      status: 'in_progress',
      priority: 'medium',
    },
  },
}

export const NoCode: Story = {
  args: {
    data: {
      code: '',
      title: 'Quick fix for date formatting',
      result: 'Applied ISO 8601 format consistently across all API responses',
      status: 'completed',
      priority: 'low',
    },
  },
}
