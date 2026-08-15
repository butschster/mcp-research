import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import ShareRowList, { type ShareRow } from './ShareRowList.vue'
import {
  includeDefault,
  includeEntriesOnly,
  includeEverything,
  mockLongLabelShareRow,
  mockManyShareRows,
  mockRecoverableShareLinks,
  mockShareRows,
  mockShareUrl,
} from '../../__mocks__/share'

const DAY = 86_400_000
const inDays = (days: number) => new Date(Date.now() + days * DAY).toISOString()
const daysAgo = (days: number) => new Date(Date.now() - days * DAY).toISOString()

/**
 * The links a research has handed out — live, expiring, expired and revoked.
 *
 * Dead rows stay, dimmed, the way lapsed invitations do. "Who did I give this
 * to, and did I take it back" is the question this list answers, and a row that
 * disappears on revocation leaves no record that it ever existed.
 *
 * The right-hand `Show link` action depends on something the server cannot tell
 * you: whether this tab still holds the URL it was handed when the link was
 * created. The server stores only a hash, so after a reload there is no way back
 * to a working link and the only honest offer is a new one. `recoverableLinks`
 * is that memory, keyed by share id.
 *
 * The label wraps and never truncates. A truncated label is not a label, and
 * this list exists to be recognised by one.
 */
const meta: Meta<typeof ShareRowList> = {
  title: 'Research/ShareRowList',
  component: ShareRowList,
  tags: ['autodocs'],
  decorators: [() => ({ template: '<div style="max-width: 680px;"><story /></div>' })],
  argTypes: {
    shares: { control: 'object' },
    busyId: { control: 'text' },
    recoverableLinks: { control: 'object' },
    onRevoke: { action: 'revoke' },
    onShowLink: { action: 'showLink' },
  },
}
export default meta
type Story = StoryObj<typeof ShareRowList>

const live = mockShareRows[0]!
const neverOpened = mockShareRows[1]!
const expiringSoon = mockShareRows[2]!
const expired = mockShareRows[3]!
const revoked = mockShareRows[4]!

/** A working link with traffic on it, created in this tab: password-protected,
 *  47 views, last opened two hours ago, and the URL still in memory so it can
 *  simply be shown again. */
export const Live: Story = {
  args: { shares: [live], recoverableLinks: mockRecoverableShareLinks },
}

/** Sent and never opened. Worth saying plainly rather than showing "0 views":
 *  the owner's next move is usually to check the recipient got the message. */
export const NeverOpened: Story = {
  args: { shares: [neverOpened], recoverableLinks: {} },
}

/** About to lapse. Counted in days rather than shown as a date, because the
 *  owner's decision is whether to reissue it before the meeting, not which
 *  Tuesday it dies on. */
export const ExpiringSoon: Story = {
  args: { shares: [expiringSoon], recoverableLinks: {} },
}

/** Already lapsed: dimmed and inert. No Revoke, because an expired link already
 *  grants nothing and offering to turn it off implies it is still on. */
export const Expired: Story = {
  args: { shares: [expired], recoverableLinks: {} },
}

/** Turned off by hand. It keeps its place in the list with no actions at all —
 *  the record of having given it out is the point. */
export const Revoked: Story = {
  args: { shares: [revoked], recoverableLinks: {} },
}

/** Password-protected. A badge on the row rather than a column of its own: it
 *  is true of a minority of links and says nothing about the rest. */
export const PasswordProtected: Story = {
  args: {
    shares: [{ ...neverOpened, id: 'shr_pw', label: 'Legal review', has_password: true }],
    recoverableLinks: {},
  },
}

/** A revoke in flight. The row dims and stops accepting clicks, and it does not
 *  say "Revoked" until the server does — telling an owner access is closed when
 *  it may not be is the one failure this list must not have. */
export const RevokeInFlight: Story = {
  args: {
    shares: mockShareRows,
    recoverableLinks: mockRecoverableShareLinks,
    busyId: live.id,
  },
}

/** A 200-character label. It wraps and pushes the row taller; the state column
 *  and the actions stay where they are. */
export const VeryLongLabel: Story = {
  args: { shares: [mockLongLabelShareRow], recoverableLinks: {} },
}

/** A link created without a label — the field is optional. The row says
 *  "Untitled link" rather than leaving the cell blank, so it can still be
 *  pointed at. */
