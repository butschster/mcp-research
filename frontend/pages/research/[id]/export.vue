<template>
  <div v-if="pending" class="skeleton-page">
    <div class="skeleton-card" style="height: 60px; margin-bottom: 2rem;"></div>
    <div class="skeleton-card" style="height: 600px;"></div>
  </div>

  <div ref="docRoot" v-else-if="exportData" class="export-page">
    <!-- Toolbar (hidden in print) -->
    <div class="export-toolbar no-print">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: exportData.research.name, to: `/research/${researchSlug}` },
        { label: 'Export' }
      ]" />
      <div class="toolbar-actions">
        <!-- "Take it with you" is one group; printing is a different kind of
             act, so a wider gap does the grouping rather than a divider. -->
        <div class="toolbar-group">
          <button class="btn btn-sm" @click="downloadMarkdown">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            Download .md
          </button>
          <ExportSplitButton
            label="Obsidian .zip"
            busy-label="Preparing…"
            options-label="Obsidian export options"
            hint="Sections become folders, entries become notes, references become links. Opens in Obsidian or any markdown editor."
            :busy="downloading"
            :with-options="hasOptionalContent"
            :options-open="optionsOpen"
            @download="downloadVault()"
            @options="optionsOpen = true"
          >
            <template #icon>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/></svg>
            </template>
          </ExportSplitButton>
        </div>
        <button class="btn btn-sm" @click="printPage">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 6 2 18 2 18 9"/><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"/><rect x="6" y="14" width="12" height="8"/></svg>
          Print / PDF
        </button>
      </div>
    </div>

    <ResearchObsidianExportDialog
      :visible="optionsOpen"
      :counts="optionCounts"
      @close="optionsOpen = false"
      @download="onDialogDownload"
    />

    <!-- Document -->
    <article class="export-document">
      <!-- Cover -->
      <header class="doc-header">
        <h1 class="doc-title">{{ exportData.research.name }}</h1>
        <p v-if="exportData.research.goal" class="doc-goal">{{ exportData.research.goal }}</p>
        <p v-if="exportData.research.description" class="doc-description">{{ exportData.research.description }}</p>
        <div v-if="exportData.research.tags?.length" class="doc-tags">
          <span v-for="tag in exportData.research.tags" :key="tag" class="doc-tag">{{ tag }}</span>
        </div>
        <div class="doc-meta">
          <span>Status: {{ exportData.research.status }}</span>
          <span>Sections: {{ exportData.sections?.length || 0 }}</span>
          <span>Entries: {{ totalEntries }}</span>
          <span>Exported: {{ new Date().toLocaleDateString() }}</span>
        </div>
      </header>

      <!-- Table of contents -->
      <nav class="doc-toc">
        <h2>Table of Contents</h2>
        <ol>
          <li v-for="(section, i) in exportData.sections" :key="section.id">
            <a :href="`#section-${i}`">{{ section.display_name || section.name }}</a>
            <span class="toc-count">{{ section.entries?.length || 0 }} entries</span>
          </li>
          <li v-if="exportData.sessions?.length">
            <a href="#sessions">Sessions</a>
            <span class="toc-count">{{ exportData.sessions.length }}</span>
          </li>
          <li v-if="exportData.tasks?.length">
            <a href="#tasks">Tasks</a>
            <span class="toc-count">{{ exportData.tasks.length }}</span>
          </li>
        </ol>
      </nav>

      <!-- Sections -->
      <section
        v-for="(section, i) in exportData.sections"
        :key="section.id"
        :id="`section-${i}`"
        class="doc-section"
      >
        <h2 class="section-heading">{{ section.display_name || section.name }}</h2>
        <p v-if="section.description" class="section-desc">{{ section.description }}</p>

        <div v-if="!section.entries?.length" class="doc-empty">No entries</div>

        <article
          v-for="entry in section.entries"
          :key="entry.id"
          class="doc-entry"
        >
          <h3 class="entry-heading">
            <span v-if="entry.code" class="entry-code">{{ entry.code }}</span>
            {{ entry.title }}
          </h3>
          <div v-if="entry.tags?.length" class="entry-meta">
            <span v-for="tag in entry.tags" :key="tag" class="doc-tag doc-tag-sm">{{ tag }}</span>
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

      <!-- Sessions -->
      <section v-if="exportData.sessions?.length" id="sessions" class="doc-section">
        <h2 class="section-heading">Sessions</h2>

        <div v-for="sess in exportData.sessions" :key="sess.id" class="doc-session">
          <h3 class="session-heading">{{ sess.title }}</h3>
          <p v-if="sess.focus" class="session-focus">Focus: {{ sess.focus }}</p>

          <div v-for="q in sess.questions" :key="q.id" class="doc-question">
            <div class="q-heading">
              <span class="q-label">Q:</span>
              <span class="q-text">{{ q.text }}</span>
              <span :class="['q-status', `q-status-${q.status}`]">{{ q.status }}</span>
            </div>
            <div v-if="q.area" class="q-meta">Area: {{ q.area }} | Priority: {{ q.priority }}</div>
            <div v-if="q.answer" class="q-answer markdown-content" v-html="renderMarkdown(q.answer)"></div>
          </div>
        </div>
      </section>

      <!-- Tasks -->
      <section v-if="exportData.tasks?.length" id="tasks" class="doc-section">
        <h2 class="section-heading">Tasks</h2>

        <div v-for="t in exportData.tasks" :key="t.id" class="doc-task">
          <div class="task-row">
            <span :class="['task-check', t.status === 'completed' ? 'task-done' : '']">
              {{ t.status === 'completed' ? '[x]' : t.status === 'failed' ? '[!]' : '[ ]' }}
            </span>
            <span class="task-title">{{ t.title }}</span>
            <span v-if="t.priority === 'high'" class="task-priority">high</span>
            <span class="task-status">{{ t.status }}</span>
          </div>
          <p v-if="t.description" class="task-desc">{{ t.description }}</p>
          <div v-if="t.result" class="task-result markdown-content" v-html="renderMarkdown(t.result)"></div>
        </div>
      </section>
    </article>
  </div>

  <EmptyState v-else icon="&#x1F50D;" title="Research not found" />
