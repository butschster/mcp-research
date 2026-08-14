import type { Meta, StoryObj } from '@storybook/vue3'
import { onUnmounted, ref } from 'vue'
import InviteDialog from './InviteDialog.vue'
import type { TeamRole } from '~/composables/useTeams'
import { mockInviteLink } from '../../__mocks__/team'

/**
 * Creates an invitation and hands over the link.
 *
 * Two states in one dialog: compose, then the link. The link is not a second
 * screen because it is the result of the first — navigating away from a token
 * that is shown exactly once is how it gets lost. Which state you see is
 * decided entirely by the `link` prop: empty composes, filled hands over.
 *
 * No email is sent. The server issues a link, the dialog shows it once, and the
 * inviter passes it along themselves — so the copy affordance is the feature,
 * not a convenience.
 *
 * Viewer is the default role. An over-permissioned colleague is silent and
 * permanent; an under-permissioned one says so in thirty seconds and the fix is
 * one select.
 */
const meta: Meta<typeof InviteDialog> = {
  title: 'Team/InviteDialog',
  component: InviteDialog,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
  argTypes: {
    visible: { control: 'boolean' },
    teamName: { control: 'text' },
    link: { control: 'text' },
    creating: { control: 'boolean' },
    error: { control: 'text' },
    prefillEmail: { control: 'text' },
    prefillRole: { control: 'select', options: ['viewer', 'editor', 'owner'] },
    onCreate: { action: 'create' },
    onClose: { action: 'close' },
  },
}
export default meta
type Story = StoryObj<typeof InviteDialog>

/** The compose state. Create stays disabled until the address contains an @ —
 *  the only validation worth doing here, since the server checks the rest. */
export const Compose: Story = {
  args: { visible: true, teamName: 'Product Research' },
}

/** Cyrillic team name in the title, which is the normal case for this product. */
export const CyrillicTeam: Story = {
  args: { visible: true, teamName: 'Отдел интеграций и партнёрских программ' },
}

/** Submitting. Both buttons lock, so a second Enter cannot issue a second
 *  invitation for the same address. */
export const Submitting: Story = {
  args: { visible: true, teamName: 'Product Research', creating: true },
}

/** The link, shown once. Focus moves to Copy when it arrives and a live region
 *  says why the dialog changed under the reader's hands. */
export const LinkReady: Story = {
  args: { visible: true, teamName: 'Product Research', link: mockInviteLink },
}

/** The server refused: that address is already a member. Said inline, next to
 *  the field that has to change, rather than as a toast over a dialog that stays
 *  open. */
export const AlreadyInTeam: Story = {
  args: {
    visible: true,
    teamName: 'Product Research',
    error: 'ekaterina.kovaleva@intruforce.com is already in this team.',
    prefillEmail: 'ekaterina.kovaleva@intruforce.com',
    prefillRole: 'editor',
  },
}

/** Re-invite: a lapsed invitation reopens the dialog with its address and role
 *  already filled, so replacing it is one click and one confirmation. */
export const Reinvite: Story = {
  args: {
    visible: true,
    teamName: 'Аналитика рынка',
    prefillEmail: 'legal@old-partner.example',
    prefillRole: 'viewer',
  },
}

/**
 * Copying is not available — a plain-HTTP deployment on a LAN, where the
 * clipboard API is simply absent. The dialog swaps the link for a selected
 * input and says so, rather than leaving a button that does nothing.
 *
 * The story removes `navigator.clipboard` for its own lifetime and presses Copy
 * for you; the property is restored when the story unmounts.
 */
export const ClipboardUnavailable: Story = {
  render: () => ({
    components: { InviteDialog },
    setup() {
      // An own property shadows the one on Navigator.prototype; deleting it
      // uncovers the real getter again when the story goes away.
      Object.defineProperty(navigator, 'clipboard', { value: undefined, configurable: true })
      onUnmounted(() => {
        delete (navigator as { clipboard?: unknown }).clipboard
      })
      return { link: mockInviteLink }
    },
    template: `<InviteDialog :visible="true" team-name="Product Research" :link="link" />`,
  }),
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    const copy = canvasElement.querySelector('.copy-btn') as HTMLElement | null
    copy?.click()
  },
}

/**
 * The whole flow against a fake server: compose, a beat of "Creating…", then the
 * link. Inviting `taken@intruforce.com` returns the already-a-member refusal, so
 * both endings are reachable without touching the args.
 */
export const Interactive: Story = {
  render: () => ({
    components: { InviteDialog },
    setup() {
      const visible = ref(false)
      const creating = ref(false)
      const link = ref('')
      const error = ref('')
      const issued = ref<{ email: string; role: TeamRole } | null>(null)

      function open() {
        link.value = ''
        error.value = ''
        issued.value = null
        visible.value = true
      }

      function create(payload: { email: string; role: TeamRole }) {
        error.value = ''
        creating.value = true
        setTimeout(() => {
          creating.value = false
          if (payload.email === 'taken@intruforce.com') {
            error.value = `${payload.email} is already in this team.`
            return
          }
          issued.value = payload
          link.value = mockInviteLink
        }, 600)
      }

      return { visible, creating, link, error, issued, open, create }
    },
    template: `
      <div style="padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-3); align-items: flex-start;">
        <button class="btn btn-sm btn-primary" @click="open">+ Invite</button>
        <p style="margin: 0; font-size: var(--type-xs); color: var(--color-text-muted);">
          Issued: <code>{{ issued ? JSON.stringify(issued) : '—' }}</code>
        </p>
        <InviteDialog
          :visible="visible"
          team-name="Product Research"
          :link="link"
          :creating="creating"
          :error="error"
          @create="create"
          @close="visible = false"
        />
      </div>
    `,
  }),
}
