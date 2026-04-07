<template>
  <ModalOverlay :visible="visible" @close="$emit('cancel')">
    <h3 class="modal-title">
      Move to <span :class="['kanban-dot', `dot-${targetStatus}`]"></span> {{ statusLabel }}
    </h3>
    <p class="modal-subtitle">
      <span class="short-code">{{ task?.code }}</span>
      {{ task?.title }}
    </p>
    <label class="modal-label">Comment (optional)</label>
    <textarea
      ref="commentInput"
      v-model="comment"
      class="modal-textarea"
      rows="3"
      placeholder="Add a note about this status change..."
    ></textarea>
    <div class="modal-actions">
      <button class="btn btn-sm" @click="$emit('cancel')">Cancel</button>
      <button class="btn btn-sm btn-primary" @click="onConfirm">Move</button>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
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

.modal-textarea {
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
  resize: vertical;
  min-height: 60px;
}
.modal-textarea:focus { outline: 2px solid var(--color-primary); outline-offset: -1px; }

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}
</style>