</template>

<script setup lang="ts">
import { marked } from 'marked'
import { renderMermaidBlocks } from '~/composables/useMermaid'
import type { ObsidianExportOptions } from '~/composables/useObsidianExport'
marked.setOptions({ gfm: true, breaks: true })

const route = useRoute()
const id = route.params.id as string

const { data: raw, pending } = await useApi<any>(`/api/researches/${id}/export`)
const exportData = computed(() => raw.value)
const researchSlug = computed(() => exportData.value?.research?.code || id)
const totalEntries = computed(() =>
  (exportData.value?.sections ?? []).reduce((sum: number, s: any) => sum + (s.entries?.length || 0), 0)
)

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

function downloadMarkdown() {
  const md = exportData.value?.markdown
  if (!md) return
  const research = exportData.value.research
  // Same saver and the same filename rules as the vault: a colon in a research
  // name produced a file Windows refuses.
  saveBlob(
    new Blob([md], { type: 'text/markdown' }),
    vaultFallbackName(research.code || '', research.name || '').replace(/\.zip$/, '.md'),
  )
}

function printPage() {
  window.print()
}

// --- Obsidian vault ---

const {
  pending: downloading,
  slow: downloadSlow,
  error: downloadError,
  filename: downloadedName,
  start: startDownload,
} = useDownload()
const toasts = useToasts()
const optionsOpen = ref(false)
const lastOptions = ref<ObsidianExportOptions>(defaultObsidianOptions())

const optionCounts = computed(() => ({
  sessions: exportData.value?.sessions?.length ?? 0,
  tasks: exportData.value?.tasks?.length ?? 0,
  roadmaps: exportData.value?.roadmap_count ?? 0,
}))

