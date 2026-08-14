<template>
  <ModalOverlay :visible="!!node" size="sm" flush @close="$emit('close')">
    <template v-if="node">
      <!-- Header (same pattern as DetailsPanel) -->
      <div class="modal-header">
        <h3 class="modal-title">{{ node.refType || node.nodeType }}</h3>
        <button class="modal-close" @click="$emit('close')">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>

      <div class="modal-body">
        <!-- Title + description -->
        <h3 class="node-title">{{ node.title }}</h3>
        <p v-if="node.description" class="node-desc">{{ node.description }}</p>

        <!-- Ref data fields (same field-group pattern as DetailsPanel) -->
        <div v-if="node.refData" class="field-group">
          <!-- Entry: content preview -->
          <div v-if="node.refType === 'entry' && node.refData.content" class="field">
            <label class="field-label">Content</label>
            <div class="field-value">{{ node.refData.content }}</div>
          </div>

          <!-- Entry: section name -->
          <div v-if="node.refData.section_name" class="field">
            <label class="field-label">Section</label>
            <div class="field-value">{{ node.refData.section_name }}</div>
          </div>

          <!-- Task: priority -->
          <div v-if="node.refData.priority" class="field">
            <label class="field-label">Priority</label>
            <div class="field-value">
              <span :class="['priority-chip', `priority-${node.refData.priority}`]">{{ node.refData.priority }}</span>
            </div>
          </div>

          <!-- Task: result -->
          <div v-if="node.refType === 'task' && node.refData.result" class="field">
            <label class="field-label">Result</label>
            <div class="field-value">{{ node.refData.result }}</div>
          </div>

          <!-- Session: question progress -->
          <div v-if="node.refType === 'session'" class="field">
            <label class="field-label">Questions</label>
            <div class="field-value">{{ node.refData.answered_questions }}/{{ node.refData.total_questions }} answered</div>
          </div>

          <!-- Research: counts -->
          <div v-if="node.refType === 'research'" class="field">
            <label class="field-label">Content</label>
            <div class="field-value">{{ node.refData.section_count }} sections, {{ node.refData.entry_count }} entries</div>
          </div>

          <!-- Entity status — clickable chips with real entity statuses -->
          <div v-if="entityStatuses.length" class="field">
            <label class="field-label">
              {{ refLabel }} status
              <span v-if="node.refData.code" class="entity-code">{{ node.refData.code }}</span>
            </label>
            <div v-if="canWrite" class="status-chips">
              <button
                v-for="s in entityStatuses"
                :key="s"
                :class="['status-chip', { active: s === node.refData?.status }]"
                @click="$emit('update-entity-status', node.refType, node.refId, s)"
              >{{ s }}</button>
            </div>
            <StatusBadge v-else-if="node.refData?.status" :status="node.refData.status" />
          </div>
        </div>

        <!-- Navigate to entity -->
        <button
          v-if="node.refType && node.refId"
          class="btn btn-sm btn-open"
          @click="$emit('navigate', node)"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" x2="21" y1="14" y2="3"/></svg>
          Open {{ refLabel }}
        </button>

        <!-- Roadmap node status chips (only for non-ref nodes) -->
        <div v-if="statuses.length && !node.refType" class="statuses-section">
          <label class="field-label">Status</label>
          <StatusBadge v-if="!canWrite && node.status" :status="node.status" />
          <div v-else-if="canWrite" class="status-chips">
            <button
              v-for="s in statuses"
              :key="s"
              :class="['status-chip', { active: s === node.status }]"
              @click="$emit('update-status', node.id, s)"
            >{{ s }}</button>
            <button
              v-if="node.status"
              class="status-chip status-chip-clear"
              @click="$emit('update-status', node.id, '')"
            >clear</button>
          </div>
        </div>
      </div>
    </template>
  </ModalOverlay>
</template>

<script setup lang="ts">
const ENTITY_STATUSES: Record<string, string[]> = {
  task:     ['pending', 'in_progress', 'blocked', 'completed', 'failed', 'deferred'],
  entry:    ['draft', 'active', 'completed', 'archived'],
  session:  ['active', 'completed', 'archived'],
  research: ['active', 'completed', 'archived'],
  question: ['pending', 'in_progress', 'answered', 'deferred', 'skipped'],
}

const props = defineProps<{
  node: {
    id: string
    title: string
    description: string
    nodeType: string
    status: string
    refType?: string
    refId?: string
    refData?: any
  } | null
  statuses: string[]
}>()

// Opening the entity stays; changing its status does not.
const { canWrite } = useResearchRole()

defineEmits<{
  'update-status': [nodeId: string, status: string]
  'update-entity-status': [refType: string, refId: string, status: string]
  'navigate': [node: any]
  'close': []
}>()

const refLabel = computed(() => {
  switch (props.node?.refType) {
    case 'entry': return 'entry'
    case 'task': return 'task'
    case 'session': return 'session'
    case 'research': return 'research'
    case 'question': return 'question'
    default: return 'entity'
  }
})

const entityStatuses = computed(() => {
  if (!props.node?.refType) return []
  return ENTITY_STATUSES[props.node.refType] || []
})
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
  font-size: var(--type-sm);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
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
  gap: var(--space-4);
  padding: var(--space-5) var(--space-6);
}

.node-title {
  font-size: var(--type-base);
  font-weight: 600;
  color: var(--color-text);
  margin: 0;
  line-height: 1.3;
}
.node-desc {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  line-height: 1.5;
  margin: 0;
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
.field-label {
  font-size: var(--type-xs);
  font-weight: 500;
  color: var(--color-text-muted);
  margin-bottom: var(--space-1);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
.field-value {
  font-size: var(--type-sm);
  color: var(--color-text);
  line-height: 1.5;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.entity-code {
  font-size: 0.5625rem;
  font-weight: 700;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.1rem 0.3rem;
  border-radius: 3px;
  font-family: 'JetBrains Mono', monospace;
}

.priority-chip {
  font-size: var(--type-xs);
  font-weight: 600;
  padding: 0.15rem 0.4rem;
  border-radius: 3px;
  text-transform: capitalize;
}
.priority-high, .priority-critical {
  background: rgba(239, 107, 107, 0.12);
  color: rgba(239, 107, 107, 1);
}
.priority-medium {
  background: rgba(240, 184, 73, 0.12);
  color: rgba(240, 184, 73, 1);
}
.priority-low {
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
}

/* Open button */
.btn-open {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-primary);
}

/* Status chips (shared for entity and node statuses) */
.statuses-section {
  border-top: 1px solid var(--color-border);
  padding-top: var(--space-3);
}
.status-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
  margin-top: var(--space-1);
}
.status-chip {
  font-size: 0.625rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  border: 1px solid var(--color-border);
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
  line-height: 1.2;
}
.status-chip:hover {
  border-color: var(--color-primary);
  color: var(--color-text);
}
.status-chip.active {
  background: var(--color-primary-muted);
  border-color: rgba(108, 197, 224, 0.4);
  color: var(--color-primary);
  font-weight: 600;
}
.status-chip-clear {
  color: var(--color-text-muted);
  opacity: 0.6;
  font-style: italic;
}
.status-chip-clear:hover {
  opacity: 1;
  border-color: rgba(239, 107, 107, 0.4);
  color: rgba(239, 107, 107, 1);
}
</style>
