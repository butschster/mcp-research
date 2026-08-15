import type { Meta, StoryObj } from '@storybook/vue3'
import ActiveSessionsGrid from './ActiveSessionsGrid.vue'
import { mockSession } from '../../__mocks__/session'
import { withShare, withoutShare } from '../../__mocks__/share'

/**
 * Interviews still running, as cards on the research overview.
 *
 * Each card links through `sessionPath()`, so the grid works under a share
 * link — but only where one includes sessions. The shell decides that; the grid
 * is never rendered with sessions switched off, rather than rendering itself
 * disabled.
 */
const meta: Meta<typeof ActiveSessionsGrid> = {
  title: 'Research/ActiveSessionsGrid',
  component: ActiveSessionsGrid,
  tags: ['autodocs'],
  // Share state is module state; this gives the ordinary stories a known
  // starting point rather than whatever the last story left behind. The
  // trade-offs are in __mocks__/share.ts.
  decorators: [withoutShare()],
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

/** Inside a share link that includes sessions: the cards point at
 *  `/s/{token}/session/SS1`. */
export const InsideAShare: Story = {
  decorators: [withShare({
    label: 'Client review, March',
    owner_name: 'Elena Marsh',
    research_id: 'res_007',
    research_code: 'R7',
    expires_at: null,
    include: { sessions: true, tasks: false, roadmaps: true, export: true },
  })],
  args: {
    sessions: [mockSession],
    researchSlug: 'R7',
  },
}
