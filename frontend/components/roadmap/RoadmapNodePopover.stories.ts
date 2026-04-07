import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapNodePopover from './RoadmapNodePopover.vue'

const meta: Meta<typeof RoadmapNodePopover> = {
  title: 'Roadmap/NodePopover',
  component: RoadmapNodePopover,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="position: relative; min-height: 300px; padding: 2rem;"><story /></div>',
    }),
  ],
  // Override fixed positioning for storybook display
  parameters: {
    layout: 'padded',
  },
}
export default meta
type Story = StoryObj<typeof RoadmapNodePopover>

export const StepWithStatus: Story = {
  args: {
    node: {
      id: 'n1',
      title: 'Learn Vue 3 Composition API',
      description: 'Understand ref, reactive, computed, watch, and lifecycle hooks in the Composition API',
      nodeType: 'step',
      status: 'in_progress',
    },
    statuses: ['not_started', 'in_progress', 'completed'],
    position: { x: 20, y: 20 },
  },
}

export const MilestoneCompleted: Story = {
  args: {
    node: {
      id: 'n3',
      title: 'Frontend Fundamentals Complete',
      description: 'All core frontend skills mastered — ready to move to advanced patterns',
      nodeType: 'milestone',
      status: 'completed',
    },
    statuses: ['not_started', 'in_progress', 'completed'],
    position: { x: 20, y: 20 },
  },
}

export const DecisionPending: Story = {
  args: {
    node: {
      id: 'n4',
      title: 'Choose state management approach',
      description: 'Decide between Pinia, Vuex, or composable-based state management',
      nodeType: 'decision',
      status: 'not_started',
    },
    statuses: ['not_started', 'evaluating', 'decided'],
    position: { x: 20, y: 20 },
  },
}

export const InfoNoStatus: Story = {
  args: {
    node: {
      id: 'n7',
      title: 'TypeScript recommended',
      description: 'TypeScript greatly improves DX with Vue 3 and Volar',
      nodeType: 'info',
      status: '',
    },
    statuses: ['not_started', 'in_progress', 'completed'],
    position: { x: 20, y: 20 },
  },
}

export const NoDescription: Story = {
  args: {
    node: {
      id: 'n10',
      title: 'Install dependencies',
      description: '',
      nodeType: 'step',
      status: 'completed',
    },
    statuses: ['not_started', 'completed'],
    position: { x: 20, y: 20 },
  },
}

export const ManyStatuses: Story = {
  args: {
    node: {
      id: 'n5',
      title: 'Database Migration',
      description: 'Migrate from SQLite to PostgreSQL for production',
      nodeType: 'step',
      status: 'review',
    },
    statuses: ['not_started', 'planning', 'in_progress', 'review', 'testing', 'deployed'],
    position: { x: 20, y: 20 },
  },
}

export const NoStatuses: Story = {
  args: {
    node: {
      id: 'n8',
      title: 'Architecture diagram node',
      description: 'Purely structural — no progress tracking on this roadmap',
      nodeType: 'group',
      status: '',
    },
    statuses: [],
    position: { x: 20, y: 20 },
  },
}

export const LongDescription: Story = {
  args: {
    node: {
      id: 'n12',
      title: 'Implement comprehensive error handling',
      description: 'Cover synchronous errors, async rejections, component-level error boundaries with Vue errorCaptured hook, global fallback via app.config.errorHandler, network failure recovery with retry logic, and user-friendly error display components. Document all error codes and recovery strategies.',
      nodeType: 'step',
      status: 'not_started',
    },
    statuses: ['not_started', 'in_progress', 'completed'],
    position: { x: 20, y: 20 },
  },
}

export const CustomDomainStatuses: Story = {
  name: 'Marketing Statuses',
  args: {
    node: {
      id: 'n1',
      title: 'Content Marketing Campaign',
      description: 'SEO blog posts, case studies, comparison landing pages',
      nodeType: 'step',
      status: 'approved',
    },
    statuses: ['planned', 'approved', 'in_progress', 'launched'],
    position: { x: 20, y: 20 },
  },
}
