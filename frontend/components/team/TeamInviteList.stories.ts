import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import TeamInviteList from './TeamInviteList.vue'
import type { TeamInvite } from '~/composables/useTeams'
import { mockInviteLink, mockInvites, mockRecoverableLinks } from '../../__mocks__/team'

/**
 * Invitations nobody has accepted yet.
 *
 * Lapsed ones stay in the list, dimmed. Filtering them away is how an owner ends
 * up wondering why a colleague never joined while the colleague is looking at a
 * page that says "expired".
 *
 * The right-hand action depends on something the server cannot tell you: whether
 * this tab still holds the token it was handed when the invitation was created.
 * The server stores only a hash, so after a reload there is no way back to a
 * working link and the only honest offer is a new invitation. `recoverableLinks`
 * is that memory, keyed by invite id.
 */
const meta: Meta<typeof TeamInviteList> = {
  title: 'Team/TeamInviteList',
  component: TeamInviteList,
  tags: ['autodocs'],
  argTypes: {
    invites: { control: 'object' },
    busyId: { control: 'text' },
    recoverableLinks: { control: 'object' },
    onRevoke: { action: 'revoke' },
    onShowLink: { action: 'showLink' },
    onReinvite: { action: 'reinvite' },
  },
}
export default meta
type Story = StoryObj<typeof TeamInviteList>

const live = mockInvites[0]!
const staleLink = mockInvites[1]!
const today = mockInvites[2]!
const expired = mockInvites[3]!

/** A live invitation created in this tab: the token is still in memory, so the
 *  link can simply be shown again. */
export const Pending: Story = {
  args: { invites: [live], recoverableLinks: mockRecoverableLinks },
}

/** The same invitation after a reload. Nothing about it changed on the server —
 *  but the token is gone from this tab, so the offer is a replacement link
 *  rather than a button that would produce nothing. */
export const LinkNotRecoverable: Story = {
  args: { invites: [staleLink], recoverableLinks: {} },
}

/** Lapsed: dimmed, "expired", and one action. Revoke is not offered because an
 *  expired link already grants nothing. */
export const Expired: Story = {
  args: { invites: [expired], recoverableLinks: {} },
}

/** The last day of a two-week window reads "expires today" rather than
 *  "expires in 1 days". */
export const ExpiresToday: Story = {
  args: { invites: [today], recoverableLinks: {} },
}

/** The list as an owner usually finds it: one link still recoverable, one not,
 *  one about to lapse and one already gone. */
export const AllStates: Story = {
  args: { invites: mockInvites, recoverableLinks: mockRecoverableLinks },
}

/** A revoke in flight. The row dims and its Revoke locks; the other actions stay
 *  reachable because they do not touch the same record. */
export const RevokeInFlight: Story = {
  args: {
    invites: mockInvites,
    recoverableLinks: mockRecoverableLinks,
    busyId: live.id,
  },
}

/** An internationalised address, which is a real thing to invite in this market
 *  and the longest string the row can be handed. It wraps rather than pushing
 *  the actions off the end. */
export const CyrillicEmail: Story = {
  args: {
    invites: [
      {
        ...live,
        id: 'inv_cyrillic',
        email: 'екатерина.ковалёва@научные-исследования.рф',
        role: 'editor',
      },
    ] as TeamInvite[],
    recoverableLinks: {},
  },
}

/** An invitation stored without an address. The API does not issue these today
 *  — the dialog requires an email — so this is the defensive branch, and it says
 *  what the link actually is rather than leaving the column blank. */
export const WithoutEmail: Story = {
  args: {
    invites: [{ ...live, id: 'inv_open', email: '' }] as TeamInvite[],
    recoverableLinks: {},
  },
}

/** Nothing pending. The row set is empty; the team page drops the whole
 *  "Pending invites" heading rather than showing an empty one. */
export const Empty: Story = {
  args: { invites: [], recoverableLinks: {} },
}

/** Wired up: Revoke removes the row after a beat, Re-invite replaces a lapsed
 *  invitation with a fresh one whose link is recoverable, and Show link prints
 *  the token this tab is holding. */
export const Interactive: Story = {
  render: () => ({
    components: { TeamInviteList },
    setup() {
      const invites = ref<TeamInvite[]>(mockInvites.map((i) => ({ ...i })))
      const links = ref<Record<string, string>>({ ...mockRecoverableLinks })
      const busyId = ref<string | null>(null)
      const shown = ref('')

      function revoke(invite: TeamInvite) {
        busyId.value = invite.id
        setTimeout(() => {
          invites.value = invites.value.filter((i) => i.id !== invite.id)
          busyId.value = null
        }, 400)
      }

      function reinvite(invite: TeamInvite) {
        const id = `inv_${Math.random().toString(36).slice(2, 8)}`
        invites.value = [
          ...invites.value.filter((i) => i.id !== invite.id),
          {
            ...invite,
            id,
            created_at: new Date().toISOString(),
            expires_at: new Date(Date.now() + 14 * 86_400_000).toISOString(),
          },
        ]
        links.value = { ...links.value, [id]: mockInviteLink }
        shown.value = mockInviteLink
      }

      return { invites, links, busyId, shown, revoke, reinvite, showLink: (i: TeamInvite) => (shown.value = links.value[i.id] ?? '') }
    },
    template: `
      <div style="display: flex; flex-direction: column; gap: var(--space-3);">
        <TeamInviteList
          :invites="invites"
          :busy-id="busyId"
          :recoverable-links="links"
          @revoke="revoke"
          @reinvite="reinvite"
          @show-link="showLink"
        />
        <p style="margin: 0; font-size: var(--type-xs); color: var(--color-text-muted); overflow-wrap: anywhere;">
          Last link shown: <code>{{ shown || '—' }}</code>
        </p>
      </div>
    `,
  }),
}
