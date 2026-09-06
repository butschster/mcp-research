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
 * Creating, listing, editing and revoking the links for one research.
 *
 * Four views in one dialog, for the same reason the invite dialog has two: the
 * link is the result of the form, and navigating away from a value shown exactly
 * once is how it gets lost.
 *
 * **Edit changes a live link in place.** Its address does not change, so the
 * recipient keeps the URL they were sent — which is the point, and also the
 * reason widening is confirmed: everybody already holding the link gets the
 * extra content the moment Save lands, with no further act by anyone and no way
 * to tell who that is. The warning names the flags going from off to on, in the
 * words the checkboxes use, and says how often the link has been opened.
 *
 * Save stays disabled until something differs. A no-op write is still a write,
 * and it would push `share.updated` to every open tab for nothing.
 *
 * **The edit view is entered by clicking Edit on a live row**, not by a prop, so
 * every edit story below drives it with a `play` function — the same way the
 * reveal stories drive `issuedUrl`.
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
    saving: { control: 'boolean' },
    saveError: { control: 'text' },
    savedTick: { control: 'number' },
    onCreate: { action: 'create' },
    onRevoke: { action: 'revoke' },
    onUpdate: { action: 'update' },
    onRefresh: { action: 'refresh' },
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
  parameters: { docs: { story: { autoplay: true } } },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    // Through the helper, which searches the document: this dialog teleports to
    // `body`, so the previous `canvasElement.querySelectorAll` matched nothing
    // and the optional-chained click made the story quietly show the list it
    // was meant to click past.
    await clickButton(canvasElement, '+ New link')
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
 * The edit form, reached by clicking Edit on the first live row.
 *
 * The same four checkboxes as the create form, from the same component — two
 * hand-maintained copies of them is how one ends up describing a flag the other
 * has renamed. There is no expiry field and no password field here: those change
 * whether the link works, not what it shows, and turning a link off is what
 * Revoke is for.
 */
export const EditLink: Story = {
  args: { shares: mockShareRows, recoverableLinks: mockRecoverableShareLinks },
  parameters: { docs: { story: { autoplay: true } } },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Edit')
  },
}

/**
 * Sessions ticked on a link that did not have them.
 *
 * The amber note is the second and last amber in this feature — the first is the
 * reveal — and for the same kind of reason: something is about to happen that
 * cannot be walked back quietly. Everyone holding this link gets the interview
 * transcripts the moment Save lands, and the sentence about 47 opens is there to
 * say that "everyone" is not hypothetical.
 *
 * Narrowing does not warn. Taking something away from a link is the safe
 * direction, and a confirmation on it would train the owner to click through
 * the one that matters.
 */
export const EditWidening: Story = {
  args: { shares: mockShareRows, recoverableLinks: mockRecoverableShareLinks },
  parameters: { docs: { story: { autoplay: true } } },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Edit')
    // Order in the fieldset: roadmaps, sessions, tasks, export.
    await clickCheckbox(canvasElement, 1)
  },
}

/** Saving. Both buttons lock, so a second Enter cannot send the same change
 *  twice — and the second send would be the one whose answer the dialog acts
 *  on. */
export const EditSaving: Story = {
  args: { shares: mockShareRows, recoverableLinks: mockRecoverableShareLinks, saving: true },
  parameters: { docs: { story: { autoplay: true } } },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Edit')
    await clickCheckbox(canvasElement, 1)
  },
}

/** The server refused for an ordinary reason. Said inline, above the actions,
 *  with the form still filled in — a toast would take the message away from the
 *  fields that have to change. */
export const EditFailed: Story = {
  args: {
    shares: mockShareRows,
    recoverableLinks: mockRecoverableShareLinks,
    saveError: 'Another link on this project already uses that label.',
  },
  parameters: { docs: { story: { autoplay: true } } },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Edit')
  },
}

/**
 * `saveError: 'dead'` — the link was revoked or expired by somebody else while
 * this form was open, and the server answered 409.
 *
 * The whole form is replaced rather than annotated. Leaving the fields on screen
 * under an error would invite another Save against a link that no longer exists,
 * and the only useful thing left to do is go back to a list that now says
 * something different. "Nothing was changed" is stated, because the owner just
 * pressed Save and has no other way to know.
 *
 * Going back emits `refresh`, so the list is refetched rather than re-showing
 * the stale row that caused this.
 */
export const EditLinkDied: Story = {
  args: {
    shares: mockShareRows,
    recoverableLinks: mockRecoverableShareLinks,
    saveError: 'dead',
  },
  parameters: { docs: { story: { autoplay: true } } },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Edit')
  },
}

