<template>
  <div :class="cardClasses">
    <div class="rmc-header">
      <div class="rmc-title-row">
        <span v-if="data.code" class="rm-code">{{ data.code }}</span>
        <span class="rmc-title">{{ displayTitle }}</span>
      </div>
      <span v-if="displayStatus" :class="['rm-status', `rm-status-${statusSlug(displayStatus)}`]">{{ displayStatus }}</span>
    </div>

    <!-- Content preview (ref content, ref result, or plain description) -->
    <template v-if="!compact">
      <p v-if="isEntryRef && data.refData?.content" class="rmc-preview">{{ truncate(data.refData.content, 80) }}</p>
      <p v-else-if="data.description" class="rmc-preview">{{ truncate(data.description, isRef ? 80 : 100) }}</p>
      <p v-if="isTaskRef && data.refData?.result" class="rmc-result">{{ truncate(data.refData.result, 80) }}</p>
    </template>

    <div class="rmc-footer">
      <!-- Ref type badge (a node referencing an entity) -->
      <span v-if="isRef" :class="['rmc-badge', `badge-${data.refType}`]">
        <svg v-if="data.refType === 'entry'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/></svg>
        <svg v-else-if="data.refType === 'task'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><path d="m9 11 3 3L22 4"/></svg>
        <svg v-else-if="data.refType === 'session'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M7.9 20A9 9 0 1 0 4 16.1L2 22z"/></svg>
        <svg v-else-if="data.refType === 'research'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
        <svg v-else-if="data.refType === 'question'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/></svg>
        {{ data.refType === 'research' ? 'project' : data.refType === 'entry' ? 'document' : data.refType }}
      </span>
      <!-- Plain node type badge -->
      <span v-else :class="['rmc-badge', `badge-${data.nodeType}`]">
        <svg v-if="data.nodeType === 'milestone'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z"/><line x1="4" x2="4" y1="22" y2="15"/></svg>
        <svg v-else-if="data.nodeType === 'decision'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M16 3h5v5"/><path d="M8 3H3v5"/><path d="M12 22v-8.3a4 4 0 0 0-1.172-2.872L3 3"/><path d="m15 9 6-6"/></svg>
        <svg v-else-if="data.nodeType === 'info'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
        <svg v-else width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="m8 12 3 3 5-5"/></svg>
        {{ data.nodeType }}
      </span>

      <span v-if="data.refData?.section_name" class="rmc-meta">{{ data.refData.section_name }}</span>
      <span v-if="data.refData?.priority" :class="['rmc-priority', `priority-${data.refData.priority}`]">{{ data.refData.priority }}</span>
      <span v-if="isSessionRef && data.refData" class="rmc-progress-text">
        {{ data.refData.answered_questions }}/{{ data.refData.total_questions }} questions
      </span>
      <span v-if="isResearchRef && data.refData" class="rmc-meta">
        {{ data.refData.section_count }} sections &middot; {{ data.refData.entry_count }} documents
      </span>
      <span v-if="data.refData?.code" class="rmc-entity-code">{{ data.refData.code }}</span>
    </div>

    <!-- Session progress bar -->
    <div v-if="isSessionRef && sessionProgress > 0" class="rmc-progress-bar">
      <div class="rmc-progress-fill" :style="{ width: sessionProgress + '%' }"></div>
    </div>

    <!-- Dependency chips (board / timeline only): predecessors this node depends on -->
    <div v-if="deps && deps.length" class="rmc-deps">
      <span class="rmc-deps-arrow">&#x21B3;</span>
      <span class="rmc-deps-label">depends on</span>
      <span v-for="d in deps" :key="d.code" class="rmc-dep-code">{{ d.code }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { truncate } from '~/utils/truncate'

/**
 * The one roadmap card, shared by all three views. The graph node components
 * wrap it and add Vue Flow <Handle>s; the stages board and timeline render it
 * directly, since <Handle> only works inside a Vue Flow provider. `data` is the
 * camelCase shape the graph already builds; board/timeline map their raw nodes
 * to it via nodeCardData() in useRoadmap.
 */
export interface RoadmapCardData {
  code?: string
  title?: string
  description?: string
  nodeType?: string
  status?: string
  refType?: string
  refId?: string
  refData?: {
    title?: string
    status?: string
    code?: string
    description?: string
    section_name?: string
    content?: string
    priority?: string
    result?: string
    total_questions?: number
    answered_questions?: number
    section_count?: number
    entry_count?: number
  }
}

const props = withDefaults(defineProps<{
  data: RoadmapCardData
  /** Predecessor codes this node depends on (board/timeline only). */
  deps?: { code: string }[]
  /** Compact strips the preview lines — used by the timeline where cells are dense. */
  compact?: boolean
  /** Related-card highlight state, driven by the board's hover/focus. */
  highlighted?: boolean
  dimmed?: boolean
}>(), {
  deps: undefined,
  compact: false,
  highlighted: false,
  dimmed: false,
})

const isRef = computed(() => !!(props.data.refType && props.data.refId))
const isEntryRef = computed(() => props.data.refType === 'entry')
const isTaskRef = computed(() => props.data.refType === 'task')
const isSessionRef = computed(() => props.data.refType === 'session')
const isResearchRef = computed(() => props.data.refType === 'research')

const cardClasses = computed(() => [
  'roadmap-node-card',
  isRef.value ? `ref-type-${props.data.refType}` : `step-type-${props.data.nodeType || 'step'}`,
  { 'is-highlighted': props.highlighted, 'is-dimmed': props.dimmed },
])

const displayTitle = computed(() => {
  if (props.data.refData?.title && !props.data.title) {
    return truncate(props.data.refData.title, 50)
  }
  return truncate(props.data.title || '', 50)
})

const displayStatus = computed(() => props.data.refData?.status || props.data.status || '')

const sessionProgress = computed(() => {
  if (!isSessionRef.value || !props.data.refData) return 0
  const total = props.data.refData.total_questions || 0
  const answered = props.data.refData.answered_questions || 0
  return total > 0 ? Math.round((answered / total) * 100) : 0
})

function statusSlug(s: string): string {
  return s.replace(/[^a-z0-9]/gi, '-').toLowerCase()
}
</script>

<style scoped>
.roadmap-node-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--space-4) var(--space-5);
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast), opacity var(--transition-fast);
  position: relative;
  text-align: left;
  width: 100%;
}
.roadmap-node-card:hover {
  border-color: var(--color-border-strong);
  box-shadow: var(--shadow-1);
}
.roadmap-node-card.is-highlighted {
  border-color: var(--color-border-strong);
  box-shadow: var(--shadow-1);
}
.roadmap-node-card.is-dimmed {
  opacity: 0.5;
}

