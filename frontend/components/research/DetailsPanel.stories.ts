import type { Meta, StoryObj } from '@storybook/vue3'
import DetailsPanel from './DetailsPanel.vue'
import { mockResearch } from '../../__mocks__/research'

const meta: Meta<typeof DetailsPanel> = {
  title: 'Research/DetailsPanel',
  component: DetailsPanel,
  tags: ['autodocs'],
  argTypes: {
    open: { control: 'boolean' },
  },
}
export default meta
type Story = StoryObj<typeof DetailsPanel>

export const Default: Story = {
  args: {
    open: true,
    research: mockResearch,
  },
}

export const Closed: Story = {
  args: {
    open: false,
    research: mockResearch,
  },
}

export const EmptyFields: Story = {
  args: {
    open: true,
    research: {
      ...mockResearch,
      goal: '',
      description: '',
      instruction: '',
      tags: [],
      memory: [],
    },
  },
}

export const FullData: Story = {
  args: {
    open: true,
    research: {
      ...mockResearch,
      goal: 'Investigate best practices for Vue 3 component design and composition patterns',
      description: 'A deep dive into Composition API patterns, component decomposition strategies, and reusable primitive extraction for large-scale Vue applications.',
      instruction: 'Focus on real-world scalability concerns. Compare Options API vs Composition API patterns. Document trade-offs.\n\nPrioritize patterns that work well with TypeScript. Include examples from open-source projects where possible.',
      memory: [
        'Composition API preferred over Options API for complex components',
        'Keep components under 200 lines for readability',
        'Extract shared primitives when pattern appears 3+ times',
        'Use provide/inject for deep prop drilling',
        'Prefer computed over watch for derived state',
      ],
      tags: ['vue', 'architecture', 'frontend', 'composables', 'patterns'],
    },
  },
}

export const LongInstruction: Story = {
  args: {
    open: true,
    research: {
      ...mockResearch,
      instruction: `You are a senior software architect helping with this research.\n\nGuidelines:\n- Be concise and technical\n- Use code examples when explaining patterns\n- Compare trade-offs objectively\n- Reference official documentation when possible\n- Flag potential pitfalls and edge cases\n\nAvoid:\n- Opinions without evidence\n- Framework evangelism\n- Ignoring performance implications`,
      memory: [],
    },
  },
}

export const ManyMemoryEntries: Story = {
  args: {
    open: true,
    research: {
      ...mockResearch,
      memory: [
        'User prefers TypeScript strict mode',
        'Project uses Nuxt 4 with auto-imports enabled',
        'Testing framework is Vitest with Vue Test Utils',
        'Avoid Vuex — project standardized on Pinia',
        'CSS approach: scoped styles + CSS custom properties',
        'Component naming: PascalCase files, kebab-case in templates',
        'State that crosses 3+ components should use composables',
        'API layer uses $fetch with typed responses',
        'Error handling follows Result pattern, not try/catch',
        'Performance budget: LCP under 2.5s, bundle under 200KB',
      ],
    },
  },
}

export const NoTagsNoMemory: Story = {
  args: {
    open: true,
    research: {
      ...mockResearch,
      tags: [],
      memory: [],
      instruction: '',
    },
  },
}
