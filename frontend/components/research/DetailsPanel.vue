<template>
  <ModalOverlay :visible="open" size="lg" flush @close="emit('update:open', false)">
    <!-- Header -->
    <div class="modal-header">
      <h3 class="modal-title">Research Details</h3>
      <button class="modal-close" @click="emit('update:open', false)">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
      </button>
    </div>

    <div class="modal-body">
      <!-- Overview -->
      <section class="section">
        <h4 class="section-title">Overview</h4>
        <div class="field-group">
          <div class="field" @dblclick="startEdit('goal')">
            <div class="field-header">
              <label class="field-label">Goal</label>
              <button v-if="canWrite && editingField !== 'goal'" class="field-edit-btn" @click="startEdit('goal')" title="Edit">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
            </div>
            <div v-if="editingField !== 'goal'" class="field-value" :class="{ 'field-empty': !research.goal }">
              {{ research.goal || 'Click the pencil to set a goal' }}
            </div>
            <div v-else class="field-edit">
              <input v-model="editValue" class="field-input" placeholder="What is this research trying to achieve?" @keydown.enter="saveEdit('goal')" @keydown.escape="cancelEdit" ref="editInput" />
              <div class="field-edit-actions">
                <button class="btn btn-sm btn-primary" @click="saveEdit('goal')">Save</button>
                <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
              </div>
            </div>
          </div>

          <div class="field" @dblclick="startEdit('description')">
            <div class="field-header">
              <label class="field-label">Description</label>
              <button v-if="canWrite && editingField !== 'description'" class="field-edit-btn" @click="startEdit('description')" title="Edit">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
            </div>
            <div v-if="editingField !== 'description'" class="field-value" :class="{ 'field-empty': !research.description }">
              {{ research.description || 'Click the pencil to add a description' }}
            </div>
            <div v-else class="field-edit">
              <textarea v-model="editValue" class="field-textarea" rows="3" placeholder="Describe what this research covers..." @keydown.escape="cancelEdit" ref="editInput"></textarea>
              <div class="field-edit-actions">
                <button class="btn btn-sm btn-primary" @click="saveEdit('description')">Save</button>
                <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
              </div>
            </div>
          </div>

          <div class="field" @dblclick="startEdit('tags')">
            <div class="field-header">
              <label class="field-label">Tags</label>
              <button v-if="canWrite && editingField !== 'tags'" class="field-edit-btn" @click="startEdit('tags')" title="Edit">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
            </div>
            <div v-if="editingField !== 'tags'">
              <div v-if="research.tags?.length" class="tags-row">
                <span v-for="tag in research.tags" :key="tag" :class="['tag', `tag-hue-${tagHue(tag)}`]">{{ tag }}</span>
              </div>
              <div v-else class="field-value field-empty">No tags yet</div>
            </div>
            <div v-else class="field-edit">
              <input v-model="editValue" class="field-input" placeholder="tag1, tag2, tag3" @keydown.enter="saveEdit('tags')" @keydown.escape="cancelEdit" ref="editInput" />
              <span class="field-hint">Comma-separated</span>
              <div class="field-edit-actions">
                <button class="btn btn-sm btn-primary" @click="saveEdit('tags')">Save</button>
                <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- AI Instruction -->
      <section class="section">
        <h4 class="section-title">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2a4 4 0 0 0-4 4c0 2 2 3 2 6H14c0-3 2-4 2-6a4 4 0 0 0-4-4z"/><line x1="10" y1="16" x2="14" y2="16"/><line x1="10" y1="20" x2="14" y2="20"/><line x1="11" y1="24" x2="13" y2="24"/></svg>
          AI Instruction
        </h4>
        <div class="field-group">
          <div class="field" @dblclick="startEdit('instruction')">
            <div class="field-header">
              <label class="field-label">System prompt for the AI assistant</label>
              <button v-if="canWrite && editingField !== 'instruction'" class="field-edit-btn" @click="startEdit('instruction')" title="Edit">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
            </div>
            <div v-if="editingField !== 'instruction'" class="field-value field-value-pre" :class="{ 'field-empty': !research.instruction }">
              {{ research.instruction || 'No instruction set. The AI will use default behavior.' }}
            </div>
            <div v-else class="field-edit">
              <textarea v-model="editValue" class="field-textarea field-textarea-tall" rows="6" placeholder="e.g. Be concise and technical. Focus on practical examples..." @keydown.escape="cancelEdit" ref="editInput"></textarea>
              <div class="field-edit-actions">
                <button class="btn btn-sm btn-primary" @click="saveEdit('instruction')">Save</button>
                <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Memory -->
      <section class="section">
        <h4 class="section-title">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
          Memory
          <span v-if="research.memory?.length" class="section-count">{{ research.memory.length }}</span>
        </h4>
        <p class="section-description">Context notes accumulated by the AI during research sessions.</p>
        <div v-if="research.memory?.length" class="memory-list">
          <div v-for="(item, i) in research.memory" :key="i" class="memory-item">
            <span class="memory-index">{{ i + 1 }}</span>
            <span class="memory-text">{{ item }}</span>
          </div>
        </div>
        <div v-else class="field-value field-empty">
          No memory entries yet. The AI will add context notes as the research progresses.
        </div>
      </section>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
