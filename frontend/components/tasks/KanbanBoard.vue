<template>
  <div class="kanban-board">
    <div
      v-for="col in columns"
      :key="col.status"
      :class="['kanban-column', `kanban-col-${col.status}`]"
      @dragover="onDragOverGuarded($event, col.status)"
      @dragleave="onDragLeave($event)"
      @drop="canWrite && onDrop($event, col.status)"
    >
      <div class="kanban-column-header">
        <div class="kanban-column-title-row">
          <span :class="['kanban-dot', `dot-${col.status}`]"></span>
          <h3 class="kanban-column-title">{{ col.label }}</h3>
          <span class="kanban-column-count">{{ columnTasks(col.status).length }}</span>
        </div>
      </div>

      <div class="kanban-column-body">
        <TasksKanbanCard
          v-for="task in columnTasks(col.status)"
          :key="task.id"
          :task="task"
          :research-slug="researchSlug"
          @click="$emit('taskClick', task)"
          @dragstart="onDragStart($event, task)"
          @dragend="onDragEnd"
        />

        <div v-if="!columnTasks(col.status).length" class="kanban-empty">
          No tasks
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  columns: Array<{ status: string; label: string }>
  tasks: any[]
  researchSlug: string
}>()

const emit = defineEmits<{
  taskClick: [task: any]
  taskDrop: [task: any, targetStatus: string]
}>()

// The column stops accepting a drop for the same reason the card stops
// lifting: the move is a write.
const { canWrite } = useResearchRole()

const draggedTask = ref<any>(null)

// `.prevent` runs whatever the guard says, and preventDefault on dragover is
// precisely what marks a column as a drop target — so a viewer's columns went
// on inviting a drop they would refuse.
function onDragOverGuarded(event: DragEvent, status: string) {
  if (!canWrite.value) return
  event.preventDefault()
  onDragOver(event, status)
}

function columnTasks(status: string): any[] {
  if (status === 'pending') {
    return props.tasks.filter((t: any) => t.status === 'pending' || t.status === 'blocked' || t.status === 'deferred')
  }
  return props.tasks.filter((t: any) => t.status === status)
}

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

  emit('taskDrop', task, targetStatus)
}
</script>

<style scoped>
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
  font-weight: var(--weight-semibold);
  letter-spacing: -0.01em;
  flex: 1;
}

.kanban-column-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  padding: 0.1rem 0.4rem;
  border-radius: var(--radius-xs);
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
