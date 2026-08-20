<template>
  <span :class="['kind-chip', `kind-chip--${kind}`, size === 'sm' && 'kind-chip--sm']">
    <span class="kind-chip__glyph" aria-hidden="true">{{ meta.glyph }}</span>
    <span v-if="!iconOnly" class="kind-chip__label">{{ meta.label }}</span>
    <span v-else class="sr-only">{{ meta.label }}</span>
  </span>
</template>

<script setup lang="ts">
/**
 * What a mark asks for.
 *
 * Separate from StatusBadge on purpose: a status answers "where is this in its
 * life", a kind answers "what is the agent supposed to do about it". Folding a
 * second vocabulary into that component would put `verify` and `open` in one
 * dictionary, and they are not the same kind of word.
 *
 * The glyph and the label both ship, always — colour alone must never be what
 * tells the two apart.
 */
import { KIND_META, type AnnotationKind } from '~/composables/useAnnotations'

const props = withDefaults(defineProps<{
  kind: AnnotationKind
  size?: 'sm' | 'md'
  iconOnly?: boolean
}>(), { size: 'md', iconOnly: false })

const meta = computed(() => KIND_META[props.kind] ?? { glyph: '•', label: props.kind, hint: '' })
</script>

<style scoped>
.kind-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  padding: 0 var(--space-2);
  min-height: var(--control-h-sm);
  border-radius: var(--radius-sm);
  font-size: var(--type-2xs);
  font-weight: 500;
  white-space: nowrap;
}

.kind-chip--sm {
  min-height: auto;
  padding: 1px var(--space-1);
  font-size: var(--type-3xs);
}

.kind-chip__glyph {
  font-weight: 700;
  line-height: 1;
}

.kind-chip--verify   { background: var(--ann-verify-wash);   color: var(--ann-verify); }
.kind-chip--dig      { background: var(--ann-dig-wash);      color: var(--ann-dig); }
.kind-chip--disagree { background: var(--ann-disagree-wash); color: var(--ann-disagree); }
</style>