/* Tint per type — mirrors the graph node tints exactly */
.ref-type-entry    { background: rgba(var(--color-primary-rgb), 0.04); }
.ref-type-task     { background: rgba(var(--color-success-rgb), 0.04); }
.ref-type-session  { background: rgba(var(--hue-5-rgb), 0.05); }
.ref-type-research { background: rgba(var(--color-warning-rgb), 0.04); }
.ref-type-question { background: rgba(var(--color-error-rgb), 0.04); }
.step-type-step      { background: rgba(var(--color-success-rgb), 0.04); }
.step-type-milestone { background: rgba(var(--hue-5-rgb), 0.06); }
.step-type-decision  { background: rgba(var(--color-warning-rgb), 0.05); }
.step-type-info      { background: rgba(var(--color-primary-rgb), 0.04); }
.step-type-group     { background: rgba(var(--color-muted-rgb), 0.04); }

.rmc-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
  margin-bottom: var(--space-1);
}
.rmc-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}
.rmc-title {
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
  line-height: 1.3;
  overflow-wrap: anywhere;
}
.rmc-preview {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  line-height: 1.4;
  margin-bottom: var(--space-2);
  overflow-wrap: anywhere;
}
.rmc-result {
  font-size: 0.6875rem;
  color: rgba(var(--color-success-rgb), 0.85);
  line-height: 1.4;
  margin-bottom: var(--space-2);
  font-style: italic;
  overflow-wrap: anywhere;
}
.rmc-footer {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.rmc-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
  line-height: 1;
  font-weight: var(--weight-semibold);
}
.badge-entry     { background: rgba(var(--color-primary-rgb), 0.12); color: rgba(var(--color-primary-rgb), 1); }
.badge-task      { background: rgba(var(--color-success-rgb), 0.12); color: rgba(var(--color-success-rgb), 1); }
.badge-session   { background: rgba(var(--hue-5-rgb), 0.12); color: rgba(var(--hue-5-rgb), 1); }
.badge-research  { background: rgba(var(--color-warning-rgb), 0.12); color: rgba(var(--color-warning-rgb), 1); }
.badge-question  { background: rgba(var(--color-error-rgb), 0.12); color: rgba(var(--color-error-rgb), 1); }
.badge-step      { background: rgba(var(--color-success-rgb), 0.12); color: rgba(var(--color-success-rgb), 1); }
.badge-milestone { background: rgba(var(--hue-5-rgb), 0.12); color: rgba(var(--hue-5-rgb), 1); }
.badge-decision  { background: rgba(var(--color-warning-rgb), 0.12); color: rgba(var(--color-warning-rgb), 1); }
.badge-info      { background: rgba(var(--color-primary-rgb), 0.12); color: rgba(var(--color-primary-rgb), 1); }
.badge-group     { background: rgba(var(--color-muted-rgb), 0.12); color: rgba(var(--color-muted-rgb), 1); }

