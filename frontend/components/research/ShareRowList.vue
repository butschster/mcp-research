<template>
  <div class="data-rows share-rows">
    <div
      v-for="share in shares"
      :key="share.id"
      class="data-row share-row"
      :class="{ 'data-row--dead': !isLive(share), 'data-row--busy': busyId === share.id }"
      :aria-busy="busyId === share.id || undefined"
      :inert="busyId === share.id || undefined"
    >
      <div class="share-cell share-cell--label">
        <span class="share-label">{{ share.label || 'Untitled link' }}</span>
        <span class="share-sub card-meta">
          {{ formatDate(share.created_at) }}
          <template v-if="share.created_by_name"> by {{ share.created_by_name }}</template>
          &middot; {{ contents(share) }}
          <span v-if="share.has_password" class="share-badge">Password</span>
        </span>
      </div>

      <div class="share-cell share-cell--state">
        <span class="share-state">{{ stateLabel(share) }}</span>
        <span class="share-sub card-meta">{{ activity(share) }}</span>
      </div>

      <div class="share-cell share-cell--actions">
        <button
          v-if="isLive(share) && recoverableLinks[share.id]"
          class="link-btn"
          :disabled="busyId === share.id"
          @click="emit('showLink', share)"
        >
          Show link
        </button>
        <button
          v-if="isLive(share)"
          class="link-btn link-btn--danger"
          :disabled="busyId === share.id"
          @click="emit('revoke', share)"
        >
          Revoke
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { shareContents, shareExpiryPhrase } from '~/composables/useShare'
/**
 * The links a research has handed out — live, expired and revoked.
 *
 * Dead rows stay, dimmed, the way lapsed invitations do. "Who did I give this
 * to, and did I take it back" is the question this list answers, and a row that
 * disappears on revocation leaves no record that it ever existed.
 */
export interface ShareRow {
  id: string
  label: string
  include: { sessions: boolean; tasks: boolean; roadmaps: boolean; export: boolean }
  has_password: boolean
  expires_at?: string | null
  revoked_at?: string | null
  last_seen_at?: string | null
  view_count: number
  created_at: string
  created_by_name?: string
}

defineProps<{
  shares: ShareRow[]
  /** The row a revoke is in flight for. */
  busyId?: string
  /** Links still in memory from this tab, by share id. */
  recoverableLinks: Record<string, string>
}>()

const emit = defineEmits<{ revoke: [share: ShareRow]; showLink: [share: ShareRow] }>()

const isLive = isShareLive

function stateLabel(s: ShareRow) {
  if (s.revoked_at) return 'Revoked'
  if (s.expires_at && parseTimestamp(s.expires_at).getTime() <= Date.now()) return 'Expired'
  if (s.expires_at) {
    const phrase = shareExpiryPhrase(s.expires_at)
    return phrase.startsWith('on ') ? `Expires ${phrase}` : `Expires ${phrase}`
  }
  return 'Active'
}

function activity(s: ShareRow) {
  if (!s.view_count) return 'Never opened'
  const views = s.view_count === 1 ? '1 view' : `${s.view_count} views`
  if (!s.last_seen_at) return views
  return `${views} · last opened ${relativeTime(s.last_seen_at)}`
}

function contents(s: ShareRow) {
  return shareContents(s.include)
}

// `new Date(value)` read the API's two timestamp shapes differently — SQLite's
// space-separated form as local, the encoder's Z-suffixed form as UTC — so a
// share's creation date could disagree by hours with the last-opened line
// directly above it, which already went through the composable.
function formatDate(value: string) {
  return absoluteTime(value, { dateStyle: 'medium' })
}
</script>

<style scoped>
.share-row {
  grid-template-columns: minmax(0, 1.6fr) minmax(0, 1fr) auto;
  align-items: start;
}
.share-cell { min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.share-cell--actions { flex-direction: row; align-items: center; gap: var(--space-3); justify-content: flex-end; }
/* A truncated label is not a label, and this list exists to be recognised by
   one. */
.share-label { font-weight: var(--weight-semibold); overflow-wrap: anywhere; }
.share-sub { font-size: var(--type-xs); overflow-wrap: anywhere; }
.share-state { font-size: var(--type-sm); }
.share-badge {
  display: inline-block;
  margin-left: var(--space-2);
  padding: 0 var(--space-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: 0.65rem;
}
/* .link-btn itself is global (system.css). Only the destructive variant is
   local, and it uses the product's own red — there is no --color-danger. */
.link-btn--danger:hover { color: var(--color-error); }
.link-btn:disabled { opacity: 0.5; cursor: default; }

@media (max-width: 768px) {
  .share-row { grid-template-columns: minmax(0, 1fr) auto; }
  .share-cell--actions { grid-column: 1 / -1; justify-content: flex-start; }
}
</style>
