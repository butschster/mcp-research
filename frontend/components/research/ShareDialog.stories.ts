import type { Meta, StoryObj } from '@storybook/vue3'
import { onMounted, ref } from 'vue'
import ShareDialog from './ShareDialog.vue'
import ToastHost from '../ToastHost.vue'
import type { ShareRow } from './ShareRowList.vue'
import {
  mockManyShareRows,
  mockRecoverableShareLinks,
  mockShareRows,
  mockShareUrl,
} from '../../__mocks__/share'

/**
 * Creating, listing and revoking the links for one research.
 *
 * Three views in one dialog, for the same reason the invite dialog has two: the
 * link is the result of the form, and navigating away from a value shown exactly
 * once is how it gets lost.
 *
 * With no links yet it opens straight onto the create form. That is the empty
 * state — the lead sentence already explains what a share link is, and a screen
 * whose only content is a button and a paragraph is a step, not information.
 *
 * **Reveal is a transition, not a state.** `issuedUrl` is watched, not read on
 * mount, so a dialog opened with a URL already in its args shows the list. Every
 * story below that shows the reveal therefore drives it: it mounts empty and
 * fills the prop a tick later, which is what the page does when the server
 * answers.
 */
const meta: Meta<typeof ShareDialog> = {
  title: 'Research/ShareDialog',
  component: ShareDialog,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
  argTypes: {
    visible: { control: 'boolean' },
    researchName: { control: 'text' },
    shares: { control: 'object' },
    loading: { control: 'boolean' },
    creating: { control: 'boolean' },
    error: { control: 'text' },
    issuedUrl: { control: 'text' },
    busyId: { control: 'text' },
    recoverableLinks: { control: 'object' },
    onCreate: { action: 'create' },
    onRevoke: { action: 'revoke' },
    onClose: { action: 'close' },
    onDismissReveal: { action: 'dismissReveal' },
  },
  args: {
    visible: true,
    researchName: 'Pricing benchmark, Q3',
    recoverableLinks: {},
  },
}
export default meta
type Story = StoryObj<typeof ShareDialog>

/**
 * Opening a research that has never been shared. There is no empty state: the
 * dialog is the create form, and focus lands in the Label field.
 *
 * Cancel closes outright here rather than returning to a list that does not
 * exist.
 */
export const NoSharesYet: Story = {
  args: { shares: [] },
}

/** Opening a research that already has links. The list leads, `+ New link`
 *  takes focus, and the lead sentence still explains what these are — an owner
 *  who last used this a month ago should not have to remember. */
export const WithShares: Story = {
  args: { shares: mockShareRows, recoverableLinks: mockRecoverableShareLinks },
}

/** The create form reached from a list. Roadmaps and downloads are on by
 *  default; sessions and tasks are off, because interview transcripts and an
 *  internal todo list are the two things an owner is most likely to hand over by
 *  accident. Entries are not a checkbox at all. */
export const CreateForm: Story = {
  args: { shares: mockShareRows },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    const buttons = Array.from(canvasElement.querySelectorAll('button'))
    const newLink = buttons.find((b) => b.textContent?.includes('New link'))
    newLink?.click()
  },
}

/** Submitting. Both buttons lock, so a second Enter cannot issue two links for
 *  the same audience. */
export const Creating: Story = {
  args: { shares: [], creating: true },
}

/** The server refused. Said inline, above the actions, next to a form that is
 *  still on screen and still filled in — a toast would take the message away
 *  from the fields that have to change. */
export const CreateFailed: Story = {
  args: {
    shares: [],
    error: 'A share link with that label already exists on this research.',
  },
}

/** Loading the list. Three skeleton rows at row height, and `+ New link` is
 *  live immediately: creating a link does not depend on knowing what is already
 *  there. */
export const LoadingList: Story = {
  args: { shares: [], loading: true },
}

/**
 * The link, shown once.
 *
 * The amber strip is the only amber in this feature, and it is here because this
 * is the only moment where something is about to be lost. Focus goes to Copy and
 * a live region says why the dialog changed under the reader's hands.
 *
 * The primary button reads `Copy and finish` until a copy has happened, and
 * closing without copying is still allowed — a dialog that refuses to close is a
 * trap, and the row keeps a `Show link` action for the life of the tab.
 */
export const Reveal: Story = {
  render: () => ({
    components: { ShareDialog, ToastHost },
    setup() {
      const issuedUrl = ref('')
      // Reveal is entered by the prop changing, which is what the page does when
      // the server answers. Setting it in args would leave the dialog on its
      // list.
      onMounted(() => setTimeout(() => (issuedUrl.value = mockShareUrl), 0))
      return { issuedUrl, shares: mockShareRows }
    },
    template: `
      <div>
        <ShareDialog
          :visible="true"
          research-name="Pricing benchmark, Q3"
          :shares="shares"
          :issued-url="issuedUrl"
          :recoverable-links="{}"
        />
        <ToastHost />
      </div>
    `,
  }),
}

