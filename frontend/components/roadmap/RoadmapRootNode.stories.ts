import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapRootNode from './RoadmapRootNode.vue'

const meta: Meta<typeof RoadmapRootNode> = {
  title: 'Roadmap/RootNode',
  component: RoadmapRootNode,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 420px;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof RoadmapRootNode>

export const Default: Story = {
  args: {
    data: {
      code: 'RM1',
      title: 'Vue 3 Learning Path',
      description: 'Step-by-step guide from zero to production-ready Vue 3 development',
      status: 'active',
      statuses: ['not_started', 'learning', 'practiced', 'mastered'],
      nodeCount: 12,
      edgeCount: 14,
    },
  },
}

export const Completed: Story = {
  args: {
    data: {
      code: 'RM2',
      title: 'Go Backend Mastery',
      description: 'Covered all major backend patterns — REST, gRPC, middleware, testing, deployment',
      status: 'completed',
      statuses: ['done'],
      nodeCount: 8,
      edgeCount: 9,
    },
  },
}

export const NoStatuses: Story = {
  args: {
    data: {
      code: 'RM3',
      title: 'Architecture Dependencies',
      description: 'Service dependency graph for the platform — no progression tracking',
      status: 'active',
      statuses: [],
      nodeCount: 6,
      edgeCount: 7,
    },
  },
}

export const NoDescription: Story = {
  args: {
    data: {
      code: 'RM4',
      title: 'Quick Brainstorm',
      description: '',
      status: 'active',
      statuses: [],
      nodeCount: 3,
      edgeCount: 2,
    },
  },
}

export const LargeGraph: Story = {
  args: {
    data: {
      code: 'RM5',
      title: 'Full-Stack Developer Roadmap 2026',
      description: 'Complete learning path covering frontend, backend, DevOps, and system design',
      status: 'active',
      statuses: ['not_started', 'in_progress', 'review', 'completed', 'skipped'],
      nodeCount: 48,
      edgeCount: 63,
    },
  },
}

export const Archived: Story = {
  args: {
    data: {
      code: 'RM6',
      title: 'Legacy Migration Plan',
      description: 'Archived after pivot to microservices — superseded by RM7',
      status: 'archived',
      statuses: ['planned', 'migrating', 'verified'],
      nodeCount: 20,
      edgeCount: 24,
    },
  },
}
