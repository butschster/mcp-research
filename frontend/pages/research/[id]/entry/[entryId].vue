<template>
  <div v-if="pending">
    <div class="skeleton-card skeleton-header"></div>
    <div class="skeleton-card skeleton-content"></div>
  </div>

  <div v-else-if="entry">
    <!-- Header -->
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: researchName, to: `/research/${researchSlug}` },
        { label: sectionName, to: `/research/${researchSlug}?section=${entry.section_id}` },
        { label: entry.title }
      ]" />
      <div class="entry-header">
        <div class="title-with-code">
          <span v-if="entry.code" class="short-code">{{ entry.code }}</span>
          <h1 class="page-title">{{ entry.title }}</h1>
        </div>
        <div class="entry-actions no-print">
          <StatusBadge :status="entry.status" />
          <button class="btn btn-sm" @click="copyMarkdown">
            <svg v-if="!copied" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
            <svg v-else width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
            {{ copied ? 'Copied' : 'Copy' }}
          </button>
        </div>
      </div>
      <p v-if="entry.description" class="card-meta mt-2" v-html="renderRefs(entry.description, researchSlug)"></p>
      <div v-if="entry.tags?.length" class="entry-tags">
        <span v-for="tag in entry.tags" :key="tag" :class="['tag', `tag-hue-${tagHue(tag)}`]">{{ tag }}</span>
      </div>
      <NuxtLink v-if="linkedSession" :to="`/research/${researchSlug}/session/${linkedSession.code || linkedSession.id}`" class="entry-session-link">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/></svg>
        {{ linkedSession.title }}
      </NuxtLink>
    </div>

    <!-- View toggle -->
    <div class="view-toggle no-print">
      <button :class="['btn btn-sm', { active: viewMode === 'rendered' }]" @click="viewMode = 'rendered'">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
        Rendered
      </button>
      <button :class="['btn btn-sm', { active: viewMode === 'source' }]" @click="viewMode = 'source'">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
        Source
      </button>
    </div>

    <!-- Content -->
    <div class="entry-content card">
      <div v-if="viewMode === 'rendered'" class="markdown-content" v-html="renderedContent"></div>
      <pre v-else class="source-view"><code>{{ entry.content }}</code></pre>
    </div>

    <!-- Cross-references -->
    <EntryCrossReferencesBlock
      :outgoing="outgoingRefs"
      :incoming="incomingRefs"
      :research-slug="researchSlug"
    />

    <!-- External links -->
    <EntryExternalLinksBlock :links="externalLinks" />

    <!-- Related by tags -->
    <EntryRelatedEntriesBlock
      :entries="relatedEntries"
      :current-tags="entry.tags ?? []"
      :research-slug="researchSlug"
      :research-id="research?.id ?? ''"
    />

    <!-- Prev / Next navigation -->
    <EntryEntryNavigation
      v-if="siblings.length > 1"
      :prev="prevEntry"
      :next="nextEntry"
      :research-slug="researchSlug"
    />
  </div>

  <EmptyState v-else icon="&#x1F50D;" title="Entry not found" />
</template>

<script setup lang="ts">
import { marked } from 'marked'
import { tagHue } from '~/composables/useTagHue'

const route = useRoute()
const id = route.params.id as string
const entryId = route.params.entryId as string

marked.setOptions({ gfm: true, breaks: true })

// Research + sections for breadcrumb and sibling navigation
const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const research = computed(() => researchData.value?.data?.research)
const researchName = computed(() => research.value?.name ?? 'Research')
const researchSlug = computed(() => research.value?.code || id)
const sections = computed(() => researchData.value?.data?.sections ?? [])

// Entry data (pass research context for code-based lookup)
const { data, pending } = await useApi<{ data: any }>(`/api/researches/${id}/entries/${entryId}`)
const entry = computed(() => data.value?.data)

const sectionName = computed(() => {
  const sec = sections.value.find((s: any) => s.id === entry.value?.section_id)
  return sec?.display_name || sec?.name || 'Section'
})

