import type { Meta, StoryObj } from '@storybook/vue3'
import EntryNode from './EntryNode.vue'

const meta: Meta<typeof EntryNode> = {
  title: 'Mindmap/EntryNode',
  component: EntryNode,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 440px;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof EntryNode>

export const Default: Story = {
  args: {
    data: {
      title: 'JWT Token Rotation Strategy',
      description: 'Implement short-lived access tokens with refresh token rotation to mitigate token theft',
      status: 'active',
      tags: ['security', 'jwt', 'auth'],
      createdAt: '2026-04-01T10:30:00Z',
      researchSlug: 'R1',
      entrySlug: 'E5',
    },
  },
}

export const ManyTags: Story = {
  args: {
    data: {
      title: 'Rate Limiting Implementation',
      description: 'Token bucket algorithm with Redis backend for distributed rate limiting',
      status: 'completed',
      tags: ['performance', 'redis', 'api', 'infrastructure', 'scalability'],
      createdAt: '2026-03-28T14:00:00Z',
      researchSlug: 'R2',
      entrySlug: 'E12',
    },
  },
}

export const NoTags: Story = {
  args: {
    data: {
      title: 'Initial observations on caching',
      description: 'Quick notes from the first pass review of the caching layer',
      status: 'draft',
      tags: [],
      createdAt: '2026-04-05T09:00:00Z',
      researchSlug: 'R1',
      entrySlug: 'E1',
    },
  },
}

export const NoDescription: Story = {
  args: {
    data: {
      title: 'CORS configuration review',
      description: '',
      status: 'active',
      tags: ['security'],
      createdAt: '2026-04-02T16:45:00Z',
      researchSlug: 'R3',
      entrySlug: 'E7',
    },
  },
}

export const LongTitle: Story = {
  args: {
    data: {
      title: 'Comprehensive analysis of the WebSocket reconnection logic and event deduplication mechanism',
      description: 'Detailed breakdown of how the client handles dropped connections and ensures no events are lost',
      status: 'in_progress',
      tags: ['websocket', 'reliability'],
      createdAt: '2026-04-03T11:20:00Z',
      researchSlug: 'R1',
      entrySlug: 'E9',
    },
  },
}

export const Pending: Story = {
  args: {
    data: {
      title: 'SQLite WAL mode tuning',
      description: 'Investigate optimal WAL checkpoint interval and page size for our workload',
      status: 'pending',
      tags: ['database', 'sqlite', 'performance'],
      createdAt: '2026-04-04T08:15:00Z',
      researchSlug: 'R4',
      entrySlug: 'E3',
    },
  },
}
