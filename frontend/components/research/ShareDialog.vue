<template>
  <ModalOverlay :visible="visible" size="lg" :labelledby="titleId" @close="requestClose">
    <!-- The link, shown once -->
    <template v-if="view === 'reveal'">
      <h3 :id="titleId" class="modal-title">Link ready</h3>
      <div class="warning-banner share-warning" role="note">
        This is the only time this link is shown. Copy it now — it cannot be shown again after you
        reload this page.
      </div>

      <CopyableSecret
        ref="secretEl"
        :value="issuedUrl"
        :hint="revealHint"
        toast="Share link copied"
        @copied="copiedOnce = true"
      />

      <div class="modal-actions">
        <button class="btn btn-sm btn-primary" @click="finishReveal">
          {{ copiedOnce ? 'Done' : 'Copy and finish' }}
        </button>
      </div>
    </template>

    <!-- Create -->
    <template v-else-if="view === 'create'">
      <h3 :id="titleId" class="modal-title">New share link</h3>

      <label class="field-label" :for="labelId">Label</label>
      <input
        :id="labelId"
        ref="labelEl"
        v-model="form.label"
        class="text-input"
        placeholder="Client review, March"
        autocomplete="off"
      />
      <p class="modal-help">Only you see this. It's how you'll recognise the link.</p>

      <ResearchShareIncludeFields v-model="form.include" />

      <label class="field-label" :for="expiryId">Stops working</label>
      <select :id="expiryId" v-model="form.expiry" class="text-input">
        <option value="7">In 7 days</option>
        <option value="30">In 30 days</option>
        <option value="90">In 90 days</option>
        <option value="">Never</option>
      </select>

      <label class="check-row mt-3"><input v-model="form.withPassword" type="checkbox" /> Require a password</label>
      <input
        v-if="form.withPassword"
        v-model="form.password"
        type="text"
        class="text-input"
        placeholder="At least 6 characters"
        autocomplete="off"
      />

      <p v-if="error" class="inline-error" role="alert">{{ error }}</p>

      <div class="modal-actions">
        <button class="btn btn-sm" :disabled="creating" @click="cancelCreate">Cancel</button>
        <button class="btn btn-sm btn-primary" :disabled="creating || !passwordOk" @click="submit">
          {{ creating ? 'Creating…' : 'Create link' }}
        </button>
      </div>
    </template>

    <!-- Edit: what a live link is called and what it shows. The address does
         not change, which is the point — and the reason widening confirms. -->
    <template v-else-if="view === 'edit' && editing">
      <template v-if="saveError === 'dead'">
        <h3 :id="titleId" class="modal-title">This link is no longer live</h3>
        <p class="modal-help">It was revoked or expired while you had this open. Nothing was changed.</p>
        <div class="modal-actions">
          <button ref="backToListEl" class="btn btn-sm btn-primary" @click="backToList">Back to links</button>
        </div>
      </template>
      <template v-else>
        <h3 :id="titleId" class="modal-title">Edit &ldquo;{{ editing.label || 'Untitled link' }}&rdquo;</h3>

        <label class="field-label" :for="editLabelId">Label</label>
        <input
          :id="editLabelId"
          ref="editLabelEl"
          v-model="editForm.label"
          class="text-input"
          placeholder="Client review, March"
          autocomplete="off"
        />
        <p class="modal-help">Only you see this. It's how you'll recognise the link.</p>

        <ResearchShareIncludeFields v-model="editForm.include" />

        <div v-if="widening.length" class="warning-banner share-warning" role="note">
          Adding {{ widening.join(', ') }} shows more to everyone who already has this link.
          {{ viewsSentence }}
        </div>

        <p v-if="saveError" class="inline-error" role="alert">{{ saveError }}</p>

        <div class="modal-actions">
          <button class="btn btn-sm" :disabled="saving" @click="cancelEdit">Cancel</button>
          <button class="btn btn-sm btn-primary" :disabled="saving || !editDirty" @click="submitEdit">
            {{ saving ? 'Saving…' : 'Save' }}
          </button>
        </div>
      </template>
    </template>

    <!-- List -->
    <template v-else>
      <h3 :id="titleId" class="modal-title">Share &ldquo;{{ researchName }}&rdquo;</h3>
      <p class="modal-help">
        Anyone with the link can read this project. They don't need an account.
      </p>

      <div class="share-list-head">
        <button ref="newLinkEl" class="btn btn-sm btn-primary" @click="startCreate">+ New link</button>
      </div>

      <div v-if="loading" class="share-skeletons">
        <div v-for="i in 3" :key="i" class="skeleton-card" style="height: 52px;"></div>
      </div>
      <template v-else>
        <ResearchShareRowList
          :shares="shares"
          :busy-id="busyId"
          :recoverable-links="recoverableLinks"
          @revoke="emit('revoke', $event)"
          @edit="startEdit"
          @show-link="showLink"
        />
        <p v-if="shares.length && !shares.some(isLive)" class="modal-help">
          All links revoked. Nobody outside the team can open this project now.
        </p>
      </template>
    </template>

    <!-- A live region announces nothing if it enters the DOM with its text
         already in it, so it lives outside all three branches. -->
    <p class="sr-only" aria-live="polite">{{ announcement }}</p>
  </ModalOverlay>
