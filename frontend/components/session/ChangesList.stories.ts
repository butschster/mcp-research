import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import ChangesList from './ChangesList.vue'
import { mockApi, neverResolves, fails } from '../../__mocks__/api'
import {
  mockSessionChanges,
  mockSessionChangesMany,
  mockChangeModified,
  mockChangeCreated,
  mockChangeCreatedSmall,
  type SessionEntryChange,
} from '../../__mocks__/revision'
import { withShare, withoutShare } from '../../__mocks__/share'

/**
 * What a session did to the research's entries — the ones it created, the ones
 * it edited, and the diff for each.
 *
 * The component fetches `GET /api/sessions/{id}/changes` itself on mount, so
 * these stories render the real component against a mocked `authFetch` (see
 * `__mocks__/api.ts`). Diffs are collapsed until "Show changes" is clicked.
 *
 * Each card's title links through `entryPath()`, so it resolves to
 * `/s/{token}/entry/…` when the session page is being read through a share
 * link.
 *
 * It emits `count` after every load — the number of entries touched, or `null`
 * when the load failed — and exposes `reload()`, which the session page calls
 * on a realtime `entry` event. Both are visible in the Actions panel of every
 * story below; `CountsIntoTheTabBadge` and `ReloadedByAnEntryEvent` show what
 * the page does with them.
 */
const meta: Meta<typeof ChangesList> = {
  title: 'Session/ChangesList',
  component: ChangesList,
  tags: ['autodocs'],
  // Share state is module state; this gives the ordinary stories a known
  // starting point rather than whatever the last story left behind. The
  // trade-offs are in __mocks__/share.ts.
  decorators: [withoutShare()],
  argTypes: {
    sessionId: { control: 'text' },
    researchSlug: { control: 'text' },
    onCount: { action: 'count' },
  },
}
export default meta
type Story = StoryObj<typeof ChangesList>

type ChangesSetup = {
  changes: SessionEntryChange[]
  pending?: boolean
  failing?: boolean
}

/** The payload the endpoint answers with for a given set of changes. */
function payload(list: SessionEntryChange[]) {
  return {
    data: {
      session_id: 'ses_4d1',
      session_code: 'SS3',
      session_title: 'Latency deep dive',
      created: list.filter((c) => c.created).length,
      modified: list.filter((c) => !c.created).length,
      changes: list,
    },
  }
}

