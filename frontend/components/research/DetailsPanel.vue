<template>
  <ModalOverlay :labelledby="titleId" :visible="open" size="lg" flush @close="emit('update:open', false)">
    <!-- Header -->
    <ModalHeader title="Research Details" :title-id="titleId" @close="emit('update:open', false)" />

    <div class="modal-body">
      <!-- Overview -->
      <section class="section">
        <h4 class="section-title">Overview</h4>
        <div class="field-group">
          <EditableField
            label="Goal"
            :value="research.goal"
            :editable="canWrite"
            :multiline="false"
            placeholder="What is this research trying to achieve?"
            empty-text="Click the pencil to set a goal"
            @save="v => emit('save', 'goal', v)"
          />

          <EditableField
            label="Description"
            :value="research.description"
            :editable="canWrite"
            :multiline="true"
            placeholder="Describe what this research covers..."
            empty-text="Click the pencil to add a description"
            @save="v => emit('save', 'description', v)"
          />

          <EditableField
            label="Tags"
            :value="(research.tags ?? []).join(', ')"
            :editable="canWrite"
            placeholder="tag1, tag2, tag3"
            empty-text="No tags yet"
            @save="v => emit('save', 'tags', v.split(',').map(t => t.trim()).filter(Boolean))"
          >
            <template #default>
              <TagList v-if="research.tags?.length" :tags="research.tags" />
              <span v-else class="field-empty">No tags yet</span>
            </template>
          </EditableField>
        </div>
      </section>

      <!-- AI Instruction -->
      <section class="section">
        <h4 class="section-title">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2a4 4 0 0 0-4 4c0 2 2 3 2 6H14c0-3 2-4 2-6a4 4 0 0 0-4-4z"/><line x1="10" y1="16" x2="14" y2="16"/><line x1="10" y1="20" x2="14" y2="20"/><line x1="11" y1="24" x2="13" y2="24"/></svg>
          AI Instruction
        </h4>
        <div class="field-group">
          <EditableField
            label="AI Instruction"
            :value="research.instruction"
            :editable="canWrite"
            :multiline="true"
            :rows="6"
            placeholder="How should the agent work on this research?"
            empty-text="No instruction set"
            @save="v => emit('save', 'instruction', v)"
          />
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
const titleId = useId()
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

</script>

<style scoped>
/* Modal header */

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
  font-weight: var(--weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  margin-bottom: var(--space-3);
}
.section-count {
  font-size: 0.65rem;
  background: var(--color-surface-hover);
  padding: 0.1rem 0.35rem;
  border-radius: var(--radius-xs);
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
.field:hover .field-edit-btn { opacity: 1; }

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
.field-hint { font-size: var(--type-xs); color: var(--color-text-muted); }

/* Tags */

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
  font-weight: var(--weight-semibold);
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
