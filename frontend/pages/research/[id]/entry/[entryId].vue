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

      <!-- View mode header -->
      <template v-if="!editing">
        <div class="entry-header">
          <div class="title-with-code">
            <span v-if="entry.code" class="short-code">{{ entry.code }}</span>
            <h1 class="page-title">{{ entry.title }}</h1>
          </div>
          <div class="entry-actions no-print">
            <!-- Status dropdown -->
            <div class="status-dropdown-wrap" @click.stop>
              <button ref="statusTriggerEl" class="status-dropdown-trigger" @click="toggleStatus">
                <StatusBadge :status="entry.status" />
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="6 9 12 15 18 9"/></svg>
              </button>
              <Teleport to="body">
                <div v-if="statusOpen" class="status-dropdown-overlay" @click="statusOpen = false">
                  <div class="status-dropdown" :style="statusDropdownStyle" @click.stop>
                    <button
                      v-for="s in statuses"
                      :key="s"
                      :class="['status-option', { active: entry.status === s }]"
                      @click="changeStatus(s)"
                    >
                      <StatusBadge :status="s" />
                    </button>
                  </div>
                </div>
              </Teleport>
            </div>
            <button class="btn btn-sm" @click="startEditing">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              Edit
            </button>
            <button class="btn btn-sm" @click="copyMarkdown">
              <svg v-if="!copied" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
              <svg v-else width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
              {{ copied ? 'Copied' : 'Copy' }}
            </button>
            <button class="btn btn-sm btn-delete" @click="showDeleteConfirm = true">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
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
      </template>

      <!-- Edit mode header -->
      <template v-else>
        <div class="edit-header-bar">
          <div class="edit-field">
            <label class="edit-label">Title</label>
            <input v-model="editForm.title" class="edit-input" placeholder="Entry title..." />
          </div>
          <div class="edit-field">
            <label class="edit-label">Description</label>
            <input v-model="editForm.description" class="edit-input" placeholder="Short description (optional)..." />
          </div>
          <div class="edit-field">
            <label class="edit-label">Tags</label>
            <input v-model="editForm.tagsRaw" class="edit-input" placeholder="comma, separated, tags" />
          </div>
          <div class="edit-actions-bar">
            <button class="btn btn-sm btn-primary" :disabled="saving" @click="saveEntry">
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
            <button class="btn btn-sm" @click="cancelEditing">Cancel</button>
          </div>
        </div>
      </template>
    </div>

    <!-- View mode content -->
    <template v-if="!editing">
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
        <div v-if="viewMode === 'rendered'" ref="contentEl" class="markdown-content" v-html="renderedContent"></div>
        <pre v-else class="source-view"><code v-html="highlightedSource"></code></pre>
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
      <EntryNavigation
        v-if="siblings.length > 1"
        :prev="prevEntry"
        :next="nextEntry"
        :research-slug="researchSlug"
      />
    </template>

    <!-- Edit mode content -->
    <div v-else class="editor-wrap">
      <ClientOnly>
        <MdEditor
          v-model="editForm.content"
          language="en-US"
          :theme="'dark'"
          :preview="true"
          preview-theme="github"
          :toolbars="editorToolbars"
          :show-code-row-number="true"
          :auto-focus="true"
          style="height: 70vh;"
          @on-save="saveEntry"
        />
      </ClientOnly>
    </div>

    <!-- Delete confirmation -->
    <ConfirmModal
      :visible="showDeleteConfirm"
      title="Delete Entry"
      :message="`Are you sure you want to delete &quot;${entry.title}&quot;? This action cannot be undone.`"
      confirm-label="Delete"
      variant="danger"
      :loading="deleting"
      @confirm="deleteEntry"
      @cancel="showDeleteConfirm = false"
    />
  </div>

  <EmptyState v-else icon="&#x1F50D;" title="Entry not found" />
</template>

<script setup lang="ts">
import { marked } from 'marked'
import { MdEditor } from 'md-editor-v3'
import type { ToolbarNames } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { tagHue } from '~/composables/useTagHue'
import { renderMermaidBlocks } from '~/composables/useMermaid'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string
const entryId = route.params.entryId as string

marked.setOptions({ gfm: true, breaks: true })

const statuses = ['draft', 'active', 'completed', 'archived']

