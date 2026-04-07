<template>
  <ModalOverlay :visible="visible" @close="$emit('cancel')">
    <h3 class="modal-title">New Task</h3>
    <label class="modal-label">Title</label>
    <input
      ref="createTitleInput"
      v-model="newTask.title"
      class="modal-input"
      placeholder="Task title..."
      @keydown.enter="onCreate"
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
      <button class="btn btn-sm" @click="$emit('cancel')">Cancel</button>
      <button class="btn btn-sm btn-primary" :disabled="!newTask.title.trim()" @click="onCreate">Create</button>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  create: [data: { title: string; description: string; priority: string }]
  cancel: []
}>()

const createTitleInput = ref<HTMLInputElement | null>(null)
const newTask = reactive({ title: '', description: '', priority: 'medium' })

watch(() => props.visible, (val) => {
  if (val) {
    newTask.title = ''
    newTask.description = ''
    newTask.priority = 'medium'
    nextTick(() => createTitleInput.value?.focus())
  }
})

function onCreate() {
  if (!newTask.title.trim()) return
  emit('create', {
    title: newTask.title.trim(),
    description: newTask.description.trim(),
    priority: newTask.priority,
  })
}
</script>

<style scoped>
.modal-title {
  font-size: var(--type-lg);
  font-weight: 600;
  letter-spacing: -0.02em;
  margin-bottom: var(--space-3);
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
</style>
