<template>
  <div>
    <Breadcrumbs :crumbs="[{ label: 'Projects', to: '/' }, { label: 'Teams', to: '/teams' }, { label: team?.name || 'Team' }]" />

    <div v-if="loading" class="skeleton-list">
      <div class="skeleton-card title-skeleton"></div>
      <div v-for="i in 3" :key="i" class="skeleton-card row-skeleton"></div>
    </div>

    <!-- 403 and 404 look the same on purpose: telling an outsider that a team
         with this id exists is itself information. -->
    <EmptyState
      v-else-if="!team"
      icon="&#x1F50D;"
      title="Team not found"
      description="It may have been deleted, or you may no longer be a member."
    >
      <NuxtLink class="btn" to="/teams">All teams</NuxtLink>
    </EmptyState>

    <template v-else>
      <PageHeader :title="team.name">
        <template #actions>
          <button v-if="canInvite" class="btn" @click="openInvite()">+ Invite</button>
          <ActionMenu v-if="canManage">
            <button v-if="!team.personal" class="action-menu-item" @click="renaming = true">Rename team</button>
            <button class="action-menu-item" @click="addingResearch = true">Move projects here</button>
          </ActionMenu>
        </template>
        <template #lead>{{ members.length }} {{ members.length === 1 ? 'member' : 'members' }} · You are {{ article(team.role) }} {{ team.role }}</template>
      </PageHeader>

      <!-- Researches lead the page. The most consequential fact about a team is
           what work is in it, and this was previously one conditional line at
           the bottom that rendered nothing at all when the count was zero —
           which is the exact moment the product has to explain itself. -->
      <div class="section-bar">
        <h2 class="section-title">Projects <span class="section-count">{{ researches.length }}</span></h2>
        <NuxtLink v-if="researches.length" class="btn" :to="`/?team=${team.id}`">Open as a list</NuxtLink>
      </div>

      <div v-if="researches.length" class="card card--list">
        <TeamResearchList :researches="researches" />
      </div>

      <!-- The owner's version names the thing nobody tells them: their existing
           researches did not come along. -->
      <EmptyState
        v-else-if="canManage"
        class="section-empty"
        icon="&#x1F4C1;"
        title="Nothing to see in here yet"
        :description="`Members of ${team.name} can only read projects that live in this team. Your other projects are still in your personal team — move one across and everyone here gets it.`"
      >
        <button class="btn btn-primary" @click="addingResearch = true">Move projects here</button>
      </EmptyState>

      <!-- A viewer cannot move anything, so the way out is a person, not a
           button. Naming them beats "ask an owner". -->
      <EmptyState
        v-else
        class="section-empty"
        icon="&#x1F4C1;"
        :title="`${team.name} has no projects yet`"
        :description="`Projects added to this team will appear here for everyone in it.${ownerName ? ` ${ownerName} can move one across.` : ''}`"
      >
        <button v-if="ownerEmail" class="btn" @click="copyOwnerEmail">
          {{ copiedOwner ? '✓ Copied' : `Copy ${ownerEmail}` }}
        </button>
        <NuxtLink class="btn" to="/">Your projects</NuxtLink>
      </EmptyState>

      <div class="section-bar">
        <h2 ref="membersHeading" tabindex="-1" class="section-title">
          Members <span class="section-count">{{ members.length }}</span>
        </h2>
        <input
          v-if="members.length > 12"
          v-model="memberFilter"
          class="text-input member-filter"
          type="search"
          placeholder="Filter members"
          aria-label="Filter members"
        />
      </div>

      <!-- An empty member list is impossible — the reader is in the team. So an
           empty one means the request failed, and saying "you are alone" would
           be a lie about who has access. -->
      <EmptyState
        v-if="membersFailed"
        class="section-empty"
        icon="&#x26A0;"
        title="Couldn't load the members"
        description="Nobody has lost access — the list just didn't arrive."
      >
        <button class="btn" @click="loadMembers()">Try again</button>
      </EmptyState>

      <template v-else>
        <div ref="memberListEl" class="card card--list">
          <TeamMemberList
            :members="filteredMembers"
            :my-user-id="user?.id || ''"
            :can-manage="canManage"
            :busy-user-id="busyUserId"
            @change-role="changeRole"
            @remove="askRemove"
          />
        </div>
        <p v-if="memberFilter" class="list-note" aria-live="polite">
          {{ filteredMembers.length }} of {{ members.length }} shown
        </p>
        <p v-else-if="members.length === 1" class="list-note">
          Only you so far. Invite someone with a link — no email is sent, you pass the link along yourself.
          <button v-if="canInvite" class="link-btn" @click="openInvite()">Invite someone</button>
        </p>
      </template>

      <template v-if="canManage">
        <div class="section-bar">
          <h2 class="section-title">
            Pending invites <span class="section-count">{{ invites.length }}</span>
          </h2>
        </div>
        <div v-if="invites.length" class="card card--list">
          <TeamInviteList
            :invites="invites"
            :busy-id="busyInviteId"
            :recoverable-links="recoverableLinks"
            @revoke="askRevoke"
            @show-link="showLink"
            @reinvite="reinvite"
          />
        </div>
        <!-- One muted line rather than nothing: an owner who revokes the last
             invitation used to watch the heading evaporate under the cursor. -->
        <p v-else class="list-note">Nobody is waiting on an invitation.</p>
      </template>

      <DangerZone v-if="!team.personal">
        <DangerRow
          label="Leave team"
          :note="team.research_count > 0
            ? `Removes your access to ${team.research_count} ${team.research_count === 1 ? 'project' : 'projects'}.`
            : undefined"
          action-label="Leave"
          :disabled="leaveBlocked"
          :disabled-reason="leaveBlocked ? lastOwnerLeaveReason : undefined"
          @action="leaving = true"
        >
          <template #escape>
            <button class="link-btn" @click="focusFirstRole">Choose a new owner</button>
          </template>
        </DangerRow>

        <DangerRow
          v-if="canManage"
          label="Delete team"
          :note="team.research_count > 0 ? undefined : 'The team is empty and can be deleted.'"
          action-label="Delete"
          :disabled="team.research_count > 0"
          :disabled-reason="team.research_count > 0 ? emptyFirstReason : undefined"
          @action="deleting = true"
        >
          <template #escape>
            <NuxtLink class="link-btn" :to="`/?team=${team.id}`">Show them</NuxtLink>
          </template>
        </DangerRow>
      </DangerZone>
    </template>

    <TeamAddResearchDialog
      :visible="addingResearch"
      :team-name="team?.name || ''"
      :candidates="candidates"
      :busy="movingResearch"
      :error="moveError"
      @move="moveResearches"
      @close="addingResearch = false"
    />

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
      :message="`${removing?.name || removing?.email} loses access to every project in ${team?.name}.`"
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
      :message="`You lose access to every project in ${team?.name}. An owner can invite you back.`"
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

