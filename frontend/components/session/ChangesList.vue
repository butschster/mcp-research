<template>
  <div class="changes">
    <div v-if="loading" class="changes-skeletons">
      <div v-for="i in 3" :key="i" class="skeleton-card change-skeleton"></div>
    </div>

    <EmptyState
      v-else-if="error"
      icon="&#x26A0;"
      title="Could not load what this session changed"
      description="The change list did not come back. Nothing about the session has changed."
    >
      <button class="btn btn-sm" @click="load()">Try again</button>
    </EmptyState>

    <EmptyState
      v-else-if="!changes.length"
      icon="&#x1F4DD;"
      title="No document changes"
      description="This session has not written to any document yet. Documents it creates or edits will show up here with what changed."
    />

    <template v-else>
      <div :class="['changes-body', { 'changes-stale': refreshing }]">
      <div class="changes-summary">
        <p class="changes-counts">
          <strong>{{ created }}</strong> created · <strong>{{ modified }}</strong> modified
        </p>
        <button v-if="changes.length > 1" class="btn btn-sm" @click="toggleAll">
          {{ allOpen ? 'Collapse all' : 'Expand all' }}
        </button>
      </div>

      <div v-for="change in changes" :key="change.entry_id" class="card change-card">
        <div class="change-head">
          <div class="change-title-row">
            <ShortCode v-if="change.entry_code" :code="change.entry_code" />
            <NuxtLink
              :to="entryPath(researchSlug, change.entry_code || change.entry_id)"
              class="change-title"
            >{{ change.title }}</NuxtLink>
          </div>
          <div class="change-meta">
            <span :class="['change-badge', change.created ? 'badge-created' : 'badge-modified']">
              {{ change.created ? 'created' : 'modified' }}
            </span>
            <EntryAuthorBadge
              v-for="kind in authorsOf(change)"
              :key="kind"
              :kind="kind"
            />
            <span v-if="change.summary" class="change-stat">{{ change.summary }}</span>
            <span class="change-revs">
              r{{ change.from_revision }}<template v-if="change.to_revision !== change.from_revision">–r{{ change.to_revision }}</template>
            </span>
            <!-- "Was it a person or the model" and "did it happen last night"
                 are one question, and the second half was arriving in the
                 payload and being dropped. Without it the reader opens each
                 entry to find a date this card already had. -->
            <time v-if="lastTouched(change)" class="change-when" :datetime="lastTouched(change)!" :title="absoluteTime(lastTouched(change)!)">
              {{ relativeTime(lastTouched(change)!) }}
            </time>
          </div>
        </div>

        <EntryFieldChanges v-if="change.fields?.length" :fields="change.fields" />

        <button
          v-if="change.diff"
          class="change-toggle"
          :aria-expanded="isOpen(change.entry_id)"
          @click="toggle(change.entry_id)"
        >
          {{ isOpen(change.entry_id) ? 'Hide changes' : 'Show changes' }}
        </button>
        <EntryDiffView v-if="change.diff && isOpen(change.entry_id)" :diff="change.diff" />
      </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { entryPath } from '~/composables/useResearchPaths'
import { relativeTime, absoluteTime } from '~/composables/useRelativeTime'

/**
 * What a session did to the research's entries: the ones it created, the ones it
 * edited, and the diff for each.
 *
 * Deliberately not the same list as the "Entries" tab. That one shows entries
 * whose session_id matches — the ones this session created. This one also
 * covers entries the session edited without creating, which is what a reader
 * asking "what happened in here" actually wants.
 */
const props = defineProps<{
  sessionId: string
  researchSlug: string
}>()

/** How many entries this session touched, so the tab can carry the count
 *  without the reader opening it to find out. Emitted on every load, including
 *  a failed one — a stale badge is a worse lie than no badge. */
const emit = defineEmits<{ count: [number | null] }>()

const { authFetch } = useAuth()
const apiBase = useRuntimeConfig().public.apiBase || ''

const changes = ref<any[]>([])
const created = ref(0)
const modified = ref(0)
const loading = ref(true)
const refreshing = ref(false)
const error = ref(false)
const open = ref<Record<string, boolean>>({})
const allOpen = ref(false)

function isOpen(entryId: string): boolean {
  return allOpen.value || !!open.value[entryId]
}