// With nothing optional in the research, the only live toggles would be "HTML"
// and "Revisions" — not worth an affordance, so the caret disappears and the
// plain button remains.
const hasOptionalContent = computed(() => {
  const c = optionCounts.value
  return c.sessions > 0 || c.tasks > 0 || c.roadmaps > 0
})

async function downloadVault(opts: ObsidianExportOptions = defaultObsidianOptions()) {
  lastOptions.value = opts
  const research = exportData.value?.research
  const ok = await startDownload(
    obsidianExportPath(researchSlug.value, opts),
    vaultFallbackName(research?.code || '', research?.name || ''),
  )
  if (ok) {
    toasts.success(downloadedName.value || 'The archive was saved.', 'Downloaded')
    return
  }
  const err = downloadError.value
  if (!err) return
  toasts.error(err.message, 'Export failed', retryAction(err.status))
}

// Each failure carries the one act that can answer it.
function retryAction(status: number) {
  if (status === 404) return { label: 'Reload the page', onClick: () => window.location.reload() }
  if (status === 401 || status === 403) return { label: 'Sign in', onClick: () => navigateTo('/login') }
  return { label: 'Try again', onClick: () => { downloadVault(lastOptions.value) } }
}

// The button says "Preparing…" on its own; this says why it is taking so long,
// and only when it is.
watch(downloadSlow, (slow) => {
  if (slow) toasts.push({ message: 'Still working — a large research takes a moment.' })
})

function onDialogDownload(opts: ObsidianExportOptions) {
  // The dialog closes first so there is one place showing download state, and
  // focus lands beside the line that will speak.
  optionsOpen.value = false
  downloadVault(opts)
}

</script>

<style scoped>
/* Toolbar */
.export-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}
/* The wide gap between the group and Print is the only grouping mechanism —
   no segmented control, no divider, both of which would be a new idiom in a
   product that has none. */
.toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2) var(--space-5);
  justify-content: flex-end;
}
.toolbar-group { display: flex; gap: var(--space-2); }


/* Document */
.export-document {
}

/* Header */
.doc-header {
  padding-bottom: var(--space-8);
  margin-bottom: var(--space-8);
  border-bottom: 2px solid var(--color-border);
}
.doc-title {
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  line-height: 1.2;
  margin-bottom: var(--space-3);
}
.doc-goal {
  font-size: var(--type-lg);
  color: var(--color-text-muted);
  line-height: 1.5;
  margin-bottom: var(--space-2);
}
.doc-description {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  line-height: 1.6;
  margin-bottom: var(--space-3);
}
.doc-tags { display: flex; gap: var(--space-2); flex-wrap: wrap; margin-bottom: var(--space-4); }
.doc-tag {
  font-size: var(--type-xs);
  padding: 0.15rem 0.5rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 3px;
  color: var(--color-text-muted);
}
.doc-tag-sm { font-size: 0.65rem; padding: 0.1rem 0.35rem; }
.doc-meta {
  display: flex;
  gap: var(--space-4);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}

/* TOC */
.doc-toc {
  padding: var(--space-6);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  margin-bottom: var(--space-8);
}
.doc-toc h2 {
  font-size: var(--type-sm);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  margin-bottom: var(--space-3);
}
.doc-toc ol {
  list-style: decimal;
  padding-left: var(--space-5);
  margin: 0;
}
.doc-toc li {
  padding: var(--space-1) 0;
  font-size: var(--type-sm);
}
.doc-toc a {
  color: var(--color-text);
  text-decoration: none;
  font-weight: 500;
}
.doc-toc a:hover { color: var(--color-primary); }
.toc-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  margin-left: var(--space-2);
}

/* Sections */
.doc-section {
  margin-bottom: var(--space-10);
}
.section-heading {
  font-size: var(--type-xl);
  font-weight: 700;
  letter-spacing: -0.02em;
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--color-border);
  margin-bottom: var(--space-4);
}
.section-desc {
  color: var(--color-text-muted);
  font-size: var(--type-sm);
  margin-bottom: var(--space-6);
  line-height: 1.6;
}
.doc-empty {
  color: var(--color-text-muted);
  font-size: var(--type-sm);
  font-style: italic;
}

