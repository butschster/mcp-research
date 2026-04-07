import type { Meta, StoryObj } from '@storybook/vue3'
import SectionNode from './SectionNode.vue'

const meta: Meta<typeof SectionNode> = {
  title: 'Mindmap/SectionNode',
  component: SectionNode,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 380px;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof SectionNode>

export const Default: Story = {
  args: {
    data: {
      code: 'S1',
      name: 'Authentication Flow',
      description: 'Analysis of the current authentication and authorization mechanisms',
      status: 'active',
      entryCount: 8,
    },
  },
}

export const Completed: Story = {
  args: {
    data: {
      code: 'S3',
      name: 'Database Schema',
      description: 'Review of table structures, indexes, and migration strategy',
      status: 'completed',
      entryCount: 12,
    },
  },
}

export const NoDescription: Story = {
  args: {
    data: {
      code: 'S2',
      name: 'Performance Benchmarks',
      description: '',
      status: 'in_progress',
      entryCount: 3,
    },
  },
}

export const LongDescription: Story = {
  args: {
    data: {
      code: 'S5',
      name: 'Error Handling Patterns',
      description: 'Comprehensive review of error propagation, retry logic, circuit breakers, and graceful degradation across all service boundaries',
      status: 'draft',
      entryCount: 0,
    },
  },
}

export const NoCode: Story = {
  args: {
    data: {
      code: '',
      name: 'Miscellaneous Notes',
      description: 'Uncategorized findings and observations',
      status: 'active',
      entryCount: 5,
    },
  },
}

export const ManyEntries: Story = {
  args: {
    data: {
      code: 'S4',
      name: 'API Endpoints',
      description: 'Documentation of all REST and WebSocket endpoints',
      status: 'active',
      entryCount: 45,
    },
  },
}
