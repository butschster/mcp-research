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
        Clear filters
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

const filterArea = ref('')
const filterPriority = ref('')

function clearFilters() {
  filterArea.value = ''
  filterPriority.value = ''
}

const STATUS_ORDER = ['pending', 'in_progress', 'answered', 'deferred', 'skipped']
const DEFAULT_OPEN = new Set(['pending', 'in_progress'])
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
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--space-2) var(--space-3);
  color: var(--color-text);
  font-size: var(--type-sm);
  cursor: pointer;
}

.question-group { margin-bottom: var(--space-2); }

.group-header {
  gap: var(--space-2);
  width: 100%;
  padding: var(--space-2) 0;
}
.group-header:hover { color: var(--color-primary); }

.group-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  border-radius: 999px;
  padding: var(--space-1) var(--space-2);
  font-weight: 600;
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
  padding-left: var(--space-1);
  border-left: 2px solid var(--color-border);
  margin-left: var(--space-2);
  margin-bottom: var(--space-3);
}

.question-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-2);
}
.q-area { margin-top: var(--space-1); }
</style>
