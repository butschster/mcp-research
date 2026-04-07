<template>
  <div v-if="open" class="card details-panel">
    <div class="details-grid">
      <!-- Goal -->
      <div class="detail-field" @dblclick="startEdit('goal')">
        <label class="detail-label">Goal</label>
        <div v-if="editingField !== 'goal'" class="detail-value" :class="{ 'detail-empty': !research.goal }">
          {{ research.goal || 'Not set — double-click to edit' }}
        </div>
        <div v-else class="detail-edit">
          <input v-model="editValue" class="detail-input" @keydown.enter="saveEdit('goal')" @keydown.escape="cancelEdit" ref="editInput" />
          <div class="detail-edit-actions">
            <button class="btn btn-sm btn-primary" @click="saveEdit('goal')">Save</button>
            <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
          </div>
        </div>
      </div>

      <!-- Description -->
      <div class="detail-field" @dblclick="startEdit('description')">
        <label class="detail-label">Description</label>
        <div v-if="editingField !== 'description'" class="detail-value" :class="{ 'detail-empty': !research.description }">
          {{ research.description || 'Not set — double-click to edit' }}
        </div>
        <div v-else class="detail-edit">
          <textarea v-model="editValue" class="detail-textarea" rows="3" @keydown.escape="cancelEdit" ref="editInput"></textarea>
          <div class="detail-edit-actions">
            <button class="btn btn-sm btn-primary" @click="saveEdit('description')">Save</button>
            <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
          </div>
        </div>
      </div>

      <!-- Instruction -->
      <div class="detail-field detail-field-wide" @dblclick="startEdit('instruction')">
        <label class="detail-label">Instruction</label>
        <div v-if="editingField !== 'instruction'" class="detail-value detail-value-pre" :class="{ 'detail-empty': !research.instruction }">
          {{ research.instruction || 'Not set — double-click to edit' }}
        </div>
        <div v-else class="detail-edit">
          <textarea v-model="editValue" class="detail-textarea" rows="6" @keydown.escape="cancelEdit" ref="editInput"></textarea>
          <div class="detail-edit-actions">
            <button class="btn btn-sm btn-primary" @click="saveEdit('instruction')">Save</button>
            <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
          </div>
        </div>
      </div>

      <!-- Memory -->
      <div class="detail-field detail-field-wide">
        <label class="detail-label">Memory <span class="detail-count">{{ research.memory?.length || 0 }}</span></label>
        <div v-if="research.memory?.length" class="memory-list">
          <div v-for="(item, i) in research.memory" :key="i" class="memory-item">
            <span class="memory-bullet">{{ i + 1 }}</span>
            <span>{{ item }}</span>
          </div>
        </div>
        <div v-else class="detail-value detail-empty">No memory entries yet</div>
      </div>

      <!-- Tags -->
      <div class="detail-field" @dblclick="startEdit('tags')">
        <label class="detail-label">Tags</label>
        <div v-if="editingField !== 'tags'">
          <div v-if="research.tags?.length" class="tags-row">
            <span v-for="tag in research.tags" :key="tag" :class="['tag', `tag-hue-${tagHue(tag)}`]">{{ tag }}</span>
          </div>
          <div v-else class="detail-value detail-empty">No tags — double-click to edit</div>
        </div>
        <div v-else class="detail-edit">
          <input v-model="editValue" class="detail-input" placeholder="tag1, tag2, tag3" @keydown.enter="saveEdit('tags')" @keydown.escape="cancelEdit" ref="editInput" />
          <span class="detail-hint">Comma-separated</span>
          <div class="detail-edit-actions">
            <button class="btn btn-sm btn-primary" @click="saveEdit('tags')">Save</button>
            <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { tagHue } from '~/composables/useTagHue'

const props = defineProps<{
  research: any
  open: boolean
}>()

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
.details-panel { margin-bottom: var(--space-6); }
.details-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-5);
}
.detail-field-wide { grid-column: 1 / -1; }
.detail-label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  margin-bottom: var(--space-2);
}
.detail-count {
  font-size: 0.625rem;
  background: var(--color-surface-hover);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  font-variant-numeric: tabular-nums;
}
.detail-value {
  font-size: var(--type-sm);
  color: var(--color-text);
  line-height: 1.6;
  cursor: default;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  transition: border-color var(--transition-fast);
}
.detail-value:hover { border-color: var(--color-border); }
.detail-value-pre { white-space: pre-wrap; }
.detail-empty {
  color: var(--color-text-muted);
  font-style: italic;
  opacity: 0.6;
}
.detail-edit { display: flex; flex-direction: column; gap: var(--space-2); }
.detail-input, .detail-textarea {
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
.detail-textarea { resize: vertical; min-height: 60px; }
.detail-input:focus, .detail-textarea:focus { outline: 2px solid var(--color-primary); outline-offset: -1px; }
.detail-edit-actions { display: flex; gap: var(--space-2); }
.detail-hint { font-size: var(--type-xs); color: var(--color-text-muted); }
.memory-list { display: flex; flex-direction: column; gap: var(--space-2); }
.memory-item {
  display: flex;
  gap: var(--space-3);
  font-size: var(--type-sm);
  line-height: 1.5;
  padding: var(--space-2) var(--space-3);
  background: var(--color-surface-hover);
  border-radius: var(--radius-sm);
}
.memory-bullet {
  font-size: var(--type-xs);
  font-weight: 700;
  color: var(--color-text-muted);
  min-width: 1.2em;
  text-align: right;
  flex-shrink: 0;
  margin-top: 2px;
}
.tags-row { display: flex; gap: var(--space-2); flex-wrap: wrap; }

/* Responsive */
@media (max-width: 768px) {
  .details-grid { grid-template-columns: 1fr; }
  .detail-field-wide { grid-column: 1; }
}
</style>
