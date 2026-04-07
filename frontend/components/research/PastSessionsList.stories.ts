import type { Meta, StoryObj } from '@storybook/vue3'
import PastSessionsList from './PastSessionsList.vue'
import { mockSessionCompleted } from '../../__mocks__/session'

const meta: Meta<typeof PastSessionsList> = {
  title: 'Research/PastSessionsList',
  component: PastSessionsList,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof PastSessionsList>

const pastSessions = [
  mockSessionCompleted,
  { ...mockSessionCompleted, id: 'sess_003', code: 'SS3', title: 'Architecture Patterns Review', status: 'completed' },
  { ...mockSessionCompleted, id: 'sess_004', code: 'SS4', title: 'CSS Design System Audit', status: 'archived' },
]

export const Collapsed: Story = {
  args: {
    sessions: pastSessions,
    researchSlug: 'R1',
  },
}

export const Expanded: Story = {
  args: {
    sessions: pastSessions,
    researchSlug: 'R1',
  },
  play: async ({ canvasElement }) => {
    const toggle = canvasElement.querySelector('.past-sessions-toggle') as HTMLElement
    if (toggle) toggle.click()
  },
}

export const Empty: Story = {
  args: {
    sessions: [],
    researchSlug: 'R1',
  },
}
