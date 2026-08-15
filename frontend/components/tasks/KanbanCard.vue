<template>
  <div
    class="kanban-card"
    role="button"
    tabindex="0"
    :draggable="canWrite"
    @dragstart="$emit('dragstart', $event)"
    @dragend="$emit('dragend', $event)"
    @click="$emit('click')"
    @keydown.enter.prevent="$emit('click')"
    @keydown.space.prevent="$emit('click')"
  >
    <div class="kanban-card-top">
      <span class="short-code">{{ task.code }}</span>
      <StatusBadge v-if="task.priority === 'high'" :status="task.priority" />
    </div>
    <div class="kanban-card-title" v-html="renderRefs(task.title, researchSlug)"></div>
  </div>
</template>

<script setup lang="ts">
import { renderRefs } from '~/composables/useCrossRefs'

// A card that lifts and snaps back is worse than a card that does not lift:
// the drop would 403, after the reader has already committed to the gesture.
//
// Dragging is the pointer's way in; Enter and Space are everyone else's. The
// card is the only route to a task's detail, so without them a keyboard reader
// cannot open a task at all.
const { canWrite } = useResearchRole()

defineProps<{
  task: any
  researchSlug: string
}>()

defineEmits<{
  click: []
  dragstart: [event: DragEvent]
  dragend: [event: DragEvent]
}>()
</script>

<style scoped>
.kanban-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: var(--space-3);
  cursor: grab;
  transition: border-color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  user-select: none;
}
/* It takes focus now, so it has to show it. */
.kanban-card:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
.kanban-card:hover {
  border-color: var(--color-border-strong);
  box-shadow: var(--shadow-1);
}
.kanban-card:deep(.dragging) {
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

.short-code {
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}

.kanban-card-title {
  font-size: var(--type-sm);
  font-weight: var(--weight-medium);
  line-height: 1.4;
  word-break: break-word;
}
</style>
