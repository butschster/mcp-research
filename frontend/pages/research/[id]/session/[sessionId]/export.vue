<template>
  <div ref="docRoot" v-if="pending" class="skeleton-page">
    <div class="skeleton-card" style="height: 60px; margin-bottom: 2rem;"></div>
    <div class="skeleton-card" style="height: 600px;"></div>
  </div>

  <div v-else-if="exportData" class="export-page">
    <!-- Toolbar (hidden in print) -->
    <div class="export-toolbar no-print">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: exportData.research.name, to: `/research/${researchSlug}` },
        { label: exportData.session.title, to: `/research/${researchSlug}/session/${sessionSlug}` },
        { label: 'Export' }
      ]" />
      <div class="toolbar-actions">
        <button class="btn btn-sm" @click="downloadMarkdown">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          Download .md
        </button>
        <button class="btn btn-sm" @click="printPage">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 6 2 18 2 18 9"/><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"/><rect x="6" y="14" width="12" height="8"/></svg>
          Print / PDF
        </button>
      </div>
    </div>

    <!-- Document -->
    <article class="export-document">
      <header class="doc-header">
        <p class="doc-parent">{{ exportData.research.name }}</p>
        <h1 class="doc-title">{{ exportData.session.title }}</h1>
        <p v-if="exportData.session.focus" class="doc-goal" v-html="renderInline(exportData.session.focus)"></p>
        <div class="doc-meta">
          <span v-if="exportData.session.code">Code: {{ exportData.session.code }}</span>
          <span>Status: {{ exportData.session.status }}</span>
          <span>Questions: {{ answeredCount }} / {{ questions.length }} answered</span>
          <span>Entries: {{ entries.length }}</span>
          <span>Exported: {{ new Date().toLocaleDateString() }}</span>
        </div>
      </header>

      <section v-if="exportData.session.notes" class="doc-section">
        <h2 class="section-heading">Notes</h2>
        <div class="markdown-content" v-html="renderMarkdown(exportData.session.notes)"></div>
      </section>

      <section id="questions" class="doc-section">
        <h2 class="section-heading">Questions</h2>

        <div v-if="!questions.length" class="doc-empty">No questions in this session</div>

        <div v-for="q in questions" :key="q.id" class="doc-question">
          <div class="q-heading">
            <span class="q-label">Q:</span>
            <span class="q-text" v-html="renderInline(q.text)"></span>
            <span :class="['q-status', `q-status-${q.status}`]">{{ humanStatus(q.status) }}</span>
          </div>
          <div class="q-meta">
            <span v-if="q.area">Area: {{ q.area }}</span>
            <span>Priority: {{ q.priority }}</span>
            <span>Status: {{ humanStatus(q.status) }}</span>
          </div>
          <p v-if="q.rationale" class="q-rationale" v-html="renderInline(q.rationale)"></p>
          <div v-if="q.answer" class="q-answer markdown-content" v-html="renderMarkdown(q.answer)"></div>
        </div>
      </section>

      <section id="entries" class="doc-section">
        <h2 class="section-heading">Entries produced in this session</h2>

        <div v-if="!entries.length" class="doc-empty">
          No entries yet — they appear here once an entry is linked to this session.
        </div>

        <article v-for="entry in entries" :key="entry.id" class="doc-entry">
          <h3 class="entry-heading">
            <span v-if="entry.code" class="entry-code">{{ entry.code }}</span>
            {{ entry.title || 'Untitled entry' }}
          </h3>
          <div class="entry-meta">
            <span v-if="sectionNames[entry.section_id]" class="entry-section">
              {{ sectionNames[entry.section_id] }}
            </span>
            <span>{{ humanStatus(entry.status) }}</span>
            <span v-for="tag in entry.tags || []" :key="tag" class="doc-tag doc-tag-sm">{{ tag }}</span>
          </div>
          <!-- A blocks entry renders through the block renderer rather than its
               markdown projection: this page is what becomes the PDF, and a
               diagram or an HTML visual is exactly what must not be flattened
               to a note here. Downloading .md still gets the text form. -->
          <BlocksBlockRenderer
            v-if="blocksOf(entry)"
            :blocks="blocksOf(entry)"
            :research-slug="researchSlug"
            readonly
          />
          <div v-else class="markdown-content" v-html="renderMarkdown(entryBody(entry))"></div>
        </article>
      </section>
    </article>
  </div>

  <EmptyState
    v-else
    icon="&#x1F50D;"
    title="Could not load this session"
    :description="`It may not exist, or the request failed. Open the research at /research/${id} and pick the session from there.`"
  />
</template>

<script setup lang="ts">
import { marked } from 'marked'
import { renderMermaidBlocks } from '~/composables/useMermaid'
import { renderRefs } from '~/composables/useCrossRefs'

marked.setOptions({ gfm: true, breaks: true })

const route = useRoute()
const id = route.params.id as string
const sessionId = route.params.sessionId as string

const { data: raw, pending } = await useApi<any>(
  `/api/researches/${id}/sessions/${sessionId}/export`
)
const exportData = computed(() => raw.value)
const researchSlug = computed(() => exportData.value?.research?.code || id)
const sessionSlug = computed(() => exportData.value?.session?.code || sessionId)
const questions = computed<any[]>(() => exportData.value?.questions ?? [])
const entries = computed<any[]>(() => exportData.value?.entries ?? [])
const sectionNames = computed<Record<string, string>>(() => exportData.value?.section_names ?? {})
const answeredCount = computed(() => questions.value.filter((q) => q.status === 'answered').length)

