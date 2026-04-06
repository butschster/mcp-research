<template>
  <div v-if="pending">
    <div class="skeleton-card skeleton-header"></div>
    <div class="skeleton-card skeleton-content"></div>
  </div>

  <div v-else-if="session">
    <!-- Header -->
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: researchName, to: `/research/${researchSlug}` },
        { label: session.title }
      ]" />
      <div class="session-header">
        <h1 class="page-title">{{ session.title }}</h1>
        <StatusBadge :status="session.status" />
      </div>
      <p v-if="session.focus" class="card-meta mt-2">Focus: {{ session.focus }}</p>
    </div>

    <!-- Notes -->
    <div v-if="session.notes" class="card notes-card">
      <h3 class="card-section-title">Session notes</h3>
      <p class="notes-text">{{ session.notes }}</p>
    </div>

    <!-- Tabs -->
    <div class="content-tabs">
      <button
        :class="['content-tab', { active: activeTab === 'questions' }]"
        @click="activeTab = 'questions'"
      >
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/></svg>
        Questions
        <span class="tab-count">{{ progress.answered }} / {{ progress.total }}</span>
      </button>
      <button
        :class="['content-tab', { active: activeTab === 'tasks' }]"
        @click="activeTab = 'tasks'"
      >
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
        Tasks
        <span v-if="tasks.length" class="tab-count">{{ completedTasks }} / {{ tasks.length }}</span>
      </button>
    </div>

    <!-- Questions panel -->
    <div v-if="activeTab === 'questions'">
      <!-- Progress inside questions -->
      <div class="panel-progress">
        <ProgressBar :value="progress.answered" :total="progress.total" />
        <div class="progress-stats">
          <span class="stat-answered">{{ progress.answered }} answered</span>
          <span class="stat-pending">{{ progress.pending }} pending</span>
          <span v-if="progress.deferred" class="stat-muted">{{ progress.deferred }} deferred</span>
          <span v-if="progress.skipped" class="stat-skipped">{{ progress.skipped }} skipped</span>
        </div>
      </div>
      <QuestionList :questions="questions" :research-slug="researchSlug" :session-id="sessionId" />
    </div>

    <!-- Tasks panel -->
    <div v-if="activeTab === 'tasks'">
      <div v-if="tasks.length">
        <div class="panel-progress">
          <ProgressBar :value="completedTasks" :total="tasks.length" />
          <div class="progress-stats">
            <span class="stat-answered">{{ completedTasks }} completed</span>
            <span class="stat-pending">{{ tasks.length - completedTasks }} remaining</span>
          </div>
        </div>
        <div class="todo-list">
          <div v-for="t in tasks" :key="t.id" class="todo-item">
            <span :class="['todo-check', `todo-${t.status}`]">
              <template v-if="t.status === 'completed'">&check;</template>
              <template v-else-if="t.status === 'failed'">&times;</template>
              <template v-else-if="t.status === 'blocked'">&block;</template>
              <template v-else-if="t.status === 'deferred'">&rarr;</template>
              <template v-else-if="t.status === 'in_progress'">&triangleright;</template>
              <template v-else>&cir;</template>
            </span>
            <div class="todo-content">
              <span :class="['todo-text', { 'todo-done': t.status === 'completed' }]">{{ t.title }}</span>
              <div v-if="t.result" class="todo-result markdown-content" v-html="renderRefs(marked.parse(t.result) as string, researchSlug)"></div>
            </div>
            <StatusBadge v-if="t.priority === 'high'" :status="t.priority" />
            <StatusBadge :status="t.status" />
          </div>
        </div>
      </div>
      <EmptyState v-else icon="&#x2705;" title="No tasks" description="Tasks will appear here when created by the AI." />
    </div>
  </div>

  <EmptyState v-else icon="&#x1F50D;" title="Session not found" />
</template>

<script setup lang="ts">
import { marked } from 'marked'
marked.setOptions({ gfm: true, breaks: true })
const route = useRoute()
const id = route.params.id as string
const sessionId = route.params.sessionId as string

const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? 'Research')
const researchSlug = computed(() => researchData.value?.data?.research?.code || id)