</template>

<script setup lang="ts">
import { useToasts } from '~/composables/useToasts'
import type { ShareInclude } from '~/composables/useShare'
import type { ShareRow } from './ShareRowList.vue'

/**
 * Creating, listing and revoking the links for one research.
 *
 * Three views in one dialog, for the same reason the invite dialog has two: the
 * link is the result of the form, and navigating away from a value shown
 * exactly once is how it gets lost.
 *
 * With no links yet it opens straight onto the form. That is the empty state —
 * the lead sentence already explains what a share link is, and a screen whose
 * only content is a button and a paragraph is a step, not information.
 */
const props = defineProps<{
  visible: boolean
  researchName: string
  shares: ShareRow[]
  loading?: boolean
  creating?: boolean
  error?: string
  /** The URL the server just issued, which puts the dialog in its reveal view. */
  issuedUrl?: string
  busyId?: string
  /** Links still in memory from this tab, by share id. */
  recoverableLinks: Record<string, string>
  /** An edit is on its way to the server. */
  saving?: boolean
  /**
   * Why the last edit did not land. The sentinel `'dead'` means the server
   * answered that the link is revoked or expired; anything else is shown as
   * the message it is.
   */
  saveError?: string
  /** Increments when an edit lands, so the dialog can return to the list. */
  savedTick?: number
}>()

const emit = defineEmits<{
  create: [payload: {
    label: string
    include: { sessions: boolean; tasks: boolean; roadmaps: boolean; export: boolean }
    expires_in_days: number | null
    password: string
  }]
  revoke: [share: ShareRow]
  /** A live link's new label and full set of flags. */
  update: [payload: { id: string; label: string; include: ShareInclude }]
  /** The list is stale — after a link died under an open edit. */
  refresh: []
  close: []
  dismissReveal: []
}>()

const uid = useId()
const titleId = `share-title-${uid}`
const labelId = `share-label-${uid}`
const expiryId = `share-expiry-${uid}`
const editLabelId = `share-edit-label-${uid}`

type View = 'list' | 'create' | 'reveal' | 'edit'
const view = ref<View>('list')
const copiedOnce = ref(false)
const announcement = ref('')
const shownLink = ref('')

const labelEl = ref<HTMLInputElement | null>(null)
const newLinkEl = ref<HTMLButtonElement | null>(null)
const secretEl = ref<{ focus: () => void; copy: () => Promise<void> } | null>(null)

