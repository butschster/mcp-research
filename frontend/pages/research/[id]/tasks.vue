<template>
  <div v-if="pending" class="skeleton-page">
    <div class="skeleton-card skeleton-header"></div>
    <div class="kanban-skeleton">
      <div v-for="i in 4" :key="i" class="skeleton-card skeleton-column"></div>
    </div>
  </div>

  <div v-else-if="research" class="tasks-page">
    <!-- Header -->
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: research.name, to: `/research/${researchSlug}` },
        { label: 'Tasks' }
      ]" />
      <div class="tasks-header">
        <div class="title-with-code">
          <span v-if="research.code" class="short-code">{{ research.code }}</span>
          <h1 class="page-title">Tasks</h1>
          <span class="task-counter">{{ tasks.length }}</span>
        </div>
        <div class="tasks-actions">
          <button class="btn btn-sm btn-primary" @click="showCreateModal = true">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            New task
          </button>
        </div>
      </div>
    </div>

    <!-- Kanban Board -->
    <div class="kanban-board">
      <div
        v-for="col in columns"
        :key="col.status"
        :class="['kanban-column', `kanban-col-${col.status}`]"
        @dragover.prevent="onDragOver($event, col.status)"
        @dragleave="onDragLeave($event)"
        @drop="onDrop($event, col.status)"
      >
        <div class="kanban-column-header">
          <div class="kanban-column-title-row">
            <span :class="['kanban-dot', `dot-${col.status}`]"></span>
            <h3 class="kanban-column-title">{{ col.label }}</h3>
            <span class="kanban-column-count">{{ columnTasks(col.status).length }}</span>
          </div>
        </div>

        <div class="kanban-column-body">
          <div
            v-for="task in columnTasks(col.status)"
            :key="task.id"
            class="kanban-card"
            draggable="true"
            @dragstart="onDragStart($event, task)"
            @dragend="onDragEnd"
            @click="openTaskDetail(task)"
          >
            <div class="kanban-card-top">
              <span class="short-code">{{ task.code }}</span>
              <StatusBadge v-if="task.priority === 'high'" :status="task.priority" />
            </div>
            <div class="kanban-card-title" v-html="renderRefs(task.title, researchSlug)"></div>
          </div>

          <div v-if="!columnTasks(col.status).length" class="kanban-empty">
            No tasks
          </div>
        </div>
      </div>
    </div>

    <!-- Task Detail Modal -->
    <Teleport to="body">
      <div v-if="detailTask" class="modal-overlay" @click.self="detailTask = null">
        <div class="modal-card modal-detail">
          <!-- Header: status + code + close -->
          <div class="detail-header">
            <span class="short-code">{{ detailTask.code }}</span>
            <StatusBadge :status="detailTask.status" />
            <button class="btn-close" @click="detailTask = null">&times;</button>
          </div>

          <!-- Title (click to edit) -->
          <div v-if="editing === 'title'" class="detail-edit-block">
            <input
              ref="editTitleInput"
              v-model="editValues.title"
              class="modal-input detail-title-input"
              @keydown.enter="saveField('title')"
              @keydown.escape="editing = null"
            />
            <div class="detail-edit-actions">
              <button class="btn btn-sm btn-primary" @click="saveField('title')">Save</button>
              <button class="btn btn-sm" @click="editing = null">Cancel</button>
            </div>
          </div>
          <h3
            v-else
            class="detail-title detail-editable"
            @dblclick="startEditing('title')"
            v-html="renderRefs(detailTask.title, researchSlug)"
          ></h3>

          <!-- Properties row -->
          <div class="detail-props">
            <div class="detail-prop">
              <label class="detail-prop-label">Priority</label>
              <div class="detail-priority-selector">
                <button
                  v-for="p in ['low', 'medium', 'high']"
                  :key="p"
                  :class="['priority-chip', `priority-${p}`, { active: detailTask.priority === p }]"
                  @click="updatePriority(p)"
                >{{ p }}</button>
              </div>
            </div>
            <div v-if="detailTask.created_at" class="detail-prop">
              <label class="detail-prop-label">Created</label>
              <span class="detail-prop-value">{{ new Date(detailTask.created_at).toLocaleDateString() }}</span>
            </div>
            <div v-if="detailTask.completed_at" class="detail-prop">
              <label class="detail-prop-label">Completed</label>
              <span class="detail-prop-value">{{ new Date(detailTask.completed_at).toLocaleDateString() }}</span>
            </div>
          </div>

          <!-- Description -->
          <div class="detail-section">
            <label class="detail-section-label">Description</label>
            <div v-if="editing === 'description'" class="detail-edit-block">
              <textarea
                ref="editDescInput"
                v-model="editValues.description"
                class="modal-textarea"
                rows="4"
                @keydown.escape="editing = null"
              ></textarea>
              <div class="detail-edit-actions">
                <button class="btn btn-sm btn-primary" @click="saveField('description')">Save</button>
                <button class="btn btn-sm" @click="editing = null">Cancel</button>
              </div>
            </div>
            <div
              v-else-if="detailTask.description"
              class="markdown-content detail-editable"
              @dblclick="startEditing('description')"
              v-html="renderRefs(marked.parse(detailTask.description) as string, researchSlug)"
            ></div>
            <div
              v-else
              class="detail-text-empty detail-editable"
              @dblclick="startEditing('description')"
            >Click to add description...</div>
          </div>

          <!-- Result / Comments -->
          <div class="detail-section">
            <label class="detail-section-label">Result / Comments</label>
            <div v-if="editing === 'result'" class="detail-edit-block">
              <textarea
                ref="editResultInput"
                v-model="editValues.result"
                class="modal-textarea"
                rows="4"
                @keydown.escape="editing = null"
              ></textarea>
              <div class="detail-edit-actions">
                <button class="btn btn-sm btn-primary" @click="saveField('result')">Save</button>
                <button class="btn btn-sm" @click="editing = null">Cancel</button>
              </div>
            </div>
            <div
              v-else-if="detailTask.result"
              class="markdown-content detail-editable"
              @dblclick="startEditing('result')"
              v-html="renderRefs(marked.parse(detailTask.result) as string, researchSlug)"
            ></div>
            <div
              v-else
              class="detail-text-empty detail-editable"
              @dblclick="startEditing('result')"
            >Click to add result...</div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Status Change Modal -->
    <Teleport to="body">
      <div v-if="statusModal.visible" class="modal-overlay" @click.self="cancelStatusChange">
        <div class="modal-card">
          <h3 class="modal-title">
            Move to <span :class="['kanban-dot', `dot-${statusModal.targetStatus}`]"></span> {{ statusLabel(statusModal.targetStatus) }}
          </h3>
          <p class="modal-subtitle">
            <span class="short-code">{{ statusModal.task?.code }}</span>
            {{ statusModal.task?.title }}
          </p>
          <label class="modal-label">Comment (optional)</label>
          <textarea
            ref="commentInput"
            v-model="statusModal.comment"
            class="modal-textarea"
            rows="3"
            placeholder="Add a note about this status change..."
          ></textarea>
          <div class="modal-actions">
            <button class="btn btn-sm" @click="cancelStatusChange">Cancel</button>
            <button class="btn btn-sm btn-primary" @click="confirmStatusChange">Move</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Create Task Modal -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
        <div class="modal-card">
          <h3 class="modal-title">New Task</h3>
          <label class="modal-label">Title</label>
          <input
            ref="createTitleInput"
            v-model="newTask.title"
            class="modal-input"
            placeholder="Task title..."
            @keydown.enter="createTask"
          />
          <label class="modal-label">Description (optional)</label>
          <textarea
            v-model="newTask.description"
            class="modal-textarea"
            rows="3"
            placeholder="Task description..."
          ></textarea>
          <label class="modal-label">Priority</label>
          <div class="modal-priority-row">
            <button
              v-for="p in ['low', 'medium', 'high']"
              :key="p"
              :class="['btn btn-sm', newTask.priority === p ? 'btn-primary' : '']"
              @click="newTask.priority = p"
            >{{ p }}</button>
          </div>
          <div class="modal-actions">
            <button class="btn btn-sm" @click="showCreateModal = false">Cancel</button>
            <button class="btn btn-sm btn-primary" :disabled="!newTask.title.trim()" @click="createTask">Create</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>

  <EmptyState v-else icon="&#x1F50D;" title="Research not found" />