/* Entries */
.doc-entry {
  margin-bottom: var(--space-8);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--color-border);
}
.doc-entry:last-child { border-bottom: none; }
.entry-heading {
  font-size: var(--type-lg);
  font-weight: 600;
  letter-spacing: -0.01em;
  margin-bottom: var(--space-2);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
.entry-code {
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  font-weight: 700;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 3px;
  flex-shrink: 0;
}
.entry-meta {
  display: flex;
  gap: var(--space-1);
  flex-wrap: wrap;
  margin-bottom: var(--space-3);
}

/* Sessions */
.doc-session {
  margin-bottom: var(--space-8);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--color-border);
}
.doc-session:last-child { border-bottom: none; }
.session-heading { font-size: var(--type-lg); font-weight: 600; margin-bottom: var(--space-2); }
.session-focus { font-size: var(--type-sm); color: var(--color-text-muted); margin-bottom: var(--space-4); }

/* Questions */
.doc-question {
  padding: var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  margin-bottom: var(--space-3);
}
.q-heading { display: flex; align-items: flex-start; gap: var(--space-2); margin-bottom: var(--space-2); }
.q-label { font-weight: 700; color: var(--color-primary); flex-shrink: 0; }
.q-text { font-weight: 500; flex: 1; }
.q-status {
  font-size: var(--type-xs); padding: 0.1rem 0.4rem;
  border-radius: 3px; flex-shrink: 0;
  background: var(--color-surface-hover); color: var(--color-text-muted);
}
.q-status-answered { background: rgba(52, 211, 153, 0.1); color: var(--color-success); }
.q-status-pending { background: rgba(251, 191, 36, 0.1); color: var(--color-warning); }
.q-meta { font-size: var(--type-xs); color: var(--color-text-muted); margin-bottom: var(--space-2); }
.q-answer { margin-top: var(--space-2); }

/* Tasks */
.doc-task {
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  margin-bottom: var(--space-2);
}
.task-row { display: flex; align-items: center; gap: var(--space-2); }
.task-check {
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  flex-shrink: 0;
}
.task-done { color: var(--color-success); }
.task-title { flex: 1; font-weight: 500; font-size: var(--type-sm); }
.task-priority {
  font-size: var(--type-xs); padding: 0.1rem 0.35rem;
  background: rgba(239, 68, 68, 0.1); color: var(--color-error);
  border-radius: 3px;
}
.task-status {
  font-size: var(--type-xs); color: var(--color-text-muted);
}
.task-desc { font-size: var(--type-xs); color: var(--color-text-muted); margin-top: var(--space-1); }
.task-result { font-size: var(--type-sm); margin-top: var(--space-2); }

/* Responsive */
@media (max-width: 768px) {
  .export-toolbar { flex-direction: column; align-items: flex-start; gap: var(--space-3); }
  /* Once the row wraps, a 1.25rem gap no longer reads as grouping — the wrap
     does it instead. */
  .toolbar-actions { gap: var(--space-2); justify-content: flex-start; }
  .doc-title { font-size: var(--type-xl); }
  .doc-meta { flex-wrap: wrap; gap: var(--space-2); }
  .doc-toc { padding: var(--space-4); }
  .task-row { flex-wrap: wrap; }
}

/* Print */
.no-print { }
@media print {
  .no-print { display: none !important; }
  .export-document { max-width: none; }
  .doc-header { border-bottom-color: #ccc; }
  .doc-toc { background: #f8f8f8; border-color: #ddd; }
  .doc-section { page-break-before: auto; }
  /* Neither a section nor an entry may refuse to break: both are routinely
     longer than a sheet, and refusing left the first page empty and started the
     document on the second. Small units still stay whole — figure, diagram,
     table, checklist. */
  .doc-section, .doc-entry { page-break-inside: auto; }
  .doc-question { background: #f8f8f8; border-color: #ddd; }
  .doc-task { border-color: #ddd; }
  .entry-code { background: #e8f4f8; }
  .doc-tag { background: #f0f0f0; border-color: #ddd; }
}
</style>
