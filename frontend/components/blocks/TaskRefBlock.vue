<template>
  <section class="b-task-ref" :aria-busy="loading || undefined">
    <p v-if="note" class="b-task-note" v-html="inline(note)"></p>
    <p v-if="truncated" class="b-task-cut" role="status">
      {{ truncated }} further {{ truncated === 1 ? 'reference was' : 'references were' }} not stored —
      this list is past the cap.
    </p>

    <!-- loading: one skeleton per reference, so the document does not jump when
         the rows arrive. -->
    <template v-if="loading">
      <span class="sr-only">Loading tasks…</span>
      <!-- Blank rather than a skeleton: it is a line of small muted text, and a
           shimmering bar where a count will be reads as content. Reserved all
           the same, or the rows below it are pushed down when it appears. -->
      <div v-if="showProgress" class="b-task-progress-hold" aria-hidden="true"></div>
      <ul class="b-task-rows" role="list">
        <li v-for="ref in refs" :key="ref" class="b-task-skeleton skeleton-card"></li>
      </ul>
    </template>

    <template v-else-if="status === 'ready'">
      <!-- The count is read aloud; the bar is not. It says the same thing in a
           shape a screen reader cannot use, and ProgressBar has two root nodes,
           so the hiding has to go on a wrapper rather than fall through. -->
      <div v-if="showProgress && rows.length" class="b-task-progress">
        <span class="b-task-count">{{ doneCount }} of {{ rows.length }} done</span>
        <span class="b-task-bar" aria-hidden="true">
          <ProgressBar :value="doneCount" :total="rows.length" tone="done" />
        </span>
      </div>

      <ul v-if="rows.length" class="b-task-rows" role="list" data-no-mark>
        <li v-for="row in rows" :key="row.id" :class="['b-task-row', { 'is-done': row.done }]">
          <!-- A label around the INPUT ALONE. The row's title is deliberately
               outside it (a label swallows clicks on a [[E3]] link inside the
               title), but without one the 26px box was decoration and only the
               native 13px square answered a tap. `aria-labelledby` still names
               the control; an empty label does not compete with it. -->
          <label class="b-task-check">
            <input
              type="checkbox"
              :checked="row.done"
              :disabled="!tickable"
              :aria-busy="pending.has(row.id) || undefined"
              :aria-labelledby="labelIds(row)"
              @change="toggle(row, $event)"
            />
          </label>
          <NuxtLink class="short-code b-task-code" :to="taskHref(row.code)">
            {{ row.code }}
          </NuxtLink>
          <span class="b-task-body">
            <span :id="`${uid}-title-${row.id}`" class="b-task-title" v-html="inline(row.title)"></span>
            <span :id="`${uid}-state-${row.id}`" class="b-task-badges">
              <StatusBadge v-if="badgeStatus(row)" :status="badgeStatus(row)!" />
              <StatusBadge v-if="row.priority === 'high'" status="high" />
            </span>
          </span>
          <p v-if="failed[row.id]" class="b-task-error" role="status">{{ failed[row.id] }}</p>
        </li>
      </ul>

      <p v-else class="b-task-notice">
        No tasks left here — {{ refList }} {{ refs.length === 1 ? 'is not' : 'are not' }} in this project any more.
        <NuxtLink v-if="canWrite && !readonly" class="b-task-notice-action" :to="boardHref">Open the task board</NuxtLink>
      </p>

      <p v-if="rows.length && unresolved.length && canWrite && !readonly" class="b-task-footnote">
        {{ unresolved.length }} {{ unresolved.length === 1 ? 'reference does' : 'references do' }} not resolve —
        {{ unresolved.join(', ') }} {{ unresolved.length === 1 ? 'is not' : 'are not' }} in this project.
      </p>
    </template>

    <template v-else>
      <!-- error and excluded both show what the block points at, so the reader
           still learns what work this document names. -->
      <ul class="b-task-rows b-task-rows--inert" role="list">
        <li v-for="ref in refs" :key="ref" class="b-task-inert"><ShortCode :code="ref" /></li>
      </ul>
      <p class="b-task-notice">
        <template v-if="status === 'excluded'">
          Tasks aren't included in this link — ask whoever shared it if you need them.
        </template>
        <template v-else>
          Couldn't load these tasks.
          <button type="button" class="btn btn-sm b-task-retry" @click="$emit('retry')">Try again</button>
        </template>
      </p>
    </template>
  </section>
</template>

