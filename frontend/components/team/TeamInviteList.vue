<template>
  <div class="data-rows data-rows--bounded invite-rows">
    <div
      v-for="invite in invites"
      :key="invite.id"
      class="data-row invite-row"
      :class="{ 'data-row--dead': isExpired(invite), 'data-row--busy': busyId === invite.id }"
      :aria-busy="busyId === invite.id || undefined"
      :inert="busyId === invite.id || undefined"
    >
      <span class="user-avatar invite-avatar" aria-hidden="true">✉</span>

      <div class="invite-identity">
        <span class="invite-email">{{ invite.email || 'Anyone with the link' }}</span>
        <span class="invite-sub">
          <span>{{ ROLE_LABELS[invite.role] }}</span>
          <!-- Already on the wire and never shown. On a team with three owners
               it is the only way to tell who is expecting this person. -->
          <span v-if="invite.inviter_name || invite.inviter_email">
            invited by {{ invite.inviter_name || invite.inviter_email }}
          </span>
        </span>
      </div>

      <!-- Expiry is a badge rather than grey text, so "expired" is not a
           difference in colour alone. -->
      <span class="badge" :class="isExpired(invite) ? 'badge-dead' : 'badge-quiet'">{{ status(invite) }}</span>

      <ActionMenu :title="`Actions for the invitation to ${invite.email || 'anyone with the link'}`">
        <button
          v-if="!isExpired(invite) && recoverableLinks[invite.id]"
          class="action-menu-item"
          @click="emit('showLink', invite)"
        >
          Show link
        </button>
        <button class="action-menu-item" @click="emit('reinvite', invite)">
          {{ isExpired(invite) ? 'Re-invite' : 'Issue a new link' }}
        </button>
        <span class="action-menu-divider"></span>
        <button
          class="action-menu-item action-menu-item--danger"
          :disabled="busyId === invite.id"
          @click="emit('revoke', invite)"
        >
          Revoke
        </button>
      </ActionMenu>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ROLE_LABELS, type TeamInvite } from '~/composables/useTeams'
import { parseTimestamp } from '~/composables/useRelativeTime'

/**
 * Invitations nobody has accepted yet.
 *
 * Lapsed ones stay in the list, dimmed. Filtering them away is how an owner
 * ends up asking why a colleague never joined while the colleague is looking at
 * a page that says "expired".
 *
 * The row is shaped like a member row on purpose — an avatar, an identity, a
 * role, one `⋯`. An invitation is a person and their access, same as the row
 * above it; it was previously an email, grey text and two bare text links with
 * 21.7px targets, distinguishable from each other by colour alone.
 */
defineProps<{
  invites: TeamInvite[]
  busyId?: string | null
  /**
   * Tokens still held in memory from this tab's session. The server hashes
   * them, so after a reload the only route to a working link is a new
   * invitation — which is what the menu then offers instead.
   */
  recoverableLinks: Record<string, string>
}>()

const emit = defineEmits<{
  revoke: [invite: TeamInvite]
  showLink: [invite: TeamInvite]
  reinvite: [invite: TeamInvite]
}>()

function expiryDate(invite: TeamInvite) {
  return parseTimestamp(invite.expires_at)
}

function isExpired(invite: TeamInvite) {
  const date = expiryDate(invite)
  return !Number.isNaN(date.getTime()) && date.getTime() <= Date.now()
}

function status(invite: TeamInvite) {
  if (isExpired(invite)) return 'expired'
  const days = Math.ceil((expiryDate(invite).getTime() - Date.now()) / 86_400_000)
  if (days <= 1) return 'expires today'
  return `expires in ${days} days`
}
</script>

<style scoped>
.invite-row { grid-template-columns: auto minmax(0, 1fr) auto auto; }
.invite-avatar {
  --avatar-size: 28px;
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
  font-size: 0.8rem;
}
.invite-identity { display: flex; flex-direction: column; gap: 1px; min-width: 0; }
.invite-email { overflow-wrap: anywhere; }
.invite-sub { display: flex; flex-wrap: wrap; gap: var(--space-1); font-size: var(--type-xs); color: var(--color-text-muted); }
.invite-sub > * + *::before { content: '· '; }
.badge-quiet { background: var(--color-surface-hover); color: var(--color-text-muted); }
.badge-dead { background: var(--color-surface-hover); color: var(--color-text-faint); }

@media (max-width: 768px) {
  .invite-row { grid-template-columns: auto minmax(0, 1fr) auto; row-gap: var(--space-2); }
  .invite-identity { grid-column: 2; }
  .action-menu { grid-column: 3; grid-row: 1; }
  .badge { grid-column: 2 / -1; grid-row: 2; justify-self: start; }
}
</style>
