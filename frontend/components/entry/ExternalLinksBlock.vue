<template>
  <EntryFoldable v-if="links.length" title="External links" :count="links.length" remember-as="links">
    <template #icon><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg></template>
    <div class="crossrefs-list">
      <a
        v-for="link in links"
        :key="link.id"
        :href="link.url"
        target="_blank"
        rel="noopener"
        class="crossref-item external-link-item"
      >
        <span class="external-link-domain">{{ link.domain }}</span>
        <span class="crossref-entry-title">{{ link.title || link.url }}</span>
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="external-link-icon"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
      </a>
    </div>
  </EntryFoldable>
</template>

<script setup lang="ts">
defineProps<{
  links: any[]
}>()
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
.crossref-entry-title {
  font-size: var(--type-sm);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.external-link-domain {
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  flex-shrink: 0;
}
.external-link-icon {
  margin-left: auto;
  opacity: 0.4;
  flex-shrink: 0;
}
.external-link-item:hover .external-link-icon { opacity: 0.8; }
</style>
