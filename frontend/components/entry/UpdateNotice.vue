<template>
  <div class="entry-update-notice">
    <div class="entry-update-notice__copy" role="status">
      <EntryUpdateBadge :kind="state.kind" :unseen-revisions="state.unseen_revisions" />
      <span v-if="state.kind === 'changed'">
        Changed since you last opened r{{ state.seen_revision }}. Current revision: r{{ state.current_revision }}.
      </span>
      <span v-else>This is the first time you’ve opened this document.</span>
    </div>
    <button v-if="state.kind === 'changed'" type="button" class="btn btn-sm" @click="$emit('review')">
      Review changes
    </button>
  </div>
</template>

<script setup lang="ts">
import type { EntryViewState } from '~/composables/useEntryUpdates'

defineProps<{ state: EntryViewState & { kind: 'new' | 'changed' } }>()
defineEmits<{ review: [] }>()
</script>

<style scoped>
.entry-update-notice {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
  margin-bottom: var(--space-4);
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius);
  background: var(--color-surface);
  font-size: var(--type-sm);
}
.entry-update-notice__copy {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
  min-width: 0;
}
</style>
