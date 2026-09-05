<template>
  <ModalOverlay :labelledby="titleId" :visible="visible" size="xl" @close="$emit('close')">
    <header class="history-head">
      <div class="history-title-row">
        <span class="history-eyebrow">History</span>
        <ShortCode v-if="entryCode" :code="entryCode" />
        <h2 :id="titleId" class="history-title">{{ entryTitle }}</h2>
      </div>
      <button class="btn btn-sm" @click="$emit('close')">Close</button>
    </header>

    <div class="history-body">
      <!-- Rail -->
      <div class="rev-rail">
        <div class="rail-head">
          <span>{{ railSummary }}</span>
          <span v-if="revisions.length > 1" class="rail-hint">Shift-click to compare</span>
        </div>

        <div v-if="loading" class="rail-skeletons">
          <div v-for="i in 4" :key="i" class="skeleton-card rail-skeleton"></div>
        </div>

        <EmptyState
          v-else-if="error"
          icon="&#x26A0;"
          title="Could not load the history"
          description="The revision list did not come back. The document itself is untouched."
        >
          <button class="btn btn-sm" @click="load">Try again</button>
        </EmptyState>

        <EmptyState
          v-else-if="!revisions.length"
          icon="&#x1F553;"
          title="No history yet"
          description="This document has not been written to since revisions were switched on. The next edit — yours or an agent's — appears here."
        />

        <ul v-else ref="railList" class="rev-list" role="listbox" aria-label="Revisions" @keydown="onRailKey">
          <EntryRevisionRow
            v-for="rev in revisions"
            :key="rev.revision"
            :revision="rev"
            :selected="rev.revision === selected"
            :is-base="rev.revision === base"
            :is-current="rev.revision === current"
            @select="select"
            @set-base="setBase"
          />
        </ul>
      </div>

      <!-- Pane -->
      <div class="rev-pane">
        <div class="pane-head">
          <div class="pane-range">
            <template v-if="diff">
              <strong v-if="diff.from">r{{ diff.from.revision }} → r{{ diff.to.revision }}</strong>
              <strong v-else>r{{ diff.to.revision }} · first version</strong>
              <span v-if="diff.summary" class="pane-stat">{{ diff.summary }}</span>
              <button v-if="base" class="btn btn-sm" @click="clearBase">Reset comparison</button>
            </template>
          </div>
          <div class="pane-actions">
            <button v-if="hasGaps" class="btn btn-sm" @click="expandAll = !expandAll">
              {{ expandAll ? 'Collapse all' : 'Expand all' }}
            </button>
            <button
              v-if="canWrite && diff && diff.to.revision !== current"
              class="btn btn-sm btn-primary"
              @click="$emit('restore', diff.to.revision)"
            >
              Restore r{{ diff.to.revision }}
            </button>
          </div>
        </div>

        <p v-if="restoreError" class="pane-error" role="alert">
          Restore failed — the document is unchanged.
        </p>
        <p class="sr-only" role="status">{{ announcement }}</p>

        <div v-if="diffLoading && !diff" class="skeleton-card pane-skeleton"></div>

        <EmptyState
          v-else-if="diffError"
          icon="&#x26A0;"
          title="Could not load this comparison"
          :description="`Revision r${selected} is still there — only the comparison failed.`"
        >
          <button class="btn btn-sm" @click="select(selected)">Try again</button>
        </EmptyState>

        <div v-else-if="diff" :class="['pane-content', { 'pane-stale': diffLoading }]">
          <EntryFieldChanges v-if="showFields" :fields="diff.fields ?? []" />

          <template v-if="bodyChanged">
            <EntryDiffView :diff="diff.content" :expand-all="expandAll" />
          </template>
          <p v-else-if="showFields" class="body-unchanged">Body unchanged.</p>
          <EmptyState
            v-else
            icon="&#x2261;"
            title="Nothing changed in this revision"
            :description="`It was written by ${diff.to.author_kind} at ${absoluteTime(diff.to.created_at)} and left both the body and the properties as they were.`"
          />
        </div>
      </div>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
const titleId = useId()
/**
 * The revision history of one entry: the rail navigates, the pane holds the
 * comparison.
 *
 * The reader's task is judgement, not browsing — they have just read a passage
 * and want to know who wrote it and what it replaced — so the rail is sized as
 * navigation and the pane gets the room. Restore lives in the pane, acting on
 * the revision being looked at: restoring something you have not read is not a
 * step worth saving.
 */
