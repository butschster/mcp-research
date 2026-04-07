<template>
  <ModalOverlay :visible="!!task" size="lg" @close="$emit('close')">
    <template v-if="task">
      <!-- Header: status + code + close -->
      <div class="detail-header">
        <span class="short-code">{{ task.code }}</span>
        <StatusBadge :status="task.status" />
        <button class="btn-close" @click="$emit('close')">&times;</button>
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
        v-html="renderRefs(task.title, researchSlug)"
      ></h3>

      <!-- Properties row -->
      <div class="detail-props">
        <div class="detail-prop">
          <label class="detail-prop-label">Priority</label>
          <div class="detail-priority-selector">
            <button
              v-for="p in ['low', 'medium', 'high']"
              :key="p"
              :class="['priority-chip', `priority-${p}`, { active: task.priority === p }]"
              @click="$emit('updatePriority', p)"
            >{{ p }}</button>
          </div>
        </div>
        <div v-if="task.created_at" class="detail-prop">
          <label class="detail-prop-label">Created</label>
          <span class="detail-prop-value">{{ new Date(task.created_at).toLocaleDateString() }}</span>
        </div>
        <div v-if="task.completed_at" class="detail-prop">
          <label class="detail-prop-label">Completed</label>
          <span class="detail-prop-value">{{ new Date(task.completed_at).toLocaleDateString() }}</span>
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
          v-else-if="task.description"
          class="markdown-content detail-editable"
          @dblclick="startEditing('description')"
          v-html="renderRefs(marked.parse(task.description) as string, researchSlug)"
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
          v-else-if="task.result"
          class="markdown-content detail-editable"
          @dblclick="startEditing('result')"
          v-html="renderRefs(marked.parse(task.result) as string, researchSlug)"
        ></div>
        <div
          v-else
          class="detail-text-empty detail-editable"
          @dblclick="startEditing('result')"
        >Click to add result...</div>
      </div>
    </template>
  </ModalOverlay>
</template>

<script setup lang="ts">
import { marked } from 'marked'
import { renderRefs } from '~/composables/useCrossRefs'
marked.setOptions({ gfm: true, breaks: true })

const props = defineProps<{
  task: any | null
  researchSlug: string
}>()

const emit = defineEmits<{
  close: []
  save: [field: string, value: string]
  updatePriority: [priority: string]
}>()

const editing = ref<string | null>(null)
const editValues = reactive({ title: '', description: '', result: '' })
const editTitleInput = ref<HTMLInputElement | null>(null)
const editDescInput = ref<HTMLTextAreaElement | null>(null)
const editResultInput = ref<HTMLTextAreaElement | null>(null)

// Reset editing state when task changes
watch(() => props.task, () => {
  editing.value = null
})

function startEditing(field: string) {
  editValues.title = props.task?.title ?? ''
  editValues.description = props.task?.description ?? ''
  editValues.result = props.task?.result ?? ''
  editing.value = field
  nextTick(() => {
    if (field === 'title') editTitleInput.value?.focus()
    else if (field === 'description') editDescInput.value?.focus()
    else if (field === 'result') editResultInput.value?.focus()
  })
}

function saveField(field: string) {
  if (!props.task) return
  emit('save', field, (editValues as any)[field])
  editing.value = null
}
</script>

<style scoped>
.detail-header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-4);
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

/* Modal inputs (needed since ModalOverlay provides the card wrapper) */
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
</style>
