<template>
  <EntryFoldable
    v-if="outgoing.length || incoming.length"
    title="Cross-references"
    :count="outgoing.length + incoming.length"
    remember-as="crossrefs"
  >
    <template #icon><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg></template>
    <div v-if="outgoing.length" class="crossrefs-section">
      <h4 class="crossrefs-subtitle">References from this entry</h4>
      <div class="crossrefs-list">
        <component
          :is="refLink(ref, 'outgoing') ? NuxtLink : 'div'"
          v-for="ref in outgoing"
          :key="ref.target_entry_id || ref.target_roadmap_id || ref.target_ref"
          :to="refLink(ref, 'outgoing') || undefined"
          :class="['crossref-item', { 'crossref-item--inert': !refLink(ref, 'outgoing') }]"
        >
          <span class="crossref-code">{{ ref.entry_code || ref.roadmap_code || ref.target_ref }}</span>
          <span v-if="ref.entry_title" class="crossref-entry-title">{{ ref.entry_title }}</span>
          <span v-else-if="ref.roadmap_title" class="crossref-entry-title">{{ ref.roadmap_title }}</span>
          <span v-if="ref.research_code && ref.research_code !== researchSlug" class="crossref-research">{{ ref.research_name || ref.research_code }}</span>
          <span v-if="!ref.resolved && !shared" class="crossref-unresolved">unresolved</span>
        </component>
      </div>
    </div>
    <div v-if="incoming.length" class="crossrefs-section">
      <h4 class="crossrefs-subtitle">Referenced by</h4>
      <div class="crossrefs-list">
        <component
          :is="refLink(ref, 'incoming') ? NuxtLink : 'div'"
          v-for="ref in incoming"
          :key="ref.source_id"
          :to="refLink(ref, 'incoming') || undefined"
          :class="['crossref-item', { 'crossref-item--inert': !refLink(ref, 'incoming') }]"
        >
          <span class="crossref-code">{{ ref.entry_code || ref.source_type }}</span>
          <span v-if="ref.entry_title" class="crossref-entry-title">{{ ref.entry_title }}</span>
          <span v-if="ref.research_code && ref.research_code !== researchSlug" class="crossref-research">{{ ref.research_name || ref.research_code }}</span>
        </component>
      </div>
    </div>
  </EntryFoldable>
</template>

<script setup lang="ts">
import { NuxtLink } from '#components'

const props = defineProps<{
  outgoing: any[]
  incoming: any[]
  researchSlug: string
}>()

const shared = computed(() => shareActive())

/**
 * Where a reference row points, or '' for a row that must not be a link.
 *
 * Under a share the second case is the important one. A reference out of the
 * shared research comes back with its target ids blanked and `resolved` false,
 * and the row still has to render — the author wrote it and the reader can see
 * the bracketed code in the text above — but as inert markup that neither
 * navigates nor hints that a destination exists.
 */
function refLink(ref: any, direction: 'outgoing' | 'incoming'): string {
  const shared = shareActive()
  if (direction === 'outgoing') {
    const rCode = ref.research_code || props.researchSlug
    const targetRef = ref.target_ref || ''
    // Roadmap references: [[RM1]] or [[RM1:N3]]
    if (targetRef.startsWith('RM') || ref.target_roadmap_id) {
      if (shared && !shareInclude().roadmaps) return ''
      return roadmapPath(rCode, targetRef.split(':')[0])
    }
    // Inert because the target is outside the link — not because the stored
    // reference happens to be stale. Inerting on `resolved` made the same
    // `[[E2]]` a working link in the entry's prose and a dead row in the panel
    // underneath it, and only the client saw the difference.
    if (shared && ref.research_code && ref.research_code !== props.researchSlug) return ''
    const eCode = ref.entry_code || targetRef
    return entryPath(rCode, eCode)
  }
  if (ref.source_type === 'entry') {
    const rCode = ref.research_code || props.researchSlug
    if (shared && ref.research_code && ref.research_code !== props.researchSlug) return ''
    return entryPath(rCode, ref.entry_code || ref.source_id)
  }
  return ''
}
</script>

<style scoped>
/* An inert row keeps the layout and loses every affordance: no pointer, no
   hover, no underline. A distinct treatment would itself say something is
   there. */
.crossref-item--inert { cursor: default; }
.crossrefs-block {
  margin-top: var(--space-6);
  padding: var(--space-6);
  border-radius: var(--radius-lg);
}
.crossrefs-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
  margin: 0 0 var(--space-4) 0;
}
.crossrefs-section + .crossrefs-section {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-border);
}
.crossrefs-subtitle {
  font-size: var(--type-xs);
  font-weight: var(--weight-medium);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 var(--space-2) 0;
}
.crossrefs-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}
.crossref-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  text-decoration: none;
  color: var(--color-text);
  transition: background 0.15s;
}
.crossref-item:hover {
  background: var(--color-surface-hover);
}
.crossref-code {
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
  flex-shrink: 0;
}
.crossref-entry-title {
  font-size: var(--type-sm);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.crossref-research {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  margin-left: auto;
  flex-shrink: 0;
}
.crossref-unresolved {
  font-size: var(--type-xs);
  color: var(--color-warning, #d4a017);
  font-style: italic;
  margin-left: auto;
}
</style>
