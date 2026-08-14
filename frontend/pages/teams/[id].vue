<template>
  <div>
    <Breadcrumbs :crumbs="[{ label: 'Research', to: '/' }, { label: 'Teams', to: '/teams' }, { label: team?.name || 'Team' }]" />

    <div v-if="loading" class="skeleton-list">
      <div class="skeleton-card title-skeleton"></div>
      <div v-for="i in 3" :key="i" class="skeleton-card member-skeleton"></div>
    </div>

    <!-- 403 and 404 look the same on purpose: telling an outsider that a team
         with this id exists is itself information. -->
    <EmptyState
      v-else-if="!team"
      icon="&#x1F50D;"
      title="Team not found"
      description="It may have been deleted, or you may no longer be a member."
    >
      <NuxtLink class="btn btn-sm" to="/teams">All teams</NuxtLink>
    </EmptyState>

    <template v-else>
      <div class="page-header">
        <div class="page-header-row">
          <h1 class="page-title">{{ team.name }}</h1>
          <div class="page-header-actions">
            <button v-if="canInvite" class="btn btn-sm btn-primary" @click="openInvite()">+ Invite</button>
            <ActionMenu v-if="canInvite">
              <button class="action-menu-item" @click="renaming = true">Rename team</button>
            </ActionMenu>
          </div>
        </div>
        <p class="page-lead">
          {{ members.length }} {{ members.length === 1 ? 'member' : 'members' }} · You are {{ article(team.role) }} {{ team.role }}
        </p>
      </div>

      <h2 ref="membersHeading" tabindex="-1" class="section-heading">Members</h2>
      <TeamMemberList
        :members="members"
        :my-user-id="user?.id || ''"
        :can-manage="canManage"
        :busy-user-id="busyUserId"
        @change-role="changeRole"
        @remove="askRemove"
      />
      <p v-if="members.length === 1" class="alone-note">
        Only you so far. Invite someone with a link — no email is sent, you pass the link along yourself.
        <button v-if="canInvite" class="btn btn-sm invite-inline" @click="openInvite()">+ Invite</button>
      </p>

      <template v-if="canManage && invites.length">
        <h2 class="section-heading">Pending invites</h2>
        <TeamInviteList
          :invites="invites"
          :busy-id="busyInviteId"
          :recoverable-links="recoverableLinks"
          @revoke="askRevoke"
          @show-link="showLink"
          @reinvite="reinvite"
        />
      </template>

      <p v-if="team.research_count > 0" class="team-work">
        <NuxtLink :to="`/?team=${team.id}`" class="text-link">
          {{ team.research_count }} {{ team.research_count === 1 ? 'research' : 'researches' }} in this team →
        </NuxtLink>
      </p>

      <div v-if="!team.personal" class="danger-zone">
        <button
          class="danger-row"
          :disabled="leaveBlocked"
          :title="leaveBlocked ? lastOwnerLeaveReason : undefined"
          :aria-describedby="leaveBlocked ? 'leave-reason' : undefined"
          @click="leaving = true"
        >
          <span>Leave team</span>
          <span class="danger-note">Removes your access to {{ team.research_count }} {{ team.research_count === 1 ? 'research' : 'researches' }}.</span>
        </button>
        <!-- A disabled control's title never reaches a keyboard or screen
             reader, and the reason is the whole point of refusing. -->
        <p v-if="leaveBlocked" id="leave-reason" class="sr-only">{{ lastOwnerLeaveReason }}</p>
        <button
          v-if="canManage"
          class="danger-row danger-strong"
          :disabled="team.research_count > 0"
          :title="team.research_count > 0 ? emptyFirstReason : undefined"
          :aria-describedby="team.research_count > 0 ? 'delete-reason' : undefined"
          @click="deleting = true"
        >
          <span>Delete team</span>
          <span class="danger-note">
            {{ team.research_count > 0 ? emptyFirstReason : 'The team is empty and can be deleted.' }}
          </span>
        </button>
        <p v-if="team.research_count > 0" id="delete-reason" class="sr-only">{{ emptyFirstReason }}</p>
      </div>
    </template>

    <TeamInviteDialog
      :visible="inviting"
      :team-name="team?.name || ''"
      :link="issuedLink"
      :creating="creatingInvite"
      :error="inviteError"
      :prefill-email="prefillEmail"
      :prefill-role="prefillRole"
      @create="createInvite"
      @close="closeInvite"
    />

    <ConfirmModal
      :visible="!!removing"
      title="Remove member"
      :message="`${removing?.name || removing?.email} loses access to every research in ${team?.name}.`"
      confirm-label="Remove"
      variant="danger"
      :loading="!!busyUserId"
      @confirm="confirmRemove"
      @cancel="removing = null"
    />

    <ConfirmModal
      :visible="!!revoking"
      title="Revoke invite"
      :message="`The link sent to ${revoking?.email || 'this person'} stops working.`"
      confirm-label="Revoke"
      variant="danger"
      :loading="!!busyInviteId"
      @confirm="confirmRevoke"
      @cancel="revoking = null"
    />

    <ConfirmModal
      :visible="leaving"
      title="Leave team"
      :message="`You lose access to every research in ${team?.name}. An owner can invite you back.`"
      confirm-label="Leave"
      variant="danger"
      :loading="busy"
      @confirm="confirmLeave"
      @cancel="leaving = false"
    />

    <ConfirmModal
      :visible="deleting"
      title="Delete team"
      :message="`${team?.name} is deleted for everyone in it.`"
      confirm-label="Delete"
      variant="danger"
      :loading="busy"
      @confirm="confirmDelete"
      @cancel="deleting = false"
    />

    <ModalOverlay :visible="renaming" size="sm" labelledby="rename-title" @close="renaming = false">
      <h3 id="rename-title" class="modal-title">Rename team</h3>
      <input v-model="newName" class="text-input" :placeholder="team?.name" @keydown.enter="confirmRename" />
      <div class="modal-actions">
        <button class="btn btn-sm" @click="renaming = false">Cancel</button>
        <button class="btn btn-sm btn-primary" :disabled="!newName.trim()" @click="confirmRename">Rename</button>
      </div>
    </ModalOverlay>
  </div>