/**
 * A link with no label. The title falls back to "Untitled link" rather than
 * rendering an empty pair of quotation marks, and the field is genuinely empty
 * so the owner can give it a name now.
 */
export const EditUnlabelledLink: Story = {
  args: {
    shares: [{ ...mockShareRows[0]!, id: 'shr_unlabelled', label: '' }] as ShareRow[],
    recoverableLinks: {},
  },
  parameters: { docs: { story: { autoplay: true } } },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Edit')
  },
}

/**
 * The whole flow against a fake server: create, a beat of "Creating…", the
 * reveal, then the new row in the list with its link still recoverable.
 *
 * Labelling a link `taken` returns a refusal, so the error branch is reachable
 * without touching the args. Revoke flips the row after the server answers, not
 * before.
 *
 * Edit is wired the same way. Save a link relabelled `taken` and the server
 * refuses inline; relabel one `dead` and it answers that the link is gone, which
 * is the 409 branch. Anything else lands: the row repaints from the server's
 * answer — not from what was typed — and `savedTick` is what sends the dialog
 * back to the list.
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
      const saving = ref(false)
      const saveError = ref('')
      const savedTick = ref(0)

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

      function update(payload: {
        id: string
        label: string
        include: { sessions: boolean; tasks: boolean; roadmaps: boolean; export: boolean }
      }) {
        lastPayload.value = payload
        saveError.value = ''
        saving.value = true
        setTimeout(() => {
          saving.value = false
          if (payload.label === 'taken') {
            saveError.value = 'Another link on this project already uses that label.'
            return
          }
          // What the server says when the row is revoked or expired: 409, and
          // the dialog turns that into the one screen with a way out.
          if (payload.label === 'dead') {
            saveError.value = 'dead'
            return
          }
          // Repaint from the answer, not from the form. The server is what
          // decides what the link now shows.
          shares.value = shares.value.map((s: ShareRow) =>
            s.id === payload.id ? { ...s, label: payload.label, include: { ...payload.include } } : s,
          )
          savedTick.value++
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
        saving,
        saveError,
        savedTick,
        open,
        create,
        revoke,
        update,
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
          :saving="saving"
          :save-error="saveError"
          :saved-tick="savedTick"
          @create="create"
          @revoke="revoke"
          @update="update"
          @refresh="saveError = ''"
          @dismiss-reveal="issuedUrl = ''"
          @close="visible = false"
        />
        <ToastHost />
      </div>
    `,
  }),
}

/**
 * Clicks the first button whose text matches, once it is on screen.
 *
 * There is no `@storybook/test` in this project, so the catalogue polls — the
 * same helper shape `ActionMenu.stories.ts` uses. Buttons are matched by their
 * visible text, because that is what the person driving this dialog matches on
 * too.
 *
 * **It searches the document, not `canvasElement`.** `ModalOverlay` teleports
 * to `body`, so nothing in this dialog is inside the story's root element —
 * a `play` that queried `canvasElement` found no buttons and, if it used
 * optional chaining, reported success while doing nothing at all.
 */
async function clickButton(root: HTMLElement, text: string, index = 0): Promise<void> {
  const doc = root.ownerDocument.body
  for (let i = 0; i < 50; i++) {
    const matches = Array.from(doc.querySelectorAll('button')).filter(
      (b) => b.textContent?.trim() === text,
    )
    const button = matches[index]
    if (button) {
      button.click()
      // Let the view swap and its focus land before the next step.
      await new Promise((resolve) => setTimeout(resolve, 20))
      return
    }
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
  throw new Error(`Button "${text}" #${index} never appeared`)
}

/**
 * Ticks the nth include checkbox in whichever form is on screen.
 *
 * By position rather than by label: the four labels are the words the product
 * shows an owner, and a story that hard-codes them would fail on a wording
 * change that is not a behaviour change. The order — roadmaps, sessions, tasks,
 * export — is `ShareIncludeFields`'s and is checked there.
 */
async function clickCheckbox(root: HTMLElement, index: number): Promise<void> {
  const doc = root.ownerDocument.body
  for (let i = 0; i < 50; i++) {
    const boxes = Array.from(doc.querySelectorAll<HTMLInputElement>('input[type="checkbox"]'))
    const box = boxes[index]
    if (box) {
      box.click()
      await new Promise((resolve) => setTimeout(resolve, 20))
      return
    }
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
  throw new Error(`Checkbox #${index} never appeared`)
}
