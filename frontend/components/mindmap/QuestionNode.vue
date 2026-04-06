<template>
  <div class="mindmap-node question-node" @click="navigate">
    <div class="q-header">
      <div class="q-title-row">
        <span v-if="data.code" class="mm-code">{{ data.code }}</span>
        <span class="q-text">{{ truncate(data.text, 70) }}</span>
      </div>
      <StatusBadge :status="data.status" />
    </div>
    <p v-if="data.answer && data.status === 'answered'" class="q-answer">{{ truncate(data.answer, 60) }}</p>
    <span class="q-session">{{ data.sessionTitle }}</span>
    <Handle type="target" :position="targetPosition" />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{
  data: {
    id: string
    code: string
    text: string
    status: string
    answer: string
    sessionId: string
    sessionTitle: string
    researchSlug: string
  }
  targetPosition?: Position
}>()

function truncate(text: string, len: number): string {
  if (!text) return ''
  return text.length > len ? text.slice(0, len) + '...' : text
}

function navigate() {
  if (props.data.researchSlug && props.data.sessionId) {
    window.open(`/research/${props.data.researchSlug}/session/${props.data.sessionId}/question/${props.data.code || props.data.id}`, '_blank')
  }
}
</script>

<style scoped>
.question-node {
  background: var(--color-surface);
  border: 1px solid rgba(240, 184, 73, 0.2);
  border-radius: var(--radius);
  padding: var(--space-4) var(--space-5);
  min-width: 340px;
  max-width: 420px;
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
.question-node:hover {
  border-color: rgba(240, 184, 73, 0.4);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}
.q-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
  margin-bottom: var(--space-1);
}
.q-title-row { display: flex; align-items: flex-start; gap: var(--space-2); min-width: 0; }
.mm-code {
  font-size: 0.625rem; font-weight: 700; color: var(--color-primary);
  background: var(--color-primary-muted); padding: 0.1rem 0.3rem;
  border-radius: 3px; font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0; line-height: 1; margin-top: 2px;
}
.q-text {
  font-size: var(--type-xs);
  font-weight: 500;
  color: var(--color-text);
  line-height: 1.35;
}
.q-answer {
  font-size: 0.6875rem;
  color: var(--color-success);
  line-height: 1.35;
  margin-bottom: var(--space-1);
}
.q-session {
  font-size: 0.625rem;
  color: var(--color-text-muted);
  opacity: 0.7;
}
</style>
