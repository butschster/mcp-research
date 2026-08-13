<template>
  <div class="block-doc">
    <template v-for="(b, i) in blocks" :key="i">
      <!-- paragraph -->
      <p v-if="b.type === 'paragraph'" class="b-paragraph" v-html="inline(b.data.text)"></p>

      <!-- heading -->
      <h2 v-else-if="b.type === 'heading' && b.data.level === 2" class="b-heading b-h2" v-html="inline(b.data.text)"></h2>
      <h3 v-else-if="b.type === 'heading' && b.data.level === 3" class="b-heading b-h3" v-html="inline(b.data.text)"></h3>
      <h4 v-else-if="b.type === 'heading'" class="b-heading b-h4" v-html="inline(b.data.text)"></h4>

      <!-- list -->
      <ol v-else-if="b.type === 'list' && b.data.style === 'ordered'" class="b-list">
        <li v-for="(it, j) in (b.data.items || [])" :key="j" v-html="inline(it)"></li>
      </ol>
      <ul v-else-if="b.type === 'list'" class="b-list">
        <li v-for="(it, j) in (b.data.items || [])" :key="j" v-html="inline(it)"></li>
      </ul>

      <!-- table -->
      <div v-else-if="b.type === 'table'" :class="['b-table-wrap', widthClass(b)]">
        <table class="b-table">
          <thead v-if="b.data.header && (b.data.rows || []).length">
            <tr>
              <th v-for="(cell, j) in b.data.rows[0]" :key="j" v-html="inline(cell)"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, j) in bodyRows(b)" :key="j">
              <td v-for="(cell, k) in row" :key="k" v-html="inline(cell)"></td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- quote -->
      <blockquote v-else-if="b.type === 'quote'" class="b-quote">
        <div v-html="inline(b.data.text)"></div>
        <cite v-if="b.data.cite">{{ b.data.cite }}</cite>
      </blockquote>

      <!-- code -->
      <pre v-else-if="b.type === 'code'" class="b-code"><code>{{ b.data.code }}</code></pre>

      <!-- callout -->
      <aside v-else-if="b.type === 'callout'" :class="['b-callout', `b-callout--${b.data.variant}`]">
        <p v-if="b.data.title" class="b-callout-title">{{ b.data.title }}</p>
        <div v-html="inline(b.data.text)"></div>
      </aside>

      <!-- divider -->
      <hr v-else-if="b.type === 'divider'" class="b-divider" />

      <!-- image -->
      <figure v-else-if="b.type === 'image'" :class="['b-figure', widthClass(b)]">
        <img :src="b.data.url" :alt="b.data.alt || ''" loading="lazy" />
        <figcaption v-if="b.data.caption" v-html="inline(b.data.caption)"></figcaption>
      </figure>

      <!-- html: a self-contained document, isolated in a sandboxed frame -->
      <figure v-else-if="b.type === 'html'" :class="['b-figure', widthClass(b)]">
        <figcaption v-if="b.data.title" class="b-html-title">{{ b.data.title }}</figcaption>
        <EntryArtifactFrame
          :html="b.data.html"
          :title="b.data.title || 'Artifact'"
          :bridge-data="bridgeData"
        />
        <figcaption v-if="b.data.caption" v-html="inline(b.data.caption)"></figcaption>
      </figure>
    </template>
  </div>
</template>

<script setup lang="ts">
import { renderInline } from '~/composables/useInlineMarkdown'

interface Block {
  type: string
  data: Record<string, any>
}

const props = withDefaults(
  defineProps<{
    /** Normalized blocks, in render order. */
    blocks?: Block[]
    /** Research short code, so [[E3]] links resolve to the right research. */
    researchSlug?: string
    /** Read-only context handed to html blocks over postMessage. */
    bridgeData?: Record<string, unknown> | null
  }>(),
  { blocks: () => [], researchSlug: '', bridgeData: null }
)

function inline(text: string): string {
  return renderInline(text, props.researchSlug)
}

function widthClass(b: Block): string {
  return b.data?.width === 'wide' ? 'is-wide' : 'is-text'
}

// With a header row the first row is the header; without one every row is body.
function bodyRows(b: Block): any[] {
  const rows = b.data?.rows || []
  return b.data?.header ? rows.slice(1) : rows
}
</script>

<style scoped>
.block-doc { display: flex; flex-direction: column; }

/* Reading rhythm: the document sets its own vertical spacing so blocks compose
   predictably regardless of the order an author picks. */
.block-doc > * + * { margin-top: var(--space-4); }
.block-doc > .b-heading { margin-top: var(--space-6); }
.block-doc > .b-heading:first-child { margin-top: 0; }

.b-paragraph { line-height: var(--line-base); }
.b-paragraph :deep(code),
.b-list :deep(code),
.b-quote :deep(code) {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.9em;
  background: var(--color-primary-muted);
  color: var(--color-primary);
  padding: 0.1rem 0.3rem;
  border-radius: var(--radius-sm);
}

.b-heading { font-weight: 650; letter-spacing: -0.015em; line-height: var(--line-tight); }
.b-h2 { font-size: var(--type-xl); }
.b-h3 { font-size: var(--type-lg); }
.b-h4 { font-size: var(--type-base); }

.b-list { padding-left: 1.35rem; }
.b-list li + li { margin-top: var(--space-1); }

.b-table-wrap { overflow-x: auto; }
.b-table { border-collapse: collapse; width: 100%; font-size: var(--type-sm); }
.b-table th,
.b-table td {
  border: 1px solid var(--color-border);
  padding: 0.45rem 0.65rem;
  text-align: left;
  vertical-align: top;
}
.b-table th { background: var(--color-surface-hover); font-weight: 600; }

.b-quote {
  border-left: 2px solid var(--color-primary);
  padding-left: var(--space-4);
  color: var(--color-text-muted);
}
.b-quote cite { display: block; margin-top: var(--space-2); font-size: var(--type-xs); font-style: normal; }

.b-code {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--space-4);
  overflow-x: auto;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: var(--type-xs);
  line-height: 1.55;
}

.b-callout {
  border: 1px solid var(--color-border);
  border-left-width: 3px;
  border-radius: var(--radius);
  padding: var(--space-3) var(--space-4);
  font-size: var(--type-sm);
}
.b-callout-title { font-weight: 650; margin-bottom: var(--space-1); }
.b-callout--info { border-left-color: var(--color-info); }
.b-callout--warning { border-left-color: var(--color-warning); }
.b-callout--success { border-left-color: var(--color-success); }
.b-callout--danger { border-left-color: var(--color-error); }

.b-divider { border: none; border-top: 1px solid var(--color-border); margin: var(--space-6) 0; }

.b-figure { margin: 0; }
.b-figure img { max-width: 100%; height: auto; display: block; border-radius: var(--radius); }
.b-figure figcaption {
  margin-top: var(--space-2);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.b-html-title {
  margin: 0 0 var(--space-2);
  font-size: var(--type-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-muted);
}

/* `wide` breaks out of the reading column. The negative margin is bounded so a
   narrow viewport degrades to the column width instead of overflowing. */
.is-wide { width: 100%; }
@media (min-width: 900px) {
  .is-wide { width: calc(100% + 2 * var(--space-8)); margin-left: calc(-1 * var(--space-8)); }
}

@media print {
  .b-code, .b-table th { background: none; }
  .b-table th, .b-table td, .b-callout, .b-divider { border-color: #ddd; }
  .b-figure figcaption, .b-quote, .b-html-title { color: #555; }
  .is-wide { width: 100%; margin-left: 0; }
}
</style>