</template>

<script setup lang="ts">
import { marked } from 'marked'
marked.setOptions({ gfm: true, breaks: true })

const route = useRoute()
const id = route.params.id as string

// Research data
const { data: researchData, pending } = await useApi<{ data: any }>(`/api/researches/${id}`)
const research = computed(() => researchData.value?.data?.research)
const researchSlug = computed(() => research.value?.code || id)

// Tasks
const { data: tasksData } = await useApi<{ data: any[] }>(`/api/researches/${id}/tasks`)
const tasks = computed(() => tasksData.value?.data ?? [])

// Kanban columns
const columns = [
  { status: 'pending', label: 'Todo' },
  { status: 'in_progress', label: 'In Progress' },
  { status: 'completed', label: 'Done' },
  { status: 'failed', label: 'Rejected' },
]

function statusLabel(status: string): string {
  return columns.find(c => c.status === status)?.label ?? status
}

function columnTasks(status: string): any[] {
  if (status === 'pending') {
    return tasks.value.filter((t: any) => t.status === 'pending' || t.status === 'blocked' || t.status === 'deferred')
  }
  return tasks.value.filter((t: any) => t.status === status)
}

// Task detail modal
const detailTask = ref<any>(null)
const editing = ref<string | null>(null)
const editValues = reactive({ title: '', description: '', result: '' })
const editTitleInput = ref<HTMLInputElement | null>(null)
const editDescInput = ref<HTMLTextAreaElement | null>(null)
const editResultInput = ref<HTMLTextAreaElement | null>(null)