// Linked session
const { data: sessionsData } = await useApi<{ data: any[] }>(`/api/researches/${id}/sessions`)
const linkedSession = computed(() => {
  if (!entry.value?.session_id) return null
  return (sessionsData.value?.data ?? []).find((s: any) => s.id === entry.value.session_id) ?? null
})

// Rendered markdown
const renderedContent = computed(() => {
  if (!entry.value?.content) return ''
  const html = marked.parse(entry.value.content) as string
  return renderRefs(html, researchSlug.value)
})

// View toggle
const viewMode = ref<'rendered' | 'source'>('rendered')

// Copy markdown
const copied = ref(false)
async function copyMarkdown() {
  if (!entry.value?.content) return
  await navigator.clipboard.writeText(entry.value.content)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}

// Cross-references
const { data: refsData } = useApi<{ outgoing: any[]; incoming: any[] }>(
  computed(() => entry.value ? `/api/entries/${entry.value.id}/crossrefs` : `/api/entries/__none__/crossrefs`)
)
const outgoingRefs = computed(() => refsData.value?.outgoing ?? [])
const incomingRefs = computed(() => refsData.value?.incoming ?? [])

// External links
const { data: linksData } = useApi<{ data: any[] }>(
  computed(() => entry.value ? `/api/entries/${entry.value.id}/links` : `/api/entries/__none__/links`)
)
const externalLinks = computed(() => linksData.value?.data ?? [])

// Related by tags
const { data: relatedData } = useApi<{ data: any[] }>(
  computed(() => entry.value ? `/api/entries/${entry.value.id}/related` : `/api/entries/__none__/related`)
)
const relatedEntries = computed(() => relatedData.value?.data ?? [])

// Sibling entries for prev/next navigation
const { data: siblingsData } = useApi<{ data: any[] }>(
  computed(() =>
    entry.value?.section_id
      ? `/api/researches/${id}/sections/${entry.value.section_id}/entries`
      : `/api/researches/__none__/sections/__none__/entries`
  )
)
const siblings = computed(() => siblingsData.value?.data ?? [])
const currIndex = computed(() => siblings.value.findIndex((e: any) => e.id === entryId))
const prevEntry = computed(() => currIndex.value > 0 ? siblings.value[currIndex.value - 1] : null)
const nextEntry = computed(() => currIndex.value < siblings.value.length - 1 ? siblings.value[currIndex.value + 1] : null)
</script>

<style scoped>
.entry-session-link {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-3);
  padding: var(--space-1) var(--space-3);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  text-decoration: none;
  transition: all var(--transition-fast);
}
.entry-session-link:hover { border-color: rgba(240, 184, 73, 0.3); color: var(--color-text); }
.entry-session-link svg { opacity: 0.6; }
.title-with-code { display: flex; align-items: center; gap: var(--space-3); }
.short-code {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
.entry-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-4);
}
.entry-actions {
  display: flex;
  gap: var(--space-2);
  align-items: center;
  flex-shrink: 0;
}
.entry-tags {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
  margin-top: var(--space-3);
}

/* View toggle */
.view-toggle {
  display: inline-flex;
  gap: 0;
  margin-bottom: var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 2px;
}
.view-toggle .btn {
  color: var(--color-text-muted);
  border: none;
  border-radius: calc(var(--radius-sm) - 2px);
  background: transparent;
}
.view-toggle .btn:hover {
  color: var(--color-text);
  background: transparent;
  transform: none;
  box-shadow: none;
}
.view-toggle .btn.active {
  background: var(--color-surface-hover);
  color: var(--color-text);
}

/* Content */
.entry-content {
  padding: var(--space-8);
  border-radius: var(--radius-lg);
}
.source-view {
  background: none;
  padding: 0;
  overflow-x: auto;
  font-size: var(--type-xs);
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--color-text-muted);
  margin: 0;
}
.source-view code { background: none; padding: 0; font-size: inherit; }

/* Skeleton */
.skeleton-header { height: 60px; margin-bottom: var(--space-4); }
.skeleton-content { height: 500px; }

/* Print */
@media print {
  .entry-content { padding: 0; border: none; }
  .entry-tags { margin-bottom: var(--space-2); }
}
</style>
