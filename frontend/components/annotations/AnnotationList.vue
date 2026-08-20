<template>
  <div class="ann-list">
    <div v-if="showFilters" class="ann-list__filters cluster">
      <select
        class="select"
        :value="filters.status || ''"
        aria-label="Status"
        @change="update('status', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Any status</option>
        <option value="open">Open</option>
        <option value="answered">Answered</option>
        <option value="closed">Closed</option>
        <option value="dismissed">Dismissed</option>
      </select>

      <select
        class="select"
        :value="filters.kind || ''"
        aria-label="Kind"
        @change="update('kind', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Any kind</option>
        <option value="verify">Verify</option>
        <option value="dig">Dig</option>
        <option value="disagree">Disagree</option>
      </select>

      <select
        class="select"
        :value="filters.anchor || ''"
        aria-label="Anchor state"
        @change="update('anchor', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Anywhere</option>
        <option value="anchored">Anchored</option>
        <option value="drifted">Drifted</option>
        <option value="moved">Moved</option>
        <option value="orphaned">Orphaned</option>
      </select>
    </div>

    <p v-if="loading" class="ann-list__status">Loading…</p>

    <div v-else-if="error" class="ann-list__status ann-list__status--error" role="alert">
      {{ error }}
      <button type="button" class="btn btn-sm" @click="$emit('retry')">Retry</button>
    </div>

    <EmptyState
      v-else-if="!groups.length"
      :icon="empty.icon"
      :title="empty.title"
      :description="empty.description"
    />

    <div v-else class="ann-list__groups">
      <section v-for="group in groups" :key="group.key" class="card card--list">
        <header v-if="grouped" class="list-head">
          <NuxtLink v-if="group.href" class="ann-list__entry" :to="group.href">
            <ShortCode :code="group.code" />
            <span class="ann-list__entry-title">{{ group.title }}</span>
          </NuxtLink>
          <span v-else class="ann-list__entry">
            <ShortCode :code="group.code" />
            <span class="ann-list__entry-title">{{ group.title }}</span>
          </span>
          <span class="ann-list__count">{{ group.items.length }}</span>
        </header>

        <div
          class="data-rows--bounded"
          role="listbox"
          :aria-label="grouped ? `Marks in ${group.code}` : 'Marks'"
          @keydown="onKey($event, group)"
        >
          <AnnotationsAnnotationRow
            v-for="(a, i) in group.items"
            :key="a.id"
            :annotation="a"
            :research-slug="researchSlug ?? ''"
            :focusable="i === 0"
            :selectable="selectable"
            :selected="selectedIds.includes(a.id)"
            :dense="dense"
            @open="$emit('open', a)"
            @toggle-select="$emit('toggle-select', a)"
          />
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * The list of marks, wherever it is shown.
 *
 * One component for the document panel and the research queue, because two
 * would be two sets of empty states — and the empty state is the whole
 * difference between "nothing is marked here" and "your filter matches
 * nothing", which is the message people actually need.
 *
 * Grouping by document is not decoration: it is the shape a pass is taken in.
 * A batch that is one document costs one read of it; a batch scattered across
 * six costs six.
 */
import type { Annotation, QueueFilters } from '~/composables/useAnnotations'

const props = withDefaults(defineProps<{
  annotations: Annotation[]
  researchSlug?: string
  loading?: boolean
  error?: string | null
  filters?: QueueFilters
  showFilters?: boolean
  grouped?: boolean
  selectable?: boolean
  selectedIds?: string[]
  dense?: boolean
  emptyVariant?: 'document' | 'research' | 'filtered'
}>(), {
  filters: () => ({}),
  selectedIds: () => [],
  emptyVariant: 'research',
})

const emit = defineEmits<{
  'update:filters': [filters: QueueFilters]
  open: [annotation: Annotation]
  'toggle-select': [annotation: Annotation]
  retry: []
}>()

