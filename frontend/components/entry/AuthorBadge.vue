<template>
  <span :class="['author', `author-${kind}`, variant === 'glyph' ? 'author-glyph-only' : '']">
    <span class="author-glyph" aria-hidden="true">{{ meta.glyph }}</span>
    <span v-if="variant !== 'glyph'" class="author-word">{{ meta.word }}</span>
    <span class="sr-only">{{ meta.label }}</span>
  </span>
</template>

<script setup lang="ts">
import { authorKind } from '~/composables/useAuthorKind'
/**
 * Who wrote a revision — the one signal in this feature a reader looks for
 * first, because "a person wrote this" and "a model wrote this" are different
 * claims about the same document.
 *
 * Only two of the four kinds carry colour. In a product where an agent writes
 * nearly everything, the human edit is the exception worth spotting and the
 * restore is the event that explains a paragraph coming back; colouring all
 * four would turn a list of revisions into a fruit salad and make none of them
 * stand out.
 *
 * Separate from StatusBadge on purpose: folding these into its map would make
 * `<StatusBadge status="agent" />` a legal call on an entry's status.
 */
const props = withDefaults(defineProps<{
  kind: 'agent' | 'human' | 'import' | 'restore' | string
  variant?: 'inline' | 'glyph'
}>(), { variant: 'inline' })

// The vocabulary lives in a composable because a menu header needs the same
// four words and cannot import them from inside a `<script setup>`.
const meta = computed(() => authorKind(props.kind))
</script>

<style scoped>
.author {
  display: inline-flex;
  align-items: baseline;
  gap: var(--space-1);
  font-size: inherit;
  color: var(--color-text-muted);
}
.author-glyph {
  font-size: 0.85em;
  line-height: 1;
}
.author-human {
  color: var(--color-primary);
}
.author-restore {
  color: var(--color-warning);
}
.author-glyph-only .author-glyph { font-size: 1em; }
</style>