</template>

<script setup lang="ts">
import type { TeamInvite, TeamMember, TeamRole } from '~/composables/useTeams'

const route = useRoute()
const teamId = computed(() => String(route.params.id))
const config = useRuntimeConfig()
const base = config.public.apiBase || ''

const { user, authFetch } = useAuth()
const { teams, load: loadTeams, refresh: refreshTeams, rename, remove: removeTeam } = useTeams()
const { success, error: errorToast } = useToasts()

const members = ref<TeamMember[]>([])
const invites = ref<TeamInvite[]>([])
const loading = ref(true)
const busy = ref(false)
const busyUserId = ref<string | null>(null)
const busyInviteId = ref<string | null>(null)

const inviting = ref(false)
const creatingInvite = ref(false)
const inviteError = ref('')
const issuedLink = ref('')
const prefillEmail = ref('')
const prefillRole = ref<TeamRole>('viewer')

/**
 * Links issued in this tab, by invite id.
 *
 * The server hashes the token, so a link exists in exactly one place after it
 * is created: here, for as long as the tab lives. After a reload the row offers
 * a replacement instead of pretending it can show the old one.
 */
const recoverableLinks = ref<Record<string, string>>({})

// The invitation a successful create is meant to replace, revoked only after
// the replacement is in hand.
const replacing = ref<TeamInvite | null>(null)

const removing = ref<TeamMember | null>(null)
const revoking = ref<TeamInvite | null>(null)
const leaving = ref(false)
const deleting = ref(false)
const renaming = ref(false)
const newName = ref('')
const membersHeading = ref<HTMLElement | null>(null)

/**
 * Puts focus somewhere real after a row is destroyed.
 *
 * ModalOverlay restores focus to whatever opened it, which after a removal is
 * a button that no longer exists — the browser then falls back to `body` and a
 * screen-reader user is returned to the top of the document.
 */
async function restoreFocus() {
  await nextTick()
  membersHeading.value?.focus()
}

