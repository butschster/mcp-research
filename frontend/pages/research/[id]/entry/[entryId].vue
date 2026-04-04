<template>
  <div v-if="pending">
    <div class="skeleton-card" style="height:60px;margin-bottom:1rem;"></div>
    <div class="skeleton-card" style="height:500px;"></div>
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
      <div class="flex-between" style="margin-top:0.25rem;">
        <h1 class="page-title">{{ entry.title }}</h1>
        <div style="display:flex;gap:0.5rem;align-items:center;">
          <StatusBadge :status="entry.status" />
          <PrintButton />
          <button class="btn btn-icon" title="Copy markdown" @click="copyMarkdown">
            {{ copied ? '✓ Copied' : '⎘ Copy' }}
          </button>
        </div>
      </div>
      <p v-if="entry.description" class="card-meta" style="margin-top:0.25rem;">
        {{ entry.description }}
      </p>
      <div v-if="entry.tags?.length" style="margin-top:0.625rem;display:flex;gap:0.375rem;flex-wrap:wrap;">
        <span v-for="tag in entry.tags" :key="tag" class="tag">{{ tag }}</span>
      </div>
    </div>

    <!-- View toggle -->
    <div class="view-toggle">
      <button :class="['toggle-btn', { active: viewMode === 'rendered' }]" @click="viewMode = 'rendered'">
        Rendered
      </button>
      <button :class="['toggle-btn', { active: viewMode === 'source' }]" @click="viewMode = 'source'">
        Source
      </button>
    </div>

    <!-- Content -->
    <div class="entry-content card">
      <!-- Rendered markdown -->
      <div
        v-if="viewMode === 'rendered'"
        class="markdown-content"
        v-html="renderedContent"
      ></div>
      <!-- Raw source -->
      <pre v-else class="source-view"><code>{{ entry.content }}</code></pre>
    </div>

    <!-- Prev / Next navigation -->
    <div v-if="siblings.length > 1" class="entry-nav">
      <NuxtLink
        v-if="prevEntry"
        :to="`/research/${id}/entry/${prevEntry.id}`"
        class="btn entry-nav-btn"
      >
        ← {{ prevEntry.title }}
      </NuxtLink>
      <span v-else class="entry-nav-placeholder"></span>
      <NuxtLink
        v-if="nextEntry"
        :to="`/research/${id}/entry/${nextEntry.id}`"
        class="btn entry-nav-btn entry-nav-next"
      >
        {{ nextEntry.title }} →
      </NuxtLink>
    </div>
  </div>

  <EmptyState v-else icon="🔍" title="Entry not found" />
</template>

<script setup lang="ts">
import { marked } from 'marked'

const route = useRoute()
const id = route.params.id as string
const entryId = route.params.entryId as string

// Configure marked once
marked.setOptions({ gfm: true, breaks: true })

// Research + sections for breadcrumb and sibling navigation
const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? 'Research')
const sections     = computed(() => researchData.value?.data?.sections ?? [])

// Entry data
const { data, pending } = await useApi<{ data: any }>(`/api/entries/${entryId}`)
const entry = computed(() => data.value?.data)

const sectionName = computed(() => {
  const sec = sections.value.find((s: any) => s.id === entry.value?.section_id)
  return sec?.display_name || sec?.name || 'Section'
})

// Rendered markdown
const renderedContent = computed(() =>
  entry.value?.content ? marked.parse(entry.value.content) as string : ''
)

// View toggle (rendered vs source)
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
const siblings  = computed(() => siblingsData.value?.data ?? [])
const currIndex = computed(() => siblings.value.findIndex((e: any) => e.id === entryId))
const prevEntry = computed(() => currIndex.value > 0 ? siblings.value[currIndex.value - 1] : null)
const nextEntry = computed(() => currIndex.value < siblings.value.length - 1 ? siblings.value[currIndex.value + 1] : null)
</script>

<style scoped>
.flex-between { display: flex; justify-content: space-between; align-items: flex-start; }

.view-toggle {
  display: flex;
  gap: 0;
  margin-bottom: 1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
  width: fit-content;
}
.toggle-btn {
  padding: 0.375rem 0.875rem;
  background: none;
  border: none;
  color: var(--color-text-muted);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.15s;
}
.toggle-btn.active {
  background: var(--color-surface);
  color: var(--color-text);
  font-weight: 500;
}
.toggle-btn:hover:not(.active) { color: var(--color-text); }

.entry-content { padding: 1.5rem; }
.source-view {
  background: var(--color-surface-hover);
  padding: 1rem;
  border-radius: var(--radius);
  overflow-x: auto;
  font-size: 0.8125rem;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--color-text-muted);
  margin: 0;
}
.source-view code { background: none; padding: 0; font-size: inherit; }

.btn-icon { font-size: 0.8125rem; }

.entry-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 1.5rem;
  gap: 1rem;
}
.entry-nav-btn {
  max-width: 45%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.8125rem;
}
.entry-nav-next   { margin-left: auto; text-align: right; }
.entry-nav-placeholder { flex: 1; }

.skeleton-card { background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius); animation: shimmer 1.5s infinite; }
@keyframes shimmer { 0%,100%{opacity:.6} 50%{opacity:1} }
</style>
