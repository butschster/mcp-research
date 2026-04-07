import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapCard from './RoadmapCard.vue'

const meta: Meta<typeof RoadmapCard> = {
  title: 'Roadmap/RoadmapCard',
  component: RoadmapCard,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 400px;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof RoadmapCard>

export const Active: Story = {
  args: {
    roadmap: {
      code: 'RM1',
      title: 'Vue 3 Learning Path',
      description: 'Step-by-step guide from basics to advanced patterns in Vue 3 with Composition API',
      status: 'active',
      statuses: ['not_started', 'learning', 'practiced', 'mastered'],
      nodeCount: 12,
      edgeCount: 14,
    },
  },
}

export const Completed: Story = {
  args: {
    roadmap: {
      code: 'RM2',
      title: 'Marketing Strategy Q1',
      description: 'Content marketing roadmap for product launch — completed all milestones',
      status: 'completed',
      statuses: ['planned', 'approved', 'launched'],
      nodeCount: 8,
      edgeCount: 9,
    },
  },
}

export const Archived: Story = {
  args: {
    roadmap: {
      code: 'RM3',
      title: 'Legacy Migration Plan',
      description: 'Migration from Angular.js to Vue 3 — archived after pivot to new architecture',
      status: 'archived',
      statuses: ['todo', 'migrating', 'verified', 'deployed'],
      nodeCount: 23,
      edgeCount: 28,
    },
  },
}

export const NoStatuses: Story = {
  args: {
    roadmap: {
      code: 'RM4',
      title: 'System Architecture Overview',
      description: 'Visual graph of service dependencies — no status tracking needed',
      status: 'active',
      statuses: [],
      nodeCount: 6,
      edgeCount: 7,
    },
  },
}

export const NoDescription: Story = {
  args: {
    roadmap: {
      code: 'RM5',
      title: 'Quick brainstorm graph',
      description: '',
      status: 'active',
      statuses: [],
      nodeCount: 3,
      edgeCount: 2,
    },
  },
}

export const ManyStatuses: Story = {
  args: {
    roadmap: {
      code: 'RM6',
      title: 'Onboarding Flow',
      description: 'New hire onboarding process with multi-stage progression tracking',
      status: 'active',
      statuses: ['not_started', 'scheduled', 'in_progress', 'review', 'completed', 'skipped'],
      nodeCount: 18,
      edgeCount: 22,
    },
  },
}

export const LongTitle: Story = {
  args: {
    roadmap: {
      code: 'RM7',
      title: 'Comprehensive full-stack web development learning roadmap from fundamentals to production deployment',
      description: 'Covers HTML/CSS, JavaScript, TypeScript, Vue 3, Node.js, databases, testing, CI/CD, and cloud deployment',
      status: 'active',
      statuses: ['not_started', 'in_progress', 'completed'],
      nodeCount: 45,
      edgeCount: 52,
    },
  },
}

export const AllStatuses: Story = {
  render: () => ({
    components: { RoadmapCard },
    template: `
      <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 1rem;">
        <RoadmapCard v-for="rm in roadmaps" :key="rm.code" :roadmap="rm" />
      </div>
    `,
    setup() {
      const roadmaps = [
        {
          code: 'RM1', title: 'Active roadmap', description: 'Currently being worked on',
          status: 'active', statuses: ['todo', 'doing', 'done'], nodeCount: 10, edgeCount: 12,
        },
        {
          code: 'RM2', title: 'Completed roadmap', description: 'All steps finished',
          status: 'completed', statuses: ['done'], nodeCount: 5, edgeCount: 4,
        },
        {
          code: 'RM3', title: 'Archived roadmap', description: 'No longer relevant',
          status: 'archived', statuses: [], nodeCount: 15, edgeCount: 18,
        },
      ]
      return { roadmaps }
    },
  }),
}
