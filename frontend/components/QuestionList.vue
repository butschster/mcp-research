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
        <option value="high">↑ High</option>
        <option value="medium">• Medium</option>
        <option value="low">↓ Low</option>
      </select>
      <button v-if="filterArea || filterPriority" class="btn q-clear" @click="clearFilters">
        Clear filters
      </button>
    </div>

    <!-- Groups -->
    <div
      v-for="group in visibleGroups"
      :key="group.status"
      class="question-group"
    >
      <!-- Group header (clickable) -->
      <button class="group-header" @click="toggle(group.status)">
        <StatusBadge :status="group.status" />
        <span class="card-meta">({{ group.questions.length }})</span>
        <span class="group-chevron" :class="{ open: isOpen(group.status) }">›</span>
      </button>

      <!-- Questions (collapsible) -->
      <div v-show="isOpen(group.status)" class="group-body">
        <div
          v-for="q in group.questions"
          :key="q.id"
          :class="['question-item', { 'question-child': q.parent_id }]"
        >
          <div class="question-row">
            <div class="question-text">{{ q.text }}</div>
            <StatusBadge v-if="q.priority" :status="q.priority" />
          </div>
          <div v-if="q.area" class="card-meta q-area">{{ q.area }}</div>
          <div v-if="q.answer" class="question-answer">{{ q.answer }}</div>
        </div>

        <EmptyState
          v-if="group.questions.length === 0"
          title="No questions match filters"
        />
      </div>
    </div>

    <EmptyState
      v-if="visibleGroups.length === 0"
      icon="📭"
      title="No questions yet"
      description="Questions will appear here once the session starts."
    />
  </div>
</template>

<script setup lang="ts">
interface Question {
  id: string
  text: string
  area: string
  priority: string
  answer: string
  parent_id: string
  status: string
}

const props = defineProps<{
  questions: Record<string, Question[]>
}>()

// Filter state
const filterArea = ref('')
const filterPriority = ref('')

function clearFilters() {
  filterArea.value = ''
  filterPriority.value = ''
}

// Collapse state — pending + in_progress open by default
const STATUS_ORDER = ['pending', 'in_progress', 'answered', 'deferred', 'skipped']
const DEFAULT_OPEN = new Set(['pending', 'in_progress'])
const openGroups = ref<Record<string, boolean>>(
  Object.fromEntries(STATUS_ORDER.map(s => [s, DEFAULT_OPEN.has(s)]))
)

function isOpen(status: string) { return openGroups.value[status] ?? false }
function toggle(status: string) { openGroups.value[status] = !openGroups.value[status] }

// Derived areas for filter
const areas = computed(() => {
  const set = new Set<string>()
  for (const qs of Object.values(props.questions))
    for (const q of qs) if (q.area) set.add(q.area)
  return [...set].sort()
})

const hasFilters = computed(() => areas.value.length > 0)

// Filter questions
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
</script>

<style scoped>
.question-filters {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}
.q-select {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 0.375rem 0.625rem;
  color: var(--color-text);
  font-size: 0.8125rem;
  cursor: pointer;
}
.q-clear { font-size: 0.8125rem; padding: 0.375rem 0.625rem; }

.question-group { margin-bottom: 0.5rem; }

.group-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  background: none;
  border: none;
  padding: 0.5rem 0;
  cursor: pointer;
  color: var(--color-text);
  text-align: left;
}
.group-header:hover { color: var(--color-primary); }

.group-chevron {
  margin-left: auto;
  font-size: 1.125rem;
  color: var(--color-text-muted);
  transition: transform 0.2s;
  display: inline-block;
}
.group-chevron.open { transform: rotate(90deg); }

.group-body {
  padding-left: 0.25rem;
  border-left: 2px solid var(--color-border);
  margin-left: 0.5rem;
  margin-bottom: 0.75rem;
}

.question-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.5rem;
}
.q-area { margin-top: 0.125rem; font-size: 0.75rem; }
</style>
