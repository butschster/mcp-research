import type { Meta, StoryObj } from '@storybook/vue3'
import ActiveSessionsGrid from './ActiveSessionsGrid.vue'
import { mockSession } from '../../__mocks__/session'

const meta: Meta<typeof ActiveSessionsGrid> = {
  title: 'Research/ActiveSessionsGrid',
  component: ActiveSessionsGrid,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof ActiveSessionsGrid>

export const SingleSession: Story = {
  args: {
    sessions: [mockSession],
    researchSlug: 'R1',
  },
}

export const ThreeSessions: Story = {
  args: {
    sessions: [
      mockSession,
      { ...mockSession, id: 'sess_010', code: 'SS10', title: 'Follow-up: Composable Testing' },
      { ...mockSession, id: 'sess_011', code: 'SS11', title: 'Performance Benchmarking Session' },
    ],
    researchSlug: 'R1',
  },
}

export const Empty: Story = {
  args: {
    sessions: [],
    researchSlug: 'R1',
  },
}
