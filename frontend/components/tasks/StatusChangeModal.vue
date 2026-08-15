<template>
  <ModalOverlay :labelledby="titleId" :visible="visible" size="lg" flush @close="$emit('cancel')">
    <!-- Header -->
    <ModalHeader :title-id="titleId" @close="$emit('cancel')">
      Move to
            <span :class="['kanban-dot', `dot-${targetStatus}`]"></span>
            {{ statusLabel }}
    </ModalHeader>

    <div class="modal-body">
      <p class="task-info">
        <span class="task-code">{{ task?.code }}</span>
        {{ task?.title }}
      </p>

      <div class="form-field">
        <label class="form-label">Comment (optional)</label>
        <textarea
          ref="commentInput"
          v-model="comment"
          class="form-textarea"
          rows="3"
          placeholder="Add a note about this status change..."
        ></textarea>
      </div>
    </div>

    <!-- Footer -->
    <div class="modal-footer">
      <button class="btn btn-sm" @click="$emit('cancel')">Cancel</button>
      <button class="btn btn-sm btn-primary" @click="onConfirm">Move</button>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
const titleId = useId()
const props = defineProps<{
  visible: boolean
  task: any | null
  targetStatus: string
  statusLabel: string
}>()

const emit = defineEmits<{
  confirm: [comment: string]
  cancel: []
}>()

const comment = ref('')
const commentInput = ref<HTMLTextAreaElement | null>(null)

watch(() => props.visible, (val) => {
  if (val) {
    comment.value = ''
    nextTick(() => commentInput.value?.focus())
  }
})

function onConfirm() {
  emit('confirm', comment.value)
}
</script>

<style scoped>

.kanban-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}
.dot-pending { background: var(--color-text-muted); }
.dot-in_progress { background: var(--color-warning); }
.dot-completed { background: var(--color-success); }
.dot-failed { background: var(--color-error); }

.modal-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  padding: var(--space-5) var(--space-6);
}


.task-info {
  font-size: var(--type-sm);
  color: var(--color-text);
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin: 0;
}
.task-code {
  font-size: var(--type-xs);
  font-weight: var(--weight-bold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.form-textarea {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-size: var(--type-sm);
  font-family: inherit;
  line-height: 1.5;
  resize: vertical;
  min-height: 60px;
}
.form-textarea:focus { outline: 2px solid var(--color-primary); outline-offset: -1px; }
</style>
