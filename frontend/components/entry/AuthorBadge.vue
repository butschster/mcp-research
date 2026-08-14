<template>
  <span :class="['author', `author-${kind}`, variant === 'glyph' ? 'author-glyph-only' : '']">
    <span class="author-glyph" aria-hidden="true">{{ meta.glyph }}</span>
    <span v-if="variant !== 'glyph'" class="author-word">{{ meta.word }}</span>
    <span class="sr-only">{{ meta.label }}</span>
  </span>
</template>

<script setup lang="ts">
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

const KINDS: Record<string, { glyph: string; word: string; label: string }> = {
  human: { glyph: '●', word: 'person', label: 'written by a person' },
  restore: { glyph: '↺', word: 'restore', label: 'restored from an earlier revision' },
  agent: { glyph: '◇', word: 'agent', label: 'written by an agent' },
  import: { glyph: '⇩', word: 'import', label: 'written by an import' },
}

const meta = computed(() => KINDS[props.kind] ?? { glyph: '◇', word: props.kind, label: `written by ${props.kind}` })
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
