import type { Meta, StoryObj } from '@storybook/vue3'
import { onMounted, ref } from 'vue'
import HistoryPanel from './HistoryPanel.vue'
import { mockApi, neverResolves, revisionParam } from '../../__mocks__/api'
import {
  mockRevisions,
  mockRevisionsMany,
  mockRevisionsSingle,
  mockEntryDiff,
  mockEntryDiffByRevision,
  mockEntryDiffSingle,
  type EntryDiff,
  type Revision,
} from '../../__mocks__/revision'

/**
 * The revision history of one entry: the list on the left, the diff of the
 * selected revision on the right, restore from any revision but the current one.
 *
 * The panel fetches its own data on open, so these stories render the real
 * component against a mocked `authFetch` (see `__mocks__/api.ts`) rather than a
 * copy of its markup. It also loads on the `visible` false → true transition,
 * which is why every story mounts closed and opens on the next tick — the same
 * path the entry page takes.
 */
const meta: Meta<typeof HistoryPanel> = {
  title: 'Entry/HistoryPanel',
  component: HistoryPanel,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
  argTypes: {
    visible: { control: false },
    entryId: { control: 'text' },
    onClose: { action: 'close' },
    onRestore: { action: 'restore' },
  },
}
export default meta
type Story = StoryObj<typeof HistoryPanel>

type PanelSetup = {
  revisions: Revision[]
  current: number
  diff?: (revision: number | null) => EntryDiff | null
  pending?: boolean
  /** Answers the first diff request and hangs on every later one. */
  slowSecondDiff?: boolean
}

/** Waits for the panel's own fetches to land before a play function acts. */
async function waitFor(root: HTMLElement, selector: string): Promise<HTMLElement | null> {
  for (let i = 0; i < 50; i++) {
    const found = root.querySelector(selector) as HTMLElement | null
    if (found) return found
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
  return null
}

/**
 * Builds a story that opens the panel over a mocked history endpoint.
 * `close` and `restore` are left unhandled on purpose: the panel stays open so
 * the emit can be read off the Actions panel.
 */
function panel(state: PanelSetup): Story['render'] {
  return (args: any) => ({
    components: { HistoryPanel },
    setup() {
      let diffCalls = 0
      mockApi({
        '/revisions': () =>
          state.pending
            ? neverResolves()
            : { data: { entry_id: 'ent_2f7c9a41', entry_code: 'E4', title: 'Streaming Ingestion Trade-offs', revisions: state.revisions, current: state.current } },
        '/diff': (url: string) => {
          diffCalls++
          if (state.slowSecondDiff && diffCalls > 1) return neverResolves()
          return { data: state.diff?.(revisionParam(url)) ?? null }
        },
      })

      const visible = ref(false)
      onMounted(() => {
        visible.value = true
      })
      return { args, visible }
    },
    template: `
      <HistoryPanel
        :visible="visible"
        entry-id="ent_2f7c9a41"
        @close="args.onClose"
        @restore="args.onRestore"
      />
    `,
  })
}

const diffByRevision = (revision: number | null) =>
  (revision === null ? mockEntryDiff : mockEntryDiffByRevision[revision]) ?? mockEntryDiff

/**
 * Five revisions written across three sessions by an agent, a person, and a
 * restore. The newest is selected on open and carries no Restore button — it is
 * already what the entry says.
 */
export const Populated: Story = {
  render: panel({ revisions: mockRevisions, current: 5, diff: diffByRevision }),
}

/**
 * The oldest revision picked from the list: nothing precedes it, so the header
 * reads "first version" and every line of the diff is an addition.
 */
export const FirstVersionSelected: Story = {
  render: panel({ revisions: mockRevisions, current: 5, diff: diffByRevision }),
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await waitFor(canvasElement, '.rev-main')
    const rows = canvasElement.querySelectorAll('.rev-main')
    ;(rows[rows.length - 1] as HTMLElement | undefined)?.click()
  },
}

/** A long edit: the diff pane collapses the unchanged runs and scrolls. */
export const LongDiff: Story = {
  render: panel({ revisions: mockRevisions, current: 5, diff: () => mockEntryDiffByRevision[4]! }),
}

/** An entry written once — by an import, with no session behind it. */
export const SingleRevision: Story = {
  render: panel({ revisions: mockRevisionsSingle, current: 1, diff: () => mockEntryDiffSingle }),
}

/** Twenty-four revisions: the list scrolls inside the modal, the diff does not. */
export const ManyRevisions: Story = {
  render: panel({ revisions: mockRevisionsMany, current: 24, diff: diffByRevision }),
}

/**
 * Waiting on `GET /api/entries/{id}/revisions`. Note that the panel also stays
 * in this state while the *first* diff loads: `load()` awaits the initial
 * `select()`, so a slow diff keeps the revision list hidden.
 */
export const Loading: Story = {
  render: panel({ revisions: [], current: 0, pending: true }),
}

/**
 * A revision picked from the list while its diff is still in flight: the list
 * stays interactive and only the right pane says "Loading diff…".
 */
export const DiffLoading: Story = {
  render: panel({ revisions: mockRevisions, current: 5, diff: diffByRevision, slowSecondDiff: true }),
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await waitFor(canvasElement, '.rev-main')
    const rows = canvasElement.querySelectorAll('.rev-main')
    ;(rows[2] as HTMLElement | undefined)?.click()
  },
}

/** No revisions at all: an entry created before the history feature existed. */
export const NoHistory: Story = {
  render: panel({ revisions: [], current: 0 }),
}

/** Closed — nothing renders, and nothing is fetched until it opens. */
export const Closed: Story = {
  render: (args: any) => ({
    components: { HistoryPanel },
    setup() {
      mockApi({})
      return { args }
    },
    template: `
      <div style="color: var(--color-text-muted); font-size: 0.875rem;">
        The panel is mounted with <code>visible: false</code> — no request is made.
        <HistoryPanel :visible="false" entry-id="ent_2f7c9a41" @close="args.onClose" @restore="args.onRestore" />
      </div>
    `,
  }),
}
