<template>
  <div class="mindmap-node answer-node" @click="navigate">
    <div class="a-header">
      <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
      <span class="a-label">Answer</span>
    </div>
    <div class="a-text" v-html="renderInline(data.answer, 120)"></div>
    <Handle type="target" :position="targetPosition" />
  </div>
</template>

<script setup lang="ts">
import { truncate } from '~/utils/truncate'
import { Handle, Position } from '@vue-flow/core'
import { parseMarkdownInline } from '~/composables/useSafeMarkdown'
import { linkRefs } from '~/composables/useCrossRefs'
import { normalizeContent } from '~/utils/normalizeContent'


const props = defineProps<{
  data: {
    answer: string
    questionCode: string
    sessionId: string
    researchSlug: string
  }
  targetPosition?: Position
}>()


function renderInline(text: string, len: number): string {
  const truncated = truncate(normalizeContent(text), len)
  return linkRefs(parseMarkdownInline(truncated) as string, props.data.researchSlug)
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
  box-shadow: var(--shadow-1);
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
  font-weight: var(--weight-bold);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.a-text {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  line-height: 1.4;
}
</style>