const form = reactive({
  label: '',
  include: { roadmaps: true, sessions: false, tasks: false, export: true } as ShareInclude,
  expiry: '30',
  withPassword: false,
  password: '',
})

/**
 * Back to defaults, every time the dialog opens.
 *
 * Without this, creating a passworded link for a client and then a second link
 * for a colleague arrived with the client's password still sitting in a
 * `type="text"` field, pre-checked. Silently reusing one recipient's password
 * for another is not something the owner asked for.
 */
function resetForm() {
  form.label = ''
  form.include = { roadmaps: true, sessions: false, tasks: false, export: true }
  form.expiry = '30'
  form.withPassword = false
  form.password = ''
}

const passwordOk = computed(() => !form.withPassword || form.password.trim().length >= 6)

const issuedUrl = computed(() => shownLink.value || props.issuedUrl || '')
const revealHint = computed(() =>
  form.withPassword ? 'Send the password separately — not in the same message.' : '',
)

const isLive = isShareLive

watch(
  () => props.visible,
  async (open) => {
    if (!open) {
      view.value = 'list'
      shownLink.value = ''
      copiedOnce.value = false
      announcement.value = ''
      return
    }
    // The list is fetched after the dialog opens, so `shares` is empty on the
    // first open of a page load. Deciding "no shares → straight to the form"
    // from that showed a blank create form to an owner whose button had just
    // said "6", with no obvious way back. While the list is loading, the list
    // is what we show.
    view.value = props.loading || props.shares.length ? 'list' : 'create'
    resetForm()
    await nextTick()
    if (view.value === 'create') labelEl.value?.focus()
    else newLinkEl.value?.focus()
  },
  { immediate: true },
)

// Once the list lands, an empty one becomes the create form after all — the
// empty state is the form, by construction.
watch(
  () => props.loading,
  async (loading) => {
    if (loading || !props.visible || view.value !== 'list' || props.shares.length) return
    view.value = 'create'
    await nextTick()
    labelEl.value?.focus()
  },
)

// The body swaps under the reader's hands when the link arrives, so focus goes
// to the one thing they came for and a live region says why.
watch(
  () => props.issuedUrl,
  async (url) => {
    if (!url) return
    view.value = 'reveal'
    shownLink.value = ''
    copiedOnce.value = false
    announcement.value = 'Link ready. Copy it now — it is shown once.'
    await nextTick()
    secretEl.value?.focus()
  },
)

async function startCreate() {
  view.value = 'create'
  await nextTick()
  labelEl.value?.focus()
}

async function cancelCreate() {
  if (!props.shares.length) return emit('close')
  view.value = 'list'
  await nextTick()
  newLinkEl.value?.focus()
}

function submit() {
  if (props.creating || !passwordOk.value) return
  emit('create', {
    label: form.label.trim(),
    include: { ...form.include },
    expires_in_days: form.expiry ? Number(form.expiry) : null,
    password: form.withPassword ? form.password.trim() : '',
  })
}

/**
 * Closing the reveal without copying is allowed.
 *
 * A dialog that refuses to close is a trap, and nothing is lost that cannot be
 * recovered: the link stays in memory for the life of this tab, and the row
 * keeps a `Show link` action until then. The one thing that cannot be recovered
 * is said in the strip above, before it happens.
 */
/**
 * "Copy and finish" copies, and then finishes.
 *
 * It used to only finish, and toast that the link had not been copied — a
 * button naming an action it did not perform, on the one screen where something
 * is actually lost. If the clipboard is unavailable the copy fails inside
 * CopyableSecret, which swaps to a selectable field, and the toast below is the
 * honest fallback.
 */
async function finishReveal() {
  if (!copiedOnce.value) {
    await secretEl.value?.copy()
  }
  if (!copiedOnce.value) {
    useToasts().push({
      // "Reload" was wrong: the links live in the research page's setup, so any
      // navigation away loses them, not only a refresh.
      message: 'Link not copied. You can still show it from the Share dialog until you leave this page.',
    })
  }
  emit('dismissReveal')
  emit('close')
}