/** A revoke in flight over the list. The row dims and locks; the rest of the
 *  dialog stays usable, because nothing else touches that record. */
export const RevokeInFlight: Story = {
  args: {
    shares: mockShareRows,
    recoverableLinks: mockRecoverableShareLinks,
    busyId: mockShareRows[0]!.id,
  },
}

/** Every link turned off. The list stays — the record of who had access is the
 *  point — and says plainly that nobody outside the team can open the research
 *  now. */
export const AllRevoked: Story = {
  args: {
    shares: mockShareRows.map((s) => ({ ...s, revoked_at: new Date().toISOString() })) as ShareRow[],
  },
}

/** Sixty links. The dialog body scrolls at 85vh and nothing paginates. */
export const ManyShares: Story = {
  args: { shares: mockManyShareRows },
}

/** A 200-character research name in the title. It wraps; the close button stays
 *  where it is. */
export const VeryLongResearchName: Story = {
  args: {
    shares: mockShareRows,
    researchName:
      'Ценообразование конкурентов в сегменте корпоративных подписок с посадочными местами, третий квартал, включая пересчёт февральских данных',
  },
}

/** Closed. Nothing renders; the trigger keeps focus. */
export const Hidden: Story = {
  args: { visible: false, shares: mockShareRows },
}

/**
 * The whole flow against a fake server: create, a beat of "Creating…", the
 * reveal, then the new row in the list with its link still recoverable.
 *
 * Labelling a link `taken` returns a refusal, so the error branch is reachable
 * without touching the args. Revoke flips the row after the server answers, not
 * before.
 */
export const Interactive: Story = {
  render: () => ({
    components: { ShareDialog, ToastHost },
    setup() {
      const visible = ref(false)
      const creating = ref(false)
      const error = ref('')
      const issuedUrl = ref('')
      const busyId = ref<string | undefined>(undefined)
      const shares = ref<ShareRow[]>([])
      const links = ref<Record<string, string>>({})
      const lastPayload = ref<unknown>(null)

      function open() {
        error.value = ''
        issuedUrl.value = ''
        visible.value = true
      }

      function create(payload: {
        label: string
        include: { sessions: boolean; tasks: boolean; roadmaps: boolean; export: boolean }
        expires_in_days: number | null
        password: string
      }) {
        lastPayload.value = payload
        error.value = ''
        creating.value = true
        setTimeout(() => {
          creating.value = false
          if (payload.label === 'taken') {
            error.value = 'A share link with that label already exists on this research.'
            return
          }
          const id = `shr_${Math.random().toString(36).slice(2, 8)}`
          const url = `https://research.intruforce.com/s/${Math.random().toString(16).slice(2).padEnd(32, '0')}`
          shares.value = [
            {
              id,
              label: payload.label,
              include: payload.include,
              has_password: !!payload.password,
              expires_at: payload.expires_in_days
                ? new Date(Date.now() + payload.expires_in_days * 86_400_000).toISOString()
                : null,
              revoked_at: null,
              last_seen_at: null,
              view_count: 0,
              created_at: new Date().toISOString(),
              created_by_name: 'Elena Marsh',
            },
            ...shares.value,
          ]
          links.value = { ...links.value, [id]: url }
          issuedUrl.value = url
        }, 600)
      }

      function revoke(share: ShareRow) {
        busyId.value = share.id
        setTimeout(() => {
          shares.value = shares.value.map((s: ShareRow) =>
            s.id === share.id ? { ...s, revoked_at: new Date().toISOString() } : s,
          )
          busyId.value = undefined
        }, 600)
      }

      return {
        visible,
        creating,
        error,
        issuedUrl,
        busyId,
        shares,
        links,
        lastPayload,
        open,
        create,
        revoke,
      }
    },
    template: `
      <div style="padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-3); align-items: flex-start;">
        <button class="btn btn-sm btn-primary" @click="open">
          Share <span v-if="shares.filter(s => !s.revoked_at).length" class="btn-count">{{ shares.filter(s => !s.revoked_at).length }}</span>
        </button>
        <p style="margin: 0; font-size: var(--type-xs); color: var(--color-text-muted); overflow-wrap: anywhere;">
          Last create payload: <code>{{ lastPayload ? JSON.stringify(lastPayload) : '—' }}</code>
        </p>
        <ShareDialog
          :visible="visible"
          research-name="Pricing benchmark, Q3"
          :shares="shares"
          :creating="creating"
          :error="error"
          :issued-url="issuedUrl"
          :busy-id="busyId"
          :recoverable-links="links"
          @create="create"
          @revoke="revoke"
          @dismiss-reveal="issuedUrl = ''"
          @close="visible = false"
        />
        <ToastHost />
      </div>
    `,
  }),
}
