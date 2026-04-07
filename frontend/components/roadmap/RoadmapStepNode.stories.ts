import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapStepNode from './RoadmapStepNode.vue'

const meta: Meta<typeof RoadmapStepNode> = {
  title: 'Roadmap/StepNode',
  component: RoadmapStepNode,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 400px;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof RoadmapStepNode>

export const Step: Story = {
  args: {
    data: {
      code: 'N1',
      title: 'Set up development environment',
      description: 'Install Node.js, configure ESLint and Prettier, set up the project scaffold with Vite',
      nodeType: 'step',
      status: 'completed',
    },
  },
}

export const StepInProgress: Story = {
  args: {
    data: {
      code: 'N2',
      title: 'Learn Vue 3 Composition API',
      description: 'Understand ref, reactive, computed, watch, and lifecycle hooks in the Composition API',
      nodeType: 'step',
      status: 'in_progress',
    },
  },
}

export const StepNotStarted: Story = {
  args: {
    data: {
      code: 'N5',
      title: 'Build a production-ready SPA',
      description: 'Apply all learned concepts to build and deploy a complete single-page application',
      nodeType: 'step',
      status: 'not_started',
    },
  },
}

export const Milestone: Story = {
  args: {
    data: {
      code: 'N3',
      title: 'Frontend Fundamentals Complete',
      description: 'All core frontend skills mastered — ready to move to advanced patterns',
      nodeType: 'milestone',
      status: 'completed',
    },
  },
}

export const MilestoneNotReached: Story = {
  args: {
    data: {
      code: 'N8',
      title: 'Production Deployment',
      description: 'Application deployed to production with CI/CD pipeline and monitoring',
      nodeType: 'milestone',
      status: 'not_started',
    },
  },
}

export const Decision: Story = {
  args: {
    data: {
      code: 'N4',
      title: 'Choose state management approach',
      description: 'Decide between Pinia, Vuex, or composable-based state management for the project',
      nodeType: 'decision',
      status: 'in_progress',
    },
  },
}

export const DecisionCompleted: Story = {
  args: {
    data: {
      code: 'N6',
      title: 'Select deployment target',
      description: 'Chose Vercel for frontend hosting — serverless functions for API',
      nodeType: 'decision',
      status: 'completed',
    },
  },
}

export const Info: Story = {
  args: {
    data: {
      code: 'N7',
      title: 'Prerequisites',
      description: 'Basic knowledge of HTML, CSS, and JavaScript is required before starting this path',
      nodeType: 'info',
      status: '',
    },
  },
}

export const Group: Story = {
  args: {
    data: {
      code: 'N9',
      title: 'Advanced Patterns',
      description: 'Optional deep-dive topics for experienced developers',
      nodeType: 'group',
      status: '',
    },
  },
}

export const NoDescription: Story = {
  args: {
    data: {
      code: 'N10',
      title: 'Install dependencies',
      description: '',
      nodeType: 'step',
      status: 'completed',
    },
  },
}

export const NoStatus: Story = {
  args: {
    data: {
      code: 'N11',
      title: 'Reference material on TypeScript generics',
      description: 'Collection of links and examples for advanced TypeScript usage',
      nodeType: 'info',
      status: '',
    },
  },
}

export const NoCode: Story = {
  args: {
    data: {
      code: '',
      title: 'Quick note about browser compatibility',
      description: 'Safari requires polyfills for some Web API features used in this project',
      nodeType: 'info',
      status: '',
    },
  },
}

export const LongTitle: Story = {
  args: {
    data: {
      code: 'N12',
      title: 'Implement comprehensive error handling strategy with custom error boundaries and global fallback mechanisms',
      description: 'Cover synchronous errors, async rejections, component-level boundaries, and network failure recovery',
      nodeType: 'step',
      status: 'not_started',
    },
  },
}

export const CustomStatuses: Story = {
  name: 'Custom Statuses (learning path)',
  args: {
    data: {
      code: 'N13',
      title: 'Master CSS Grid Layout',
      description: 'Complete exercises on grid-template-areas, auto-fit/auto-fill, and subgrid',
      nodeType: 'step',
      status: 'learning',
    },
  },
}

export const MasteredStatus: Story = {
  args: {
    data: {
      code: 'N14',
      title: 'HTML Semantic Elements',
      description: 'Proper usage of article, section, aside, nav, header, footer, and main',
      nodeType: 'step',
      status: 'mastered',
    },
  },
}

export const AllNodeTypes: Story = {
  render: () => ({
    components: { RoadmapStepNode },
    template: `
      <div style="display: flex; flex-direction: column; gap: 1rem; max-width: 400px;">
        <RoadmapStepNode v-for="n in nodes" :key="n.code" :data="n" />
      </div>
    `,
    setup() {
      const nodes = [
        { code: 'N1', title: 'Step node', description: 'Regular step in the path', nodeType: 'step', status: 'completed' },
        { code: 'N2', title: 'Milestone node', description: 'Key achievement point', nodeType: 'milestone', status: 'in_progress' },
        { code: 'N3', title: 'Decision node', description: 'Fork in the road — pick a direction', nodeType: 'decision', status: 'not_started' },
        { code: 'N4', title: 'Info node', description: 'Supplementary reference material', nodeType: 'info', status: '' },
        { code: 'N5', title: 'Group node', description: 'Container for related steps', nodeType: 'group', status: '' },
      ]
      return { nodes }
    },
  }),
}