import { absoluteTime } from '~/composables/useRelativeTime'

const props = defineProps<{
  visible: boolean
  entryId: string
  entryTitle?: string
  entryCode?: string
  /** Set by the page while a restore is in flight, cleared on success. */
  restoreError?: boolean
}>()

const emit = defineEmits<{ close: []; restore: [number] }>()

// Reading the history is a read. Writing an old revision back over the entry
// is not, so only that button goes.
const { canWrite } = useResearchRole()

const { authFetch } = useAuth()
const apiBase = useRuntimeConfig().public.apiBase || ''

const revisions = ref<any[]>([])
const current = ref(0)
const selected = ref(0)
const base = ref(0)
const diff = ref<any>(null)
const loading = ref(false)
const error = ref(false)
const diffLoading = ref(false)
const diffError = ref(false)
const expandAll = ref(false)
const announcement = ref('')
const railList = ref<HTMLElement | null>(null)
const requestedRange = ref<{ from: number; to: number } | null>(null)

const railSummary = computed(() => {
  const n = revisions.value.length
  // A placeholder rather than an empty string: collapsing the sticky header to
  // nothing shifts the whole rail up the moment the list lands.
  if (!n) return loading.value ? 'Loading revisions…' : 'No revisions'
  const sessions = new Set(revisions.value.map((r) => r.session_code).filter(Boolean))
  const noun = n === 1 ? 'revision' : 'revisions'
  if (sessions.size === 1) return `${n} ${noun} · all in ${[...sessions][0]}`
  return `${n} ${noun}`
})

// Revision 1 reports every field as changed-from-empty, which is the entry
// being born rather than a change to it.
const showFields = computed(() => !!diff.value?.from && !!diff.value?.fields?.length)
const bodyChanged = computed(() => {
  const c = diff.value?.content
  return !!c && (c.added > 0 || c.removed > 0 || c.truncated)
})
const hasGaps = computed(() => {
  const c = diff.value?.content
  return !!c && c.unchanged > 6 && bodyChanged.value
})

async function load() {
  if (!props.entryId) return
  loading.value = true
  error.value = false
  try {
    const res: any = await authFetch(`${apiBase}/api/entries/${props.entryId}/revisions`)
    revisions.value = res?.data?.revisions ?? []
    current.value = res?.data?.current ?? 0
    const range = requestedRange.value
    requestedRange.value = null
    base.value = range?.from ?? 0
    if (revisions.value.length) {
      const newest = revisions.value[0].revision
      const target = range?.to && revisions.value.some((revision: any) => revision.revision === range.to)
        ? range.to
        : newest
      await select(target)
    }
  } catch (e) {
    error.value = true
    revisions.value = []
  } finally {
    loading.value = false
  }
}

async function select(revision: number) {
  selected.value = revision
  diffLoading.value = true
  diffError.value = false
  expandAll.value = false
  try {
    const query = base.value && base.value !== revision ? `?from=${base.value}&to=${revision}` : `?to=${revision}`
    const res: any = await authFetch(`${apiBase}/api/entries/${props.entryId}/diff${query}`)
    diff.value = res?.data ?? null
    announcement.value = diff.value?.from
      ? `r${diff.value.from.revision} to r${diff.value.to.revision}, ${diff.value.summary}`
      : `r${revision}, first version`
  } catch (e) {
    diffError.value = true
    diff.value = null
  } finally {
    diffLoading.value = false
  }
}

/**
 * Re-fetch the revision list without disturbing what the reader is looking at.
 *
 * `load()` is the open-the-panel path: it clears the comparison base and jumps
 * to the newest revision, which is right after a restore and wrong for a remote
 * write. A reader who shift-clicked r2 as a base and is reading r2 → r5 must not
 * be moved to r7 because an agent saved something. History is append-only, so a
 * selection is still valid after a refresh; it only needs re-resolving if the
 * revision it pointed at is gone.
 */
async function refreshList() {
  if (!props.entryId) return
  try {
    const res: any = await authFetch(`${apiBase}/api/entries/${props.entryId}/revisions`)
    const next = res?.data?.revisions ?? []
    revisions.value = next
    current.value = res?.data?.current ?? current.value
    error.value = false
    if (!next.some((r: any) => r.revision === selected.value)) {
      base.value = 0
      if (next.length) await select(next[0].revision)
    }
  } catch {
    // A refresh that fails leaves the list that is on screen: it was correct a
    // moment ago, and replacing it with an error would be the panel losing data
    // it still has.
  }
}

