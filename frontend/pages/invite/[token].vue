<template>
  <div class="centered-gate invite-page">
    <div class="invite-card">
      <!-- No partially-rendered team name: a skeleton, then the real thing. -->
      <div v-if="loading" class="skeleton-card invite-skeleton"></div>

      <template v-else-if="status === 'pending' && !preview?.already_member">
        <p class="invite-from">
          <span class="inviter">{{ preview?.inviter_name || 'Someone' }}</span><br />
          invited you to
        </p>
        <h1 class="invite-team">{{ preview?.team_name }}</h1>
        <p class="invite-role">as {{ article }} {{ preview?.role }}</p>
        <p class="invite-role-desc">{{ roleDescription }}</p>

        <p v-if="signedIn && !preview?.email_matches" class="mismatch">
          This invite was sent to {{ preview?.email }}. You're signed in as {{ user?.email }}.
        </p>

        <template v-if="signedIn">
          <button class="btn btn-primary invite-action" :disabled="joining" @click="join">
            {{ joining ? 'Joining…' : preview?.email_matches ? 'Join team' : `Join as ${user?.email}` }}
          </button>
          <p v-if="joinError" class="inline-error" role="alert">{{ joinError }}</p>
          <p class="signed-in-as">
            Signed in as {{ user?.email }}.
            <button class="text-link link-btn" @click="switchAccount">Not you? Sign in as someone else</button>
          </p>
        </template>

        <template v-else>
          <NuxtLink class="btn btn-primary invite-action" :to="loginLink">Sign in to join</NuxtLink>
          <!-- Offered even where registration is closed: a valid invitation is
               the authorization, and the server accepts one on this path. -->
          <p class="signed-in-as">
            <NuxtLink :to="registerLink" class="text-link">Create an account</NuxtLink>
          </p>
        </template>
      </template>

      <template v-else-if="preview?.already_member">
        <h1 class="invite-team">You're already in {{ preview.team_name }}</h1>
        <!-- Deliberately silent about which role: the preview carries the
             invitation's, which is not what an existing member holds. -->
        <NuxtLink class="btn btn-primary invite-action" :to="`/?team=${preview.team_id}`">
          Open the team's researches
        </NuxtLink>
        <p class="signed-in-as"><NuxtLink to="/" class="text-link">All researches</NuxtLink></p>
      </template>

      <template v-else-if="status === 'expired' || status === 'revoked' || status === 'accepted'">
        <h1 class="invite-team">{{ deadTitle }}</h1>
        <p class="invite-role-desc">{{ deadReason }}</p>
        <button v-if="preview?.inviter_email" class="btn invite-action" @click="copyInviter">
          {{ copied ? '✓ Copied' : `Copy ${preview.inviter_email}` }}
        </button>
        <p v-if="preview?.inviter_email && clipboardBroken" class="invite-role-desc selectable">
          {{ preview.inviter_email }}
        </p>
        <p v-if="signedIn" class="signed-in-as"><NuxtLink to="/" class="text-link">Go to Research</NuxtLink></p>
      </template>

      <template v-else-if="loadFailed">
        <h1 class="invite-team">Couldn't check this invite</h1>
        <p class="invite-role-desc">The server didn't answer. The link is probably fine — try again in a moment.</p>
        <button class="btn btn-primary invite-action" @click="loadPreview">Try again</button>
      </template>

      <template v-else>
        <h1 class="invite-team">This link isn't valid</h1>
        <p class="invite-role-desc">Check that you copied the whole link — they're long and easy to cut short.</p>
        <NuxtLink class="btn btn-primary invite-action" :to="signedIn ? '/' : loginLink">
          {{ signedIn ? 'Go to Research' : 'Sign in' }}
        </NuxtLink>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ROLE_DESCRIPTIONS, type TeamRole } from '~/composables/useTeams'

/**
 * Where an invitation link lands.
 *
 * Full-page and chrome-free like the auth pages — app.vue drops the nav for
 * `/invite/*`: the recipient may not have an account, and a nav bar full of
 * links they cannot follow is a worse welcome than none.
 *
 * Every outcome — valid, wrong account, already in, expired, revoked, bogus —
 * is one component with a status, because they differ by two sentences and a
 * button, not by layout.
 */
interface InvitePreview {
  status: 'pending' | 'expired' | 'revoked' | 'accepted' | 'unknown'
  team_id?: string
  team_name?: string
  role?: TeamRole
  email?: string
  inviter_name?: string
  inviter_email?: string
  already_member: boolean
  email_matches: boolean
  signed_in: boolean
}

const route = useRoute()
const token = computed(() => String(route.params.token))
const config = useRuntimeConfig()
const base = config.public.apiBase || ''