<script setup lang="ts">
/**
 * Existing tasks, projected into a document as a checklist.
 *
 * The block holds references and nothing else. A tick here is a status change on
 * the task — so the board, `task_list` over MCP and this document can never
 * disagree, because there is only one place the state lives. Contrast
 * `checklist`, whose state genuinely belongs to the document.
 *
 * A component rather than a branch in BlockRenderer: it owns three keyed state
 * maps and makes its own network calls, which is the line this repository draws.
 *
 * Two rules that look like bugs and are not:
 *
 *   - **Every status is tickable**, including `blocked` and `failed`. A checkbox
 *     that looks live and refuses the click is worse than one that does the
 *     thing; the untick below makes it reversible.
 *   - **No confirmation dialog**, though the board opens one on every drag. The
 *     modal exists there because a drag is easy to do by accident and because
 *     the board wants an optional comment. A checkbox in a document is the
 *     explicit gesture, and "mark it done and move on" is the whole block.
 */
import { renderInline } from '~/composables/useInlineMarkdown'
import { taskPath, tasksPath } from '~/composables/useResearchPaths'
import { useOptimisticWrite } from '~/composables/useOptimisticWrite'

interface TaskLike {
  id: string
  code: string
  title: string
  status: string
  priority?: string
}

interface Row extends TaskLike {
  done: boolean
}

const props = withDefaults(
  defineProps<{
    /** The block's stable id; data-block-id lands on the root via fallthrough. */
    blockId?: string
    /** References as the document stores them — codes or uuids. */
    refs?: string[]
    note?: string
    showProgress?: boolean
    /** The research's tasks, fetched and owned by the page. */
    tasks?: TaskLike[]
    status?: 'idle' | 'loading' | 'ready' | 'error' | 'excluded'
    researchSlug?: string
    /** The caller's own context — an export preview — not a permission. */
    readonly?: boolean
    /** References the server had to drop. */
    truncated?: number
  }>(),
  {
    blockId: '', refs: () => [], note: '', showProgress: true,
    tasks: () => [], status: 'idle', researchSlug: '', readonly: false, truncated: 0,
  }
)

defineEmits<{ retry: [] }>()

const { canWrite } = useResearchRole()
const { authFetch } = useAuth()
const config = useRuntimeConfig()

const uid = useId()

function inline(text: string): string {
  return renderInline(text, props.researchSlug)
}

const loading = computed(() => props.status === 'idle' || props.status === 'loading')

/**
 * Statuses written locally before the server confirmed them, keyed by task id.
 * A checkbox that waits for a round trip before it moves feels broken. Shared
 * with the checklist branch of BlockRenderer, so one refusal reads the same in
 * both halves of one document.
 */
const { optimistic, pending, failed, write, reset } = useOptimisticWrite<string>()
/**
 * The status a row held immediately before it was ticked.
 *
 * Unticking sends this back rather than a flat `pending`: tick a `failed` task
 * by accident, untick it, and a flat `pending` would have erased the failure
 * with nothing said. Page lifetime only — an unremembered row falls back.
 */
const previous = ref<Record<string, string>>({})

const byRef = computed(() => {
  const m: Record<string, TaskLike> = {}
  for (const t of props.tasks || []) {
    if (t?.code) m[t.code.toUpperCase()] = t
    if (t?.id) m[t.id.toLowerCase()] = t
  }
  return m
})

function lookup(ref: string): TaskLike | undefined {
  return byRef.value[ref.toUpperCase()] ?? byRef.value[ref.toLowerCase()]
}

/**
 * Resolved rows, in the order the document wrote them.
 *
 * A reference that resolves to nothing is dropped rather than drawn as a ghost:
 * a deleted task is not an unfinished one, and a row that cannot be ticked is
 * worse than no row.
 */
const rows = computed<Row[]>(() =>
  (props.refs || []).flatMap((ref) => {
    const t = lookup(ref)
    if (!t) return []
    const status = optimistic.value[t.id] ?? t.status
    return [{ ...t, status, done: status === 'completed' }]
  })
)

const unresolved = computed(() => (props.refs || []).filter((r) => !lookup(r)))
const doneCount = computed(() => rows.value.filter((r) => r.done).length)
const refList = computed(() => (props.refs || []).join(', '))

const boardHref = computed(() => tasksPath(props.researchSlug))
function taskHref(code: string) {
  return taskPath(props.researchSlug, code)
}

// `completed` needs no badge — the tick says it — and `pending` is the default,
// so a badge on every row would be noise.
function badgeStatus(row: Row): string | null {
  return row.status === 'completed' || row.status === 'pending' ? null : row.status
}

// The checkbox is named rather than wrapped in a <label>: a label swallows
// clicks on a [[E3]] link inside the title and toggles the box instead of
// following it.
function labelIds(row: Row): string {
  return `${uid}-title-${row.id} ${uid}-state-${row.id}`
}

const tickable = computed(() => canWrite.value && !props.readonly)

// Statuses whose refusal means something particular here. The rest — 401, an
// unreachable server, an unexplained failure — read the same as everywhere else.
const TASK_ERRORS = {
  403: "Not saved — you can't change tasks in this project.",
  404: 'Not saved — that task no longer exists.',
}

