<template>
  <div class="action-menu-header prov">
    <p class="prov-who">
      <EntryAuthorBadge :kind="authorKind" variant="glyph" class="prov-glyph" />
      <span class="prov-name" :title="who">{{ who }}</span>
    </p>
    <p class="prov-when">
      <span v-if="revision" class="prov-rev">r{{ revision }}</span>
      <span v-if="revision" class="prov-sep" aria-hidden="true">·</span>
      <span>{{ relativeTime(revisedAt) }}</span>
    </p>
  </div>
</template>

<script setup lang="ts">
/**
 * Who wrote this document and when — a fact at the top of the actions panel,
 * and nothing more than a fact.
 *
 * It was a button once, the whole block clickable, on the reasoning that a dead
 * row in a list of verbs reads as broken. That produced the opposite complaint:
 * a link-shaped `View history →` inside a list of buttons, and three lines of
 * near-identical text that looked pasted in rather than laid out. So the fact
 * is a fact, `Revision history` is an ordinary item beneath it, and this block
 * carries no hover and no pointer at all — inertness has to be provable by the
 * thing a person actually tests it with.
 *
 * It renders with or without a revision: a document nobody has revised still has
 * an author kind and a timestamp, and two structurally different menus for one
 * page is worse than one that degrades.
 */
import { computed } from 'vue'
// Aliased: the prop is also called `authorKind`, and in the template the
// import would win.
import { authorKind as kindWords } from '~/composables/useAuthorKind'
import { relativeTime } from '~/composables/useRelativeTime'

const props = defineProps<{
  /** Absent until the first revision exists. */
  revision?: number
  authorKind: string
  revisedAt: string
  /** Absent with auth off, and for anything an agent wrote. */
  authorName?: string
}>()

// The name when there is one, the kind when there is not. The glyph already
// carries the kind, so repeating it beside a name would spend the panel's one
// strong line on something said twice.
const who = computed(() => props.authorName || kindWords(props.authorKind).phrase)
</script>

<style scoped>
.prov-who,
.prov-when { margin: 0; }

.prov-who {
  display: grid;
  /* With a fallback: outside an ActionMenu the property is undefined, and
     an invalid grid-template-columns collapses the row to a stack. */
  grid-template-columns: var(--menu-icon, 14px) 1fr;
  align-items: center;
  /* --space-3, matching `.action-menu-item`'s gap after its icon. At --space-2
     the glyph column lined up and the text after it did not: the name sat four
     pixels left of every verb label, which is precisely the alignment this
     column exists to produce. */
  gap: var(--space-3);
  font-size: var(--type-sm);
  color: var(--color-text);
}

/* The glyph sits in the same column the verbs' icons use, so the name and the
   labels below it start at one x. `variant="glyph"` is the badge's own way of
   dropping the word — no override needed. */
.prov-glyph { justify-self: center; }

.prov-name {
  font-weight: var(--weight-medium);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.prov-when {
  /* Indented to the text column, not to the panel, so the two lines read as one
     block rather than as two rows that happen to be adjacent. */
  margin-left: calc(var(--menu-icon, 14px) + var(--space-3));
  margin-top: 0.1rem;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-xs);
  color: var(--color-text-faint);
}
/* A revision number is a short code, and every short code in this product is
   set in the mono face. `font-variant-numeric` alone was a no-op — body already
   sets it. */
.prov-rev { font-family: 'JetBrains Mono', monospace; }
.prov-sep { opacity: 0.6; }
</style>
