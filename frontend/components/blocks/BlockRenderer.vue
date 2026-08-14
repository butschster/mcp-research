<template>
  <div ref="root" class="block-doc">
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
      <!-- A mermaid source is a diagram, not code: the viewer is mounted into
           the container from script, so Vue never patches what is inside it. -->
      <figure v-else-if="isMermaid(b)" class="b-figure">
        <div class="b-mermaid" :data-mermaid="i"></div>
        <figcaption v-if="b.data.caption" v-html="inline(b.data.caption)"></figcaption>
      </figure>

      <div v-else-if="b.type === 'code'" class="b-code-wrap">
        <span v-if="b.data.language" class="b-code-lang">{{ b.data.language }}</span>
        <pre class="b-code"><code>{{ b.data.code }}</code></pre>
      </div>

      <!-- checklist: the one block a reader writes to -->
      <div v-else-if="b.type === 'checklist'" class="b-checklist">
        <p v-if="b.data.title" class="b-checklist-title">{{ b.data.title }}</p>
        <template v-for="item in checklistItems(b)" :key="item.key">
          <label :class="['b-check', { 'is-done': item.checked, 'is-busy': pending.has(item.token) }]">
            <input
              type="checkbox"
              :checked="item.checked"
              :disabled="!tickable(b)"
              :aria-busy="pending.has(item.token)"
              @change="toggle(b, item, ($event.target as HTMLInputElement).checked)"
            />
            <span v-html="inline(item.text)"></span>
          </label>
          <!-- Beside the item that failed, not under the whole block: with two
               checklists in a document a shared line said nothing about which
               box did not save. -->
          <p v-if="failed[item.token]" class="b-check-error" role="status">{{ failed[item.token] }}</p>
        </template>
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
import { createMermaidFallback, createMermaidViewer } from '~/composables/useMermaidViewer'

const root = ref<HTMLElement | null>(null)

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
    /** Entry the blocks belong to. Required for a checklist to be tickable. */
    entryId?: string
    /** A viewer who may read but not write gets checkboxes that show state and
     *  do nothing. The server is the control; this is the courtesy. */
    readonly?: boolean
  }>(),
  { blocks: () => [], researchSlug: '', bridgeData: null, entryId: '', readonly: false }
)

// The reader's role in the research on screen. `readonly` is the caller's own
// choice of context — an export preview, say — and this is the permission; a
// tick needs both.
const { canWrite } = useResearchRole()

function inline(text: string): string {
  return renderInline(text, props.researchSlug)
}


// Both spellings draw: the dedicated block type, and a code block that declares
// mermaid — agents reach for the latter, and it worked before the type existed.
function isMermaid(b: Block): boolean {
  const declared = b.type === 'mermaid' || (b.type === 'code' && b.data?.language === 'mermaid')
  return declared && !!b.data?.code
}

// What each container currently holds, so a re-render only redraws diagrams
// whose source actually changed. Keyed by the element, not by index: the
// element goes away with the block, and so does its entry.
const drawn = new WeakMap<HTMLElement, string>()

// Rendering a diagram takes long enough for a second update to arrive mid-flight
// (a save, then the WebSocket echo of it). Only the newest run may paint, and it
// records what it drew *after* drawing — otherwise a slow first run could land
// on top of a fast second one and the guard would call it current.
let drawRun = 0

async function drawDiagrams() {
  const run = ++drawRun
  await nextTick()
  const host = root.value
  if (!host) return
  for (const el of host.querySelectorAll<HTMLElement>('.b-mermaid')) {
    const source = props.blocks[Number(el.dataset.mermaid)]?.data?.code || ''
    if (!source || drawn.get(el) === source) continue
    const node = (await createMermaidViewer(source)) ?? (await createMermaidFallback(source))
    if (run !== drawRun) return
    drawn.set(el, source)
    el.replaceChildren(node)
  }
}

onMounted(drawDiagrams)
watch(() => props.blocks, drawDiagrams, { deep: true })


const { authFetch } = useAuth()
const config = useRuntimeConfig()

// Ticks applied locally before the server has confirmed them, keyed by
// block+item. A checkbox that waits for a round trip before moving feels broken,
// and a checklist is ticked in bursts.
const optimistic = ref<Record<string, boolean>>({})
const pending = ref<Set<string>>(new Set())
// Message per item, not a flag: the server says what went wrong — a stale
// document, an item that no longer exists, an expired session — and "reload the
// page" was wrong for most of them.
const failed = ref<Record<string, string>>({})

interface CheckItem {
  key: string
  token: string
  text: string
  checked: boolean
}

