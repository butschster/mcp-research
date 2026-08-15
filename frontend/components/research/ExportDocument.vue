<template>
  <!-- Document -->
  <article ref="docRoot" class="export-document">
    <!-- Cover -->
    <header class="doc-header">
      <h1 class="doc-title">{{ data.research.name }}</h1>
      <p v-if="data.research.goal" class="doc-goal">{{ data.research.goal }}</p>
      <p v-if="data.research.description" class="doc-description">{{ data.research.description }}</p>
      <div v-if="data.research.tags?.length" class="doc-tags">
        <span v-for="tag in data.research.tags" :key="tag" class="doc-tag">{{ tag }}</span>
      </div>
      <div class="doc-meta">
        <span>Status: {{ data.research.status }}</span>
        <span>Sections: {{ data.sections?.length || 0 }}</span>
        <span>Entries: {{ totalEntries }}</span>
        <span>Exported: {{ new Date().toLocaleDateString() }}</span>
      </div>
    </header>

    <!-- Table of contents -->
    <nav class="doc-toc">
      <h2>Table of Contents</h2>
      <ol>
        <li v-for="(section, i) in data.sections" :key="section.id">
          <a :href="`#section-${i}`">{{ section.display_name || section.name }}</a>
          <span class="toc-count">{{ section.entries?.length || 0 }} entries</span>
        </li>
        <li v-if="data.sessions?.length">
          <a href="#sessions">Sessions</a>
          <span class="toc-count">{{ data.sessions.length }}</span>
        </li>
        <li v-if="data.tasks?.length">
          <a href="#tasks">Tasks</a>
          <span class="toc-count">{{ data.tasks.length }}</span>
        </li>
      </ol>
    </nav>

    <!-- Sections -->
    <section
      v-for="(section, i) in data.sections"
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
          :blocks="blocksOf(entry) || []"
          :research-slug="researchSlug"
          readonly
        />
        <div v-else class="markdown-content" v-html="renderMarkdown(entryBody(entry))"></div>
      </article>
    </section>

    <!-- Sessions -->
    <section v-if="data.sessions?.length" id="sessions" class="doc-section">
      <h2 class="section-heading">Sessions</h2>

      <div v-for="sess in data.sessions" :key="sess.id" class="doc-session">
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
    <section v-if="data.tasks?.length" id="tasks" class="doc-section">
      <h2 class="section-heading">Tasks</h2>

      <div v-for="t in data.tasks" :key="t.id" class="doc-task">
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
</template>

<script setup lang="ts">
/**
 * The printable research document.
 *
 * Extracted from the export page so the shared view can render the same thing.
 * Two copies would have drifted the first time an entry type was added, and the
 * one that drifted would have been the one nobody was looking at — the copy a
 * client sees.
 *
 * It renders whatever it is handed and asks no questions about permission: the
 * server decides what is in the payload, and under a share it has already left
 * out `instruction`, `memory`, and any section the link does not include.
 */
import { parseMarkdown } from '~/composables/useSafeMarkdown'
import { linkRefs } from '~/composables/useCrossRefs'
import { renderMermaidBlocks } from '~/composables/useMermaid'
import { normalizeContent } from '~/utils/normalizeContent'

const props = defineProps<{
  data: any
  researchSlug: string
}>()

const totalEntries = computed(() =>
  (props.data?.sections ?? []).reduce((sum: number, s: any) => sum + (s.entries?.length || 0), 0),
)

// A blocks entry stores JSON; the server ships a markdown rendering beside it so
// the printable document does not have to know the block format.
function entryBody(entry: any): string {
  return entry?.content_markdown || entry?.content || ''
}

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

function renderMarkdown(content: string): string {
  if (!content) return ''
  const html = parseMarkdown(normalizeContent(content)) as string
  return linkRefs(html, props.researchSlug)
}

// Markdown entries carry their diagrams as ```mermaid fences. Draw them, or the
// printed page shows the source of a diagram instead of the diagram.
const docRoot = ref<HTMLElement | null>(null)
async function drawDiagrams() {
  await nextTick()
  if (docRoot.value) await renderMermaidBlocks(docRoot.value)
}
onMounted(drawDiagrams)
watch(() => props.data, drawDiagrams)
</script>

<style scoped>
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