const { user, authFetch, isAuthenticated, logout } = useAuth()
const { refresh: refreshTeams } = useTeams()
const { success } = useToasts()

const preview = ref<InvitePreview | null>(null)
const loading = ref(true)
const joining = ref(false)
const joinError = ref('')
const copied = ref(false)
const clipboardBroken = ref(false)

const status = computed(() => preview.value?.status ?? 'unknown')
const signedIn = computed(() => !!isAuthenticated.value)
const article = computed(() => (preview.value?.role === 'owner' || preview.value?.role === 'editor' ? 'an' : 'a'))
const roleDescription = computed(() => (preview.value?.role ? ROLE_DESCRIPTIONS[preview.value.role] : ''))

// A link works once. Pasted into a group chat, the first person takes it and
// everyone else must be told that — not that they mis-copied a link that is
// perfectly intact.
const sender = computed(() => preview.value?.inviter_name || 'the sender')
const deadTitle = computed(() => {
  if (status.value === 'expired') return 'This invite has expired'
  if (status.value === 'revoked') return 'This invite was cancelled'
  return 'This invite has already been used'
})
const deadReason = computed(() => {
  if (status.value === 'expired') return `Invites last two weeks. Ask ${sender.value} to send you a new link.`
  if (status.value === 'revoked') return `This invite was cancelled. Ask ${sender.value} for a new one if you still need access.`
  return `Someone has already joined with this link — each one works once. Ask ${sender.value} for your own.`
})

// The destination rides along so signing in or registering comes back here
// rather than dropping the reader on the research list with no explanation.
const loginLink = computed(() => `/login?next=${encodeURIComponent(route.fullPath)}`)
const registerLink = computed(() => {
  // The token travels with the reader: registration presents it, the server
  // accepts the account, and the join happens in the same request.
  const params = new URLSearchParams({ next: route.fullPath, invite: token.value })
  if (preview.value?.email) params.set('email', preview.value.email)
  return `/register?${params.toString()}`
})

// A transport failure is not a malformed link, and telling a reader to check
// how they copied it is the wrong instruction for a server that is down.
const loadFailed = ref(false)

async function loadPreview() {
  loading.value = true
  loadFailed.value = false
  try {
    // The route tolerates having no session; when there is one it answers with
    // more, which is what makes "already a member" visible before pressing Join.
    const res = await authFetch<{ data: InvitePreview }>(`${base}/api/invites/${token.value}`)
    preview.value = res.data
  } catch {
    loadFailed.value = true
    preview.value = null
  } finally {
    loading.value = false
  }
}

onMounted(loadPreview)

async function join() {
  joining.value = true
  joinError.value = ''
  try {
    const res = await authFetch<{ data: { team_id: string; team_name: string; role: string } }>(
      `${base}/api/invites/${token.value}/accept`,
      { method: 'POST' },
    )
    await refreshTeams()
    success(`You joined ${res.data.team_name} as ${res.data.role}.`, 'Welcome')
    navigateTo(`/?team=${res.data.team_id}`)
  } catch (e: any) {
    joinError.value = e?.data?.error || 'Could not join the team'
  } finally {
    joining.value = false
  }
}

// The guard sends an authenticated visitor straight back here, so a bare link
// to /login is a no-op. The session has to go first.
function switchAccount() {
  logout(`/login?next=${encodeURIComponent(route.fullPath)}`)
}

async function copyInviter() {
  const email = preview.value?.inviter_email
  if (!email) return
  try {
    if (!navigator.clipboard) throw new Error('no clipboard')
    await navigator.clipboard.writeText(email)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  } catch {
    clipboardBroken.value = true
  }
}
</script>

<style scoped>
.invite-card { width: 100%; max-width: 440px; text-align: center; }
.invite-skeleton { height: 260px; }
.invite-from { font-size: var(--type-sm); color: var(--color-text-muted); margin-bottom: var(--space-4); }
.inviter { color: var(--color-text); }
.invite-team {
  font-size: var(--type-2xl);
  font-weight: var(--weight-bold);
  letter-spacing: -0.03em;
  line-height: var(--line-tight);
  overflow-wrap: anywhere;
}
.invite-role { font-size: var(--type-sm); color: var(--color-text-muted); margin-top: var(--space-3); }
.invite-role-desc {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  width: 100%;
  max-width: 40ch;
  margin: var(--space-2) auto 0;
}
.mismatch { font-size: var(--type-xs); color: var(--color-warning); margin-top: var(--space-4); }
.invite-action { width: 100%; justify-content: center; margin-top: var(--space-6); }
.signed-in-as { font-size: var(--type-xs); color: var(--color-text-muted); margin-top: var(--space-4); }
.text-link { color: var(--color-primary); }
.selectable { user-select: all; }
</style>