const { data, pending } = await useApi<{ data: any }>(`/api/sessions/${sessionId}`)

const session  = computed(() => data.value?.data?.session ?? data.value?.data?.Session)
const questions = computed(() => data.value?.data?.questions ?? data.value?.data?.Questions ?? {})
const progress  = computed(() => ({
  total:    data.value?.data?.progress?.total    ?? 0,
  answered: data.value?.data?.progress?.answered ?? 0,
  pending:  data.value?.data?.progress?.pending  ?? 0,
  deferred: data.value?.data?.progress?.deferred ?? 0,
  skipped:  data.value?.data?.progress?.skipped  ?? 0,
}))

// Tasks
const { data: tasksData } = await useApi<{ data: any[] }>(`/api/researches/${id}/tasks`)
const tasks = computed(() => tasksData.value?.data ?? [])
const completedTasks = computed(() => tasks.value.filter((t: any) => t.status === 'completed').length)

// Active tab
const activeTab = ref<'questions' | 'tasks'>('questions')

const { authFetch } = useAuth()
const rtBase = useRuntimeConfig().public.apiBase || ''
useRealtimeUpdates(async (event) => {
  if (event.research_id && event.research_id !== id) return
  if (['question', 'session'].includes(event.entity)) {
    data.value = await authFetch<any>(`${rtBase}/api/sessions/${sessionId}`)
  }
  if (event.entity === 'task') {
    tasksData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/tasks`)
  }
})
</script>

<style scoped>
.session-header { display: flex; justify-content: space-between; align-items: center; gap: var(--space-4); }
.card-section-title { font-size: var(--type-base); font-weight: 600; margin-bottom: var(--space-3); letter-spacing: -0.01em; }
.notes-card { margin-bottom: var(--space-6); }
.notes-text { white-space: pre-wrap; color: var(--color-text-muted); font-size: var(--type-sm); line-height: 1.6; }

/* Panel progress (inside tab) */
.panel-progress {
  margin-bottom: var(--space-5);
}
.progress-stats {
  display: flex;
  gap: var(--space-4);
  font-size: var(--type-xs);
  margin-top: var(--space-2);
  color: var(--color-text-muted);
}
.stat-answered { color: var(--color-success); font-weight: 500; }
.stat-pending  { color: var(--color-warning); font-weight: 500; }
.stat-skipped  { color: var(--color-error); font-weight: 500; }
.stat-muted    { color: var(--color-text-muted); }

/* Content tabs */
.content-tabs {
  display: flex;
  gap: 0;
  margin-bottom: var(--space-5);
  border-bottom: 1px solid var(--color-border);
}
.content-tab {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-5);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--color-text-muted);
  font-size: var(--type-sm);
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition: all var(--transition-fast);
  margin-bottom: -1px;
}
.content-tab:hover { color: var(--color-text); }
.content-tab.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}
.tab-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  border-radius: 4px;
  padding: 0.15rem 0.4rem;
  font-variant-numeric: tabular-nums;
}
.content-tab.active .tab-count {
  background: var(--color-primary-muted);
  color: var(--color-primary);
}

/* Todo list (tasks) */
.todo-list { display: flex; flex-direction: column; gap: var(--space-1); }
.todo-item {
  display: flex; align-items: flex-start; gap: var(--space-3); font-size: var(--type-sm);
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  transition: border-color var(--transition-fast);
}
.todo-item:hover { border-color: var(--color-border-strong); }
.todo-check { width: var(--space-5); text-align: center; flex-shrink: 0; color: var(--color-text-muted); margin-top: 1px; }
.todo-content { flex: 1; min-width: 0; }
.todo-done { text-decoration: line-through; color: var(--color-text-muted); }
.todo-result { font-size: var(--type-xs); margin-top: var(--space-2); }
.todo-completed .todo-check { color: var(--color-success); }
.todo-failed .todo-check { color: var(--color-error); }
.todo-blocked .todo-check { color: var(--color-error); }
.todo-in_progress .todo-check { color: var(--color-warning); }

.skeleton-header { height: 60px; margin-bottom: var(--space-4); }
.skeleton-content { height: 400px; }
</style>
