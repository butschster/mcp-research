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
      <p v-if="session.focus" class="card-meta mt-2" v-html="'Focus: ' + renderRefs(session.focus, researchSlug)"></p>
    </div>

    <!-- Notes -->
    <div v-if="session.notes" class="card notes-card">
      <h3 class="card-section-title">Session notes</h3>
      <div class="notes-text markdown-content" v-html="renderRefs(marked.parse(session.notes) as string, researchSlug)"></div>
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
      <div class="panel-progress">
        <ProgressBar :value="progress.answered" :total="progress.total" />
        <div class="progress-stats">
          <span class="stat-answered">{{ progress.answered }} answered</span>
          <span class="stat-pending">{{ progress.pending }} pending</span>
          <span v-if="progress.deferred" class="stat-muted">{{ progress.deferred }} deferred</span>
          <span v-if="progress.skipped" class="stat-skipped">{{ progress.skipped }} skipped</span>
        </div>
      </div>

      <!-- Add question -->
      <button v-if="!showAddQuestion" class="btn btn-sm add-btn" @click="showAddQuestion = true">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        Add question
      </button>
      <form v-else class="add-form" @submit.prevent="submitQuestion">
        <input v-model="newQuestion.text" class="add-input" placeholder="Question text..." required autofocus />
        <div class="add-form-row">
          <input v-model="newQuestion.area" class="add-input add-input-sm" placeholder="Area (optional)" />
          <select v-model="newQuestion.priority" class="add-select">
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="low">Low</option>
          </select>
          <button type="submit" class="btn btn-sm" :disabled="addingQuestion || !newQuestion.text.trim()">
            {{ addingQuestion ? 'Adding...' : 'Add' }}
          </button>
          <button type="button" class="btn btn-sm" @click="showAddQuestion = false">Cancel</button>
        </div>
      </form>

      <QuestionList :questions="questions" :research-slug="researchSlug" :session-id="sessionId" />
    </div>

    <!-- Tasks panel -->
    <div v-if="activeTab === 'tasks'">
      <div v-if="tasks.length" class="panel-progress">
        <ProgressBar :value="completedTasks" :total="tasks.length" />
        <div class="progress-stats">
          <span class="stat-answered">{{ completedTasks }} completed</span>
          <span class="stat-pending">{{ tasks.length - completedTasks }} remaining</span>
        </div>
      </div>

      <!-- Add task -->
      <button v-if="!showAddTask" class="btn btn-sm add-btn" @click="showAddTask = true">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        Add task
      </button>
      <form v-else class="add-form" @submit.prevent="submitTask">
        <input v-model="newTask.title" class="add-input" placeholder="Task title..." required autofocus />
        <div class="add-form-row">
          <input v-model="newTask.description" class="add-input add-input-sm" placeholder="Description (optional)" />
          <select v-model="newTask.priority" class="add-select">
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="low">Low</option>
          </select>
          <button type="submit" class="btn btn-sm" :disabled="addingTask || !newTask.title.trim()">
            {{ addingTask ? 'Adding...' : 'Add' }}
          </button>
          <button type="button" class="btn btn-sm" @click="showAddTask = false">Cancel</button>
        </div>
      </form>

      <!-- Task groups -->
      <div v-if="tasks.length">
        <template v-for="group in taskGroups" :key="group.status">
          <div v-if="group.tasks.length" class="task-group">
            <button class="btn-ghost group-header" @click="toggleTaskGroup(group.status)">
              <StatusBadge :status="group.status" />
              <span class="group-count">{{ group.tasks.length }}</span>
              <span class="group-chevron" :class="{ open: taskGroupOpen[group.status] }">&rsaquo;</span>
            </button>
            <div v-show="taskGroupOpen[group.status]" class="task-group-body">
              <div v-for="t in group.tasks" :key="t.id" class="todo-item">
                <span :class="['todo-check', `todo-${t.status}`]">
                  <svg v-if="t.status === 'completed'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                  <svg v-else-if="t.status === 'failed'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
                  <svg v-else-if="t.status === 'blocked'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/></svg>
                  <svg v-else-if="t.status === 'deferred'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                  <svg v-else-if="t.status === 'in_progress'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v4"/><path d="M12 18v4"/><path d="m4.93 4.93 2.83 2.83"/><path d="m16.24 16.24 2.83 2.83"/><path d="M2 12h4"/><path d="M18 12h4"/><path d="m4.93 19.07 2.83-2.83"/><path d="m16.24 7.76 2.83-2.83"/></svg>
                  <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/></svg>
                </span>
                <div class="todo-content">
                  <span :class="['todo-text', { 'todo-done': t.status === 'completed' }]" v-html="renderRefs(t.title, researchSlug)"></span>
                  <div v-if="t.description && t.status !== 'completed'" class="todo-desc" v-html="renderRefs(t.description, researchSlug)"></div>
                  <div v-if="t.result" class="todo-result markdown-content" v-html="renderRefs(marked.parse(t.result) as string, researchSlug)"></div>
                </div>
                <StatusBadge v-if="t.priority === 'high'" :status="t.priority" />
              </div>
            </div>
          </div>
        </template>
      </div>
      <EmptyState v-else-if="!showAddTask" icon="&#x2705;" title="No tasks" description="Add a task or let the AI create them." />
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
const researchId = computed(() => researchData.value?.data?.research?.id || id)