function setBase(revision: number) {
  base.value = revision === base.value ? 0 : revision
  select(selected.value)
}

function clearBase() {
  base.value = 0
  select(selected.value)
}

// The rail is one tab stop; arrows move within it. Without this a keyboard user
// tabs through every revision to reach the pane.
function onRailKey(event: KeyboardEvent) {
  const idx = revisions.value.findIndex((r) => r.revision === selected.value)
  if (idx < 0) return
  let next = idx
  if (event.key === 'ArrowDown') next = Math.min(idx + 1, revisions.value.length - 1)
  else if (event.key === 'ArrowUp') next = Math.max(idx - 1, 0)
  else if (event.key === 'Home') next = 0
  else if (event.key === 'End') next = revisions.value.length - 1
  else return
  event.preventDefault()
  select(revisions.value[next].revision)
  nextTick(() => railList.value?.querySelector<HTMLElement>('.rev-selected .rev-button')?.focus())
}

watch(() => props.visible, (open) => {
  if (open) load()
})

// reload: after a restore, where landing on the new head is the point.
// refreshList: after somebody else's write, where it would be an interruption.
/**
 * Open the panel already comparing a revision against the newest one.
 *
 * The caller is an annotation whose text drifted or vanished: the question is
 * "what happened to this sentence since I marked it", and the answer is the
 * diff from the revision it was anchored to. Without this the only route was
 * open the panel, find revision N in the rail, and shift-click it as a base —
 * a convention nothing on screen explains.
 */
/**
 * Queue links know both ends of the snapshot they displayed. Preparing the
 * range before the modal opens lets its ordinary visible watcher issue one
 * exact comparison instead of first jumping to whatever is newest now.
 */
function prepareRange(from: number, to: number) {
  requestedRange.value = { from, to }
}

defineExpose({ reload: load, refreshList, prepareRange })
</script>

<style scoped>
.history-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.history-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}
.history-eyebrow {
  font-size: var(--type-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-muted);
}
.history-title {
  margin: 0;
  font-size: var(--type-base, 1rem);
  font-weight: var(--weight-semibold);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-body {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: 300px 1fr;
  /* stretch, not start: the rail's rule has to run the full height of the
     dialog rather than stopping under the last revision. */
  align-items: stretch;
}
.rev-rail {
  border-inline-end: 1px solid var(--color-border-strong);
  overflow-y: auto;
  min-height: 0;
}
.rail-head {
  position: sticky;
  top: 0;
  z-index: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: var(--space-3) var(--space-4);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.rail-hint { opacity: 0.75; }
.rail-skeletons {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  padding: var(--space-3) var(--space-4);
}
.rail-skeleton { height: 3rem; }
.rev-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.rev-pane {
  overflow-y: auto;
  min-height: 0;
  padding: var(--space-5) var(--space-6);
}
.pane-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}
.pane-range {
  display: flex;
  align-items: baseline;
  gap: var(--space-3);
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  color: var(--color-text);
}
.pane-stat { color: var(--color-text-muted); }
.pane-actions {
  display: flex;
  gap: var(--space-2);
}
.pane-error {
  margin: 0 0 var(--space-3);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  background: color-mix(in srgb, var(--color-error) 12%, transparent);
  color: var(--color-error);
  font-size: var(--type-sm);
}
.pane-skeleton { height: 12rem; }
/* A new comparison dims the old one instead of blanking it — arrowing down the
   rail must not strobe. */
.pane-stale {
  opacity: 0.5;
  transition: opacity var(--transition-base, 0.2s);
}
.body-unchanged {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  margin: 0;
}

@media (max-width: 1024px) {
  .history-body { grid-template-columns: clamp(240px, 28%, 300px) 1fr; }
}
@media (max-width: 768px) {
  .history-body {
    grid-template-columns: 1fr;
    grid-template-rows: minmax(0, 40%) minmax(0, 60%);
  }
  .rev-rail {
    border-inline-end: none;
    border-bottom: 1px solid var(--color-border-strong);
  }
  .rev-pane { padding: var(--space-4); }
}
</style>
