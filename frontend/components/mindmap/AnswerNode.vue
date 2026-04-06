<template>
  <div class="mindmap-node answer-node" @click="navigate">
    <div class="a-header">
      <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
      <span class="a-label">Answer</span>
    </div>
    <div class="a-text" v-html="renderRefs(truncate(data.answer, 120), data.researchSlug)"></div>
    <Handle type="target" :position="targetPosition" />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{
  data: {
    answer: string
    questionCode: string
    sessionId: string
    researchSlug: string
  }
  targetPosition?: Position
}>()

function truncate(text: string, len: number): string {
  if (!text) return ''
  return text.length > len ? text.slice(0, len) + '...' : text
}

function navigate() {
  if (props.data.researchSlug && props.data.sessionId && props.data.questionCode) {
    window.open(`/research/${props.data.researchSlug}/session/${props.data.sessionId}/question/${props.data.questionCode}`, '_blank')
  }
}
</script>

<style scoped>
.answer-node {
  background: var(--color-surface);
  border: 1px solid rgba(52, 211, 153, 0.2);
  border-radius: var(--radius);
  padding: var(--space-3) var(--space-4);
  min-width: 280px;
  max-width: 380px;
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
.answer-node:hover {
  border-color: rgba(52, 211, 153, 0.4);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}
.a-header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-2);
  color: var(--color-success);
}
.a-label {
  font-size: 0.625rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.a-text {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  line-height: 1.4;
}
</style>
