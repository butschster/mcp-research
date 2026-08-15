<template>
  <div>
    <!-- Filter bar -->
    <div v-if="hasFilters" class="question-filters">
      <select v-model="filterArea" class="q-select">
        <option value="">All areas</option>
        <option v-for="a in areas" :key="a" :value="a">{{ a }}</option>
      </select>
      <select v-model="filterPriority" class="q-select">
        <option value="">All priorities</option>
        <option value="high">&uarr; High</option>
        <option value="medium">&bull; Medium</option>
        <option value="low">&darr; Low</option>
      </select>
      <button v-if="filterArea || filterPriority" class="btn btn-sm" @click="clearFilters">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
        Clear
      </button>
    </div>

    <!-- Groups -->
    <div v-for="group in visibleGroups" :key="group.status" class="question-group">
      <button class="btn-ghost group-header" @click="toggle(group.status)">
        <StatusBadge :status="group.status" />
        <span class="group-count">{{ group.questions.length }}</span>
        <span class="group-chevron" :class="{ open: isOpen(group.status) }">&rsaquo;</span>
      </button>

      <div v-show="isOpen(group.status)" class="group-body">
        <component
          :is="questionLink(q) ? NuxtLink : 'div'"
          v-for="q in group.questions"
          :key="q.id"
          :to="questionLink(q) || undefined"
          :class="['question-card', { 'question-child': q.parent_id, 'question-answered': q.status === 'answered', 'question-inert': !questionLink(q) }]"
        >
          <div class="q-top">
            <span v-if="q.code" class="q-code">{{ q.code }}</span>
            <span class="q-text" v-html="renderRefs(q.text, researchSlug || '')"></span>
            <div class="q-badges">
              <StatusBadge v-if="q.priority" :status="q.priority" />
            </div>
          </div>
          <div v-if="q.area" class="q-area">{{ q.area }}</div>
          <div v-if="q.answer" class="q-answer-preview">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
            <span>{{ truncateAnswer(q.answer, 120) }}</span>
          </div>
        </component>

        <EmptyState v-if="group.questions.length === 0" title="No questions match filters" />
      </div>
    </div>

    <EmptyState
      v-if="visibleGroups.length === 0"
      icon="&#x1F4ED;"
      title="No questions yet"
      description="Questions will appear here once the session starts."
    />
  </div>
</template>

<script setup lang="ts">
import { NuxtLink } from '#components'

import { renderRefs } from '~/composables/useCrossRefs'
import { normalizeContent } from '~/utils/normalizeContent'

interface Question {
  id: string
  code: string
  text: string
  area: string
  priority: string
  // A pending question has no answer and a top-level one has no parent; both
  // arrive as null. Declaring them as plain strings made the type disagree with
  // every payload the component is actually handed.
  answer?: string | null
  parent_id?: string | null
  status: string
}

const props = defineProps<{
  questions: Record<string, Question[]>
  researchSlug?: string
  sessionId?: string
}>()

const filterArea = ref('')
const filterPriority = ref('')

function clearFilters() {
  filterArea.value = ''
  filterPriority.value = ''
}

const STATUS_ORDER = ['pending', 'in_progress', 'answered', 'deferred', 'skipped']
const DEFAULT_OPEN = new Set(['pending', 'in_progress', 'answered'])
const openGroups = ref<Record<string, boolean>>(
  Object.fromEntries(STATUS_ORDER.map(s => [s, DEFAULT_OPEN.has(s)]))
)

function isOpen(status: string) { return openGroups.value[status] ?? false }
function toggle(status: string) { openGroups.value[status] = !openGroups.value[status] }

const areas = computed(() => {
  const set = new Set<string>()
  for (const qs of Object.values(props.questions))
    for (const q of qs) if (q.area) set.add(q.area)
  return [...set].sort()
})

const hasFilters = computed(() => areas.value.length > 0)

function filterQ(qs: Question[]) {
  return qs.filter(q => {
    if (filterArea.value && q.area !== filterArea.value) return false
    if (filterPriority.value && q.priority !== filterPriority.value) return false
    return true
  })
}