const EMPTY = {
  document: {
    icon: '◌',
    title: 'Nothing marked in this document',
    description: 'Select a sentence you do not believe and mark it. The agent works these when you ask it to.',
  },
  research: {
    icon: '◌',
    title: 'No marks yet',
    description: 'Open a document, select a sentence that needs checking, and mark it.',
  },
  filtered: {
    icon: '⌕',
    title: 'No marks match',
    description: 'Nothing here fits the current filters.',
  },
}

const { canWrite } = useResearchRole()

const empty = computed(() => {
  // A filtered-empty list is a different message from an empty research, and
  // showing the wrong one sends somebody looking for a feature that is working.
  const filtered = Object.values(props.filters ?? {}).some(Boolean)
  if (filtered && props.emptyVariant !== 'document') return EMPTY.filtered
  const base = EMPTY[props.emptyVariant]
  // Telling a viewer to go and mark a sentence describes a button they do not
  // have.
  if (!canWrite.value) {
    return { ...base, description: 'Nothing here has been marked as needing a second look.' }
  }
  return base
})

interface Group {
  key: string
  code: string
  title: string
  href: string
  items: Annotation[]
}

const groups = computed<Group[]>(() => {
  if (!props.annotations.length) return []
  if (!props.grouped) {
    return [{ key: 'all', code: '', title: '', href: '', items: props.annotations }]
  }
  const map = new Map<string, Group>()
  for (const a of props.annotations) {
    const key = a.entry_id
    if (!map.has(key)) {
      map.set(key, {
        key,
        code: a.entry_code || '',
        title: a.entry_title || 'Untitled',
        href: props.researchSlug && a.entry_code ? entryPath(props.researchSlug, a.entry_code) : '',
        items: [],
      })
    }
    map.get(key)!.items.push(a)
  }
  // Within a document, marks read in the order of the page rather than the
  // order they were made. An orphan has no place there, and the server says so
  // with -1 —
  // which sorts FIRST unless it is translated. Reading top to bottom, a mark on
  // text that no longer exists belongs at the end, not above the first
  // paragraph.
  const at = (a: Annotation) => {
    const i = a.anchor?.block_index ?? -1
    return i < 0 ? Number.MAX_SAFE_INTEGER : i
  }
  for (const group of map.values()) {
    group.items.sort((a, b) => at(a) - at(b))
  }
  return [...map.values()]
})

function update(key: keyof QueueFilters, value: string) {
  emit('update:filters', { ...props.filters, [key]: value })
}

/** Roving selection inside a group, the same shape HistoryPanel's rail uses. */
function onKey(event: KeyboardEvent, group: Group) {
  const keys = ['ArrowDown', 'ArrowUp', 'Home', 'End']
  if (!keys.includes(event.key)) return
  const container = event.currentTarget as HTMLElement
  const rows = [...container.querySelectorAll<HTMLElement>('[role="option"]')]
  if (!rows.length) return
  const current = rows.findIndex((r) => r === document.activeElement)
  let next = current
  if (event.key === 'ArrowDown') next = Math.min(rows.length - 1, current + 1)
  if (event.key === 'ArrowUp') next = Math.max(0, current - 1)
  if (event.key === 'Home') next = 0
  if (event.key === 'End') next = rows.length - 1
  if (next !== current && rows[next]) {
    event.preventDefault()
    // Roving: the row that has focus is the row that is a tab stop, so tabbing
    // away and back returns to where the reader was rather than to the top.
    rows.forEach((r, i) => r.setAttribute('tabindex', i === next ? '0' : '-1'))
    rows[next]!.focus()
    rows[next]!.scrollIntoView({ block: 'nearest' })
  }
}
</script>

<style scoped>
.ann-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.ann-list__filters { flex-wrap: wrap; }

.ann-list__status {
  margin: 0;
  padding: var(--space-3);
  font-size: var(--type-sm);
  color: var(--color-text-muted);
}

.ann-list__status--error {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-error);
}

.ann-list__groups {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.ann-list__entry {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}

.ann-list__entry-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ann-list__count {
  font-size: var(--type-2xs);
  color: var(--color-text-faint);
}
</style>
