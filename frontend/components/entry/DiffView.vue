<template>
  <div class="diff-view">
    <div v-if="diff.truncated" class="diff-notice">
      <p class="diff-notice-title">Too large to align line by line.</p>
      <p class="diff-notice-body">{{ diff.removed }} lines replaced by {{ diff.added }}.</p>
      <slot name="truncated-action" />
    </div>

    <p v-else-if="!diff.lines?.length" class="diff-empty">Nothing to compare.</p>

    <ol v-else class="diff-body">
      <template v-for="(row, i) in rows" :key="i">
        <li v-if="row.gap" class="diff-gap">
          <button
            class="diff-gap-btn"
            :aria-expanded="isOpen(row.id)"
            @click="toggleGap(row.id)"
          >
            {{ isOpen(row.id) ? 'Hide' : 'Show' }} {{ row.hidden }} unchanged
            {{ row.hidden === 1 ? 'line' : 'lines' }}
          </button>
        </li>
        <li v-else :class="['diff-line', `diff-${row.op}`]">
          <span v-if="row.op !== 'equal'" class="sr-only">{{ row.op === 'add' ? 'Added: ' : 'Removed: ' }}</span>
          <span class="diff-marker" aria-hidden="true">{{ marker(row.op) }}</span>
          <span class="diff-text">
            <template v-if="row.words?.length">
              <template v-for="(w, wi) in row.words" :key="wi">
                <ins v-if="w.op === 'add'" class="diff-word-add">{{ w.text }}</ins>
                <del v-else-if="w.op === 'remove'" class="diff-word-remove">{{ w.text }}</del>
                <template v-else>{{ w.text }}</template>
              </template>
            </template>
            <template v-else>{{ row.text || ' ' }}</template>
          </span>
        </li>
      </template>
    </ol>
  </div>
</template>

<script setup lang="ts">
/**
 * Renders a DiffResult from the API.
 *
 * Three rules the markup exists to satisfy:
 *
 * - **A change is never carried by colour alone.** Each changed line gets an
 *   inset edge bar — solid for an addition, dashed for a removal — which
 *   survives every wrapped row, and a screen-reader-only "Added:" / "Removed:"
 *   prefix. Equal lines get no prefix: nobody wants "unchanged" read sixty
 *   times.
 * - **Gaps are individually expandable.** One flag for the whole document meant
 *   opening one gap opened all of them, with no way back.
 * - **Unchanged lines recede.** They are context, not content, so they drop to
 *   the muted colour and the changed lines keep the foreground. This is what
 *   gives the diff a figure and a ground.
 */
type DiffOp = 'equal' | 'add' | 'remove'
type DiffWord = { op: DiffOp; text: string }
type DiffLine = { op: DiffOp; text: string; words?: DiffWord[] }

const props = withDefaults(defineProps<{
  diff: { lines?: DiffLine[]; truncated?: boolean; added?: number; removed?: number }
  context?: number
  /** Drives every gap open from outside — the pane header's "Expand all". */
  expandAll?: boolean
}>(), { context: 3, expandAll: false })

const openGaps = ref<Set<number>>(new Set())

// A new comparison is a new document: gaps opened in the previous one must not
// stay open in this one.
watch(() => props.diff, () => { openGaps.value = new Set() })
function isOpen(id: number): boolean {
  return props.expandAll || openGaps.value.has(id)
}

function toggleGap(id: number) {
  const next = new Set(openGaps.value)
  next.has(id) ? next.delete(id) : next.add(id)
  openGaps.value = next
}

type Row =
  | (DiffLine & { gap?: false; id?: never; hidden?: never })
  | { gap: true; id: number; hidden: number; op?: never; text?: never; words?: never }

const rows = computed<Row[]>(() => {
  const lines = (props.diff.lines ?? []).map(collapseWordRuns)

  const keep = new Array(lines.length).fill(false)
  lines.forEach((l, i) => {
    if (l.op === 'equal') return
    for (let j = i - props.context; j <= i + props.context; j++) {
      if (j >= 0 && j < lines.length) keep[j] = true
    }
  })
  // A comparison with no changes at all: show it as it is rather than as one
  // enormous gap the reader has to open to see nothing happened.
  if (!keep.some(Boolean)) return lines as Row[]

  const out: Row[] = []
  let hidden: DiffLine[] = []
  let gapId = 0

  const flush = () => {
    if (!hidden.length) return
    const id = gapId++
    if (isOpen(id)) out.push(...(hidden as Row[]))
    else out.push({ gap: true, id, hidden: hidden.length })
    hidden = []
  }

  lines.forEach((l, i) => {
    if (keep[i]) {
      flush()
      out.push(l as Row)
    } else {
      hidden.push(l)
    }
  })
  flush()
  return out
})

