<template>
  <div class="data-rows data-rows--bounded member-rows">
    <div
      v-for="member in members"
      :key="member.user_id"
      class="data-row member-row"
      :class="{ 'data-row--busy': busyUserId === member.user_id }"
      :aria-busy="busyUserId === member.user_id || undefined"
      :inert="busyUserId === member.user_id || undefined"
    >
      <span class="user-avatar" aria-hidden="true">{{ initial(member) }}</span>

      <div class="member-identity">
        <span class="member-line">
          <span class="member-name">{{ member.name || member.email }}</span>
          <span v-if="member.user_id === myUserId" class="tag member-you">you</span>
        </span>
        <span class="member-sub">
          <span class="member-email">{{ member.email }}</span>
          <span v-if="member.created_at" class="member-joined">joined {{ joined(member.created_at) }}</span>
        </span>
      </div>

      <TeamRoleSelect
        v-if="canManage"
        :model-value="member.role"
        :busy="busyUserId === member.user_id"
        :disabled="isLastOwner(member)"
        :disabled-reason="lastOwnerReason"
        :label="`Role of ${member.name || member.email}`"
        @update:model-value="emit('changeRole', member.user_id, $event)"
      />
      <span v-else class="member-role-static">{{ ROLE_LABELS[member.role] }}</span>

      <!-- The removal lives behind the same `⋯` as every other row action in
           the product. It used to be a 28px unlabelled × — the loudest control
           in the row was also the only irreversible one. -->
      <ActionMenu v-if="canManage" :title="`Actions for ${member.name || member.email}`">
        <button
          class="action-menu-item action-menu-item--danger"
          :disabled="isLastOwner(member)"
          :title="isLastOwner(member) ? lastOwnerReason : undefined"
          @click="emit('remove', member)"
        >
          Remove from team
        </button>
      </ActionMenu>
      <span v-else class="member-action-placeholder"></span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ROLE_LABELS, type TeamMember, type TeamRole } from '~/composables/useTeams'
import { absoluteTime } from '~/composables/useRelativeTime'

/**
 * The people in a team.
 *
 * Name and email were a column each, which made the row 72px against the
 * invite list's 48px under two identical headings — one page, two lists about
 * the same thing, half again the height. They are one identity block now: the
 * name a reader scans for, and underneath the two facts that qualify it. That
 * is 56px, the same rhythm as every other rule list on the page.
 *
 * Neither string truncates. A shortened email is not an email, and these are
 * the two strings a reader is here to check.
 */
const props = defineProps<{
  members: TeamMember[]
  myUserId: string
  canManage: boolean
  busyUserId?: string | null
}>()

const emit = defineEmits<{
  changeRole: [userId: string, role: TeamRole]
  remove: [member: TeamMember]
}>()

const lastOwnerReason = 'A team needs at least one owner. Make someone else an owner first.'

const ownerCount = computed(() => props.members.filter((m) => m.role === 'owner').length)

// The guard is the server's, and it is repeated here so the control is never
// offered in a state the server would refuse.
function isLastOwner(member: TeamMember) {
  return member.role === 'owner' && ownerCount.value <= 1
}

function initial(member: TeamMember) {
  return (member.name || member.email || '?').trim().charAt(0).toUpperCase()
}

function joined(iso: string) {
  return absoluteTime(iso, { day: 'numeric', month: 'short' })
}
</script>

<style scoped>
/* Bigger than the nav's, which is a 22px afterthought beside a name. */
.member-row .user-avatar { --avatar-size: 28px; }
.member-row { grid-template-columns: auto minmax(0, 1fr) auto auto; }
.member-identity { display: flex; flex-direction: column; gap: 1px; min-width: 0; }
.member-line { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }
.member-name { font-weight: var(--weight-medium); overflow-wrap: anywhere; }
.member-you { color: var(--color-text-muted); }
.member-sub { display: flex; flex-wrap: wrap; gap: var(--space-1); font-size: var(--type-xs); color: var(--color-text-muted); }
.member-email { overflow-wrap: anywhere; }
.member-joined::before { content: '· '; }
.member-role-static { font-size: var(--type-xs); color: var(--color-text-muted); white-space: nowrap; }
/* The column has to keep its width for a viewer, or every row's role would sit
   at a different distance from the edge depending on who is reading. */
.member-action-placeholder { width: var(--control-h); }

@media (max-width: 768px) {
  .member-row { grid-template-columns: auto minmax(0, 1fr) auto; row-gap: var(--space-2); }
  .member-identity { grid-column: 2; }
  /* The menu stays top-right rather than sitting under a wrapping email. */
  .action-menu, .member-action-placeholder { grid-column: 3; grid-row: 1; }
  .member-role-static, :deep(.role-select) { grid-column: 2 / -1; grid-row: 2; justify-self: start; }
}
</style>
