import type { Meta, StoryObj } from '@storybook/vue3'
import { onMounted, ref } from 'vue'
import HistoryPanel from './HistoryPanel.vue'
import { mockApi, neverResolves, fails, revisionParam } from '../../__mocks__/api'
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
 *
 * It also exposes `reload()`, which the entry page calls on a realtime event
 * while the panel is open, so a remote write does not leave a revision list
 * presenting itself as complete. There is no story for it: it re-runs exactly
 * the fetch `Populated` already shows.
 */
const meta: Meta<typeof HistoryPanel> = {
  title: 'Entry/HistoryPanel',
  component: HistoryPanel,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
  argTypes: {
    visible: { control: false },
    entryId: { control: 'text' },
    entryTitle: { control: 'text' },
    entryCode: { control: 'text' },
    restoreError: { control: 'boolean' },
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
  /** The revision list itself fails: the rail has nothing and says so. */
  listFails?: boolean
  /** The list lands, the comparison for the selected revision does not. */
  diffFails?: boolean
  /** The page's restore attempt came back an error, and says so in the pane. */
  restoreError?: boolean
}

/**
 * The header's title and code are props, not part of the fetched payload — the
 * entry page already holds them and passes them down, so the header is filled
 * in before the revision list arrives. Every story passes them for that reason:
 * a panel headed by an empty `<h2>` is not a state the product reaches.
 */
const ENTRY = {
  id: 'ent_2f7c9a41',
  code: 'E4',
  title: 'Streaming Ingestion Trade-offs',
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
          state.listFails
            ? fails()
            : state.pending
            ? neverResolves()
            : { data: { entry_id: ENTRY.id, entry_code: ENTRY.code, title: ENTRY.title, revisions: state.revisions, current: state.current } },
        '/diff': (url: string) => {
          diffCalls++
          if (state.diffFails) return fails()
          if (state.slowSecondDiff && diffCalls > 1) return neverResolves()
          return { data: state.diff?.(revisionParam(url)) ?? null }
        },
      })

      const visible = ref(false)
      onMounted(() => {
        visible.value = true
      })
      return { args, visible, entry: ENTRY, restoreError: !!state.restoreError }
    },
    template: `
      <HistoryPanel
        :visible="visible"
        :entry-id="entry.id"
        :entry-title="entry.title"
        :entry-code="entry.code"
        :restore-error="restoreError"
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

/**
 * `GET /api/entries/{id}/revisions` failed.
 *
 * The rail used to fall through to "No history yet" here, which is a claim the
 * panel is in no position to make: migration 018 backfilled revision 1 for every
 * pre-existing entry and one is written on create, so an entry with a genuinely
 * empty history is essentially unreachable. Almost every appearance of that
 * empty state was a 500 or an expired token wearing its clothes.
 */
export const ListFailed: Story = {
  render: panel({ revisions: [], current: 0, listFails: true }),
}

/**
 * The rail loaded, the comparison did not.
 *
 * The failure is confined to the pane: the revision stays selected and the list
 * keeps working, so the retry re-requests one diff rather than the panel. What
 * this replaces is worse than blank — it read "Pick a revision to see what
 * changed" while a revision sat visibly selected.
 */
export const DiffFailed: Story = {
  render: panel({ revisions: mockRevisions, current: 5, diffFails: true }),
}

/**
 * The restore itself failed.
 *
 * `restore-error` is a prop, not internal state: the panel emits `restore` and
 * the entry page owns the request, so this is the page telling the panel how it
 * went. r3 is picked here because that is the shape of the incident — somebody
 * chose an old revision, pressed Restore, and needs to be told the entry is
 * still what it was, next to the button they pressed rather than behind a toast
 * on the page underneath.
 */
export const RestoreFailed: Story = {
  render: panel({ revisions: mockRevisions, current: 5, diff: diffByRevision, restoreError: true }),
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await waitFor(canvasElement, '.rev-main')
    const rows = canvasElement.querySelectorAll('.rev-main')
    ;(rows[2] as HTMLElement | undefined)?.click()
  },
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
      return { args, entry: ENTRY }
    },
    template: `
      <div style="color: var(--color-text-muted); font-size: 0.875rem;">
        The panel is mounted with <code>visible: false</code> — no request is made.
        <HistoryPanel
          :visible="false"
          :entry-id="entry.id"
          :entry-title="entry.title"
          :entry-code="entry.code"
          @close="args.onClose"
          @restore="args.onRestore"
        />
      </div>
    `,
  }),
}