const visibleGroups = computed(() =>
  STATUS_ORDER
    .filter(s => props.questions[s]?.length)
    .map(s => ({ status: s, questions: filterQ(props.questions[s] ?? []) }))
    .filter(g => g.questions.length > 0 || isOpen(g.status))
)

/**
 * Where a question card points, or '' when it must not be a link.
 *
 * A shared view has no per-question page: the question, its answer and its
 * children are all on the session page already, and a route per question is one
 * more payload to get the redaction right on. Under a share the card renders as
 * a card and stays put.
 */
function questionLink(q: Question): string {
  if (shareActive()) return ''
  if (props.sessionId && props.researchSlug) {
    return `/research/${props.researchSlug}/session/${props.sessionId}/question/${q.code || q.id}`
  }
  return ''
}

function truncateAnswer(text: string, max: number): string {
  // Strip markdown formatting for preview
  const plain = normalizeContent(text)
    .replace(/#{1,6}\s+/g, '')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/\[\[([^\]]+)\]\]/g, '$1')
    .replace(/\n/g, ' ')
    .replace(/\|[^|]+\|/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
  return plain.length > max ? plain.slice(0, max) + '...' : plain
}
</script>

<style scoped>
.question-filters {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-4);
  flex-wrap: wrap;
}
.q-select {
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-sm);
  padding: 0.35rem 1.5rem 0.35rem 0.625rem;
  color: var(--color-text-muted);
  font-size: var(--type-sm);
  font-weight: var(--weight-medium);
  font-family: inherit;
  cursor: pointer;
  transition: all var(--transition-fast);
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='6' fill='none'%3E%3Cpath d='M1 1l4 4 4-4' stroke='%237f8ea3' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.625rem center;
}
.q-select:hover { border-color: rgba(108, 197, 224, 0.25); color: var(--color-text); }

.question-group { margin-bottom: var(--space-3); }

.group-header {
  gap: var(--space-2);
  width: 100%;
  padding: var(--space-2) var(--space-2);
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
}
.group-header:hover { background: var(--color-surface-hover); }

.group-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  border-radius: var(--radius-xs);
  padding: 0.15rem 0.375rem;
  font-weight: var(--weight-semibold);
  min-width: 1.25rem;
  text-align: center;
  font-variant-numeric: tabular-nums;
}

.group-chevron {
  margin-left: auto;
  font-size: var(--type-lg);
  color: var(--color-text-muted);
  transition: transform var(--transition-base);
  display: inline-block;
}
.group-chevron.open { transform: rotate(90deg); }

.group-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  margin-top: var(--space-2);
  margin-left: var(--space-2);
}

/* Question card */
.question-card {
  display: block;
  padding: var(--space-3) var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  text-decoration: none;
  color: inherit;
  transition: border-color var(--transition-fast);
}
.question-card:hover {
  border-color: var(--color-border-strong);
}
/* A card that does not navigate must not offer a hover affordance saying it
   might. There is no question page inside a shared view. */
.question-inert { cursor: default; }
.question-inert:hover { border-color: var(--color-border); }
.question-answered {
  border-left: 2px solid var(--color-success);
}
.question-child {
  margin-left: var(--space-6);
}

.q-top {
  display: flex;
  align-items: flex-start;
  gap: var(--space-2);
}
.q-code {
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.625rem;
  font-weight: var(--weight-bold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.1rem 0.3rem;
  border-radius: var(--radius-xs);
  flex-shrink: 0;
  margin-top: 2px;
}
.q-text {
  flex: 1;
  font-size: var(--type-sm);
  font-weight: var(--weight-medium);
  line-height: 1.4;
}
.q-badges {
  display: flex;
  gap: var(--space-1);
  flex-shrink: 0;
}
.q-area {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  margin-top: var(--space-1);
  margin-left: calc(0.625rem + 0.6rem + var(--space-2));
}
.q-answer-preview {
  display: flex;
  align-items: flex-start;
  gap: var(--space-2);
  margin-top: var(--space-2);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  line-height: 1.5;
  margin-left: calc(0.625rem + 0.6rem + var(--space-2));
}
.q-answer-preview svg {
  flex-shrink: 0;
  color: var(--color-success);
  margin-top: 2px;
}
.q-answer-preview span {
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
</style>