function checklistItems(b: Block): CheckItem[] {
  const state = (b.data?.state || {}) as Record<string, boolean>
  return ((b.data?.items || []) as any[]).map((it) => {
    const key = it?.key || ''
    const token = `${b.id || ''}:${key}`
    return {
      key,
      token,
      text: it?.text || '',
      checked: token in optimistic.value ? optimistic.value[token]! : !!state[key],
    }
  })
}

// A checkbox that cannot be saved must be disabled rather than guarded in the
// handler: the browser flips the box before any handler runs, and returning
// early left it flipped, saying "done" about something nobody recorded.
function tickable(b: Block): boolean {
  // A viewer's tick would be a write — the box posts a patch to the entry —
  // so the role has to reach this far down into rendered content. This is the
  // edit affordance most easily missed: it does not look like a button.
  return canWrite.value && !props.readonly && !!props.entryId && !!b.id
}

async function toggle(b: Block, item: CheckItem, checked: boolean) {
  if (!tickable(b)) return
  // Not disabled while in flight: disabling a focused control blurs it, and a
  // keyboard reader was thrown back to the top of the page on every tick.
  if (pending.value.has(item.token)) return

  optimistic.value = { ...optimistic.value, [item.token]: checked }
  pending.value = new Set(pending.value).add(item.token)
  const cleared = { ...failed.value }
  delete cleared[item.token]
  failed.value = cleared

  const base = config.public.apiBase || ''
  try {
    await authFetch(`${base}/api/entries/${props.entryId}/patch`, {
      method: 'POST',
      body: {
        ops: [{ op: 'set_state', id: b.id, item: item.key, checked }],
      },
    })
  } catch (e: any) {
    // Revert visibly. A silent revert is worse than an error: the reader
    // re-ticks and assumes it stuck.
    const next = { ...optimistic.value }
    delete next[item.token]
    optimistic.value = next
    failed.value = { ...failed.value, [item.token]: saveError(e) }
  } finally {
    const p = new Set(pending.value)
    p.delete(item.token)
    pending.value = p
  }
}

function saveError(e: any): string {
  const status = e?.status ?? e?.response?.status
  const detail = e?.data?.error ?? e?.response?._data?.error
  if (status === 409) return 'Not saved — the document changed. Reload the page.'
  if (status === 401 || status === 403) return 'Not saved — your session expired. Sign in again.'
  if (!status) return 'Not saved — the server could not be reached.'
  return detail ? `Not saved — ${detail}` : 'Not saved.'
}

// A fresh document from the server is the truth again.
watch(() => props.blocks, () => {
  optimistic.value = {}
  failed.value = {}
})

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

/* The viewer is built outside the component and carries no scoped attribute,
   hence :deep. Its own vertical margin is dropped so document rhythm stays with
   `.block-doc > * + *`, the way every other block spaces itself. */
.b-mermaid :deep(.mermaid-diagram),
.b-mermaid :deep(.mermaid-broken) { margin: 0; }

.b-checklist { display: flex; flex-direction: column; gap: var(--space-1); }
.b-checklist-title { font-weight: 600; margin-bottom: var(--space-1); }
.b-check {
  display: flex;
  align-items: flex-start;
  gap: var(--space-2);
  padding: 0.15rem 0;
  line-height: var(--line-base);
  cursor: pointer;
}
.b-check input { margin-top: 0.35rem; accent-color: var(--color-primary); cursor: pointer; }
.b-check.is-done span { color: var(--color-text-muted); text-decoration: line-through; }
.b-check.is-busy { opacity: 0.6; }
.b-check input:disabled { cursor: default; }
.b-check-error {
  margin-left: 1.6rem;
  font-size: var(--type-xs);
  color: var(--color-warning);
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
  .b-mermaid :deep(.mermaid-diagram) { background: none; border: none; }
  /* A wide table scrolls on screen; on paper the columns past the margin were
     simply absent, with the scrollbar track printed where they should be. */
  .b-table-wrap { overflow: visible; }
  /* Neither a diagram nor a sandboxed visual can be split across a sheet: one
     used to leave an empty bordered rectangle on the page before it. */
  .b-figure,
  .b-mermaid,
  .b-checklist { break-inside: avoid; }
  /* The record of what was DONE was the palest ink on the page: muted grey on a
     grey box, while the unfinished items printed solid black. */
  .b-check.is-done span { color: #111; text-decoration-color: #999; }
  /* These two set their own light colour for the dark UI, so on white they came
     out as pale grey on pale grey — the code block was the worst of it. */
  .b-code,
  .b-code code { color: #111; }
  .b-quote,
  .b-quote :deep(*) { color: #222; }
  .b-check input { filter: none; }
}
</style>
