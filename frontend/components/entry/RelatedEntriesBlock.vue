<template>
  <div v-if="entries.length" class="crossrefs-block card no-print">
    <h3 class="crossrefs-title">
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg>
      Related by tags
    </h3>
    <div class="crossrefs-list">
      <component
        :is="relatedPath(rel) ? NuxtLink : 'div'"
        v-for="rel in entries"
        :key="rel.id"
        :to="relatedPath(rel) || undefined"
        :class="['crossref-item', { 'crossref-item--inert': !relatedPath(rel) }]"
      >
        <span class="crossref-code">{{ rel.code }}</span>
        <span class="crossref-entry-title">{{ rel.title }}</span>
        <div class="related-tags">
          <span
            v-for="tag in sharedTags(rel)"
            :key="tag"
            :class="['tag-dot', `tag-hue-${tagHue(tag)}`]"
            :title="tag"
          ></span>
        </div>
      </component>
    </div>
  </div>
</template>

<script setup lang="ts">
import { NuxtLink } from '#components'
import { tagHue } from '~/composables/useTagHue'

const props = defineProps<{
  entries: any[]
  currentTags: string[]
  researchSlug: string
  researchId: string
}>()

/**
 * Where a related entry points, or '' when it must not be a link.
 *
 * The related-by-tags query is scoped to the shared research server-side, so
 * under a share the foreign branch should be unreachable. It renders inert
 * rather than linking anyway: an endpoint that starts returning more than it
 * should must not turn into navigation here.
 */
function relatedPath(rel: any): string {
  const own = rel.research_id === props.researchId
  if (shareActive() && !own) return ''
  return entryPath(own ? props.researchSlug : rel.research_id, rel.code || rel.id)
}

function sharedTags(rel: any): string[] {
  if (!props.currentTags?.length || !rel.tags) return []
  const mine = new Set(props.currentTags)
  return rel.tags.filter((t: string) => mine.has(t))
}
</script>

<style scoped>
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
.related-tags {
  display: flex;
  gap: 3px;
  align-items: center;
  margin-left: auto;
  flex-shrink: 0;
}
.tag-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-text-muted);
}
.tag-dot.tag-hue-0 { background: hsl(200, 60%, 55%); }
.tag-dot.tag-hue-1 { background: hsl(160, 50%, 50%); }
.tag-dot.tag-hue-2 { background: hsl(280, 45%, 55%); }
.tag-dot.tag-hue-3 { background: hsl(30, 60%, 55%); }
.tag-dot.tag-hue-4 { background: hsl(340, 50%, 55%); }
.tag-dot.tag-hue-5 { background: hsl(60, 50%, 50%); }
</style>
