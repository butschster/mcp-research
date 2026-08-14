<template>
  <ModalOverlay :visible="visible" size="sm" :labelledby="titleId" @close="emit('close')">
    <!-- Compose -->
    <template v-if="!link">
      <h3 :id="titleId" class="modal-title">Invite to {{ teamName }}</h3>

      <label class="field-label" :for="emailId">Email</label>
      <input
        :id="emailId"
        ref="emailInput"
        v-model="email"
        type="email"
        class="text-input"
        placeholder="colleague@example.com"
        autocomplete="off"
        @keydown.enter="submit"
      />

      <label class="field-label" :for="roleId">Role</label>
      <TeamRoleSelect :input-id="roleId" v-model="role" describe label="Role for the invited person" />

      <p v-if="error" class="inline-error" role="alert">{{ error }}</p>

      <div class="modal-actions">
        <button class="btn btn-sm" :disabled="creating" @click="emit('close')">Cancel</button>
        <button class="btn btn-sm btn-primary" :disabled="!canSubmit || creating" @click="submit">
          {{ creating ? 'Creating…' : 'Create' }}
        </button>
      </div>
    </template>

    <!-- The link, shown once -->
    <template v-else>
      <h3 :id="titleId" class="modal-title">Link ready</h3>
      <p class="modal-help">
        Send this link to {{ email || 'whoever should join' }}. It works once, and expires in two weeks.
      </p>

      <div class="empty-command link-block">
        <input
          v-if="clipboardBroken"
          ref="linkInput"
          class="command-text link-input"
          :value="link"
          readonly
          @focus="($event.target as HTMLInputElement).select()"
        />
        <code v-else class="command-text">{{ link }}</code>
        <button v-if="!clipboardBroken" ref="copyBtn" class="copy-btn" :class="{ copied }" @click="copy">
          {{ copied ? '✓ Copied' : 'Copy' }}
        </button>
      </div>
      <p v-if="clipboardBroken" class="modal-help">
        Copying isn't available on this connection — select the link above and copy it manually.
      </p>

      <div class="modal-actions">
        <button class="btn btn-sm btn-primary" @click="emit('close')">Done</button>
      </div>
    </template>

    <!-- Outside both branches: a live region announces nothing if it enters
         the DOM with its text already in it. -->
    <p class="sr-only" aria-live="polite">{{ announcement }}</p>
  </ModalOverlay>
</template>

<script setup lang="ts">
import type { TeamRole } from '~/composables/useTeams'
// Explicit: Nuxt's auto-import does not exist in Storybook, and a component
// that reaches for a composable without it crashes in the catalogue.
import { useToasts } from '~/composables/useToasts'

/**
 * Creates an invitation and hands over the link.
 *
 * Two states, one dialog: compose, then the link. The link is not a second
 * screen because it is the result of the first — navigating away from a token
 * that is shown exactly once is how it gets lost.
 *
 * Viewer is the default role. An over-permissioned colleague is silent and
 * permanent; an under-permissioned one says so in thirty seconds and the fix is
 * one select.
 */
const props = defineProps<{
  visible: boolean
  teamName: string
  /** The link, once the server has issued it. Empty keeps the compose state. */
  link?: string
  creating?: boolean
  error?: string
  /** Prefills for "re-invite", so a lapsed invitation is one click to replace. */
  prefillEmail?: string
  prefillRole?: TeamRole
}>()

const emit = defineEmits<{ create: [payload: { email: string; role: TeamRole }]; close: [] }>()

const uid = useId()
const titleId = `invite-title-${uid}`
const emailId = `invite-email-${uid}`
const roleId = `invite-role-${uid}`

const email = ref('')
const role = ref<TeamRole>('viewer')
const copied = ref(false)
const clipboardBroken = ref(false)
const announcement = ref('')

const emailInput = ref<HTMLInputElement | null>(null)
const copyBtn = ref<HTMLButtonElement | null>(null)
const linkInput = ref<HTMLInputElement | null>(null)

const canSubmit = computed(() => email.value.includes('@'))

watch(
  () => props.visible,
  async (open) => {
    if (!open) {
      copied.value = false
      clipboardBroken.value = false
      announcement.value = ''
      return
    }
    email.value = props.prefillEmail ?? ''
    role.value = props.prefillRole ?? 'viewer'
    await nextTick()
    emailInput.value?.focus()
  },
  { immediate: true },
)

// When the link arrives the body swaps under the reader's hands, so focus goes
// to the one thing they came for and a live region says why.
watch(
  () => props.link,
  async (link) => {
    if (!link) return
    announcement.value = 'Invite link ready. Copy it now — it is shown once.'
    await nextTick()
    copyBtn.value?.focus()
  },
)

function submit() {
  if (!canSubmit.value || props.creating) return
  emit('create', { email: email.value.trim(), role: role.value })
}

async function copy() {
  const { success } = useToasts()
  try {
    // Deployments here include plain HTTP on a LAN, where the clipboard API is
    // simply absent. Falling back to a selected input beats a dead button.
    if (!navigator.clipboard) throw new Error('no clipboard')
    await navigator.clipboard.writeText(props.link!)
    copied.value = true
    success('Invite link copied')
    setTimeout(() => (copied.value = false), 2000)
  } catch {
    clipboardBroken.value = true
    await nextTick()
    linkInput.value?.focus()
    linkInput.value?.select()
  }
}
</script>

<style scoped>
.link-block { margin-top: 0; max-width: none; }
.link-input { background: none; border: none; color: var(--color-primary); font-family: inherit; width: 100%; }
.link-input:focus { outline: none; }
</style>
