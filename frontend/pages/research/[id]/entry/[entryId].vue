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
        { label: researchName, to: `/research/${id}` },
        { label: sectionName, to: `/research/${id}?section=${entry.section_id}` },
        { label: entry.title }
      ]" />
      <div class="entry-header">
        <h1 class="page-title">{{ entry.title }}</h1>
        <div class="entry-actions">
          <StatusBadge :status="entry.status" />
          <PrintButton />
          <button class="btn btn-sm" @click="copyMarkdown">
            {{ copied ? '&#x2713; Copied' : 'Copy' }}
          </button>
        </div>
      </div>
      <p v-if="entry.description" class="card-meta mt-2">{{ entry.description }}</p>
      <div v-if="entry.tags?.length" class="entry-tags">
        <span v-for="tag in entry.tags" :key="tag" :class="['tag', `tag-hue-${tagHue(tag)}`]">{{ tag }}</span>
      </div>
    </div>

    <!-- View toggle -->
    <div class="view-toggle">
      <button :class="['btn btn-sm', { active: viewMode === 'rendered' }]" @click="viewMode = 'rendered'">
        Rendered
      </button>
      <button :class="['btn btn-sm', { active: viewMode === 'source' }]" @click="viewMode = 'source'">
        Source
      </button>
    </div>

    <!-- Content -->
    <div class="entry-content card">
      <div v-if="viewMode === 'rendered'" class="markdown-content" v-html="renderedContent"></div>
      <pre v-else class="source-view"><code>{{ entry.content }}</code></pre>
    </div>

    <!-- Prev / Next navigation -->
    <div v-if="siblings.length > 1" class="entry-nav">
      <NuxtLink v-if="prevEntry" :to="`/research/${id}/entry/${prevEntry.id}`" class="btn btn-sm entry-nav-btn">
        &larr; {{ prevEntry.title }}
      </NuxtLink>
      <span v-else class="entry-nav-placeholder"></span>
      <NuxtLink v-if="nextEntry" :to="`/research/${id}/entry/${nextEntry.id}`" class="btn btn-sm entry-nav-btn entry-nav-next">
        {{ nextEntry.title }} &rarr;
      </NuxtLink>
    </div>
  </div>

  <EmptyState v-else icon="&#x1F50D;" title="Entry not found" />
</template>

<script setup lang="ts">
import { marked } from 'marked'

const route = useRoute()
const id = route.params.id as string
const entryId = route.params.entryId as string

marked.setOptions({ gfm: true, breaks: true })

// Research + sections for breadcrumb and sibling navigation
const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? 'Research')
const sections = computed(() => researchData.value?.data?.sections ?? [])

// Entry data
const { data, pending } = await useApi<{ data: any }>(`/api/entries/${entryId}`)
const entry = computed(() => data.value?.data)

const sectionName = computed(() => {
  const sec = sections.value.find((s: any) => s.id === entry.value?.section_id)
  return sec?.display_name || sec?.name || 'Section'
})

// Tag color
function tagHue(tag: string): number {
  return [...tag].reduce((acc, c) => acc + c.charCodeAt(0), 0) % 6
}

// Rendered markdown
const renderedContent = computed(() =>
  entry.value?.content ? marked.parse(entry.value.content) as string : ''
)

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
  display: flex;
  gap: var(--space-1);
  margin-bottom: var(--space-4);
}
.view-toggle .btn { color: var(--color-text-muted); }
.view-toggle .btn.active {
  background: var(--color-surface);
  color: var(--color-text);
  border-color: var(--color-primary);
}

/* Content */
.entry-content { padding: var(--space-6); }
.source-view {
  background: var(--color-surface-hover);
  padding: var(--space-4);
  border-radius: var(--radius);
  overflow-x: auto;
  font-size: var(--type-xs);
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--color-text-muted);
  margin: 0;
}
.source-view code { background: none; padding: 0; font-size: inherit; }

/* Navigation */
.entry-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--space-6);
  gap: var(--space-4);
}
.entry-nav-btn {
  max-width: 45%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.entry-nav-next { margin-left: auto; text-align: right; }
.entry-nav-placeholder { flex: 1; }

/* Skeleton */
.skeleton-card { background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius); opacity: 0.5; }
.skeleton-header { height: 60px; margin-bottom: var(--space-4); }
.skeleton-content { height: 500px; }
</style>
