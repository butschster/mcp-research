<template>
  <div>
    <div class="section-header">
      <h2 class="section-title">External links</h2>
    </div>

    <div v-if="loading">
      <div v-for="i in 3" :key="i" class="skeleton-card skeleton-entry"></div>
    </div>

    <div v-else-if="groups.length">
      <div v-for="group in groups" :key="group.domain" class="links-domain-group">
        <h3 class="links-domain-title">{{ group.domain }}</h3>
        <div class="links-list">
          <a
            v-for="link in group.links"
            :key="link.url"
            :href="link.url"
            target="_blank"
            rel="noopener"
            class="card link-card"
          >
            <div class="link-card-header">
              <span class="link-title">{{ link.title || link.url }}</span>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="link-external-icon"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
            </div>
            <div v-if="link.entry_code" class="link-source">
              <span class="short-code">{{ link.entry_code }}</span>
              <span class="card-meta">{{ link.entry_title }}</span>
            </div>
          </a>
        </div>
      </div>
    </div>

    <EmptyState v-else icon="&#x1F517;" title="No external links" description="Links from document content will appear here." />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  groups: any[]
  loading: boolean
}>()
</script>

<style scoped>
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--space-4); }
.section-title { font-size: var(--type-xl); font-weight: var(--weight-semibold); letter-spacing: -0.02em; }
.links-domain-group { margin-bottom: var(--space-5); }
.links-domain-title {
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  color: var(--color-text-muted);
  margin-bottom: var(--space-2);
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--color-border);
}
.links-list { display: flex; flex-direction: column; gap: var(--space-2); }
.link-card {
  display: block;
  text-decoration: none;
  color: inherit;
  transition: border-color var(--transition-fast);
}
.link-card:hover { border-color: var(--color-border-strong); }
.link-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-2);
}
.link-title {
  font-size: var(--type-sm);
  font-weight: var(--weight-medium);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.link-external-icon { opacity: 0.3; flex-shrink: 0; }
.link-card:hover .link-external-icon { opacity: 0.7; }
.link-source {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-2);
  font-size: var(--type-xs);
}
.short-code {
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
.skeleton-entry { height: 90px; margin-bottom: var(--space-3); }
</style>