function changes(state: ChangesSetup): Story['render'] {
  return (args: any) => ({
    components: { ChangesList },
    setup() {
      mockApi({
        '/changes': () =>
          state.failing ? fails() : state.pending ? neverResolves() : payload(state.changes),
      })
      return { args }
    },
    // `count` is forwarded to the Actions panel, which is the only place a
    // reader can see what the tab badge would say — including the `null` a
    // failed load sends.
    template: '<ChangesList session-id="ses_4d1" research-slug="R1" @count="args.onCount" />',
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

/**
 * The request failed — a 500, an expired token, a dropped connection.
 *
 * Worth comparing side by side with `Empty`: this used to render as that one,
 * telling the reader "this session has not written to any entry yet" about a
 * session that may have written twenty. The distinction between "nothing
 * happened" and "we could not find out" is the whole point of the state, and
 * the retry is what makes it actionable.
 *
 * `count` fires as `null` here — read it off the Actions panel — which is what
 * clears the tab badge instead of leaving the previous number standing.
 */
export const LoadFailed: Story = {
  render: changes({ changes: [], failing: true }),
}

/**
 * What the `count` emit is for.
 *
 * The session page keeps it in `changesCount` and prints it on the Changes tab,
 * so a reader learns something happened without opening the tab. Three sessions
 * here, one per answer the emit can give: a load that landed sends a number, an
 * empty session sends `0`, a failed load sends `null`.
 *
 * The tab strip is reproduced rather than imported — `.tab-count` is scoped to
 * the session page — and it mirrors what that page renders for each: a number
 * when the count is known and non-zero, **nothing** at zero, and a red `!` when
 * the count is `null`.
 *
 * The third case is the one worth looking at. Hiding the badge at zero and
 * hiding it on failure would make "nothing happened" and "we could not find
 * out" pixel-identical, and "nothing happened" is the single wrong conclusion
 * this badge exists to prevent.
 */
export const CountsIntoTheTabBadge: Story = {
  render: (args: any) => ({
    components: { ChangesList },
    setup() {
      mockApi({
        '/sessions/ses_ok/changes': () => payload(mockSessionChanges),
        '/sessions/ses_empty/changes': () => payload([]),
        '/sessions/ses_bad/changes': () => fails(),
      })
      const cases = [
        { id: 'ses_ok', label: 'A session that touched three entries' },
        { id: 'ses_empty', label: 'A session that touched none' },
        { id: 'ses_bad', label: 'A session whose changes did not load' },
      ]
      const counts = ref<Record<string, number | null | undefined>>({})
      // Two arguments rather than a curried handler: `@count="record(c.id)"`
      // compiles to an inline statement, so the function it returned would
      // never see the emitted value.
      const record = (id: string, n: number | null) => {
        counts.value = { ...counts.value, [id]: n }
        args.onCount?.(n)
      }
      return { args, cases, counts, record }
    },
    template: `
      <div style="display: grid; gap: 2.5rem;">
        <section v-for="c in cases" :key="c.id">
          <p style="font-size: var(--type-xs); color: var(--color-text-muted); margin: 0 0 0.5rem;">{{ c.label }}</p>
          <div style="display: flex; align-items: center; gap: 0.5rem; border-bottom: 1px solid var(--color-border); margin-bottom: 1rem;">
            <span style="display: flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1.25rem; margin-bottom: -1px; border-bottom: 2px solid var(--color-primary); color: var(--color-primary); font-size: var(--type-sm); font-weight: var(--weight-medium);">
              Changes
              <span
                v-if="counts[c.id] === null"
                title="Could not load what this session changed — open the tab to retry"
                style="min-width: 1.6rem; text-align: center; padding: 0.15rem 0.4rem; border-radius: var(--radius-xs); background: var(--color-primary-muted); color: var(--color-error); font-size: var(--type-xs);"
              >!</span>
              <span
                v-else-if="counts[c.id]"
                style="min-width: 1.6rem; text-align: center; padding: 0.15rem 0.4rem; border-radius: var(--radius-xs); background: var(--color-primary-muted); color: var(--color-primary); font-size: var(--type-xs); font-variant-numeric: tabular-nums;"
              >{{ counts[c.id] }}</span>
            </span>
            <span style="margin-left: auto; font-size: var(--type-xs); color: var(--color-text-muted); font-family: 'JetBrains Mono', monospace;">
              count: {{ c.id in counts ? String(counts[c.id]) : '—' }}
            </span>
          </div>
          <ChangesList :session-id="c.id" research-slug="R1" @count="record(c.id, $event)" />
        </section>
      </div>
    `,
  }),
}

/**
 * The list after `reload()` — the method the session page calls on a realtime
 * `entry` event, through a template ref.
 *
 * An agent writing entries is exactly when somebody is watching this page, and
 * this was the one list that stayed frozen through it. Press the button: the
 * first load answered with one created entry, the second answers with the full
 * three and re-emits `count`, so the tab badge moves with the list.
 */
export const ReloadedByAnEntryEvent: Story = {
  render: (args: any) => ({
    components: { ChangesList },
    setup() {
      let loads = 0
      mockApi({
        '/changes': () => payload(loads++ === 0 ? [mockChangeCreated] : mockSessionChanges),
      })
      const list = ref<{ reload: () => Promise<void> } | null>(null)
      // `undefined` until the first load lands — printing `null` before then
      // would show the failure state to a story that has not failed.
      const count = ref<number | null | undefined>(undefined)
      const onCount = (n: number | null) => {
        count.value = n
        args.onCount?.(n)
      }
      return { args, list, count, onCount }
    },
    template: `
      <div>
        <div style="display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem;">
          <button class="btn btn-sm btn-primary" @click="list?.reload()">Simulate an entry event</button>
          <span style="font-size: var(--type-xs); color: var(--color-text-muted); font-family: 'JetBrains Mono', monospace;">
            count: {{ count === undefined ? '—' : String(count) }}
          </span>
        </div>
        <ChangesList ref="list" session-id="ses_4d1" research-slug="R1" @count="onCount" />
      </div>
    `,
  }),
}

/** The same list on a shared session page: every entry title points at
 *  `/s/{token}/entry/…` instead of `/research/R1/entry/…`. */
export const InsideAShare: Story = {
  decorators: [withShare()],
  render: changes({ changes: mockSessionChanges }),
}
