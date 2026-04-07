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

export const Expanded: Story = {
  args: {
    open: true,
    research: mockResearch,
  },
}

export const Collapsed: Story = {
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

export const EditingGoal: Story = {
  args: {
    open: true,
    research: mockResearch,
  },
  play: async ({ canvasElement }) => {
    const goalField = canvasElement.querySelector('.detail-field')
    if (goalField) {
      goalField.dispatchEvent(new MouseEvent('dblclick', { bubbles: true }))
    }
  },
}
