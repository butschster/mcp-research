<template>
  <div
    class="kanban-card"
    draggable="true"
    @dragstart="$emit('dragstart', $event)"
    @dragend="$emit('dragend', $event)"
    @click="$emit('click')"
  >
    <div class="kanban-card-top">
      <span class="short-code">{{ task.code }}</span>
      <StatusBadge v-if="task.priority === 'high'" :status="task.priority" />
    </div>
    <div class="kanban-card-title" v-html="renderRefs(task.title, researchSlug)"></div>
  </div>
</template>

<script setup lang="ts">
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
.kanban-card:hover {
  border-color: var(--color-border-strong);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
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
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}

.kanban-card-title {
  font-size: var(--type-sm);
  font-weight: 500;
  line-height: 1.4;
  word-break: break-word;
}
</style>