interface TeamResearch {
  id: string
  code?: string
  name: string
  goal?: string
  status: string
  updated_at?: string
  team_id?: string
  team_name?: string
  role?: TeamRole
}

const route = useRoute()
const teamId = computed(() => String(route.params.id))
const config = useRuntimeConfig()
const base = config.public.apiBase || ''

const { user, authFetch } = useAuth()
const { teams, load: loadTeams, refresh: refreshTeams, rename, remove: removeTeam } = useTeams()
const { success, error: errorToast } = useToasts()

const members = ref<TeamMember[]>([])
const membersFailed = ref(false)
const memberFilter = ref('')
const invites = ref<TeamInvite[]>([])
const researches = ref<TeamResearch[]>([])
const candidates = ref<TeamResearch[]>([])
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

const addingResearch = ref(false)
const movingResearch = ref(false)
const moveError = ref('')
const copiedOwner = ref(false)

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
const memberListEl = ref<HTMLElement | null>(null)

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

/**
 * The way out of "you are the only owner".
 *
 * A refusal that states only the rule leaves the reader to work out that the
 * fix is somewhere else on the page, in a control they have not noticed.
 */
async function focusFirstRole() {
  membersHeading.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  await nextTick()
  const select = memberListEl.value?.querySelector<HTMLSelectElement>('select:not(:disabled)')
  select?.focus()
}

const team = computed(() => teams.value.find((t) => t.id === teamId.value) ?? null)
const canManage = computed(() => team.value?.role === 'owner')
// A personal team admits nobody: it refuses every removal, so an invitation
// into it would be a door that locks behind whoever walks through.
const canInvite = computed(() => canManage.value && !team.value?.personal)
const emptyFirstReason = computed(
  () => `Move its ${team.value?.research_count} ${team.value?.research_count === 1 ? 'project' : 'projects'} out first.`,
)
const lastOwnerLeaveReason = 'You are the only owner. Make someone else an owner before leaving.'
const leaveBlocked = computed(
  () => team.value?.role === 'owner' && members.value.filter((m) => m.role === 'owner').length <= 1,
)

// Whom a viewer should ask. The members endpoint needs only viewer rights, so
// this is answerable for exactly the people who need the answer.
const firstOwner = computed(() => members.value.find((m) => m.role === 'owner') ?? null)
const ownerName = computed(() => firstOwner.value?.name || firstOwner.value?.email || '')
const ownerEmail = computed(() => firstOwner.value?.email || '')

