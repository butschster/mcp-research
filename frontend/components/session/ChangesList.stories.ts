import type { Meta, StoryObj } from '@storybook/vue3'
import ChangesList from './ChangesList.vue'
import { mockApi, neverResolves } from '../../__mocks__/api'
import {
  mockSessionChanges,
  mockSessionChangesMany,
  mockChangeModified,
  mockChangeCreated,
  mockChangeCreatedSmall,
  type SessionEntryChange,
} from '../../__mocks__/revision'

/**
 * What a session did to the research's entries — the ones it created, the ones
 * it edited, and the diff for each.
 *
 * The component fetches `GET /api/sessions/{id}/changes` itself on mount, so
 * these stories render the real component against a mocked `authFetch` (see
 * `__mocks__/api.ts`). Diffs are collapsed until "Show changes" is clicked.
 */
const meta: Meta<typeof ChangesList> = {
  title: 'Session/ChangesList',
  component: ChangesList,
  tags: ['autodocs'],
  argTypes: {
    sessionId: { control: 'text' },
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof ChangesList>

type ChangesSetup = {
  changes: SessionEntryChange[]
  pending?: boolean
}

function changes(state: ChangesSetup): Story['render'] {
  return () => ({
    components: { ChangesList },
    setup() {
      mockApi({
        '/changes': () =>
          state.pending
            ? neverResolves()
            : {
                data: {
                  session_id: 'ses_4d1',
                  session_code: 'SS3',
                  session_title: 'Latency deep dive',
                  created: state.changes.filter((c) => c.created).length,
                  modified: state.changes.filter((c) => !c.created).length,
                  changes: state.changes,
                },
              },
      })
      return {}
    },
    template: '<ChangesList session-id="ses_4d1" research-slug="R1" />',
  })
}

/** A session that created two entries and edited one. */
export const Populated: Story = {
  render: changes({ changes: mockSessionChanges }),
}

/** Only edits: nothing new came out of this session, but three entries moved. */
export const ModifiedOnly: Story = {
  render: changes({
    changes: [
      mockChangeModified,
      { ...mockChangeModified, entry_id: 'ent_a1', entry_code: 'E2', title: 'Latency budget', from_revision: 4, to_revision: 5 },
      { ...mockChangeModified, entry_id: 'ent_a2', entry_code: 'E3', title: 'Benchmark harness', from_revision: 1, to_revision: 2 },
    ],
  }),
}

/** Only new entries: every card starts at r1, and the badge reads "created". */
export const CreatedOnly: Story = {
  render: changes({ changes: [mockChangeCreated, mockChangeCreatedSmall] }),
}


/** One card with its diff already open, the state after a "Show changes" click. */
export const DiffExpanded: Story = {
  render: changes({ changes: [mockChangeModified] }),
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    for (let i = 0; i < 50 && !canvasElement.querySelector('.change-toggle'); i++) {
      await new Promise((resolve) => setTimeout(resolve, 20))
    }
    const toggle = canvasElement.querySelector('.change-toggle') as HTMLElement | null
    toggle?.click()
  },
}

/** A twelve-entry session: the summary line and the card list under load. */
export const ManyChanges: Story = {
  render: changes({ changes: mockSessionChangesMany }),
}

/** A single revision that changed one word — the smallest honest card. */
export const SingleSmallChange: Story = {
  render: changes({ changes: [{ ...mockChangeModified, from_revision: 5, to_revision: 5, summary: '+1 −1' }] }),
}

/** The session has not written to any entry yet. */
export const Empty: Story = {
  render: changes({ changes: [] }),
}

/** Waiting on `GET /api/sessions/{id}/changes`. */
export const Loading: Story = {
  render: changes({ changes: [], pending: true }),
}
