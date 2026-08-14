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
      <button class="btn btn-sm" @click="load">Try again</button>
    </EmptyState>

    <EmptyState
      v-else-if="!changes.length"
      icon="&#x1F4DD;"
      title="No entry changes"
      description="This session has not written to any entry yet. Entries it creates or edits will show up here with what changed."
    />

    <template v-else>
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
              :to="`/research/${researchSlug}/entry/${change.entry_code || change.entry_id}`"
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
    </template>
  </div>
</template>

<script setup lang="ts">
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

const { authFetch } = useAuth()
const apiBase = useRuntimeConfig().public.apiBase || ''

const changes = ref<any[]>([])
const created = ref(0)
const modified = ref(0)
const loading = ref(true)
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

async function load() {
  loading.value = true
  error.value = false
  try {
    const res: any = await authFetch(`${apiBase}/api/sessions/${props.sessionId}/changes`)
    changes.value = res?.data?.changes ?? []
    created.value = res?.data?.created ?? 0
    modified.value = res?.data?.modified ?? 0
    // One change is a session that did one thing; make the reader click nothing.
    allOpen.value = changes.value.length === 1
  } catch (e) {
    error.value = true
    changes.value = []
  } finally {
    loading.value = false
  }
}

watch(() => props.sessionId, load, { immediate: true })
</script>

<style scoped>
.changes-skeletons {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
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
  font-weight: 600;
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
  border-radius: 4px;
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