export const WithoutLabel: Story = {
  args: {
    shares: [{ ...live, id: 'shr_unlabelled', label: '' } as ShareRow],
    recoverableLinks: {},
  },
}

/** The list as an owner usually finds it: one live and recoverable, one never
 *  opened, one about to lapse, one expired, one revoked. */
export const AllStates: Story = {
  args: { shares: mockShareRows, recoverableLinks: mockRecoverableShareLinks },
}

/** Sixty links on one research. Nothing paginates and revoked rows are not
 *  swept up — the dialog body scrolls, exactly as the team pages do. */
export const SixtyRows: Story = {
  args: { shares: mockManyShareRows, recoverableLinks: {} },
}

/** Nothing handed out yet. The list renders empty; the dialog opens on the
 *  create form instead of showing this. */
export const Empty: Story = {
  args: { shares: [], recoverableLinks: {} },
}

/** Wired up: Revoke flips the row after a beat rather than before, and Show
 *  link prints the URL this tab is holding. */
export const Interactive: Story = {
  render: () => ({
    components: { ShareRowList },
    setup() {
      const shares = ref<ShareRow[]>(mockShareRows.map((s) => ({ ...s })))
      const links = ref<Record<string, string>>({ ...mockRecoverableShareLinks })
      const busyId = ref<string | undefined>(undefined)
      const shown = ref('')

      function revoke(share: ShareRow) {
        busyId.value = share.id
        setTimeout(() => {
          shares.value = shares.value.map((s: ShareRow) =>
            s.id === share.id ? { ...s, revoked_at: new Date().toISOString() } : s,
          )
          delete links.value[share.id]
          links.value = { ...links.value }
          busyId.value = undefined
        }, 600)
      }

      return {
        shares,
        links,
        busyId,
        shown,
        revoke,
        showLink: (s: ShareRow) => (shown.value = links.value[s.id] ?? ''),
      }
    },
    template: `
      <div style="display: flex; flex-direction: column; gap: var(--space-3);">
        <ShareRowList
          :shares="shares"
          :busy-id="busyId"
          :recoverable-links="links"
          @revoke="revoke"
          @show-link="showLink"
        />
        <p style="margin: 0; font-size: var(--type-xs); color: var(--color-text-muted); overflow-wrap: anywhere;">
          Last link shown: <code>{{ shown || '—' }}</code>
        </p>
      </div>
    `,
  }),
}

/** ≤768px: the row collapses to two columns and the actions move to a line of
 *  their own, the pattern `TeamMemberList` already uses. */
export const Mobile: Story = {
  parameters: { viewport: { defaultViewport: 'mobile' } },
  args: {
    shares: [live, expiringSoon, revoked],
    recoverableLinks: mockRecoverableShareLinks,
  },
}

/** The include column, one row per shape. Entries lead every line because they
 *  are never optional. */
export const IncludeVariants: Story = {
  args: {
    shares: [
      { ...live, id: 'shr_inc_all', label: 'Everything', include: includeEverything, has_password: false },
      { ...live, id: 'shr_inc_default', label: 'The default', include: includeDefault, has_password: false },
      { ...live, id: 'shr_inc_min', label: 'Entries only', include: includeEntriesOnly, has_password: false },
    ] as ShareRow[],
    recoverableLinks: {},
  },
}

/** Boundary wording, side by side: today, tomorrow, in six days. "Expires in 1
 *  days" is the bug this row set exists to catch. */
export const ExpiryWording: Story = {
  args: {
    shares: [
      { ...live, id: 'shr_today', label: 'Expires today', expires_at: inDays(0.3), has_password: false, view_count: 4, last_seen_at: daysAgo(0.1) },
      { ...live, id: 'shr_tomorrow', label: 'Expires tomorrow', expires_at: inDays(1.2), has_password: false, view_count: 1, last_seen_at: null },
      { ...live, id: 'shr_week', label: 'Expires in six days', expires_at: inDays(6), has_password: false, view_count: 0, last_seen_at: null },
      { ...live, id: 'shr_never', label: 'No end date', expires_at: null, has_password: false, view_count: 130, last_seen_at: daysAgo(30) },
    ] as ShareRow[],
    recoverableLinks: {},
  },
}
