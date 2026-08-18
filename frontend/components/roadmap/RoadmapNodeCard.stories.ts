import type { Meta, StoryObj } from '@storybook/vue3'
import RoadmapNodeCard from './RoadmapNodeCard.vue'

// The one card shared by all three roadmap views (graph nodes wrap it with Vue
// Flow <Handle>s; the board and timeline render it directly). Each story feeds
// the camelCase RoadmapCardData the graph's buildGraph produces.
const meta: Meta<typeof RoadmapNodeCard> = {
  title: 'Roadmap/NodeCard',
  component: RoadmapNodeCard,
  tags: ['autodocs'],
  argTypes: {
    compact: { control: 'boolean' },
    highlighted: { control: 'boolean' },
    dimmed: { control: 'boolean' },
  },
  decorators: [
    () => ({ template: '<div style="max-width: 340px;"><story /></div>' }),
  ],
}
export default meta
type Story = StoryObj<typeof RoadmapNodeCard>

export const EntryRef: Story = {
  name: 'Entry reference (content preview)',
  args: {
    data: {
      code: 'N1',
      title: 'Landscape review',
      nodeType: 'step',
      status: 'completed',
      refType: 'entry',
      refId: 'entry-uuid-1',
      refData: {
        title: 'Competitive Landscape Review',
        status: 'completed',
        code: 'E3',
        section_name: 'Discovery',
        content: 'Service mesh gives transparent routing and observability without touching application code, which is where the incumbents stop.',
      },
    },
  },
}

export const TaskRef: Story = {
  name: 'Task reference (result)',
  args: {
    data: {
      code: 'N5',
      title: 'Wire the API',
      nodeType: 'step',
      status: 'in_progress',
      refType: 'task',
      refId: 'task-uuid-1',
      refData: {
        title: 'Configure GitHub Actions CI/CD pipeline',
        status: 'in_progress',
        code: 'T5',
        priority: 'high',
        result: 'Build, test and deploy stages green. Average build time 3m 20s, deploys to staging on merge.',
      },
    },
  },
}

export const SessionRef: Story = {
  name: 'Session reference (progress bar)',
  args: {
    data: {
      code: 'N2',
      title: 'User interviews',
      nodeType: 'step',
      status: 'active',
      refType: 'session',
      refId: 'session-uuid-1',
      refData: {
        title: 'Deep-dive User Interviews',
        status: 'active',
        code: 'SS2',
        total_questions: 12,
        answered_questions: 8,
      },
    },
  },
}

export const PlainStep: Story = {
  name: 'Plain step',
  args: {
    data: {
      code: 'N4',
      title: 'Design the component system',
      description: 'Design the shared card, badges, and layout primitives before any screen is built.',
      nodeType: 'step',
      status: 'in_progress',
    },
  },
}

export const Milestone: Story = {
  args: {
    data: {
      code: 'N6',
      title: 'Public beta',
      description: 'First cohort of external users onboarded.',
      nodeType: 'milestone',
      status: 'pending',
    },
  },
}

export const Decision: Story = {
  args: {
    data: {
      code: 'N3',
      title: 'Pick a state layer',
      description: 'Decide between Pinia, composables, or a bespoke store.',
      nodeType: 'decision',
      status: 'pending',
    },
  },
}

export const WithDependencies: Story = {
  name: 'With dependency chips',
  args: {
    data: {
      code: 'N5',
      title: 'Wire the API',
      description: 'Connect the frontend to the write endpoints.',
      nodeType: 'step',
      status: 'pending',
    },
    deps: [{ code: 'N3' }, { code: 'N4' }],
  },
}

export const Compact: Story = {
  name: 'Compact (timeline density)',
  args: {
    compact: true,
    data: {
      code: 'N1',
      title: 'Landscape review',
      nodeType: 'step',
      status: 'completed',
      refType: 'entry',
      refId: 'entry-uuid-1',
      refData: {
        title: 'Competitive Landscape Review',
        status: 'completed',
        code: 'E3',
        section_name: 'Discovery',
        content: 'This preview line is stripped in compact mode.',
      },
    },
  },
}

export const Highlighted: Story = {
  args: {
    highlighted: true,
    data: {
      code: 'N4',
      title: 'Design the component system',
      description: 'Highlighted by the board when a related card is hovered.',
      nodeType: 'step',
      status: 'in_progress',
    },
  },
}

export const Dimmed: Story = {
  args: {
    dimmed: true,
    data: {
      code: 'N7',
      title: 'Unrelated work',
      description: 'Dimmed while another card and its neighbours are highlighted.',
      nodeType: 'step',
      status: 'pending',
    },
  },
}

export const Overloaded: Story = {
  name: 'Overloaded (long title, deps)',
  args: {
    data: {
      code: 'N12',
      title: '',
      nodeType: 'step',
      status: 'in_progress',
      refType: 'task',
      refId: 'task-long',
      refData: {
        title: 'Implement comprehensive error handling strategy with custom boundaries and a global fallback for every API endpoint',
        status: 'in_progress',
        code: 'T12',
        priority: 'critical',
        result: 'Covered synchronous errors, async rejections, component-level boundaries via errorCaptured, and network retry with backoff across the whole surface.',
      },
    },
    deps: [{ code: 'N3' }, { code: 'N4' }, { code: 'N5' }, { code: 'N8' }],
  },
}

export const AllTypes: Story = {
  render: () => ({
    components: { RoadmapNodeCard },
    setup() {
      const cards = [
        { code: 'N1', title: 'Landscape review', nodeType: 'step', status: 'completed', refType: 'entry', refId: 'e1', refData: { title: 'Landscape Review', status: 'completed', code: 'E3', content: 'Service mesh gives transparent routing...' } },
        { code: 'N5', title: 'Wire the API', nodeType: 'step', status: 'in_progress', refType: 'task', refId: 't1', refData: { title: 'CI/CD pipeline', status: 'in_progress', code: 'T5', priority: 'high', result: 'Deploys to staging on merge.' } },
        { code: 'N2', title: 'User interviews', nodeType: 'step', status: 'active', refType: 'session', refId: 's1', refData: { title: 'User Interviews', status: 'active', code: 'SS2', total_questions: 12, answered_questions: 8 } },
        { code: 'N9', title: 'Market analysis', nodeType: 'info', status: '', refType: 'research', refId: 'r1', refData: { title: 'Market Analysis 2026', status: 'in_progress', code: 'R3', section_count: 5, entry_count: 23 } },
        { code: 'N4', title: 'Design the component system', description: 'Shared card, badges, layout.', nodeType: 'step', status: 'in_progress' },
        { code: 'N6', title: 'Public beta', description: 'First external cohort.', nodeType: 'milestone', status: 'pending' },
        { code: 'N3', title: 'Pick a state layer', description: 'Pinia vs composables.', nodeType: 'decision', status: 'pending' },
      ]
      return { cards }
    },
    template: `
      <div style="display: grid; grid-template-columns: repeat(2, 320px); gap: 1rem;">
        <RoadmapNodeCard v-for="c in cards" :key="c.code" :data="c" />
      </div>
    `,
  }),
}
