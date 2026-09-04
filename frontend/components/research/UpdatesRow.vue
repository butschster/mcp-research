<template>
  <article class="data-row update-row">
    <div class="update-row__inner">
      <div class="update-row__body">
        <div class="update-row__title-line">
          <EntryUpdateBadge :kind="update.kind" :unseen-revisions="update.unseen_revisions" />
          <ShortCode v-if="update.entry_code" :code="update.entry_code" />
          <NuxtLink :to="target" class="update-row__title">{{ update.title }}</NuxtLink>
        </div>
        <div class="update-row__meta">
          <span>{{ sectionName }}</span>
          <span aria-hidden="true">·</span>
          <span>{{ revisionRange }}</span>
          <span aria-hidden="true">·</span>
          <time :datetime="update.updated_at" :title="absoluteTime(update.updated_at)">
            {{ relativeTime(update.updated_at) }}
          </time>
        </div>
      </div>
      <div class="update-row__actions">
        <StatusBadge :status="update.status" />
        <NuxtLink v-if="update.kind === 'changed'" :to="reviewTarget" class="btn btn-sm">
          Review changes
        </NuxtLink>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import type { EntryUpdate } from '~/composables/useEntryUpdates'
import { entryPath } from '~/composables/useResearchPaths'
import { absoluteTime, relativeTime } from '~/composables/useRelativeTime'

const props = defineProps<{
  update: EntryUpdate
  researchSlug: string
  sectionName: string
}>()

const target = computed(() => entryPath(props.researchSlug, props.update.entry_code || props.update.entry_id))
const reviewTarget = computed(() => ({
  path: target.value,
  query: {
    changes_from: String(props.update.seen_revision),
    changes_to: String(props.update.current_revision),
    review_changes: '1',
  },
}))
const revisionRange = computed(() => props.update.kind === 'new'
  ? `r${props.update.current_revision}`
  : `r${props.update.seen_revision} → r${props.update.current_revision}`)
</script>

<style scoped>
.update-row { display: block; }
.update-row__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  width: 100%;
}
.update-row__body { min-width: 0; }
.update-row__title-line,
.update-row__meta,
.update-row__actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}
.update-row__title { color: var(--color-text); font-weight: var(--weight-semibold); text-decoration: none; overflow-wrap: anywhere; }
.update-row__title:hover { color: var(--color-primary); }
.update-row__meta { margin-top: var(--space-2); color: var(--color-text-muted); font-size: var(--type-xs); }
.update-row__actions { justify-content: flex-end; flex: none; }

@media (max-width: 768px) {
  .update-row__inner { align-items: flex-start; flex-direction: column; gap: var(--space-3); }
  .update-row__actions { justify-content: flex-start; width: 100%; }
}
</style>
