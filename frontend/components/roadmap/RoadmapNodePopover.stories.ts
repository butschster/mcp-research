import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapNodePopover from './RoadmapNodePopover.vue'

// The node detail modal (a ModalOverlay, not a positioned popover any more).
// It shows ref-entity fields, entity/node status chips, and — new — the stage
// and date placement controls. With auth off in Storybook canWrite is true, so
// the placement select + date input render.
const meta: Meta<typeof RoadmapNodePopover> = {
  title: 'Roadmap/NodePopover',
  component: RoadmapNodePopover,
  tags: ['autodocs'],
  args: {
    // The declared stage columns the placement <select> offers.
    stages: ['Discovery', 'Design', 'Build', 'Launch'],
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
  },
}

export const WithPlacement: Story = {
  name: 'Placement controls (stage + date range)',
  args: {
    node: {
      id: 'n2',
      title: 'Wire the API',
      description: 'Connect the frontend to the write endpoints.',
      nodeType: 'step',
      status: 'in_progress',
      stage: 'Build',
      node_date: '2026-03-10',
      node_end_date: '2026-05-05',
    },
    statuses: ['not_started', 'in_progress', 'completed'],
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
      stage: 'Launch',
      node_date: '2026-04-01',
    },
    statuses: ['not_started', 'in_progress', 'completed'],
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
  },
}

// --- Reference node popovers ---

export const EntryRefPopover: Story = {
  name: 'Entry Reference',
  args: {
    node: {
      id: 'n20',
      title: 'Architecture Overview',
      description: 'Main architecture document for the project',
      nodeType: 'step',
      status: 'in_progress',
      refType: 'entry',
      refId: 'entry-uuid',
      refData: {
        title: 'Microservices Architecture Overview',
        status: 'draft',
        code: 'E3',
        section_name: 'Architecture',
        content: 'Service mesh provides transparent routing, load balancing, and observability between microservices without requiring application-level changes...',
      },
    },
    statuses: ['not_started', 'in_progress', 'completed'],
  },
}

export const TaskRefPopover: Story = {
  name: 'Task Reference (with result)',
  args: {
    node: {
      id: 'n21',
      title: 'Set up CI/CD pipeline',
      description: '',
      nodeType: 'step',
      status: 'done',
      refType: 'task',
      refId: 'task-uuid',
      refData: {
        title: 'Configure GitHub Actions CI/CD',
        status: 'done',
        code: 'T5',
        priority: 'high',
        result: 'Pipeline configured with build, test, and deploy stages. Average build time: 3m 20s. Deploys to staging automatically on merge.',
      },
    },
    statuses: ['todo', 'in_progress', 'review', 'done'],
  },
}

export const SessionRefPopover: Story = {
  name: 'Session Reference',
  args: {
    node: {
      id: 'n22',
      title: 'User interviews',
      description: '',
      nodeType: 'step',
      status: 'active',
      refType: 'session',
      refId: 'session-uuid',
      refData: {
        title: 'Deep-dive User Interviews',
        status: 'active',
        code: 'SS2',
        total_questions: 12,
        answered_questions: 8,
      },
    },
    statuses: ['planned', 'active', 'completed'],
  },
}

export const ResearchRefPopover: Story = {
  name: 'Research Reference',
  args: {
    node: {
      id: 'n23',
      title: 'Related market analysis',
      description: '',
      nodeType: 'info',
      status: '',
      refType: 'research',
      refId: 'research-uuid',
      refData: {
        title: 'Competitive Market Analysis 2025',
        status: 'in_progress',
        code: 'R3',
        section_count: 5,
        entry_count: 23,
      },
    },
    statuses: [],
  },
}

export const QuestionRefPopover: Story = {
  name: 'Question Reference',
  args: {
    node: {
      id: 'n24',
      title: 'Key scalability question',
      description: '',
      nodeType: 'step',
      status: 'answered',
      refType: 'question',
      refId: 'question-uuid',
      refData: {
        title: 'What are the main scalability bottlenecks in the current architecture?',
        status: 'answered',
        code: 'Q3',
        description: 'Database connection pooling and lack of caching layer are the primary bottlenecks. Recommendation: add Redis cache and switch to connection pool with max 50 connections.',
      },
    },
    statuses: ['pending', 'answered', 'verified'],
  },
}