function toggle(entryId: string) {
  open.value = { ...open.value, [entryId]: !isOpen(entryId) }
  if (allOpen.value) {
    // Leaving "expand all" on would immediately re-open what was just closed.
    allOpen.value = false
    open.value = Object.fromEntries(changes.value.map((c) => [c.entry_id, c.entry_id !== entryId]))
  }
}

function toggleAll() {
  allOpen.value = !allOpen.value
  open.value = {}
}

/** The distinct author kinds behind a card — the only place outside the history
 *  panel where "a person did this" is visible. */
function authorsOf(change: any): string[] {
  const kinds = (change.revisions ?? []).map((r: any) => r.author_kind).filter(Boolean)
  return [...new Set(kinds)] as string[]
}

/** When this session last wrote to the entry. */
function lastTouched(change: any): string | null {
  const stamps = (change.revisions ?? []).map((r: any) => r.created_at).filter(Boolean)
  return stamps.length ? stamps[stamps.length - 1] : null
}

// Two reloads can be in flight at once — a realtime burst arriving while the
// first is still out — and the responses are not ordered. Without a token the
// older one can land last and the list settles on a snapshot that is already
// wrong, with nothing to correct it until the next event.
let requestToken = 0

async function load({ quiet = false } = {}) {
  const token = ++requestToken
  // A refresh keeps what is on screen. Blanking four rendered cards to three
  // skeletons on every write an agent makes is a ~200px collapse under a
  // reader's eyes, and the panel beside this one already knows better: it dims
  // rather than tears down.
  if (quiet && changes.value.length) refreshing.value = true
  else loading.value = true
  try {
    const res: any = await authFetch(`${apiBase}/api/sessions/${props.sessionId}/changes`)
    if (token !== requestToken) return
    changes.value = res?.data?.changes ?? []
    created.value = res?.data?.created ?? 0
    modified.value = res?.data?.modified ?? 0
    error.value = false
    // One change is a session that did one thing; make the reader click nothing.
    // Only on a first load: doing it on a refresh silently undoes an explicit
    // "Expand all" while individually opened cards survive.
    if (!quiet) allOpen.value = changes.value.length === 1
    emit('count', changes.value.length)
  } catch (e) {
    if (token !== requestToken) return
    // A refresh that fails must not destroy a list that is correct. Only a load
    // with nothing behind it falls through to the error state.
    if (changes.value.length) {
      emit('count', changes.value.length)
    } else {
      error.value = true
      changes.value = []
      emit('count', null)
    }
  } finally {
    if (token === requestToken) {
      loading.value = false
      refreshing.value = false
    }
  }
}

watch(() => props.sessionId, () => load(), { immediate: true })

// The session page reloads this on an `entry` event: an agent writing entries is
// exactly when the reader is watching, and it was the one list that froze.
defineExpose({ reload: () => load({ quiet: true }) })
</script>

<style scoped>
.changes-skeletons {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}
/* A refresh dims the list instead of blanking it — an agent writing five
   entries must not strobe the page five times. Same treatment the history
   pane already uses for the same reason. */
.changes-stale {
  opacity: 0.5;
  transition: opacity var(--transition-base, 0.2s);
}
.change-when {
  color: var(--color-text-muted);
  font-size: var(--type-xs);
  white-space: nowrap;
}
.change-skeleton { height: 5rem; }
.changes-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin: 0 0 var(--space-4);
}
.changes-counts {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  margin: 0;
}
.change-card {
  padding: var(--space-4);
  margin-bottom: var(--space-3);
}
.change-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}
.change-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}
.change-title {
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
  text-decoration: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.change-title:hover { color: var(--color-primary); }
.change-meta {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.change-badge {
  padding: 0.1rem 0.4rem;
  border-radius: var(--radius-xs);
  font-size: var(--type-xs);
}
.badge-created {
  color: var(--color-success);
  background: color-mix(in srgb, var(--color-success) 15%, transparent);
}
.badge-modified {
  color: var(--color-primary);
  background: var(--color-primary-muted);
}
.change-stat,
.change-revs { font-family: 'JetBrains Mono', monospace; }
.change-toggle {
  margin-top: var(--space-3);
  background: none;
  border: none;
  padding: 0;
  font-size: var(--type-xs);
  color: var(--color-primary);
  cursor: pointer;
}
.change-toggle:hover { text-decoration: underline; }
</style>