const { data, pending } = await useApi<{ data: any }>(`/api/researches/${id}/sessions/${sessionId}`)

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

// Task groups
const TASK_STATUS_ORDER = ['in_progress', 'pending', 'blocked', 'deferred', 'failed', 'completed']
const taskGroupOpen = ref<Record<string, boolean>>({
  pending: true,
  in_progress: true,
  blocked: true,
  deferred: true,
  failed: true,
  completed: false,
})
function toggleTaskGroup(status: string) {
  taskGroupOpen.value[status] = !taskGroupOpen.value[status]
}
const taskGroups = computed(() =>
  TASK_STATUS_ORDER
    .map(s => ({ status: s, tasks: tasks.value.filter((t: any) => t.status === s) }))
    .filter(g => g.tasks.length > 0)
)

// Active tab
const activeTab = ref<'questions' | 'tasks'>('questions')

// Add task
const showAddTask = ref(false)
const addingTask = ref(false)
const newTask = ref({ title: '', description: '', priority: 'medium' })
const { authFetch } = useAuth()
const rtBase = useRuntimeConfig().public.apiBase || ''

async function submitTask() {
  if (!newTask.value.title.trim()) return
  addingTask.value = true
  try {
    await authFetch(`${rtBase}/api/tasks`, {
      method: 'POST',
      body: JSON.stringify({
        research_id: researchId.value,
        title: newTask.value.title,
        description: newTask.value.description,
        priority: newTask.value.priority,
      }),
      headers: { 'Content-Type': 'application/json' },
    })
    newTask.value = { title: '', description: '', priority: 'medium' }
    showAddTask.value = false
    tasksData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/tasks`)
  } catch { /* ignore */ }
  addingTask.value = false
}

// Add question
const showAddQuestion = ref(false)
const addingQuestion = ref(false)
const newQuestion = ref({ text: '', area: '', priority: 'medium' })
const realSessionId = computed(() => session.value?.id || sessionId)

async function submitQuestion() {
  if (!newQuestion.value.text.trim()) return
  addingQuestion.value = true
  try {
    await authFetch(`${rtBase}/api/sessions/${realSessionId.value}/questions`, {
      method: 'POST',
      body: JSON.stringify({
        questions: [{
          text: newQuestion.value.text,
          area: newQuestion.value.area,
          priority: newQuestion.value.priority,
        }],
      }),
      headers: { 'Content-Type': 'application/json' },
    })
    newQuestion.value = { text: '', area: '', priority: 'medium' }
    showAddQuestion.value = false
    data.value = await authFetch<any>(`${rtBase}/api/researches/${id}/sessions/${sessionId}`)
  } catch { /* ignore */ }
  addingQuestion.value = false
}

// Real-time updates
useRealtimeUpdates(async (event) => {
  if (event.research_id && event.research_id !== id) return
  if (['question', 'session'].includes(event.entity)) {
    data.value = await authFetch<any>(`${rtBase}/api/researches/${id}/sessions/${sessionId}`)
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

/* Panel progress */
.panel-progress { margin-bottom: var(--space-5); }
.progress-stats {
  display: flex; gap: var(--space-4); font-size: var(--type-xs);
  margin-top: var(--space-2); color: var(--color-text-muted);
}
.stat-answered { color: var(--color-success); font-weight: 500; }
.stat-pending  { color: var(--color-warning); font-weight: 500; }
.stat-skipped  { color: var(--color-error); font-weight: 500; }
.stat-muted    { color: var(--color-text-muted); }

/* Tabs */
.content-tabs {
  display: flex; gap: 0; margin-bottom: var(--space-5);
  border-bottom: 1px solid var(--color-border);
}
.content-tab {
  display: flex; align-items: center; gap: var(--space-2);
  padding: var(--space-3) var(--space-5);
  background: none; border: none; border-bottom: 2px solid transparent;
  color: var(--color-text-muted); font-size: var(--type-sm); font-weight: 500;
  font-family: inherit; cursor: pointer; transition: all var(--transition-fast);
  margin-bottom: -1px;
}
.content-tab:hover { color: var(--color-text); }
.content-tab.active { color: var(--color-primary); border-bottom-color: var(--color-primary); }
.tab-count {
  font-size: var(--type-xs); color: var(--color-text-muted);
  background: var(--color-surface-hover); border-radius: 4px;
  padding: 0.15rem 0.4rem; font-variant-numeric: tabular-nums;
}
.content-tab.active .tab-count { background: var(--color-primary-muted); color: var(--color-primary); }

/* Add form */
.add-btn { margin-bottom: var(--space-4); }
.add-form { margin-bottom: var(--space-5); }
.add-input {
  width: 100%; padding: var(--space-2) var(--space-3);
  background: var(--color-surface); border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-sm); color: var(--color-text);
  font-size: var(--type-sm); font-family: inherit;
  margin-bottom: var(--space-2);
}
.add-input:focus { outline: 2px solid var(--color-primary); outline-offset: -1px; }
.add-input-sm { flex: 1; margin-bottom: 0; }
.add-form-row { display: flex; gap: var(--space-2); align-items: center; }
.add-select {
  padding: var(--space-2) var(--space-3);
  background: var(--color-surface); border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-sm); color: var(--color-text-muted);
  font-size: var(--type-sm); font-family: inherit; cursor: pointer;
}

/* Task groups */
.task-group { margin-bottom: var(--space-3); }
.group-header {
  gap: var(--space-2); width: 100%;
  padding: var(--space-2); border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
}
.group-header:hover { background: var(--color-surface-hover); }
.group-count {
  font-size: var(--type-xs); color: var(--color-text-muted);
  background: var(--color-surface-hover); border-radius: 4px;
  padding: 0.15rem 0.375rem; font-weight: 600;
  min-width: 1.25rem; text-align: center; font-variant-numeric: tabular-nums;
}
.group-chevron {
  margin-left: auto; font-size: var(--type-lg);
  color: var(--color-text-muted); transition: transform var(--transition-base);
  display: inline-block;
}
.group-chevron.open { transform: rotate(90deg); }
.task-group-body {
  display: flex; flex-direction: column; gap: var(--space-2);
  margin-top: var(--space-2); margin-left: var(--space-2);
}

/* Task items */
.todo-item {
  display: flex; align-items: flex-start; gap: var(--space-3); font-size: var(--type-sm);
  padding: var(--space-3) var(--space-4);
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  transition: border-color var(--transition-fast);
}
.todo-item:hover { border-color: var(--color-border-strong); }
.todo-check { display: flex; align-items: center; justify-content: center; width: 16px; flex-shrink: 0; color: var(--color-text-muted); margin-top: 2px; }
.todo-content { flex: 1; min-width: 0; }
.todo-done { text-decoration: line-through; color: var(--color-text-muted); }
.todo-desc { font-size: var(--type-xs); color: var(--color-text-muted); margin-top: var(--space-1); }
.todo-result { font-size: var(--type-xs); margin-top: var(--space-2); }
.todo-completed .todo-check { color: var(--color-success); }
.todo-failed .todo-check { color: var(--color-error); }
.todo-blocked .todo-check { color: var(--color-error); }
.todo-in_progress .todo-check { color: var(--color-warning); }

.skeleton-header { height: 60px; margin-bottom: var(--space-4); }
.skeleton-content { height: 400px; }
</style>