function openTaskDetail(task: any) {
  detailTask.value = task
  editing.value = null
}

function startEditing(field: string) {
  editValues.title = detailTask.value?.title ?? ''
  editValues.description = detailTask.value?.description ?? ''
  editValues.result = detailTask.value?.result ?? ''
  editing.value = field
  nextTick(() => {
    if (field === 'title') editTitleInput.value?.focus()
    else if (field === 'description') editDescInput.value?.focus()
    else if (field === 'result') editResultInput.value?.focus()
  })
}

async function saveField(field: string) {
  if (!detailTask.value) return
  const body: Record<string, any> = {}
  body[field] = (editValues as any)[field]
  await authFetch(`${rtBase}/api/tasks/${detailTask.value.id}`, { method: 'PUT', body })
  tasksData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/tasks`)
  // Refresh detailTask reference
  detailTask.value = tasks.value.find((t: any) => t.id === detailTask.value.id) ?? null
  editing.value = null
}

async function updatePriority(priority: string) {
  if (!detailTask.value || detailTask.value.priority === priority) return
  await authFetch(`${rtBase}/api/tasks/${detailTask.value.id}`, {
    method: 'PUT',
    body: { priority },
  })
  tasksData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/tasks`)
  detailTask.value = tasks.value.find((t: any) => t.id === detailTask.value.id) ?? null
}

// Drag & drop
const draggedTask = ref<any>(null)

function onDragStart(e: DragEvent, task: any) {
  draggedTask.value = task
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', task.id)
  }
  ;(e.target as HTMLElement).classList.add('dragging')
}

function onDragEnd(e: DragEvent) {
  draggedTask.value = null
  ;(e.target as HTMLElement).classList.remove('dragging')
  document.querySelectorAll('.kanban-column').forEach(el => el.classList.remove('drag-over'))
}

function onDragOver(e: DragEvent, status: string) {
  if (!draggedTask.value) return
  const col = (e.currentTarget as HTMLElement)
  col.classList.add('drag-over')
}

function onDragLeave(e: DragEvent) {
  const col = (e.currentTarget as HTMLElement)
  if (!col.contains(e.relatedTarget as Node)) {
    col.classList.remove('drag-over')
  }
}

function onDrop(e: DragEvent, targetStatus: string) {
  document.querySelectorAll('.kanban-column').forEach(el => el.classList.remove('drag-over'))
  if (!draggedTask.value) return

  const task = draggedTask.value
  const currentStatus = task.status

  // Map blocked/deferred to pending column
  const currentColumn = (currentStatus === 'blocked' || currentStatus === 'deferred') ? 'pending' : currentStatus

  if (currentColumn === targetStatus) return

  // Show modal for comment
  statusModal.visible = true
  statusModal.task = task
  statusModal.targetStatus = targetStatus
  statusModal.comment = ''
  nextTick(() => commentInput.value?.focus())
}

