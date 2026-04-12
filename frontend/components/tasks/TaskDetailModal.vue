<template>
  <ModalOverlay :visible="!!task" size="lg" flush @close="$emit('close')">
    <template v-if="task">
      <!-- Header -->
      <div class="modal-header">
        <h3 class="modal-title">
          <span class="header-code">{{ task.code }}</span>
          Task Details
        </h3>
        <button class="modal-close" @click="$emit('close')">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>

      <div class="modal-body">
        <!-- Title -->
        <section class="section">
          <div class="field-group">
            <div class="field" @dblclick="startEditing('title')">
              <div class="field-header">
                <label class="field-label">Title</label>
                <button v-if="editing !== 'title'" class="field-edit-btn" @click="startEditing('title')" title="Edit">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                </button>
              </div>
              <div v-if="editing !== 'title'" class="field-value field-value-title" v-html="renderRefs(task.title, researchSlug)"></div>
              <div v-else class="field-edit">
                <input
                  ref="editTitleInput"
                  v-model="editValues.title"
                  class="field-input"
                  @keydown.enter="saveField('title')"
                  @keydown.escape="cancelEdit"
                />
                <div class="field-edit-actions">
                  <button class="btn btn-sm btn-primary" @click="saveField('title')">Save</button>
                  <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Properties -->
        <section class="section">
          <h4 class="section-title">Properties</h4>
          <div class="field-group">
            <div class="field">
              <div class="field-header">
                <label class="field-label">Status</label>
              </div>
              <div class="field-value">
                <StatusBadge :status="task.status" />
              </div>
            </div>
            <div class="field">
              <div class="field-header">
                <label class="field-label">Priority</label>
              </div>
              <div class="priority-selector">
                <button
                  v-for="p in ['low', 'medium', 'high']"
                  :key="p"
                  :class="['priority-chip', `priority-${p}`, { active: task.priority === p }]"
                  @click="$emit('updatePriority', p)"
                >{{ p }}</button>
              </div>
            </div>
            <div v-if="task.created_at" class="field">
              <div class="field-header">
                <label class="field-label">Created</label>
              </div>
              <div class="field-value">{{ new Date(task.created_at).toLocaleDateString() }}</div>
            </div>
            <div v-if="task.completed_at" class="field">
              <div class="field-header">
                <label class="field-label">Completed</label>
              </div>
              <div class="field-value">{{ new Date(task.completed_at).toLocaleDateString() }}</div>
            </div>
          </div>
        </section>

        <!-- Description -->
        <section class="section">
          <h4 class="section-title">Description</h4>
          <div class="field-group">
            <div class="field" @dblclick="startEditing('description')">
              <div class="field-header">
                <label class="field-label">Task description</label>
                <button v-if="editing !== 'description'" class="field-edit-btn" @click="startEditing('description')" title="Edit">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                </button>
              </div>
              <div v-if="editing !== 'description'">
                <div
                  v-if="task.description"
                  ref="descContentEl"
                  class="field-value field-value-pre markdown-content"
                  v-html="renderRefs(marked.parse(normalizeContent(task.description)) as string, researchSlug)"
                ></div>
                <div v-else class="field-value field-empty">Click the pencil to add a description</div>
              </div>
              <div v-else class="field-edit">
                <textarea
                  ref="editDescInput"
                  v-model="editValues.description"
                  class="field-textarea"
                  rows="4"
                  placeholder="Describe the task..."
                  @keydown.escape="cancelEdit"
                ></textarea>
                <div class="field-edit-actions">
                  <button class="btn btn-sm btn-primary" @click="saveField('description')">Save</button>
                  <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Result / Comments -->
        <section class="section">
          <h4 class="section-title">Result / Comments</h4>
          <div class="field-group">
            <div class="field" @dblclick="startEditing('result')">
              <div class="field-header">
                <label class="field-label">Outcome or notes</label>
                <button v-if="editing !== 'result'" class="field-edit-btn" @click="startEditing('result')" title="Edit">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                </button>
              </div>
              <div v-if="editing !== 'result'">
                <div
                  v-if="task.result"
                  ref="resultContentEl"
                  class="field-value field-value-pre markdown-content"
                  v-html="renderRefs(marked.parse(normalizeContent(task.result)) as string, researchSlug)"
                ></div>
                <div v-else class="field-value field-empty">Click the pencil to add a result</div>
              </div>
              <div v-else class="field-edit">
                <textarea
                  ref="editResultInput"
                  v-model="editValues.result"
                  class="field-textarea"
                  rows="4"
                  placeholder="What was the outcome?"
                  @keydown.escape="cancelEdit"
                ></textarea>
                <div class="field-edit-actions">
                  <button class="btn btn-sm btn-primary" @click="saveField('result')">Save</button>
                  <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </template>
  </ModalOverlay>
</template>

<script setup lang="ts">
import { marked } from 'marked'
import { renderRefs } from '~/composables/useCrossRefs'
import { renderMermaidBlocks } from '~/composables/useMermaid'
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

// Mermaid rendering for task description/result
const descContentEl = ref<HTMLElement | null>(null)
const resultContentEl = ref<HTMLElement | null>(null)
function renderTaskMermaid() {
  nextTick(() => {
    if (descContentEl.value) renderMermaidBlocks(descContentEl.value)
    if (resultContentEl.value) renderMermaidBlocks(resultContentEl.value)
  })
}

watch(() => props.task, () => {
  editing.value = null
  renderTaskMermaid()
})
onMounted(renderTaskMermaid)

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

function cancelEdit() {
  editing.value = null
}

function saveField(field: string) {
  if (!props.task) return
  emit('save', field, (editValues as any)[field])
  editing.value = null
}
</script>

<style scoped>
/* Header — same as DetailsPanel */
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3) var(--space-6);
  border-bottom: 1px solid var(--color-border);
}
.modal-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-sm);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}
.header-code {
  font-size: var(--type-xs);
  font-weight: 700;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  letter-spacing: 0;
  text-transform: none;
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

/* Body — same as DetailsPanel */
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

/* Field group — same as DetailsPanel */
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
.field-label {
  font-size: var(--type-xs);
  font-weight: 500;
  color: var(--color-text-muted);
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
.field-value-title {
  font-size: var(--type-lg);
  font-weight: 600;
  letter-spacing: -0.02em;
  line-height: 1.35;
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
.field-input:focus, .field-textarea:focus { outline: 2px solid var(--color-primary); outline-offset: -1px; }
.field-edit-actions { display: flex; gap: var(--space-2); }

/* Priority selector */
.priority-selector {
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

/* Responsive */
@media (max-width: 768px) {
  .modal-body { gap: var(--space-5); }
}
</style>
