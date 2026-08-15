import type { Meta, StoryObj } from '@storybook/vue3'
import PastSessionsList from './PastSessionsList.vue'
import { mockSessionCompleted } from '../../__mocks__/session'
import { includeEverything, mockShareMeta, withShare, withoutShare } from '../../__mocks__/share'

/**
 * Closed interviews, folded away behind a count.
 *
 * Same destination helper as the active grid — `sessionPath()` — so the list
 * behaves under a share link, where it is rendered only when the link includes
 * sessions.
 */
const meta: Meta<typeof PastSessionsList> = {
  title: 'Research/PastSessionsList',
  component: PastSessionsList,
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
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
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

/** Inside a share link that includes sessions, expanded so the destinations are
 *  visible: every row points at `/s/{token}/session/…`. */
export const InsideAShare: Story = {
  decorators: [withShare({ ...mockShareMeta, include: includeEverything })],
  args: {
    sessions: pastSessions,
    researchSlug: 'R7',
  },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    const toggle = canvasElement.querySelector('.past-sessions-toggle') as HTMLElement
    if (toggle) toggle.click()
  },
}
