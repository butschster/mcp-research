<template>
  <button
    type="button"
    :class="['rm-bar', `bar-type-${data.nodeType || 'step'}`, { 'is-highlighted': highlighted, 'is-dimmed': dimmed }]"
    :title="fullTitle"
    @click="emit('click')"
    @mouseenter="emit('hover', true)"
    @mouseleave="emit('hover', false)"
    @focus="emit('hover', true)"
    @blur="emit('hover', false)"
  >
    <span v-if="data.code" class="rm-bar-code">{{ data.code }}</span>
    <span class="rm-bar-title">{{ data.title }}</span>
    <span v-if="deps && deps.length" class="rm-bar-deps">&#x21B3;{{ deps.map(d => d.code).join(' ') }}</span>
  </button>
</template>

<script setup lang="ts">
import type { RoadmapCardData } from './RoadmapNodeCard.vue'

// A compact horizontal bar spanning several axis cells — what RoadmapNodeCard
// cannot be, since a card is a full vertical block. The parent applies the grid
// span (grid-column / grid-row); this only renders the bar's content and state.
const props = defineProps<{
  data: RoadmapCardData
  startLabel: string
  endLabel: string
  deps?: { code: string }[]
  highlighted?: boolean
  dimmed?: boolean
}>()

const emit = defineEmits<{ click: []; hover: [boolean] }>()

const fullTitle = computed(() => {
  const range = props.startLabel === props.endLabel ? props.startLabel : `${props.startLabel} – ${props.endLabel}`
  return `${props.data.code ? props.data.code + ' ' : ''}${props.data.title || ''} (${range})`
})
</script>

<style scoped>
.rm-bar {
  appearance: none;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  width: 100%;
  height: 100%;
  padding: var(--space-1) var(--space-2);
  border: 1px solid var(--color-border);
  border-left: 3px solid var(--color-border-strong);
  border-radius: var(--radius-xs);
  background: var(--color-surface);
  color: var(--color-text);
  font: inherit;
  text-align: left;
  cursor: pointer;
  overflow: hidden;
  white-space: nowrap;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast), opacity var(--transition-fast);
}
.rm-bar:hover,
.rm-bar.is-highlighted {
  border-color: var(--color-border-strong);
  box-shadow: var(--shadow-1);
}
.rm-bar.is-dimmed { opacity: 0.5; }
.rm-bar:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

/* Left-accent tint per node type, mirroring the card's tints */
.bar-type-step      { border-left-color: rgba(107, 203, 119, 0.7); background: rgba(107, 203, 119, 0.06); }
.bar-type-milestone { border-left-color: rgba(168, 130, 255, 0.7); background: rgba(168, 130, 255, 0.08); }
.bar-type-decision  { border-left-color: rgba(240, 184, 73, 0.7);  background: rgba(240, 184, 73, 0.07); }
.bar-type-info      { border-left-color: rgba(108, 197, 224, 0.7); background: rgba(108, 197, 224, 0.06); }
.bar-type-group     { border-left-color: rgba(160, 160, 160, 0.7); background: rgba(160, 160, 160, 0.06); }

.rm-bar-code {
  flex-shrink: 0;
  font-size: 0.5625rem;
  font-weight: var(--weight-bold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.05rem 0.25rem;
  border-radius: var(--radius-xs);
  font-family: 'JetBrains Mono', monospace;
}
.rm-bar-title {
  font-size: var(--type-xs);
  overflow: hidden;
  text-overflow: ellipsis;
}
.rm-bar-deps {
  flex-shrink: 0;
  margin-left: auto;
  font-size: 0.5625rem;
  color: var(--color-text-faint);
  font-family: 'JetBrains Mono', monospace;
}
</style>