function requestClose() {
  if (view.value === 'reveal') return finishReveal()
  // Escape out of a filled-in form means "back", not "throw it away".
  if (view.value === 'create') return cancelCreate()
  if (view.value === 'edit') return saveError.value === 'dead' ? backToList() : cancelEdit()
  emit('close')
}

// --- Edit ---

const editing = ref<ShareRow | null>(null)
const editForm = reactive({
  label: '',
  include: { roadmaps: true, sessions: false, tasks: false, export: true } as ShareInclude,
})
const editLabelEl = ref<HTMLInputElement | null>(null)
const backToListEl = ref<HTMLButtonElement | null>(null)

const saveError = computed(() => props.saveError || '')

async function startEdit(share: ShareRow) {
  editing.value = share
  editForm.label = share.label || ''
  editForm.include = { ...share.include }
  view.value = 'edit'
  await nextTick()
  editLabelEl.value?.focus()
}

/** Save is off while nothing differs: a no-op write is still a write. */
const editDirty = computed(() => {
  const s = editing.value
  if (!s) return false
  if (editForm.label.trim() !== (s.label || '')) return true
  return (['sessions', 'tasks', 'roadmaps', 'export'] as const).some(k => editForm.include[k] !== s.include[k])
})

const INCLUDE_WORDS: Record<keyof ShareInclude, string> = {
  roadmaps: 'roadmaps',
  sessions: 'sessions',
  tasks: 'tasks',
  export: 'downloading the project as a file',
}

/** The flags going from off to on, in the words the checkboxes use. */
const widening = computed(() => {
  const s = editing.value
  if (!s) return [] as string[]
  return (Object.keys(INCLUDE_WORDS) as (keyof ShareInclude)[])
    .filter(k => editForm.include[k] && !s.include[k])
    .map(k => INCLUDE_WORDS[k])
})

const viewsSentence = computed(() => {
  const n = editing.value?.view_count ?? 0
  if (!n) return "It hasn't been opened yet, but the link is already out."
  return `It has been opened ${n === 1 ? 'once' : n + ' times'}.`
})

function submitEdit() {
  if (!editing.value || props.saving || !editDirty.value) return
  emit('update', {
    id: editing.value.id,
    label: editForm.label.trim(),
    include: { ...editForm.include },
  })
}

async function cancelEdit() {
  editing.value = null
  view.value = 'list'
  await nextTick()
  newLinkEl.value?.focus()
}

async function backToList() {
  emit('refresh')
  await cancelEdit()
}

// The edit landed: back to the list, which the parent has already repainted
// from the server's answer, and say so where a screen reader will hear it.
watch(
  () => props.savedTick,
  async () => {
    if (view.value !== 'edit') return
    announcement.value = 'Link updated.'
    await cancelEdit()
  },
)

// A link that died under the open edit needs the way out to be reachable.
watch(saveError, async (err) => {
  if (err === 'dead') {
    await nextTick()
    backToListEl.value?.focus()
  }
})

async function showLink(share: ShareRow) {
  shownLink.value = props.recoverableLinks[share.id] || ''
  if (!shownLink.value) return
  view.value = 'reveal'
  copiedOnce.value = false
  await nextTick()
  secretEl.value?.focus()
}
</script>

<style scoped>
.share-list-head { display: flex; justify-content: flex-end; margin: var(--space-3) 0; }
.check-row { display: flex; align-items: center; gap: var(--space-2); font-size: var(--type-sm); padding: var(--space-1) 0; }
.share-skeletons { display: flex; flex-direction: column; gap: var(--space-2); }
/* .warning-banner is the product's amber strip (system.css). Amber belongs here
   and nowhere else in this feature: this is the one moment where something is
   about to be lost. */
.share-warning { margin: var(--space-3) 0; }
.mt-3 { margin-top: var(--space-3); }
</style>