.rmc-meta {
  font-size: 0.5625rem;
  color: var(--color-text-muted);
  line-height: 1;
}
.rmc-entity-code {
  font-size: 0.5625rem;
  font-weight: var(--weight-bold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.1rem 0.3rem;
  border-radius: var(--radius-xs);
  font-family: 'JetBrains Mono', monospace;
  line-height: 1;
  margin-left: auto;
}
.rmc-priority {
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 0.1rem 0.3rem;
  border-radius: var(--radius-xs);
  line-height: 1;
}
.priority-high, .priority-critical { background: rgba(var(--color-error-rgb), 0.12); color: rgba(var(--color-error-rgb), 1); }
.priority-medium { background: rgba(var(--color-warning-rgb), 0.12); color: rgba(var(--color-warning-rgb), 1); }
.priority-low { background: var(--color-surface-hover); color: var(--color-text-muted); }
.rmc-progress-text {
  font-size: 0.5625rem;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}
.rmc-progress-bar {
  margin-top: var(--space-2);
  width: 100%;
  height: 3px;
  background: var(--color-surface-hover);
  border-radius: var(--radius-hair);
  overflow: hidden;
}
.rmc-progress-fill {
  height: 100%;
  background: rgba(var(--hue-5-rgb), 0.6);
  border-radius: var(--radius-hair);
  transition: width 0.3s ease;
}

.rmc-deps {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  flex-wrap: wrap;
  margin-top: var(--space-2);
  padding-top: var(--space-2);
  border-top: 1px solid var(--color-border);
  font-size: 0.5625rem;
  color: var(--color-text-faint);
}
.rmc-deps-arrow { color: var(--color-text-muted); }
.rmc-deps-label { text-transform: uppercase; letter-spacing: 0.04em; }
.rmc-dep-code {
  /* Informational, not a link: the whole card is the click target, so the code
     must not read as separately clickable. Muted mono, no primary tint. */
  font-weight: var(--weight-semibold);
  color: var(--color-text-muted);
  font-family: 'JetBrains Mono', monospace;
}

.rm-code {
  font-size: 0.625rem;
  font-weight: var(--weight-bold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.1rem 0.3rem;
  border-radius: var(--radius-xs);
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
.rm-status {
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-weight: var(--weight-semibold);
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
  line-height: 1;
  white-space: nowrap;
  flex-shrink: 0;
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
}
.rm-status-completed,
.rm-status-mastered,
.rm-status-done,
.rm-status-final {
  background: rgba(var(--color-success-rgb), 0.15);
  color: rgba(var(--color-success-rgb), 1);
}
.rm-status-in-progress,
.rm-status-in_progress,
.rm-status-learning,
.rm-status-active {
  background: rgba(var(--color-primary-rgb), 0.15);
  color: rgba(var(--color-primary-rgb), 1);
}
.rm-status-not-started,
.rm-status-not_started,
.rm-status-pending,
.rm-status-todo,
.rm-status-draft,
.rm-status-planned {
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
}
.rm-status-skipped,
.rm-status-blocked {
  background: rgba(var(--color-error-rgb), 0.12);
  color: rgba(var(--color-error-rgb), 1);
}
.rm-status-answered {
  background: rgba(var(--hue-5-rgb), 0.15);
  color: rgba(var(--hue-5-rgb), 1);
}
</style>