// A blocks entry stores JSON; the server ships a markdown rendering beside it so
// the printable document does not have to know the block format.
function entryBody(entry: any): string {
  return entry?.content_markdown || entry?.content || ''
}

const docRoot = ref<HTMLElement | null>(null)

// A blocks entry ships its document as JSON; parse it once per entry so the
// renderer gets blocks rather than a string.
function blocksOf(entry: any): any[] | null {
  if (entry?.entry_type !== 'blocks' || !entry?.content) return null
  try {
    const doc = JSON.parse(entry.content)
    const blocks = Array.isArray(doc) ? doc : doc?.blocks
    return Array.isArray(blocks) && blocks.length ? blocks : null
  } catch {
    return null
  }
}

// Markdown entries carry their diagrams as ```mermaid fences. Draw them, or the
// printed page shows the source of a diagram instead of the diagram.
async function drawDiagrams() {
  await nextTick()
  if (docRoot.value) await renderMermaidBlocks(docRoot.value)
}
onMounted(drawDiagrams)
watch(() => exportData.value, drawDiagrams)

function renderMarkdown(content: string): string {
  if (!content) return ''
  const html = marked.parse(normalizeContent(content)) as string
  return renderRefs(html, researchSlug.value)
}

// Plain-text fields still carry [[E3]] references, and they are rendered as
// links everywhere else in the app. renderRefs does not escape, so escape first.
function renderInline(text: string): string {
  if (!text) return ''
  const escaped = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
  return renderRefs(escaped, researchSlug.value)
}

function humanStatus(status: string): string {
  if (!status) return ''
  return status.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}

function downloadMarkdown() {
  const md = exportData.value?.markdown
  if (!md) return
  const blob = new Blob([md], { type: 'text/markdown' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${exportData.value.session.title.replace(/\s+/g, '_')}.md`
  a.click()
  URL.revokeObjectURL(url)
}

function printPage() {
  window.print()
}
</script>

<style scoped>
.export-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
  margin-bottom: var(--space-5);
}
.toolbar-actions { display: flex; gap: var(--space-2); }

/* No max-width here: the document fills the page .container, the same way the
   research export page does. Constraining it made the content column narrower
   than the rest of the app and jump on load, because the skeleton above uses
   the full container width. */

.doc-header {
  padding-bottom: var(--space-4);
  margin-bottom: var(--space-5);
  border-bottom: 1px solid var(--color-border);
}
.doc-parent {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  margin-bottom: var(--space-1);
}
.doc-title {
  font-size: var(--type-2xl);
  font-weight: 700;
  letter-spacing: -0.03em;
  line-height: var(--line-tight);
  text-wrap: balance;
  overflow-wrap: anywhere;
}
.doc-goal {
  margin-top: var(--space-2);
  color: var(--color-text-muted);
  font-style: italic;
}
.doc-meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  margin-top: var(--space-3);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

.doc-section { margin-bottom: var(--space-6); }
.section-heading {
  font-size: var(--type-lg);
  font-weight: 650;
  margin-bottom: var(--space-3);
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--color-border);
}
.doc-empty {
  color: var(--color-text-muted);
  font-style: italic;
  font-size: var(--type-sm);
}

.doc-question {
  padding: var(--space-3) 0;
  border-bottom: 1px solid var(--color-border);
  break-inside: avoid;
}
.doc-question:last-child { border-bottom: none; }
.q-heading {
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
  flex-wrap: wrap;
}
.q-label { color: var(--color-primary); font-weight: 700; }
.q-text { font-weight: 600; flex: 1; min-width: 0; overflow-wrap: anywhere; }
.q-status {
  font-size: var(--type-xs);
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
  flex-shrink: 0;
}
.q-status-answered { color: var(--color-success); }
.q-status-in_progress { color: var(--color-info); }
.q-status-deferred { color: var(--color-warning); }
.q-status-skipped { color: var(--color-text-muted); text-decoration: line-through; }
.q-meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  margin-top: var(--space-1);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.q-rationale {
  margin-top: var(--space-1);
  font-size: var(--type-sm);
  font-style: italic;
  color: var(--color-text-muted);
}
.q-answer { margin-top: var(--space-2); }

.doc-entry {
  padding: var(--space-3) 0;
  border-bottom: 1px solid var(--color-border);
  break-inside: avoid;
}
.doc-entry:last-child { border-bottom: none; }
.entry-heading {
  font-size: var(--type-base);
  font-weight: 650;
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
}
.entry-code {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: var(--type-xs);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.1rem 0.35rem;
  border-radius: var(--radius-sm);
}
.entry-meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin: var(--space-1) 0 var(--space-2);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.doc-tag {
  padding: 0.1rem 0.4rem;
  border: 1px solid var(--color-border);
  border-radius: 3px;
}
.doc-tag-sm { font-size: var(--type-xs); }

@media (max-width: 768px) {
  .export-toolbar { align-items: flex-start; }
  .doc-title { font-size: var(--type-xl); }
  .doc-meta { gap: var(--space-2); }
}

@media print {
  .no-print { display: none !important; }
  .export-document { max-width: none; }
  .doc-section { break-inside: auto; }
  /* The dark-theme border and muted-text tokens are near-invisible on white. */
  .doc-header,
  .section-heading,
  .doc-question,
  .doc-entry { border-color: #ddd; }
  .doc-parent,
  .doc-goal,
  .doc-meta,
  .q-meta,
  .q-rationale,
  .entry-meta,
  .doc-empty { color: #555; }
  .entry-code { background: #e8f4f8; }
  .doc-tag { background: #f0f0f0; border-color: #ddd; }
}
</style>
