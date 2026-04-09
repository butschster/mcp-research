import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapRefNode from './RoadmapRefNode.vue'

const meta: Meta<typeof RoadmapRefNode> = {
  title: 'Roadmap/RefNode',
  component: RoadmapRefNode,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 420px;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof RoadmapRefNode>

// --- Entry references ---

export const EntryRef: Story = {
  name: 'Entry Reference',
  args: {
    data: {
      code: 'N1',
      title: 'Architecture Overview',
      description: '',
      nodeType: 'step',
      status: 'in_progress',
      refType: 'entry',
      refId: 'entry-uuid-1',
      refData: {
        title: 'Microservices Architecture Overview',
        status: 'draft',
        code: 'E3',
        description: 'Comprehensive analysis of microservices patterns',
        section_name: 'Architecture',
        content: 'Service mesh provides transparent routing, load balancing, and observability between microservices without requiring application-level changes...',
      },
    },
  },
}

export const EntryRefCompleted: Story = {
  name: 'Entry Reference (completed)',
  args: {
    data: {
      code: 'N2',
      title: 'Database Design Doc',
      description: '',
      nodeType: 'step',
      status: 'completed',
      refType: 'entry',
      refId: 'entry-uuid-2',
      refData: {
        title: 'Database Schema Design',
        status: 'final',
        code: 'E7',
        section_name: 'Data Layer',
        content: 'The normalized schema uses 8 tables with foreign key constraints and cascade deletes for referential integrity...',
      },
    },
  },
}

export const EntryRefNoContent: Story = {
  name: 'Entry Reference (no content)',
  args: {
    data: {
      code: 'N3',
      title: 'API Endpoints',
      description: 'REST API documentation entry',
      nodeType: 'info',
      status: '',
      refType: 'entry',
      refId: 'entry-uuid-3',
      refData: {
        title: 'API Endpoints Reference',
        status: 'draft',
        code: 'E12',
        section_name: 'Documentation',
      },
    },
  },
}

// --- Task references ---

export const TaskRef: Story = {
  name: 'Task Reference',
  args: {
    data: {
      code: 'N4',
      title: 'Analyze competitors',
      description: '',
      nodeType: 'step',
      status: 'in_progress',
      refType: 'task',
      refId: 'task-uuid-1',
      refData: {
        title: 'Analyze top 5 competitor products',
        status: 'in_progress',
        code: 'T2',
        research_id: 'research-uuid',
        priority: 'high',
      },
    },
  },
}

export const TaskRefCompleted: Story = {
  name: 'Task Reference (completed with result)',
  args: {
    data: {
      code: 'N5',
      title: 'Set up CI/CD',
      description: '',
      nodeType: 'step',
      status: 'done',
      refType: 'task',
      refId: 'task-uuid-2',
      refData: {
        title: 'Configure GitHub Actions CI/CD pipeline',
        status: 'done',
        code: 'T5',
        priority: 'medium',
        result: 'Pipeline configured with build, test, and deploy stages. Average build time: 3m 20s.',
      },
    },
  },
}

export const TaskRefLowPriority: Story = {
  name: 'Task Reference (low priority)',
  args: {
    data: {
      code: 'N6',
      title: 'Update README',
      description: '',
      nodeType: 'step',
      status: 'todo',
      refType: 'task',
      refId: 'task-uuid-3',
      refData: {
        title: 'Update README with deployment instructions',
        status: 'todo',
        code: 'T8',
        priority: 'low',
      },
    },
  },
}

export const TaskRefCritical: Story = {
  name: 'Task Reference (critical)',
  args: {
    data: {
      code: 'N7',
      title: 'Fix auth vulnerability',
      description: '',
      nodeType: 'step',
      status: 'in_progress',
      refType: 'task',
      refId: 'task-uuid-4',
      refData: {
        title: 'Patch session token storage vulnerability',
        status: 'in_progress',
        code: 'T1',
        priority: 'critical',
      },
    },
  },
}

// --- Session references ---

export const SessionRef: Story = {
  name: 'Session Reference',
  args: {
    data: {
      code: 'N8',
      title: 'User interviews',
      description: '',
      nodeType: 'step',
      status: 'active',
      refType: 'session',
      refId: 'session-uuid-1',
      refData: {
        title: 'Deep-dive User Interviews',
        status: 'active',
        code: 'SS2',
        description: 'Understanding user pain points with current onboarding flow',
        total_questions: 12,
        answered_questions: 8,
      },
    },
  },
}

export const SessionRefComplete: Story = {
  name: 'Session Reference (all answered)',
  args: {
    data: {
      code: 'N9',
      title: 'Initial exploration',
      description: '',
      nodeType: 'step',
      status: 'completed',
      refType: 'session',
      refId: 'session-uuid-2',
      refData: {
        title: 'Initial Exploration Session',
        status: 'completed',
        code: 'SS1',
        description: 'Broad survey of the problem space',
        total_questions: 6,
        answered_questions: 6,
      },
    },
  },
}

export const SessionRefEmpty: Story = {
  name: 'Session Reference (no answers yet)',
  args: {
    data: {
      code: 'N10',
      title: 'Follow-up session',
      description: '',
      nodeType: 'step',
      status: 'not_started',
      refType: 'session',
      refId: 'session-uuid-3',
      refData: {
        title: 'Follow-up Technical Deep-dive',
        status: 'planned',
        code: 'SS3',
        description: 'Technical feasibility analysis',
        total_questions: 10,
        answered_questions: 0,
      },
    },
  },
}

// --- Research references ---

export const ResearchRef: Story = {
  name: 'Research Reference',
  args: {
    data: {
      code: 'N11',
      title: 'Market Analysis',
      description: '',
      nodeType: 'info',
      status: '',
      refType: 'research',
      refId: 'research-uuid-1',
      refData: {
        title: 'Competitive Market Analysis 2025',
        status: 'in_progress',
        code: 'R3',
        description: 'Deep analysis of market positioning and competitor strategies',
        section_count: 5,
        entry_count: 23,
      },
    },
  },
}

export const ResearchRefSmall: Story = {
  name: 'Research Reference (small)',
  args: {
    data: {
      code: 'N12',
      title: 'Tech stack research',
      description: '',
      nodeType: 'info',
      status: '',
      refType: 'research',
      refId: 'research-uuid-2',
      refData: {
        title: 'Frontend Framework Evaluation',
        status: 'completed',
        code: 'R1',
        description: 'Comparing React, Vue, and Svelte for our use case',
        section_count: 3,
        entry_count: 8,
      },
    },
  },
}

// --- Question references ---

export const QuestionRef: Story = {
  name: 'Question Reference',
  args: {
    data: {
      code: 'N13',
      title: 'Key question',
      description: '',
      nodeType: 'step',
      status: 'answered',
      refType: 'question',
      refId: 'question-uuid-1',
      refData: {
        title: 'What are the main scalability bottlenecks in the current architecture?',
        status: 'answered',
        code: 'Q3',
        description: 'Database connection pooling and lack of caching layer are the primary bottlenecks...',
      },
    },
  },
}

export const QuestionRefUnanswered: Story = {
  name: 'Question Reference (unanswered)',
  args: {
    data: {
      code: 'N14',
      title: 'Open question',
      description: '',
      nodeType: 'step',
      status: 'pending',
      refType: 'question',
      refId: 'question-uuid-2',
      refData: {
        title: 'How should we handle data migration from the legacy system?',
        status: 'pending',
        code: 'Q7',
      },
    },
  },
}

// --- Showcase: all ref types ---

export const AllRefTypes: Story = {
  render: () => ({
    components: { RoadmapRefNode },
    template: `
      <div style="display: flex; flex-direction: column; gap: 1rem; max-width: 420px;">
        <RoadmapRefNode v-for="n in nodes" :key="n.code" :data="n" />
      </div>
    `,
    setup() {
      const nodes = [
        {
          code: 'N1', title: 'Architecture doc', description: '', nodeType: 'step', status: 'draft',
          refType: 'entry', refId: 'e1',
          refData: { title: 'Architecture Overview', status: 'draft', code: 'E3', section_name: 'Design', content: 'Service mesh provides transparent routing...' },
        },
        {
          code: 'N2', title: 'Implement auth', description: '', nodeType: 'step', status: 'in_progress',
          refType: 'task', refId: 't1',
          refData: { title: 'Implement OAuth2 flow', status: 'in_progress', code: 'T2', priority: 'high' },
        },
        {
          code: 'N3', title: 'User interviews', description: '', nodeType: 'step', status: 'active',
          refType: 'session', refId: 's1',
          refData: { title: 'UX Research Session', status: 'active', code: 'SS1', total_questions: 8, answered_questions: 5 },
        },
        {
          code: 'N4', title: 'Related research', description: '', nodeType: 'info', status: '',
          refType: 'research', refId: 'r1',
          refData: { title: 'Market Analysis 2025', status: 'in_progress', code: 'R3', section_count: 5, entry_count: 23 },
        },
        {
          code: 'N5', title: 'Key question', description: '', nodeType: 'step', status: 'answered',
          refType: 'question', refId: 'q1',
          refData: { title: 'What are the scalability bottlenecks?', status: 'answered', code: 'Q3', description: 'DB pooling and caching are the main issues...' },
        },
      ]
      return { nodes }
    },
  }),
}

// --- Edge cases ---

export const RefWithNoRefData: Story = {
  name: 'Reference (entity not found)',
  args: {
    data: {
      code: 'N15',
      title: 'Deleted entry reference',
      description: 'The referenced entry may have been removed',
      nodeType: 'step',
      status: 'blocked',
      refType: 'entry',
      refId: 'deleted-uuid',
    },
  },
}

export const RefWithLongTitle: Story = {
  name: 'Reference (long title from entity)',
  args: {
    data: {
      code: 'N16',
      title: '',
      description: '',
      nodeType: 'step',
      status: 'in_progress',
      refType: 'task',
      refId: 'task-long',
      refData: {
        title: 'Implement comprehensive error handling strategy with custom error boundaries and global fallback mechanisms for all API endpoints',
        status: 'in_progress',
        code: 'T12',
        priority: 'medium',
      },
    },
  },
}
