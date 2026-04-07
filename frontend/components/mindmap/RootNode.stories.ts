import type { Meta, StoryObj } from '@storybook/vue3'
import RootNode from './RootNode.vue'

const meta: Meta<typeof RootNode> = {
  title: 'Mindmap/RootNode',
  component: RootNode,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 420px;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof RootNode>

export const Default: Story = {
  args: {
    data: {
      name: 'OAuth2 Security Audit',
      goal: 'Evaluate the security posture of the OAuth2 implementation including token handling and PKCE flow',
      status: 'active',
      sectionCount: 5,
      entryCount: 23,
      questionCount: 12,
      taskCount: 7,
    },
  },
}

export const Completed: Story = {
  args: {
    data: {
      name: 'API Rate Limiting Strategy',
      goal: 'Design and document rate limiting approach for public REST API endpoints',
      status: 'completed',
      sectionCount: 3,
      entryCount: 14,
      questionCount: 6,
      taskCount: 4,
    },
  },
}

export const NoGoal: Story = {
  args: {
    data: {
      name: 'Quick Investigation',
      goal: '',
      status: 'draft',
      sectionCount: 1,
      entryCount: 2,
      questionCount: 0,
      taskCount: 0,
    },
  },
}

export const LongGoal: Story = {
  args: {
    data: {
      name: 'Distributed Systems Patterns',
      goal: 'Investigate consensus algorithms, eventual consistency models, and partition tolerance strategies for a multi-region deployment architecture with strict latency requirements',
      status: 'in_progress',
      sectionCount: 8,
      entryCount: 47,
      questionCount: 31,
      taskCount: 15,
    },
  },
}

export const Archived: Story = {
  args: {
    data: {
      name: 'Legacy Migration Plan',
      goal: 'Document migration path from monolith to microservices',
      status: 'archived',
      sectionCount: 6,
      entryCount: 38,
      questionCount: 19,
      taskCount: 12,
    },
  },
}