// Status change modal
const commentInput = ref<HTMLTextAreaElement | null>(null)
const statusModal = reactive({
  visible: false,
  task: null as any,
  targetStatus: '' as string,
  comment: '',
})

const { authFetch } = useAuth()
const rtBase = useRuntimeConfig().public.apiBase || ''

function cancelStatusChange() {
  statusModal.visible = false
  statusModal.task = null
  statusModal.comment = ''
}

async function confirmStatusChange() {
  if (!statusModal.task) return

  const body: Record<string, any> = { status: statusModal.targetStatus }

  // Append comment to result if provided
  if (statusModal.comment.trim()) {
    const existing = statusModal.task.result || ''
    const commentBlock = `**[${statusLabel(statusModal.targetStatus)}]** ${statusModal.comment.trim()}`
    body.result = existing ? `${existing}\n\n${commentBlock}` : commentBlock
  }

  await authFetch(`${rtBase}/api/tasks/${statusModal.task.id}`, {
    method: 'PUT',
    body,
  })

  tasksData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/tasks`)
  cancelStatusChange()
}

// Create task
const showCreateModal = ref(false)
const createTitleInput = ref<HTMLInputElement | null>(null)
const newTask = reactive({ title: '', description: '', priority: 'medium' })

watch(showCreateModal, (val) => {
  if (val) {
    newTask.title = ''
    newTask.description = ''
    newTask.priority = 'medium'
    nextTick(() => createTitleInput.value?.focus())
  }
})

async function createTask() {
  if (!newTask.title.trim()) return

  await authFetch(`${rtBase}/api/tasks`, {
    method: 'POST',
    body: {
      research_id: id,
      title: newTask.title.trim(),
      description: newTask.description.trim(),
      priority: newTask.priority,
    },
  })

  tasksData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/tasks`)
  showCreateModal.value = false
}

