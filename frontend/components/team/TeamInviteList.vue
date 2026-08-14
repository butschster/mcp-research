<template>
  <div class="invite-rows">
    <div
      v-for="invite in invites"
      :key="invite.id"
      class="invite-row"
      :class="{ expired: isExpired(invite), busy: busyId === invite.id }"
    >
      <span class="invite-email">{{ invite.email || 'Anyone with the link' }}</span>
      <span class="invite-role">{{ ROLE_LABELS[invite.role] }}</span>
      <span class="invite-status">{{ status(invite) }}</span>

      <div class="invite-actions">
        <template v-if="isExpired(invite)">
          <button class="link-btn" @click="emit('reinvite', invite)">Re-invite</button>
        </template>
        <template v-else>
          <button
            v-if="recoverableLinks[invite.id]"
            class="link-btn"
            @click="emit('showLink', invite)"
          >Show link</button>
          <button
            v-else
            class="link-btn"
            title="The link is shown only once, so a replacement means revoking this one."
            @click="emit('reinvite', invite)"
          >New link</button>
          <button class="link-btn link-danger" :disabled="busyId === invite.id" @click="emit('revoke', invite)">
            Revoke
          </button>
        </template>
      </div>
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
 */
defineProps<{
  invites: TeamInvite[]
  busyId?: string | null
  /**
   * Tokens still held in memory from this tab's session. The server hashes
   * them, so after a reload the only route to a working link is a new
   * invitation — which is what the button then offers.
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
.invite-rows { display: flex; flex-direction: column; border-top: 1px solid var(--color-border); }
.invite-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto auto;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-1);
  border-bottom: 1px solid var(--color-border);
}
.invite-row.expired { color: var(--color-text-muted); }
.invite-row.busy { opacity: 0.6; }
.invite-email { overflow-wrap: anywhere; }
.invite-role, .invite-status { font-size: var(--type-xs); color: var(--color-text-muted); white-space: nowrap; }
.invite-actions { display: flex; gap: var(--space-3); }
.link-danger { color: var(--color-error); }

@media (max-width: 768px) {
  .invite-row { grid-template-columns: minmax(0, 1fr) auto; row-gap: var(--space-2); }
  .invite-email { grid-column: 1 / -1; }
  .invite-role { grid-column: 1; }
  .invite-status { grid-column: 2; justify-self: end; }
  .invite-actions { grid-column: 1 / -1; }
}
</style>
