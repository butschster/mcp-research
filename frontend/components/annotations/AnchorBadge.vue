<template>
  <span :class="['anchor-badge', `anchor-badge--${state}`]" :title="hint">
    <span class="anchor-badge__mark" aria-hidden="true">{{ glyph }}</span>
    {{ label }}<span v-if="showConfidence" class="anchor-badge__conf">{{ confidenceText }}</span>
  </span>
</template>

<script setup lang="ts">
/**
 * Where the marked text is now.
 *
 * A third axis, and deliberately not folded into StatusBadge: `drifted` and
 * `open` are not two values of one thing. A mark can be open and drifted, or
 * answered and orphaned, and a single badge would have to pick one to show.
 *
 * Confidence is printed only when the placement was recovered rather than
 * exact. Showing "100%" next to every ordinary mark trains people to stop
 * reading the number, which is the number that matters on the one row where it
 * is 60.
 */
import { ANCHOR_META, type AnchorState } from '~/composables/useAnnotations'

const props = defineProps<{
  state: AnchorState
  confidence?: number
  entryType?: string
}>()

const GLYPHS: Record<AnchorState, string> = {
  anchored: '⚓',
  drifted: '≈',
  moved: '⤳',
  orphaned: '⊘',
}

// A state this build does not know is reported as itself. Falling back to
// `anchored` printed the most reassuring word in the vocabulary for the one
// condition the badge cannot vouch for.
const known = computed(() => !!ANCHOR_META[props.state])
const meta = computed(() => ANCHOR_META[props.state] ?? { label: props.state, underline: 'none', hint: 'Unrecognised anchor state' })
const label = computed(() => meta.value.label)
const glyph = computed(() => (known.value ? GLYPHS[props.state] : '?'))

// A markdown document has no blocks to pin to, so its marks are found by quote
// alone. Saying that out loud beats implying a precision the document cannot
// give.
const hint = computed(() => props.entryType === 'markdown'
  ? `${meta.value.hint}. This document has no blocks, so the mark is found by its text alone.`
  : meta.value.hint)

const showConfidence = computed(() =>
  props.state === 'moved' && typeof props.confidence === 'number' && props.confidence < 1)
const confidenceText = computed(() => ` ${Math.round((props.confidence ?? 0) * 100)}%`)
</script>

<style scoped>
.anchor-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  padding: 1px var(--space-2);
  border-radius: var(--radius-sm);
  font-size: var(--type-3xs);
  background: var(--color-surface-sunken);
  color: var(--color-text-muted);
  white-space: nowrap;
}

.anchor-badge__mark { line-height: 1; }
.anchor-badge__conf { color: var(--color-text-faint); }

.anchor-badge--drifted  { background: rgba(var(--color-warning-rgb), 0.10); color: var(--color-warning); }
.anchor-badge--orphaned { background: rgba(var(--color-error-rgb), 0.10); color: var(--color-error); }
</style>