const team = computed(() => teams.value.find((t) => t.id === teamId.value) ?? null)
const canManage = computed(() => team.value?.role === 'owner')
// A personal team admits nobody: it refuses every removal, so an invitation
// into it would be a door that locks behind whoever walks through.
const canInvite = computed(() => canManage.value && !team.value?.personal)
const emptyFirstReason = computed(() =>
  `Move its ${team.value?.research_count} ${team.value?.research_count === 1 ? 'research' : 'researches'} out first.`,
)
const lastOwnerLeaveReason = 'You are the only owner. Make someone else an owner before leaving.'
const leaveBlocked = computed(
  () => team.value?.role === 'owner' && members.value.filter((m) => m.role === 'owner').length <= 1,
)

function article(role: string) {
  return role === 'owner' || role === 'editor' ? 'an' : 'a'
}

async function loadAll(showSkeleton = true) {
  if (showSkeleton) loading.value = true
  await loadTeams()
  if (!team.value) {
    loading.value = false
    return
  }
  try {
    const res = await authFetch<{ data: TeamMember[] }>(`${base}/api/teams/${teamId.value}/members`)
    members.value = res.data ?? []
  } catch {
    members.value = []
  }
  if (canManage.value) await loadInvites()
  loading.value = false
}

async function loadInvites() {
  try {
    const res = await authFetch<{ data: TeamInvite[] }>(`${base}/api/teams/${teamId.value}/invites`)
    invites.value = res.data ?? []
  } catch {
    invites.value = []
  }
}

onMounted(loadAll)

// A membership change from anywhere — another tab, another person — repaints
// this screen, because a stale member list is the one thing this page must not
// show.
// The hub echoes the reader's own changes back to them; repainting into
// skeletons would blink the page right after they used a control on it.
useRealtimeUpdates((event) => {
  if (event.entity === 'team' && event.entity_id === teamId.value) loadAll(false)
})

async function changeRole(userId: string, role: TeamRole) {
  const previous = members.value.find((m) => m.user_id === userId)?.role
  busyUserId.value = userId
  try {
    await authFetch(`${base}/api/teams/${teamId.value}/members/${userId}`, {
      method: 'PUT',
      body: { role },
    })
    const member = members.value.find((m) => m.user_id === userId)
    if (member) member.role = role
    success('Role updated')
  } catch (e: any) {
    // Never leave a permission showing a value the server did not accept.
    const member = members.value.find((m) => m.user_id === userId)
    if (member && previous) member.role = previous
    errorToast(e?.data?.error || 'The server refused the change', 'Role not changed', {
      label: 'Try again',
      onClick: () => { void changeRole(userId, role) },
    })
  } finally {
    busyUserId.value = null
  }
}

function askRemove(member: TeamMember) {
  removing.value = member
}

async function confirmRemove() {
  const member = removing.value
  if (!member) return
  busyUserId.value = member.user_id
  try {
    await authFetch(`${base}/api/teams/${teamId.value}/members/${member.user_id}`, { method: 'DELETE' })
    members.value = members.value.filter((m) => m.user_id !== member.user_id)
    removing.value = null
    success(`${member.name || member.email} removed`)
    await restoreFocus()
    await refreshTeams()
  } catch (e: any) {
    errorToast(e?.data?.error || 'The server refused the removal', 'Not removed', {
      label: 'Try again',
      onClick: () => { void confirmRemove() },
    })
  } finally {
    busyUserId.value = null
  }
}

function openInvite(email = '', role: TeamRole = 'viewer') {
  prefillEmail.value = email
  prefillRole.value = role
  issuedLink.value = ''
  inviteError.value = ''
  inviting.value = true
}

function closeInvite() {
  inviting.value = false
  issuedLink.value = ''
  replacing.value = null
}

async function createInvite(payload: { email: string; role: TeamRole }) {
  creatingInvite.value = true
  inviteError.value = ''
  try {
    const res = await authFetch<{ data: { invite: TeamInvite; url: string } }>(
      `${base}/api/teams/${teamId.value}/invites`,
      { method: 'POST', body: payload },
    )
    issuedLink.value = res.data.url
    recoverableLinks.value[res.data.invite.id] = res.data.url
    if (replacing.value) {
      const old = replacing.value
      replacing.value = null
      try {
        await authFetch(`${base}/api/invites/${old.id}`, { method: 'DELETE' })
      } catch {
        // A link that had already lapsed is fine to fail here.
      }
    }
    await loadInvites()
  } catch (e: any) {
    // Inline, not a toast: the email is still in the field and the reader is
    // looking at the dialog.
    inviteError.value = e?.data?.error || 'Could not create the invitation'
  } finally {
    creatingInvite.value = false
  }
}

