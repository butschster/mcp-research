<template>
  <div class="member-rows">
    <div
      v-for="member in members"
      :key="member.user_id"
      class="member-row"
      :class="{ busy: busyUserId === member.user_id }"
    >
      <span class="user-avatar" aria-hidden="true">{{ initial(member) }}</span>

      <div class="member-identity">
        <span class="member-name">{{ member.name || member.email }}</span>
        <span v-if="member.user_id === myUserId" class="member-note">you</span>
        <span v-else-if="member.created_at" class="member-note">joined {{ joined(member.created_at) }}</span>
      </div>

      <span class="member-email">{{ member.email }}</span>

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

      <button
        v-if="canManage"
        class="btn-icon remove-btn"
        :disabled="isLastOwner(member) || busyUserId === member.user_id"
        :title="isLastOwner(member) ? lastOwnerReason : `Remove ${member.name || member.email}`"
        :aria-label="`Remove ${member.name || member.email}`"
        @click="emit('remove', member)"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" /></svg>
      </button>
      <span v-else class="remove-placeholder"></span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ROLE_LABELS, type TeamMember, type TeamRole } from '~/composables/useTeams'
import { absoluteTime } from '~/composables/useRelativeTime'

/**
 * The people in a team.
 *
 * Name and email each get their own column and both wrap rather than
 * truncating: a shortened email is not an email, and these are the two strings
 * a reader is here to check.
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
.member-rows { display: flex; flex-direction: column; border-top: 1px solid var(--color-border); }
/* Bigger than the nav's, which is a 22px afterthought beside a name. */
.member-row .user-avatar { --avatar-size: 28px; }
.member-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1.2fr) minmax(0, 1.4fr) auto auto;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-1);
  border-bottom: 1px solid var(--color-border);
  transition: opacity var(--transition-fast);
}
.member-row.busy { opacity: 0.6; }
.member-identity { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.member-name { font-weight: 500; overflow-wrap: anywhere; }
.member-note { font-size: var(--type-xs); color: var(--color-text-muted); }
.member-email { font-size: var(--type-xs); color: var(--color-text-muted); overflow-wrap: anywhere; }
.member-role-static { font-size: var(--type-xs); color: var(--color-text-muted); white-space: nowrap; }
.remove-placeholder { width: 0; }
.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: none;
  color: var(--color-text-muted);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}
.btn-icon:hover:not(:disabled) { background: var(--color-surface-hover); color: var(--color-error); }
.btn-icon:disabled { opacity: 0.35; cursor: not-allowed; }

@media (max-width: 768px) {
  .member-row {
    grid-template-columns: auto minmax(0, 1fr) auto;
    row-gap: var(--space-2);
  }
  .member-identity { grid-column: 2; }
  /* The one destructive control stays top-right rather than sitting under a
     wrapping email. */
  .remove-btn, .remove-placeholder { grid-column: 3; grid-row: 1; }
  .member-email { grid-column: 2 / -1; grid-row: 2; }
  .member-role-static, :deep(.role-select) { grid-column: 2 / -1; grid-row: 3; justify-self: start; }
}
</style>