/**
 * Merges a run of consecutive changed words into one element and pushes the
 * run's trailing space outside it.
 *
 * The server's tokenizer keeps the trailing space inside each token, so two
 * adjacent changed words rendered as two elements with struck-through gaps
 * between them. In Russian a case change alters several adjacent tokens at
 * once, which made that the normal case rather than the edge one.
 */
function collapseWordRuns(line: DiffLine): DiffLine {
  if (!line.words?.length) return line
  const out: DiffWord[] = []
  for (const word of line.words) {
    const prev = out[out.length - 1]
    if (prev && prev.op === word.op && word.op !== 'equal') {
      prev.text += word.text
    } else {
      out.push({ ...word })
    }
  }
  return {
    ...line,
    words: out.flatMap((w) => {
      if (w.op === 'equal') return [w]
      const trimmed = w.text.replace(/\s+$/, '')
      const trailing = w.text.slice(trimmed.length)
      return trailing ? [{ ...w, text: trimmed }, { op: 'equal' as DiffOp, text: trailing }] : [w]
    }),
  }
}

function marker(op: string): string {
  if (op === 'add') return '+'
  if (op === 'remove') return '−'
  return ' '
}
</script>

<style scoped>
.diff-view {
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  line-height: 1.6;
}
.diff-notice {
  padding: var(--space-5);
  text-align: center;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
}
.diff-notice-title {
  margin: 0;
  font-size: var(--type-sm);
  font-weight: 500;
  color: var(--color-text);
}
.diff-notice-body {
  margin: var(--space-1) 0 var(--space-3);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.diff-empty {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  padding: var(--space-4);
  text-align: center;
}
.diff-body {
  list-style: none;
  margin: 0;
  padding: var(--space-2) 0;
  max-width: 90ch;
  /* Page ground rather than surface: the diff reads as a well sunk into the
     panel, and the 12% row tints sit on a darker base, which is what makes them
     visible at all. */
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
}
.diff-line {
  display: flex;
  gap: var(--space-2);
  padding: 0 var(--space-3);
  border-left: 2px solid transparent;
  white-space: pre-wrap;
  /* Only unbreakable strings break mid-token; ordinary prose keeps its words. */
  overflow-wrap: anywhere;
  color: var(--color-text-muted);
}
.diff-marker {
  flex-shrink: 0;
  width: 1ch;
  user-select: none;
}
.diff-text { flex: 1; min-width: 0; }

/* The edge bar is what survives wrapping, a bad monitor and colour blindness:
   the shape differs (solid vs dashed), not only the hue. */
.diff-add {
  color: var(--color-text);
  background: color-mix(in srgb, var(--color-success) 12%, transparent);
  border-left-color: var(--color-success);
}
.diff-add .diff-marker { color: var(--color-success); }
.diff-remove {
  color: var(--color-text);
  background: color-mix(in srgb, var(--color-error) 12%, transparent);
  border-left: 2px dashed var(--color-error);
}
.diff-remove .diff-marker { color: var(--color-error); }

.diff-word-add {
  background: color-mix(in srgb, var(--color-success) 30%, transparent);
  border-radius: 2px;
  text-decoration: underline;
  text-underline-offset: 2px;
}
.diff-word-remove {
  background: color-mix(in srgb, var(--color-error) 30%, transparent);
  border-radius: 2px;
  text-decoration: line-through;
}

.diff-gap {
  border-block: 1px solid var(--color-border);
  background: var(--color-surface-hover);
}
.diff-gap + .diff-line,
.diff-line + .diff-gap { margin-top: 1px; }
.diff-gap-btn {
  display: block;
  width: 100%;
  background: none;
  border: none;
  color: var(--color-text-muted);
  font-family: inherit;
  font-size: var(--type-xs);
  padding: var(--space-2) var(--space-3);
  text-align: left;
  cursor: pointer;
}
.diff-gap-btn:hover { color: var(--color-primary); }
</style>
