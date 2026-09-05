<template>
  <ModalOverlay :labelledby="titleId" :visible="visible" size="lg" flush @close="$emit('cancel')">
    <!-- Header -->
    <ModalHeader title="New Task" :title-id="titleId" @close="$emit('cancel')" />

    <div class="modal-body">
      <div class="form-field">
        <label class="form-label">Title</label>
        <input
          ref="createTitleInput"
          v-model="newTask.title"
          class="form-input"
          placeholder="Task title..."
          @keydown.enter="onCreate"
        />
      </div>
      <div class="form-field">
        <label class="form-label">Description (optional)</label>
        <textarea
          v-model="newTask.description"
          class="form-textarea"
          rows="3"
          placeholder="Task description..."
        ></textarea>
      </div>
      <div class="form-field">
        <label class="form-label">Priority</label>
        <div class="priority-selector">
          <button
            v-for="p in ['low', 'medium', 'high']"
            :key="p"
            :class="['priority-chip', `priority-${p}`, { active: newTask.priority === p }]"
            @click="newTask.priority = p"
          >{{ p }}</button>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="modal-footer">
      <button class="btn btn-sm" @click="$emit('cancel')">Cancel</button>
      <button class="btn btn-sm btn-primary" :disabled="!newTask.title.trim()" @click="onCreate">Create</button>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
const titleId = useId()
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

.modal-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  padding: var(--space-5) var(--space-6);
}


.form-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.form-textarea { resize: vertical; min-height: 60px; }

.priority-selector {
  display: flex;
  gap: var(--space-1);
}
.priority-chip {
  font-size: 0.75rem;
  font-weight: var(--weight-medium);
  padding: 0.15rem 0.5rem;
  border-radius: var(--radius-xs);
  border: 1px solid var(--color-border);
  background: none;
  color: var(--color-text-muted);
  cursor: pointer;
  text-transform: capitalize;
  transition: all var(--transition-fast);
  font-family: inherit;
}
.priority-chip:hover { border-color: var(--color-border-strong); color: var(--color-text); }
.priority-chip.active { color: var(--color-text); font-weight: var(--weight-semibold); }
.priority-low.active { background: rgba(var(--color-primary-rgb), 0.1); border-color: var(--color-primary); color: var(--color-primary); }
.priority-medium.active { background: rgba(var(--color-warning-rgb), 0.1); border-color: var(--color-warning); color: var(--color-warning); }
.priority-high.active { background: rgba(var(--color-error-rgb), 0.1); border-color: var(--color-error); color: var(--color-error); }
</style>