function toggle(row: Row, event: Event) {
  const box = event.target as HTMLInputElement
  if (!tickable.value) return
  // The input is not disabled while in flight — disabling a focused control
  // blurs it — so the browser has already flipped the box by the time this
  // runs. Ignoring the second click without putting the box back left it
  // showing the opposite of the row it belongs to, and nothing re-rendered it.
  if (pending.value.has(row.id)) {
    box.checked = row.done
    return
  }
  const checked = box.checked

  const next = checked ? 'completed' : (previous.value[row.id] ?? 'pending')
  if (checked) previous.value = { ...previous.value, [row.id]: row.status }

  const base = config.public.apiBase || ''
  void write(row.id, next, () =>
    authFetch(`${base}/api/tasks/${row.id}`, { method: 'PUT', body: { status: next } }),
  TASK_ERRORS)
}

// A fresh list from the server is the truth again — except for a row still in
// flight, whose optimistic value is the only thing keeping the box where the
// reader put it. `reset` keeps exactly those.
watch(() => props.tasks, reset)
</script>

<style scoped>
.b-task-ref { display: flex; flex-direction: column; }

.b-task-note { margin: 0 0 var(--space-2); font-size: var(--type-sm); }
.b-task-cut {
  margin: 0 0 var(--space-2);
  font-size: var(--type-2xs);
  color: var(--color-warning);
}

.b-task-progress {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-2) var(--space-3);
  margin-bottom: var(--space-3);
}
.b-task-count {
  font-size: var(--type-2xs);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
.b-task-bar { display: flex; flex: 1; min-width: 8rem; }
/* ProgressBar has two root nodes, so it cannot take flex:1 by fallthrough. */
.b-task-bar :deep(.progress-wrap) { flex: 1; }

.b-task-rows { display: flex; flex-direction: column; gap: var(--space-1); list-style: none; margin: 0; padding: 0; }

.b-task-row {
  display: grid;
  grid-template-columns: var(--control-h-sm) auto 1fr;
  align-items: start;
  column-gap: var(--space-2);
  row-gap: var(--space-1);
  padding: 0.15rem 0;
  line-height: var(--line-base);
}

/* A bare 13px checkbox is not a hit target. */
.b-task-check {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: var(--control-h-sm);
  height: var(--control-h-sm);
}
.b-task-check input { cursor: pointer; }
.b-task-check input:disabled { cursor: default; }
/* Dimmed on the box alone: dimming the row puts the title under AA for as long
   as the request runs. */
.b-task-check input[aria-busy='true'] { opacity: 0.6; }

.b-task-code { align-self: center; }

.b-task-body {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: var(--space-2);
  min-width: 0;
}
/* Inherited, not --type-sm: a checklist and a task list can sit a paragraph
   apart in one document, and the checklist's items are body size. Two sizes for
   the same kind of line reads as one of them being an aside. */
.b-task-title { overflow-wrap: anywhere; }
.b-task-badges { display: inline-flex; flex-wrap: wrap; gap: var(--space-1); }
.is-done .b-task-title { color: var(--color-text-muted); text-decoration: line-through; }

/* Same sentence as the checklist's .b-check-error, so the same size and the
   same colour: one document can hold both, and two type sizes for "this did not
   save" reads as one of them being less true. */
.b-task-error {
  grid-column: 1 / -1;
  margin: 0;
  font-size: var(--type-xs);
  color: var(--color-warning);
}

/* The height of the row it stands in for, padding included — a skeleton that is
   short by five pixels a row moves the whole document when the rows arrive, which
   is the one thing a skeleton exists to prevent. */
.b-task-skeleton { height: calc(var(--control-h-sm) + 0.3rem); border-radius: var(--radius-sm); }
.b-task-progress-hold { height: var(--type-2xs); margin-bottom: var(--space-3); }

/* Codes with no rows behind them are chips, not rows: stacked one per line they
   take a screen of height to say four things the reader cannot act on. */
.b-task-rows--inert { flex-direction: row; flex-wrap: wrap; gap: var(--space-2); }
.b-task-inert { padding: 0.15rem 0; }

.b-task-notice,
.b-task-footnote {
  margin: var(--space-2) 0 0;
  font-size: var(--type-2xs);
  color: var(--color-text-muted);
}
.b-task-notice-action { margin-left: var(--space-1); }
.b-task-retry { margin-left: var(--space-2); vertical-align: middle; }

@media print {
  .b-task-ref { break-inside: avoid; }
  /* A 4px tint prints as a smudge; the count carries the same fact in words. */
  .b-task-bar { display: none; }
  .is-done .b-task-title { color: #111; text-decoration-color: #999; }
}
</style>