const filteredMembers = computed(() => {
  const needle = memberFilter.value.trim().toLowerCase()
  if (!needle) return members.value
  return members.value.filter(
    (m) => m.name?.toLowerCase().includes(needle) || m.email?.toLowerCase().includes(needle),
  )
})

function article(role: string) {
  return role === 'owner' || role === 'editor' ? 'an' : 'a'
}

async function copyOwnerEmail() {
  if (!ownerEmail.value) return
  try {
    await navigator.clipboard.writeText(ownerEmail.value)
    copiedOwner.value = true
    setTimeout(() => { copiedOwner.value = false }, 2000)
  } catch {
    errorToast('Copy the address from the members list below', 'Clipboard refused')
  }
}

async function loadAll(showSkeleton = true) {
  if (showSkeleton) loading.value = true
  await loadTeams()
  if (!team.value) {
    loading.value = false
    return
  }
  await Promise.all([loadMembers(), loadResearches(), canManage.value ? loadInvites() : Promise.resolve()])
  loading.value = false
}

async function loadMembers() {
  try {
    const res = await authFetch<{ data: TeamMember[] }>(`${base}/api/teams/${teamId.value}/members`)
    members.value = res.data ?? []
    membersFailed.value = false
  } catch {
    members.value = []
    membersFailed.value = true
  }
}

// No `status` filter on purpose: a team holding eight archived researches is
// not an empty team, and the list defaults to active everywhere else.
async function loadResearches() {
  try {
    const res = await authFetch<{ data: TeamResearch[] }>(`${base}/api/researches?team=${teamId.value}`)
    researches.value = res.data ?? []
  } catch {
    researches.value = []
  }
}

// Everything the reader owns that is not already here. Owner rather than
// editor, because moving a research out of a team is an owner's call in the
// team it is leaving.
async function loadCandidates() {
  try {
    const res = await authFetch<{ data: TeamResearch[] }>(`${base}/api/researches`)
    candidates.value = (res.data ?? []).filter((r) => r.team_id !== teamId.value && r.role === 'owner')
  } catch {
    candidates.value = []
  }
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

// The candidate list is a second full read of the research list, so it waits
// until somebody actually opens the dialog.
watch(addingResearch, (open) => {
  if (open) {
    moveError.value = ''
    void loadCandidates()
  }
})

// A membership change from anywhere — another tab, another person — repaints
// this screen, because a stale member list is the one thing this page must not
// show. A transfer moves work in or out of the team, which this page now shows.
// The hub echoes the reader's own changes back to them; repainting into
// skeletons would blink the page right after they used a control on it.
useRealtimeUpdates(
  (event) => {
    if (event.entity === 'team' && event.entity_id === teamId.value) loadAll(false)
    else if (event.entity === 'research') void loadResearches()
  },
  { onResync: () => loadAll(false) },
)

async function moveResearches(ids: string[]) {
  movingResearch.value = true
  moveError.value = ''
  try {
    const res = await authFetch<{ data: { moved: number } }>(`${base}/api/teams/${teamId.value}/researches`, {
      method: 'POST',
      body: { research_ids: ids },
    })
    addingResearch.value = false
    const moved = res.data?.moved ?? ids.length
    success(
      `${moved} ${moved === 1 ? 'project is' : 'projects are'} now visible to everyone in ${team.value?.name}.`,
      'Moved',
    )
    await Promise.all([loadResearches(), refreshTeams()])
  } catch (e: any) {
    moveError.value = e?.data?.error || 'The server refused the move'
  } finally {
    movingResearch.value = false
  }
}

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
/* The heading over a rule list, with its count and whatever that section lets
   you do. Not promoted to system.css: two other components declare
   `.section-heading` and both are a document heading with a rule under it —
   the same name for a different design. Sharing the name and overriding the
   rule is precisely the `.danger-zone` bug this page just lost. */
.section-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  margin: var(--space-8) 0 var(--space-3);
}
.section-title {
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
  font-size: var(--type-lg);
  font-weight: var(--weight-semibold);
  margin: 0;
}
.section-count {
  font-size: var(--type-sm);
  font-weight: var(--weight-normal);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}
.member-filter { max-width: 16rem; }
/* An empty state between two lists, rather than instead of a page. The full
   padding put the Members heading below the fold on a team that has no work in
   it yet — which is the one team where the reader needs to see both. */
.section-empty { padding: var(--space-8) var(--space-4); }
.list-note {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  padding: var(--space-3) var(--space-1);
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}
.title-skeleton { height: 40px; }
/* The same height as the rows it stands in for, so the page does not
   reassemble itself when the data lands. */
.row-skeleton { height: 56px; }
</style>