// Real-time updates
useRealtimeUpdates(async (event) => {
  if (event.research_id && event.research_id !== id) return
  if (event.entity === 'task') {
    tasksData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/tasks`)
  }
})
</script>

<style scoped>
/* Use default container width */

/* Reduce page-header spacing on tasks page */
.page-header {
  margin-bottom: var(--space-3);
  padding-bottom: var(--space-2);
}

/* Header */
.tasks-header { display: flex; justify-content: space-between; align-items: center; gap: var(--space-4); }
.tasks-actions { display: flex; align-items: center; gap: var(--space-3); }
.title-with-code { display: flex; align-items: center; gap: var(--space-3); }
.short-code {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
.task-counter {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  font-variant-numeric: tabular-nums;
}

/* Kanban Board */
.kanban-board {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
  align-items: start;
}

.kanban-column {
  background: var(--color-surface);
  border-radius: var(--radius);
  border: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  transition: border-color var(--transition-base), box-shadow var(--transition-base);
}

.kanban-column.drag-over {
  border-color: var(--color-primary);
  box-shadow: inset 0 0 0 1px var(--color-primary), 0 0 20px rgba(108, 197, 224, 0.08);
}

.kanban-column-header {
  padding: var(--space-4) var(--space-4) var(--space-3);
  border-bottom: 1px solid var(--color-border);
}

.kanban-column-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.kanban-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.dot-pending { background: var(--color-text-muted); }
.dot-in_progress { background: var(--color-warning); }
.dot-completed { background: var(--color-success); }
.dot-failed { background: var(--color-error); }

.kanban-column-title {
  font-size: var(--type-sm);
  font-weight: 600;
  letter-spacing: -0.01em;
  flex: 1;
}

.kanban-column-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  font-variant-numeric: tabular-nums;
}

.kanban-column-body {
  padding: var(--space-3);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}

/* Kanban Card — compact: code + title only */
.kanban-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: var(--space-3);
  cursor: grab;
  transition: border-color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  user-select: none;
}
.kanban-card:hover {
  border-color: var(--color-border-strong);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}
.kanban-card.dragging {
  opacity: 0.4;
  transform: scale(0.95);
}
.kanban-card:active { cursor: grabbing; }

.kanban-card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-1);
}

.kanban-card-title {
  font-size: var(--type-sm);
  font-weight: 500;
  line-height: 1.4;
  word-break: break-word;
}

.kanban-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  opacity: 0.5;
  min-height: 80px;
}

/* Task Detail Modal */

.detail-header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-4);
}

.btn-close {
  margin-left: auto;
  background: none;
  border: none;
  color: var(--color-text-muted);
  font-size: var(--type-xl);
  cursor: pointer;
  padding: var(--space-1) var(--space-2);
  line-height: 1;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}
.btn-close:hover { color: var(--color-text); background: var(--color-surface-hover); }

/* Title */
.detail-title {
  font-size: var(--type-xl);
  font-weight: 600;
  letter-spacing: -0.02em;
  line-height: 1.35;
  margin-bottom: var(--space-6);
}
.detail-title-input {
  font-size: var(--type-xl);
  font-weight: 600;
  letter-spacing: -0.02em;
  margin-bottom: var(--space-2);
}

/* Editable fields */
.detail-editable {
  cursor: pointer;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  transition: all var(--transition-fast);
  margin: calc(-1 * var(--space-2)) calc(-1 * var(--space-3));
}
.detail-editable:hover {
  border-color: var(--color-border);
  background: var(--color-surface-hover);
}

.detail-edit-block {
  margin-bottom: var(--space-2);
}
.detail-edit-actions {
  display: flex;
  gap: var(--space-2);
  margin-top: var(--space-2);
}

.detail-text-empty {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  font-style: italic;
  opacity: 0.6;
}

/* Properties row */
.detail-props {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-5);
  padding: var(--space-4);
  margin-top: var(--space-4);
  background: var(--color-bg);
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border);
  margin-bottom: var(--space-5);
}

.detail-prop {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.detail-prop-label {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-muted);
}

.detail-prop-value {
  font-size: var(--type-sm);
  color: var(--color-text);
}

/* Priority selector */
.detail-priority-selector {
  display: flex;
  gap: var(--space-1);
}

.priority-chip {
  font-size: 0.75rem;
  font-weight: 500;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  border: 1px solid var(--color-border);
  background: none;
  color: var(--color-text-muted);
  cursor: pointer;
  text-transform: capitalize;
  transition: all var(--transition-fast);
  font-family: inherit;
}
.priority-chip:hover { border-color: var(--color-border-strong); color: var(--color-text); }
.priority-chip.active { color: var(--color-text); font-weight: 600; }
.priority-low.active { background: rgba(108, 197, 224, 0.1); border-color: var(--color-primary); color: var(--color-primary); }
.priority-medium.active { background: rgba(240, 184, 73, 0.1); border-color: var(--color-warning); color: var(--color-warning); }
.priority-high.active { background: rgba(239, 107, 107, 0.1); border-color: var(--color-error); color: var(--color-error); }

/* Sections */
.detail-section {
  margin-bottom: var(--space-5);
}

.detail-section-label {
  display: block;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-muted);
  margin-bottom: var(--space-2);
}

/* markdown-content class from main.css handles all markdown rendering */

/* Modals (shared) */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-overlay);
}

.modal-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  width: 100%;
  max-width: 460px;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.3);
}
.modal-card.modal-detail {
  max-width: 720px;
  max-height: 85vh;
  overflow-y: auto;
}

.modal-title {
  font-size: var(--type-lg);
  font-weight: 600;
  letter-spacing: -0.02em;
  margin-bottom: var(--space-3);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.modal-subtitle {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  margin-bottom: var(--space-5);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.modal-label {
  display: block;
  font-size: var(--type-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  margin-bottom: var(--space-2);
}

.modal-input, .modal-textarea {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  background: var(--color-bg);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-size: var(--type-sm);
  font-family: inherit;
  line-height: 1.5;
  margin-bottom: var(--space-4);
}
.modal-textarea { resize: vertical; min-height: 60px; }
.modal-input:focus, .modal-textarea:focus { outline: 2px solid var(--color-primary); outline-offset: -1px; }

.modal-priority-row {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-5);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}

/* Skeleton */
.kanban-skeleton {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
}
.skeleton-column { height: 400px; }

/* Responsive */
@media (max-width: 1024px) {
  .kanban-board {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 640px) {
  .kanban-board {
    grid-template-columns: 1fr;
  }
}
</style>
