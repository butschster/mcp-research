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

      <fieldset class="share-fieldset">
        <legend class="field-label">What the link shows</legend>
        <p class="modal-help">Sections, entries and cross-references &middot; always</p>
        <label class="check-row"><input v-model="form.roadmaps" type="checkbox" /> Roadmaps</label>
        <label class="check-row"><input v-model="form.sessions" type="checkbox" /> Interview sessions, with questions and answers</label>
        <label class="check-row"><input v-model="form.tasks" type="checkbox" /> Tasks</label>
        <label class="check-row"><input v-model="form.export" type="checkbox" /> Downloading the research as a file</label>
      </fieldset>

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

    <!-- List -->
    <template v-else>
      <h3 :id="titleId" class="modal-title">Share &ldquo;{{ researchName }}&rdquo;</h3>
      <p class="modal-help">
        Anyone with the link can read this research. They don't need an account.
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
          @show-link="showLink"
        />
        <p v-if="shares.length && !shares.some(isLive)" class="modal-help">
          All links revoked. Nobody outside the team can open this research now.
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
}>()

const emit = defineEmits<{
  create: [payload: {
    label: string
    include: { sessions: boolean; tasks: boolean; roadmaps: boolean; export: boolean }
    expires_in_days: number | null
    password: string
  }]
  revoke: [share: ShareRow]
  close: []
  dismissReveal: []
}>()

const uid = useId()
const titleId = `share-title-${uid}`
const labelId = `share-label-${uid}`
const expiryId = `share-expiry-${uid}`

type View = 'list' | 'create' | 'reveal'
const view = ref<View>('list')
const copiedOnce = ref(false)
const announcement = ref('')
const shownLink = ref('')

const labelEl = ref<HTMLInputElement | null>(null)
const newLinkEl = ref<HTMLButtonElement | null>(null)
const secretEl = ref<{ focus: () => void; copy: () => Promise<void> } | null>(null)

const form = reactive({
  label: '',
  roadmaps: true,
  sessions: false,
  tasks: false,
  export: true,
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
  form.roadmaps = true
  form.sessions = false
  form.tasks = false
  form.export = true
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
    include: {
      sessions: form.sessions,
      tasks: form.tasks,
      roadmaps: form.roadmaps,
      export: form.export,
    },
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
  emit('close')
}

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
.share-fieldset { border: none; padding: 0; margin: var(--space-4) 0; }
.check-row { display: flex; align-items: center; gap: var(--space-2); font-size: var(--type-sm); padding: var(--space-1) 0; }
.share-skeletons { display: flex; flex-direction: column; gap: var(--space-2); }
/* .warning-banner is the product's amber strip (system.css). Amber belongs here
   and nowhere else in this feature: this is the one moment where something is
   about to be lost. */
.share-warning { margin: var(--space-3) 0; }
.mt-3 { margin-top: var(--space-3); }
</style>
