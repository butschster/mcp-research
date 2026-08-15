<template>
  <div :class="['roadmap-node', 'ref-node', `ref-type-${data.refType}`]">
    <div class="ref-header">
      <div class="ref-title-row">
        <span v-if="data.code" class="rm-code">{{ data.code }}</span>
        <span class="ref-title">{{ displayTitle }}</span>
      </div>
      <span v-if="displayStatus" :class="['rm-status', `rm-status-${statusSlug(displayStatus)}`]">{{ displayStatus }}</span>
    </div>

    <!-- Entry ref: show content preview -->
    <p v-if="data.refType === 'entry' && data.refData?.content" class="ref-preview">{{ truncate(data.refData.content, 80) }}</p>
    <p v-else-if="data.description" class="ref-preview">{{ truncate(data.description, 80) }}</p>

    <!-- Task ref: show result preview -->
    <p v-if="data.refType === 'task' && data.refData?.result" class="ref-result">{{ truncate(data.refData.result, 80) }}</p>

    <div class="ref-footer">
      <!-- Type badge with icon -->
      <span :class="['ref-type-badge', `badge-${data.refType}`]">
        <svg v-if="data.refType === 'entry'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/></svg>
        <svg v-else-if="data.refType === 'task'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><path d="m9 11 3 3L22 4"/></svg>
        <svg v-else-if="data.refType === 'session'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M7.9 20A9 9 0 1 0 4 16.1L2 22z"/></svg>
        <svg v-else-if="data.refType === 'research'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
        <svg v-else-if="data.refType === 'question'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/></svg>
        {{ data.refType }}
      </span>

      <!-- Entry: show section -->
      <span v-if="data.refData?.section_name" class="ref-meta">{{ data.refData.section_name }}</span>
      <!-- Task: show priority -->
      <span v-if="data.refData?.priority" :class="['ref-priority', `priority-${data.refData.priority}`]">{{ data.refData.priority }}</span>
      <!-- Session: show question progress -->
      <span v-if="data.refType === 'session' && data.refData" class="ref-progress-text">
        {{ data.refData.answered_questions }}/{{ data.refData.total_questions }} questions
      </span>
      <!-- Research: show counts -->
      <span v-if="data.refType === 'research' && data.refData" class="ref-meta">
        {{ data.refData.section_count }}s &middot; {{ data.refData.entry_count }}e
      </span>
      <!-- Ref code badge -->
      <span v-if="data.refData?.code" class="ref-entity-code">{{ data.refData.code }}</span>
    </div>

    <!-- Session progress bar -->
    <div v-if="data.refType === 'session' && sessionProgress > 0" class="ref-progress-bar">
      <div class="ref-progress-fill" :style="{ width: sessionProgress + '%' }"></div>
    </div>

    <Handle type="target" :position="targetPosition" />
    <Handle type="source" :position="sourcePosition" />
  </div>
</template>

<script setup lang="ts">
import { truncate } from '~/utils/truncate'
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{
  data: {
    code: string
    title: string
    description: string
    nodeType: string
    status: string
    refType?: string
    refId?: string
    metadata?: string
    refData?: {
      title?: string
      status?: string
      code?: string
      description?: string
      research_id?: string
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
  sourcePosition?: Position
  targetPosition?: Position
}>()

const displayTitle = computed(() => {
  if (props.data.refData?.title && !props.data.title) {
    return truncate(props.data.refData.title, 50)
  }
  return truncate(props.data.title, 50)
})

const displayStatus = computed(() => {
  if (props.data.refData?.status) {
    return props.data.refData.status
  }
  return props.data.status
})

const sessionProgress = computed(() => {
  if (props.data.refType !== 'session' || !props.data.refData) return 0
  const total = props.data.refData.total_questions || 0
  const answered = props.data.refData.answered_questions || 0
  return total > 0 ? Math.round((answered / total) * 100) : 0
})


function statusSlug(s: string): string {
  return s.replace(/[^a-z0-9]/gi, '-').toLowerCase()
}
</script>

<style scoped>
.ref-node {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--space-4) var(--space-5);
  min-width: 280px;
  max-width: 380px;
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
  position: relative;
}
.ref-node:hover {
  border-color: var(--color-border-strong);
  box-shadow: var(--shadow-1);
}

/* Subtle background tint per ref type */
.ref-type-entry    { background: rgba(108, 197, 224, 0.04); }
.ref-type-task     { background: rgba(107, 203, 119, 0.04); }
.ref-type-session  { background: rgba(168, 130, 255, 0.05); }
.ref-type-research { background: rgba(240, 184, 73, 0.04); }
.ref-type-question { background: rgba(239, 107, 107, 0.04); }

.ref-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
  margin-bottom: var(--space-1);
}
.ref-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}
.ref-title {
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
  line-height: 1.3;
}
.ref-preview {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  line-height: 1.4;
  margin-bottom: var(--space-2);
}
.ref-result {
  font-size: 0.6875rem;
  color: rgba(107, 203, 119, 0.85);
  line-height: 1.4;
  margin-bottom: var(--space-2);
  font-style: italic;
}
.ref-footer {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

/* Type badge with icon and color */
.ref-type-badge {
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
.badge-entry {
  background: rgba(108, 197, 224, 0.12);
  color: rgba(108, 197, 224, 1);
}
.badge-task {
  background: rgba(107, 203, 119, 0.12);
  color: rgba(107, 203, 119, 1);
}
.badge-session {
  background: rgba(168, 130, 255, 0.12);
  color: rgba(168, 130, 255, 1);
}
.badge-research {
  background: rgba(240, 184, 73, 0.12);
  color: rgba(240, 184, 73, 1);
}
.badge-question {
  background: rgba(239, 107, 107, 0.12);
  color: rgba(239, 107, 107, 1);
}

.ref-meta {
  font-size: 0.5625rem;
  color: var(--color-text-muted);
  line-height: 1;
}
.ref-entity-code {
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
.ref-priority {
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 0.1rem 0.3rem;
  border-radius: var(--radius-xs);
  line-height: 1;
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
.ref-progress-text {
  font-size: 0.5625rem;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}
.ref-progress-bar {
  margin-top: var(--space-2);
  width: 100%;
  height: 3px;
  background: var(--color-surface-hover);
  border-radius: var(--radius-hair);
  overflow: hidden;
}
.ref-progress-fill {
  height: 100%;
  background: rgba(168, 130, 255, 0.6);
  border-radius: var(--radius-hair);
  transition: width 0.3s ease;
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
  background: rgba(107, 203, 119, 0.15);
  color: rgba(107, 203, 119, 1);
}
.rm-status-in-progress,
.rm-status-in_progress,
.rm-status-learning,
.rm-status-active {
  background: rgba(108, 197, 224, 0.15);
  color: rgba(108, 197, 224, 1);
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
  background: rgba(239, 107, 107, 0.12);
  color: rgba(239, 107, 107, 1);
}
.rm-status-answered {
  background: rgba(168, 130, 255, 0.15);
  color: rgba(168, 130, 255, 1);
}
</style>
