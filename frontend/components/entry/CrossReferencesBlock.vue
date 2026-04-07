<template>
  <div v-if="outgoing.length || incoming.length" class="crossrefs-block card no-print">
    <h3 class="crossrefs-title">
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
      Cross-references
    </h3>
    <div v-if="outgoing.length" class="crossrefs-section">
      <h4 class="crossrefs-subtitle">References from this entry</h4>
      <div class="crossrefs-list">
        <NuxtLink
          v-for="ref in outgoing"
          :key="ref.target_entry_id || ref.target_ref"
          :to="refLink(ref, 'outgoing')"
          class="crossref-item"
        >
          <span class="crossref-code">{{ ref.entry_code || ref.target_ref }}</span>
          <span v-if="ref.entry_title" class="crossref-entry-title">{{ ref.entry_title }}</span>
          <span v-if="ref.research_code && ref.research_code !== researchSlug" class="crossref-research">{{ ref.research_name || ref.research_code }}</span>
          <span v-if="!ref.resolved" class="crossref-unresolved">unresolved</span>
        </NuxtLink>
      </div>
    </div>
    <div v-if="incoming.length" class="crossrefs-section">
      <h4 class="crossrefs-subtitle">Referenced by</h4>
      <div class="crossrefs-list">
        <NuxtLink
          v-for="ref in incoming"
          :key="ref.source_id"
          :to="refLink(ref, 'incoming')"
          class="crossref-item"
        >
          <span class="crossref-code">{{ ref.entry_code || ref.source_type }}</span>
          <span v-if="ref.entry_title" class="crossref-entry-title">{{ ref.entry_title }}</span>
          <span v-if="ref.research_code && ref.research_code !== researchSlug" class="crossref-research">{{ ref.research_name || ref.research_code }}</span>
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  outgoing: any[]
  incoming: any[]
  researchSlug: string
}>()

function refLink(ref: any, direction: 'outgoing' | 'incoming'): string {
  if (direction === 'outgoing') {
    const rCode = ref.research_code || props.researchSlug
    const targetRef = ref.target_ref || ''
    // Roadmap references: [[RM1]] or [[RM1:N3]]
    if (targetRef.startsWith('RM') || ref.target_roadmap_id) {
      const rmCode = targetRef.split(':')[0]
      return `/research/${rCode}/roadmap/${rmCode}`
    }
    const eCode = ref.entry_code || targetRef
    return `/research/${rCode}/entry/${eCode}`
  }
  if (ref.source_type === 'entry') {
    const rCode = ref.research_code || props.researchSlug
    return `/research/${rCode}/entry/${ref.entry_code || ref.source_id}`
  }
  return '#'
}
</script>

<style scoped>
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
  font-weight: 600;
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
  font-weight: 500;
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
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
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