import { tagHue } from '~/composables/useTagHue'

const props = defineProps<{
  research: any
  open: boolean
}>()

// A viewer sees the same fields, without the pencils. The panel is a display of
// the research either way; only the way in changes.
const { canWrite } = useResearchRole()

const emit = defineEmits<{
  save: [field: string, value: any]
  'update:open': [value: boolean]
}>()

const editingField = ref<string | null>(null)
const editValue = ref('')
const editInput = ref<HTMLElement | null>(null)

function startEdit(field: string) {
  if (field === 'tags') {
    editValue.value = (props.research?.tags ?? []).join(', ')
  } else {
    editValue.value = props.research?.[field] ?? ''
  }
  editingField.value = field
  nextTick(() => editInput.value?.focus?.())
}

function cancelEdit() {
  editingField.value = null
  editValue.value = ''
}

function saveEdit(field: string) {
  let value: any
  if (field === 'tags') {
    value = editValue.value.split(',').map((t: string) => t.trim()).filter(Boolean)
  } else {
    value = editValue.value
  }
  emit('save', field, value)
  editingField.value = null
  editValue.value = ''
}
</script>

<style scoped>
/* Modal header */
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3) var(--space-6);
  border-bottom: 1px solid var(--color-border);
}
.modal-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: var(--radius-sm);
  background: none;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast);
}
.modal-close:hover {
  background: var(--color-surface-hover);
  color: var(--color-text);
}

/* Body */
.modal-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  padding: var(--space-5) var(--space-6);
}

/* Sections */
.section-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-sm);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  margin-bottom: var(--space-3);
}
.section-count {
  font-size: 0.65rem;
  background: var(--color-surface-hover);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  font-variant-numeric: tabular-nums;
}
.section-description {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  margin-bottom: var(--space-3);
  opacity: 0.7;
}

/* Field group */
.field-group {
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
}
.field {
  padding: var(--space-3) var(--space-4);
}
.field + .field {
  border-top: 1px solid var(--color-border);
}
.field-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-1);
}
.field-edit-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: none;
  border-radius: var(--radius-sm);
  background: none;
  color: var(--color-text-muted);
  cursor: pointer;
  opacity: 0;
  transition: opacity var(--transition-fast), background var(--transition-fast), color var(--transition-fast);
}
.field:hover .field-edit-btn { opacity: 1; }
.field-edit-btn:hover {
  background: var(--color-surface-hover);
  color: var(--color-primary);
}

/* Field values */
.field-value {
  font-size: var(--type-sm);
  color: var(--color-text);
  line-height: 1.6;
}
.field-value-pre { white-space: pre-wrap; }
.field-empty {
  color: var(--color-text-muted);
  font-style: italic;
  opacity: 0.5;
}

/* Field editing */
.field-edit { display: flex; flex-direction: column; gap: var(--space-2); }
.field-input, .field-textarea {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-size: var(--type-sm);
  font-family: inherit;
  line-height: 1.5;
}
.field-textarea { resize: vertical; min-height: 60px; }
.field-textarea-tall { min-height: 120px; }
.field-input:focus, .field-textarea:focus { outline: 2px solid var(--color-primary); outline-offset: -1px; }
.field-edit-actions { display: flex; gap: var(--space-2); }
.field-hint { font-size: var(--type-xs); color: var(--color-text-muted); }

/* Tags */
.tags-row { display: flex; gap: var(--space-2); flex-wrap: wrap; padding-top: var(--space-1); }

/* Memory */
.memory-list {
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
}
.memory-item {
  display: flex;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  font-size: var(--type-sm);
  line-height: 1.5;
}
.memory-item + .memory-item {
  border-top: 1px solid var(--color-border);
}
.memory-index {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-text-muted);
  min-width: 1.4em;
  text-align: right;
  flex-shrink: 0;
  padding-top: 2px;
  font-variant-numeric: tabular-nums;
}
.memory-text {
  color: var(--color-text);
}

/* Responsive */
@media (max-width: 768px) {
  .modal-body { gap: var(--space-5); }
}
</style>
