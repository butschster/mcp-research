<template>
  <NuxtLink :to="entryPath(researchSlug, entry.code || entry.id)" class="card entry-card">
    <div class="entry-card-header">
      <div class="entry-title-row">
        <ShortCode v-if="entry.code" :code="entry.code" />
        <span v-if="entry.entry_type === 'artifact'" class="entry-artifact-badge" title="HTML artifact">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
          artifact
        </span>
        <h3 class="card-title">{{ entry.title }}</h3>
      </div>
      <div class="entry-card-flags">
        <!-- A blank cell next to filled ones is the strongest force there is on
             whether an optional field ever gets answered, so the gap is shown
             where the documents are listed, not only on the document. -->
        <span
          v-if="missingRequired"
          class="badge badge-draft"
          :title="`${missingRequired} required ${missingRequired === 1 ? 'field is' : 'fields are'} unanswered`"
        >{{ missingRequired }} missing</span>
        <StatusBadge :status="entry.status" />
      </div>
    </div>
    <p v-if="entry.description" class="card-meta mt-2" v-html="renderRefs(entry.description, researchSlug)"></p>
    <TagList v-if="entry.tags?.length" :tags="entry.tags" class="mt-3" />
  </NuxtLink>
</template>

<script setup lang="ts">
import { renderRefs } from '~/composables/useCrossRefs'
import { entryPath } from '~/composables/useResearchPaths'

defineProps<{
  entry: {
    id: string
    code?: string
    title: string
    status: string
    description?: string
    tags?: string[]
    entry_type?: string
  }
  researchSlug: string
  /** How many required fields this document does not answer. Zero hides it. */
  missingRequired?: number
}>()
</script>

<style scoped>
.entry-card { display: block; text-decoration: none; color: inherit; }
.entry-card-header { display: flex; justify-content: space-between; align-items: flex-start; gap: var(--space-2); }
.entry-card-flags { display: flex; align-items: center; gap: var(--space-2); flex-shrink: 0; }
.entry-title-row { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }
.entry-artifact-badge {
  display: inline-flex; align-items: center; gap: 0.25rem;
  padding: 0.1rem 0.4rem; border-radius: var(--radius-sm);
  background: var(--color-primary-muted); color: var(--color-primary);
  font-size: var(--type-xs); font-weight: var(--weight-semibold); letter-spacing: 0.02em;
  flex-shrink: 0;
}
</style>
