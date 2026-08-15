<template>
  <li
    :class="['rev-row', { 'rev-selected': selected, 'rev-base': isBase }]"
    role="option"
    :aria-selected="selected"
  >
    <button
      class="rev-button"
      :tabindex="selected ? 0 : -1"
      :title="revision.summary || undefined"
      @click="onClick"
      @keydown.enter.shift.prevent="$emit('setBase', revision.revision)"
    >
      <span class="rev-main">
        <span class="rev-summary">{{ revision.summary || 'Created' }}</span>
        <span class="rev-meta">
          <EntryAuthorBadge :kind="revision.author_kind" />
          <span aria-hidden="true">·</span>
          <time :datetime="revision.created_at" :title="absoluteTime(revision.created_at)">
            {{ relativeTime(revision.created_at) }}
          </time>
          <span v-if="isCurrent" class="rev-current">current</span>
        </span>
      </span>
      <span class="rev-number" aria-hidden="true">r{{ revision.revision }}</span>
      <span class="sr-only">revision {{ revision.revision }}</span>
    </button>
  </li>
</template>

<script setup lang="ts">
/**
 * One revision in the rail.
 *
 * Two lines, never three: the summary is what a reader scans for, so it leads,
 * and author and time recede beneath it. The revision number sits in a
 * fixed-width mono spine on the right — a straight vertical the ragged
 * summaries hang off, which is what stops six similar rows reading as a wall.
 *
 * Weight 400 on the number is deliberate: this project loads JetBrains Mono at
 * 400 and 500 only, so a heavier weight synthesises faux-bold and smears at
 * this size, worst in Cyrillic.
 */
import { relativeTime, absoluteTime } from '~/composables/useRelativeTime'

const props = defineProps<{
  revision: {
    revision: number
    summary?: string
    author_kind: string
    created_at: string
    session_code?: string
  }
  selected: boolean
  isBase?: boolean
  isCurrent?: boolean
}>()

const emit = defineEmits<{ select: [number]; setBase: [number] }>()

// Shift-click picks the far end of a comparison; a plain click moves the near
// end. One target, two meanings, no mode to enter.
function onClick(event: MouseEvent) {
  if (event.shiftKey) emit('setBase', props.revision.revision)
  else emit('select', props.revision.revision)
}
</script>

<style scoped>
.rev-row {
  border-left: 2px solid transparent;
  border-bottom: 1px solid var(--color-border);
}
.rev-row:last-child { border-bottom: none; }
.rev-selected {
  background: var(--color-surface-hover);
  border-left-color: var(--color-primary);
}
.rev-base {
  background: color-mix(in srgb, var(--color-primary) 8%, transparent);
  border-left-color: var(--color-primary);
  border-left-style: dashed;
}
.rev-button {
  display: flex;
  align-items: flex-start;
  gap: var(--space-2);
  width: 100%;
  padding: var(--space-3) var(--space-4);
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  color: inherit;
}
.rev-row:not(.rev-selected):hover { background: var(--color-surface-hover); }
.rev-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}
.rev-summary {
  font-size: var(--type-sm);
  font-weight: var(--weight-medium);
  color: var(--color-text);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.rev-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-1);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.rev-current {
  color: var(--color-text-muted);
  font-style: italic;
}
.rev-number {
  flex-shrink: 0;
  width: 3.5ch;
  text-align: right;
  font-family: 'JetBrains Mono', monospace;
  font-variant-numeric: tabular-nums;
  font-weight: var(--weight-normal);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  padding-top: 0.15em;
}
</style>