function showLink(invite: TeamInvite) {
  issuedLink.value = recoverableLinks.value[invite.id] ?? ''
  prefillEmail.value = invite.email
  inviting.value = true
}

/**
 * Replaces an invitation with a fresh one.
 *
 * The old link is revoked only once the new one exists. Revoking first — which
 * is what "revoke and create as one action" literally means — killed a working
 * link the moment the reader pressed the button, and cancelling the dialog
 * then left the colleague with nothing.
 */
function reinvite(invite: TeamInvite) {
  replacing.value = invite
  openInvite(invite.email, invite.role)
}

function askRevoke(invite: TeamInvite) {
  revoking.value = invite
}

async function confirmRevoke() {
  const invite = revoking.value
  if (!invite) return
  busyInviteId.value = invite.id
  try {
    await authFetch(`${base}/api/invites/${invite.id}`, { method: 'DELETE' })
    revoking.value = null
    delete recoverableLinks.value[invite.id]
    await loadInvites()
    success('Invitation revoked')
    await restoreFocus()
  } catch (e: any) {
    errorToast(e?.data?.error || 'The server refused', 'Not revoked', {
      label: 'Try again',
      onClick: () => { void confirmRevoke() },
    })
  } finally {
    busyInviteId.value = null
  }
}

async function confirmLeave() {
  busy.value = true
  try {
    await authFetch(`${base}/api/teams/${teamId.value}/members/${user.value?.id}`, { method: 'DELETE' })
    await refreshTeams()
    success('You left the team')
    navigateTo('/teams')
  } catch (e: any) {
    errorToast(e?.data?.error || 'The server refused', 'Could not leave')
  } finally {
    busy.value = false
    leaving.value = false
  }
}

async function confirmDelete() {
  busy.value = true
  try {
    await removeTeam(teamId.value)
    success('Team deleted')
    navigateTo('/teams')
  } catch (e: any) {
    errorToast(e?.data?.error || 'The server refused', 'Could not delete the team')
  } finally {
    busy.value = false
    deleting.value = false
  }
}

async function confirmRename() {
  if (!newName.value.trim()) return
  const name = newName.value.trim()
  renaming.value = false
  try {
    await rename(teamId.value, name)
    success('Team renamed')
  } catch (e: any) {
    await refreshTeams()
    errorToast(e?.data?.error || 'The server refused', 'Not renamed')
  }
}

watch(renaming, (open) => {
  if (open) newName.value = team.value?.name ?? ''
})
</script>

<style scoped>
.page-lead { font-size: var(--type-sm); color: var(--color-text-muted); margin-top: var(--space-2); }
.section-heading { font-size: var(--type-lg); font-weight: 600; margin: var(--space-8) 0 var(--space-3); }
.team-work { margin-top: var(--space-6); font-size: var(--type-sm); }
.text-link { color: var(--color-primary); }
.alone-note {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  padding: var(--space-4) var(--space-1);
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}
.invite-inline { flex-shrink: 0; }
.title-skeleton { height: 40px; }
.member-skeleton { height: 52px; }

.danger-zone {
  margin-top: var(--space-12);
  border-top: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
}
/* Text rows rather than buttons: these should be reachable, not tempting. */
.danger-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-1);
  background: none;
  border: none;
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text);
  font: inherit;
  font-size: var(--type-sm);
  text-align: left;
  cursor: pointer;
}
.danger-row:hover:not(:disabled) { background: var(--color-surface-hover); }
.danger-row:disabled { opacity: 0.45; cursor: not-allowed; }
.danger-strong { color: var(--color-error); }
.danger-note { font-size: var(--type-xs); color: var(--color-text-muted); text-align: right; }

.modal-title {
  font-size: var(--type-sm);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
  margin: 0 0 var(--space-4);
}
.text-input {
  width: 100%;
  padding: 0.5rem 0.7rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font: inherit;
  font-size: var(--type-sm);
}
.modal-actions { display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-5); }

@media (max-width: 768px) {
  .danger-row { flex-direction: column; gap: var(--space-1); }
  .danger-note { text-align: left; }
}
</style>
