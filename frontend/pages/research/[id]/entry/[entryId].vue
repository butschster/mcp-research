<template>
  <div v-if="pending" class="empty-state">Loading...</div>
  <div v-else-if="!entry" class="empty-state">Entry not found.</div>
  <div v-else>
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: researchName, to: `/research/${id}` },
        { label: sectionName, to: `/research/${id}?section=${entry.section_id}` },
        { label: entry.title },
      ]" />
      <div style="display: flex; justify-content: space-between; align-items: center; margin-top: 0.25rem;">
        <h1 class="page-title">{{ entry.title }}</h1>
        <div style="display: flex; gap: 0.5rem; align-items: center;">
          <StatusBadge :status="entry.status" />
          <PrintButton />
        </div>
      </div>
      <p v-if="entry.description" class="card-meta" style="margin-top: 0.25rem;">{{ entry.description }}</p>
      <div v-if="entry.tags?.length" style="margin-top: 0.75rem; display: flex; gap: 0.375rem;">
        <span v-for="tag in entry.tags" :key="tag" class="tag">{{ tag }}</span>
      </div>
    </div>

    <div class="card">
      <div class="markdown-content" v-html="renderedContent" />
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const id = route.params.id as string
const entryId = route.params.entryId as string

const { data: researchData } = await useApi<{ data: { research: any; sections: any[] } }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? 'Project')
const researchSections = computed(() => researchData.value?.data?.sections ?? [])
const sectionName = computed(() => {
  const sec = researchSections.value.find((s: any) => s.id === entry.value?.section_id)
  return sec?.display_name || sec?.name || 'Section'
})

const { data, pending } = await useApi<{ data: any }>(`/api/entries/${entryId}`)
const entry = computed(() => data.value?.data)

// Simple markdown to HTML (basic rendering)
const renderedContent = computed(() => {
  if (!entry.value?.content) return ''
  return basicMarkdown(entry.value.content)
})

function basicMarkdown(md: string): string {
  return md
    // Code blocks
    .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code class="language-$1">$2</code></pre>')
    // Inline code
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    // Headers
    .replace(/^### (.+)$/gm, '<h3>$1</h3>')
    .replace(/^## (.+)$/gm, '<h2>$1</h2>')
    .replace(/^# (.+)$/gm, '<h1>$1</h1>')
    // Bold and italic
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    // Links
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank">$1</a>')
    // Blockquotes
    .replace(/^> (.+)$/gm, '<blockquote>$1</blockquote>')
    // Unordered lists
    .replace(/^- (.+)$/gm, '<li>$1</li>')
    // Paragraphs
    .replace(/\n\n/g, '</p><p>')
    .replace(/^(?!<[huprolb])(.+)$/gm, '<p>$1</p>')
}
</script>