const editorToolbars: ToolbarNames[] = [
  'bold', 'underline', 'italic', 'strikeThrough', '-',
  'title', 'sub', 'sup', 'quote', '-',
  'unorderedList', 'orderedList', 'task', '-',
  'codeRow', 'code', 'link', 'image', 'table', 'mermaid', '-',
  'revoke', 'next', '=',
  'preview', 'catalog',
]

// Research + sections for breadcrumb and sibling navigation
const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const research = computed(() => researchData.value?.data?.research)
const researchName = computed(() => research.value?.name ?? 'Research')
const researchSlug = computed(() => research.value?.code || id)
const sections = computed(() => researchData.value?.data?.sections ?? [])

// Entry data (pass research context for code-based lookup)
const { data, pending, refresh } = await useApi<{ data: any }>(`/api/researches/${id}/entries/${entryId}`)
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
  const html = marked.parse(normalizeContent(entry.value.content)) as string
  return renderRefs(html, researchSlug.value)
})

// Syntax-highlighted markdown source
const highlightedSource = computed(() => {
  if (!entry.value?.content) return ''
  return highlightMarkdown(entry.value.content)
})

function highlightMarkdown(src: string): string {
  const esc = (s: string) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

  const lines = src.split('\n')
  const result: string[] = []
  let inCodeBlock = false
  let codeLang = ''

  for (const raw of lines) {
    const line = esc(raw)

    // Fenced code blocks
    if (/^```/.test(raw)) {
      if (!inCodeBlock) {
        inCodeBlock = true
        codeLang = raw.slice(3).trim()
        result.push(`<span class="md-fence">${line}</span>`)
      } else {
        inCodeBlock = false
        codeLang = ''
        result.push(`<span class="md-fence">${line}</span>`)
      }
      continue
    }
    if (inCodeBlock) {
      result.push(`<span class="md-code-line">${line}</span>`)
      continue
    }

    // Headings
    if (/^#{1,6}\s/.test(raw)) {
      const match = line.match(/^(#{1,6})\s(.*)$/)
      if (match) {
        result.push(`<span class="md-heading-marker">${match[1]}</span> <span class="md-heading">${inlineHighlight(match[2])}</span>`)
        continue
      }
    }

    // Blockquotes
    if (/^&gt;\s?/.test(line)) {
      result.push(`<span class="md-blockquote">${inlineHighlight(line)}</span>`)
      continue
    }

    // Horizontal rule
    if (/^(---|\*\*\*|___)$/.test(raw)) {
      result.push(`<span class="md-hr">${line}</span>`)
      continue
    }

    // Unordered list
    if (/^(\s*)([-*+])\s/.test(raw)) {
      const match = line.match(/^(\s*)([-*+])\s(.*)$/)
      if (match) {
        result.push(`${match[1]}<span class="md-list-marker">${match[2]}</span> ${inlineHighlight(match[3])}`)
        continue
      }
    }

    // Ordered list
    if (/^\s*\d+\.\s/.test(raw)) {
      const match = line.match(/^(\s*)(\d+\.)\s(.*)$/)
      if (match) {
        result.push(`${match[1]}<span class="md-list-marker">${match[2]}</span> ${inlineHighlight(match[3])}`)
        continue
      }
    }

    // Regular line with inline highlighting
    result.push(inlineHighlight(line))
  }

  return result.join('\n')
}

function inlineHighlight(line: string): string {
  return line
    // Images ![alt](url) — before links
    .replace(/!\[([^\]]*)\]\(([^)]+)\)/g, '<span class="md-image">![$1]($2)</span>')
    // Links [text](url)
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<span class="md-link">[$1](<span class="md-url">$2</span>)</span>')
    // Cross-references [[...]]
    .replace(/\[\[([^\]]+)\]\]/g, '<span class="md-crossref">[[$1]]</span>')
    // Bold **text** or __text__
    .replace(/(\*\*|__)(.+?)\1/g, '<span class="md-bold">$1$2$1</span>')
    // Italic *text* or _text_ (not inside bold markers)
    .replace(/(?<!\*)(\*|_)(?!\*)(.+?)\1(?!\*)/g, '<span class="md-italic">$1$2$1</span>')
    // Inline code `code`
    .replace(/`([^`]+)`/g, '<span class="md-inline-code">`$1`</span>')
}

// View toggle
const viewMode = ref<'rendered' | 'source'>('rendered')

// Mermaid rendering
const contentEl = ref<HTMLElement | null>(null)
watch([renderedContent, viewMode], () => {
  if (viewMode.value !== 'rendered') return
  nextTick(() => {
    if (contentEl.value) renderMermaidBlocks(contentEl.value)
  })
}, { immediate: false })
onMounted(() => {
  if (contentEl.value) renderMermaidBlocks(contentEl.value)
})

// Copy markdown
const copied = ref(false)
async function copyMarkdown() {
  if (!entry.value?.content) return
  await navigator.clipboard.writeText(entry.value.content)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}

// --- Edit mode ---
const { authFetch } = useAuth()
const rtBase = useRuntimeConfig().public.apiBase || ''
const editing = ref(false)
const saving = ref(false)
const editForm = reactive({
  title: '',
  description: '',
  content: '',
  tagsRaw: '',
})

function startEditing() {
  editForm.title = entry.value?.title ?? ''
  editForm.description = entry.value?.description ?? ''
  editForm.content = entry.value?.content ?? ''
  editForm.tagsRaw = (entry.value?.tags ?? []).join(', ')
  editing.value = true
}

function cancelEditing() {
  editing.value = false
}

async function saveEntry() {
  if (!entry.value || saving.value) return
  saving.value = true
  try {
    const tags = editForm.tagsRaw
      .split(',')
      .map((t: string) => t.trim())
      .filter(Boolean)

    await authFetch(`${rtBase}/api/entries/${entry.value.id}`, {
      method: 'PUT',
      body: {
        title: editForm.title,
        description: editForm.description || null,
        content: editForm.content,
        tags,
      },
    })
    editing.value = false
    await refresh()
  } catch (e: any) {
    console.error('Failed to save entry:', e)
    alert(e?.data?.error || e?.message || 'Failed to save')
  } finally {
    saving.value = false
  }
}

// --- Status change ---
const statusOpen = ref(false)
const statusTriggerEl = ref<HTMLElement | null>(null)
const statusDropdownStyle = ref<Record<string, string>>({})

function toggleStatus() {
  statusOpen.value = !statusOpen.value
  if (statusOpen.value && statusTriggerEl.value) {
    const rect = statusTriggerEl.value.getBoundingClientRect()
    statusDropdownStyle.value = {
      position: 'fixed',
      top: `${rect.bottom + 4}px`,
      left: `${rect.left}px`,
    }
  }
}

async function changeStatus(newStatus: string) {
  statusOpen.value = false
  if (!entry.value || entry.value.status === newStatus) return
  const prev = entry.value.status
  // Optimistic update
  data.value.data = { ...entry.value, status: newStatus }
  try {
    await authFetch(`${rtBase}/api/entries/${entry.value.id}`, {
      method: 'PUT',
      body: { status: newStatus },
    })
  } catch (e: any) {
    // Rollback on error
    data.value.data = { ...entry.value, status: prev }
    console.error('Failed to change status:', e)
  }
}


// --- Delete ---
const showDeleteConfirm = ref(false)
const deleting = ref(false)

async function deleteEntry() {
  if (!entry.value) return
  deleting.value = true
  try {
    await authFetch(`${rtBase}/api/entries/${entry.value.id}`, {
      method: 'DELETE',
    })
    showDeleteConfirm.value = false
    router.push(`/research/${researchSlug.value}?section=${entry.value.section_id}`)
  } catch (e: any) {
    console.error('Failed to delete entry:', e)
    alert(e?.data?.error || e?.message || 'Failed to delete')
  } finally {
    deleting.value = false
  }
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

/* Delete button */
.btn-delete {
  color: var(--color-text-muted);
  padding: 0.35rem 0.5rem;
}
.btn-delete:hover { color: #ef4444; }

/* Status dropdown */
.status-dropdown-wrap {
  position: relative;
}
.status-dropdown-trigger {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.2rem 0.4rem;
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
}
.status-dropdown-trigger:hover { background: var(--color-surface-hover); }
.status-dropdown-trigger svg { color: var(--color-text-muted); }

/* Edit mode */
.edit-header-bar {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  margin-top: var(--space-4);
}
.edit-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}
.edit-label {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.edit-input {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: var(--space-2) var(--space-3);
  color: var(--color-text);
  font-size: var(--type-sm);
  font-family: inherit;
  transition: border-color var(--transition-fast);
}
.edit-input:focus {
  outline: none;
  border-color: var(--color-primary);
}
.edit-actions-bar {
  display: flex;
  gap: var(--space-2);
  padding-top: var(--space-2);
}

/* Editor wrapper */
.editor-wrap {
  margin-top: var(--space-4);
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

/* md-editor-v3 dark theme overrides */
.editor-wrap :deep(.md-editor) {
  --md-bk-color: var(--color-bg) !important;
  --md-border-color: var(--color-border) !important;
  --md-color: var(--color-text) !important;
  --md-hover-color: var(--color-surface-hover) !important;
  --md-bk-hover-color: var(--color-surface-hover) !important;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}
.editor-wrap :deep(.md-editor-toolbar-wrapper) {
  background: var(--color-surface) !important;
  border-bottom: 1px solid var(--color-border) !important;
}
.editor-wrap :deep(.md-editor-content) {
  background: var(--color-bg) !important;
}
.editor-wrap :deep(.md-editor-preview-wrapper) {
  background: var(--color-bg) !important;
  padding: var(--space-6) !important;
}
.editor-wrap :deep(.md-editor-input-wrapper) {
  background: var(--color-bg) !important;
}
.editor-wrap :deep(.cm-editor) {
  background: var(--color-bg) !important;
  font-family: 'JetBrains Mono', 'Fira Code', monospace !important;
}
.editor-wrap :deep(.cm-gutters) {
  background: var(--color-surface) !important;
  border-right: 1px solid var(--color-border) !important;
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

/* Markdown syntax highlighting */
.source-view :deep(.md-heading-marker) { color: #e06c75; font-weight: 700; }
.source-view :deep(.md-heading) { color: #e5c07b; font-weight: 600; }
.source-view :deep(.md-bold) { color: #d19a66; font-weight: 600; }
.source-view :deep(.md-italic) { color: #c678dd; font-style: italic; }
.source-view :deep(.md-inline-code) { color: #98c379; background: rgba(152, 195, 121, 0.08); border-radius: 3px; padding: 0 0.15em; }
.source-view :deep(.md-link) { color: #61afef; }
.source-view :deep(.md-url) { color: #56b6c2; opacity: 0.7; }
.source-view :deep(.md-image) { color: #c678dd; }
.source-view :deep(.md-crossref) { color: #6cc5e0; background: rgba(108, 197, 224, 0.1); border-radius: 3px; padding: 0 0.15em; }
.source-view :deep(.md-blockquote) { color: #5c6370; font-style: italic; border-left: 2px solid #5c6370; padding-left: 0.75em; display: inline-block; }
.source-view :deep(.md-list-marker) { color: #e06c75; font-weight: 600; }
.source-view :deep(.md-hr) { color: #5c6370; }
.source-view :deep(.md-fence) { color: #98c379; }
.source-view :deep(.md-code-line) { color: #abb2bf; opacity: 0.85; }

/* Skeleton */
.skeleton-header { height: 60px; margin-bottom: var(--space-4); }
.skeleton-content { height: 500px; }

/* Responsive */
@media (max-width: 768px) {
  .entry-header { flex-direction: column; align-items: flex-start; gap: var(--space-3); }
  .entry-actions { flex-wrap: wrap; }
  .title-with-code { flex-wrap: wrap; gap: var(--space-2); }
  .entry-content { padding: var(--space-4); }
}

/* Print */
@media print {
  .entry-content { padding: 0; border: none; }
  .entry-tags { margin-bottom: var(--space-2); }
}
</style>

<style>
/* Teleported status dropdown — must be non-scoped */
.status-dropdown-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
}
.status-dropdown {
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius);
  padding: var(--space-1);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  min-width: 140px;
}
.status-option {
  display: block;
  width: 100%;
  background: none;
  border: none;
  padding: var(--space-2) var(--space-3);
  cursor: pointer;
  border-radius: var(--radius-sm);
  text-align: left;
  transition: background 0.15s;
}
.status-option:hover { background: var(--color-surface-hover); }
.status-option.active { background: var(--color-surface-hover); }
</style>
