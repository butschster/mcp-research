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
      <div v-else-if="b.type === 'table'" class="b-table-wrap">
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
      <div v-else-if="b.type === 'code'" class="b-code-wrap">
        <span v-if="b.data.language" class="b-code-lang">{{ b.data.language }}</span>
        <pre class="b-code"><code>{{ b.data.code }}</code></pre>
      </div>

      <!-- callout -->
      <aside v-else-if="b.type === 'callout'" :class="['b-callout', `b-callout--${b.data.variant}`]">
        <p v-if="b.data.title" class="b-callout-title">{{ b.data.title }}</p>
        <div v-html="inline(b.data.text)"></div>
      </aside>

      <!-- divider -->
      <hr v-else-if="b.type === 'divider'" class="b-divider" />

      <!-- image -->
      <figure v-else-if="b.type === 'image'" class="b-figure">
        <img :src="b.data.url" :alt="b.data.alt || ''" loading="lazy" />
        <figcaption v-if="b.data.caption" v-html="inline(b.data.caption)"></figcaption>
      </figure>

      <!-- html: a self-contained document, isolated in a sandboxed frame -->
      <figure v-else-if="b.type === 'html'" class="b-figure">
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
/* Only h2 gets a rule: it separates the document's top-level parts, and putting
   one on every level turns the page into a stack of boxes. */
.b-h2 {
  font-size: var(--type-xl);
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--color-border);
}
.b-h3 { font-size: var(--type-lg); }
.b-h4 { font-size: var(--type-base); color: var(--color-text-muted); }

.b-list { padding-left: 1.35rem; }
.b-list li + li { margin-top: var(--space-1); }

/* The table is a bordered card so a wide one reads as a contained object and its
   horizontal scroll is obviously its own, not the page's. */
.b-table-wrap {
  overflow-x: auto;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-surface);
}
.b-table { border-collapse: collapse; width: 100%; font-size: var(--type-sm); }
.b-table th,
.b-table td {
  padding: 0.5rem 0.75rem;
  text-align: left;
  vertical-align: top;
  border-bottom: 1px solid var(--color-border);
}
.b-table tr:last-child td { border-bottom: none; }
.b-table th {
  font-size: var(--type-xs);
  font-weight: 650;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  background: rgba(0, 0, 0, 0.2);
  white-space: nowrap;
}
/* Zebra rows: cheaper to scan than gridlines on every cell. */
.b-table tbody tr:nth-child(even) { background: rgba(255, 255, 255, 0.015); }

.b-quote {
  border-left: 2px solid var(--color-primary);
  padding: var(--space-1) 0 var(--space-1) var(--space-4);
  font-size: var(--type-base);
  font-style: italic;
  color: var(--color-text);
}
.b-quote cite {
  display: block;
  margin-top: var(--space-2);
  font-size: var(--type-xs);
  font-style: normal;
  color: var(--color-text-muted);
}
.b-quote cite::before { content: '— '; }

/* Code sits BELOW the page surface rather than above it: a terminal reads as a
   recess, and --color-surface (lighter than the page) made it float. There is no
   darker token, so the recess is a black overlay that composites over whatever
   the page background is. */
.b-code-wrap {
  position: relative;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: rgba(0, 0, 0, 0.35);
  overflow: hidden;
}
.b-code-lang {
  position: absolute;
  top: 0;
  right: 0;
  padding: 0.15rem 0.5rem;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: var(--type-xs);
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
  background: rgba(0, 0, 0, 0.35);
  border-bottom-left-radius: var(--radius-sm);
  pointer-events: none;
}
.b-code {
  margin: 0;
  padding: var(--space-4);
  overflow-x: auto;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: var(--type-xs);
  line-height: 1.6;
  color: var(--color-text);
  background: none;
  border: none;
}

/* A callout is tinted, not just bordered: at a glance the colour should carry the
   severity without reading the title. */
.b-callout {
  border: 1px solid var(--color-border);
  border-left-width: 3px;
  border-radius: var(--radius);
  padding: var(--space-3) var(--space-4);
  font-size: var(--type-sm);
  line-height: var(--line-base);
}
.b-callout-title {
  font-weight: 650;
  margin-bottom: var(--space-1);
  font-size: var(--type-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}
.b-callout--info {
  border-left-color: var(--color-info);
  background: rgba(107, 157, 240, 0.07);
}
.b-callout--info .b-callout-title { color: var(--color-info); }
.b-callout--warning {
  border-left-color: var(--color-warning);
  background: rgba(240, 184, 73, 0.07);
}
.b-callout--warning .b-callout-title { color: var(--color-warning); }
.b-callout--success {
  border-left-color: var(--color-success);
  background: rgba(52, 211, 153, 0.07);
}
.b-callout--success .b-callout-title { color: var(--color-success); }
.b-callout--danger {
  border-left-color: var(--color-error);
  background: rgba(239, 107, 107, 0.07);
}
.b-callout--danger .b-callout-title { color: var(--color-error); }

/* A hairline that fades at both ends reads as a pause rather than a cut. */
.b-divider {
  border: none;
  height: 1px;
  margin: var(--space-6) 0;
  background: linear-gradient(
    90deg,
    transparent,
    var(--color-border-strong) 20%,
    var(--color-border-strong) 80%,
    transparent
  );
}

/* No `margin: 0` here: main.css already zeroes the figure's UA margins, and
   repeating it would beat `.block-doc > * + *` on source order at equal
   specificity — which is exactly how the figure lost its spacing from the block
   above it. */
.b-figure img {
  max-width: 100%;
  height: auto;
  display: block;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}
.b-figure figcaption {
  margin-top: var(--space-2);
  font-size: var(--type-xs);
  line-height: 1.5;
  color: var(--color-text-muted);
}
/* The html block's own name, above the frame, styled as a label rather than a
   heading so it does not compete with the document's headings. */
.b-html-title {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  margin: 0 0 var(--space-2);
  padding: 0.15rem 0.5rem;
  font-size: var(--type-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  border-radius: var(--radius-sm);
}

@media print {
  /* The dark tints are invisible on paper and the recessed code block turns into
     a black slab, so everything falls back to hairlines on white. */
  .b-code-wrap,
  .b-code-lang,
  .b-table-wrap,
  .b-table th,
  .b-table tbody tr:nth-child(even),
  .b-callout { background: none; }
  .b-table-wrap,
  .b-code-wrap,
  .b-table th,
  .b-table td,
  .b-callout,
  .b-figure img { border-color: #ddd; }
  .b-divider { background: none; border-top: 1px solid #ddd; }
  .b-h2 { border-bottom-color: #ddd; }
  .b-figure figcaption,
  .b-quote cite,
  .b-code-lang,
  .b-table th { color: #555; }
  .b-html-title { color: #333; border: 1px solid #ddd; }
  .b-callout-title { color: #333 !important; }
}
</style>
