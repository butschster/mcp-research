<template>
  <div
    class="data-row ann-row"
    :class="[selected && 'is-selected', dense && 'ann-row--dense']"
    role="option"
    :aria-selected="selected"
    :tabindex="focusable ? 0 : -1"
    @click="$emit('open')"
    @keydown.enter.prevent="$emit('open')"
    @keydown.space.prevent="$emit('toggle-select')"
  >
    <input
      v-if="selectable"
      type="checkbox"
      class="ann-row__check"
      :checked="selected"
      :aria-label="`Select ${annotation.code}`"
      @click.stop
      @change="$emit('toggle-select')"
    >

    <ShortCode :code="annotation.code" />
    <AnnotationsKindChip :kind="annotation.kind" size="sm" />

    <p class="ann-row__quote">
      <span class="ann-row__quote-text">{{ annotation.quote.exact }}</span>
      <!-- A note reading "contradicts [[E7]]" is a reference like any other. -->
      <span v-if="annotation.body" class="ann-row__note" v-html="renderRefs(annotation.body, researchSlug)" />
    </p>

    <span class="ann-row__meta">
      <!-- Only when it says something. "Anchored" on every row is a word that
           teaches the reader to stop looking at the column. -->
      <AnnotationsAnchorBadge
        v-if="annotation.anchor && annotation.anchor.state !== 'anchored'"
        :state="annotation.anchor.state"
        :confidence="annotation.anchor.confidence"
        :entry-type="annotation.entry_type"
      />
      <StatusBadge :status="annotation.status" />
      <span v-if="annotation.attempts > 0" class="ann-row__attempts" :title="`Sent back ${annotation.attempts} times`">
        ↩{{ annotation.attempts }}
      </span>
    </span>
  </div>
</template>

<script setup lang="ts">
/**
 * One mark, as a line.
 *
 * A component rather than the `.data-row` classes on their own, because three
 * places draw this line — the document panel, the research queue and the pass
 * review — and a mark that looks different in the queue than it does in the
 * review is a mark somebody accepts twice.
 */
import type { Annotation } from '~/composables/useAnnotations'
import { renderRefs } from '~/composables/useCrossRefs'

withDefaults(defineProps<{
  annotation: Annotation
  researchSlug: string
  selectable?: boolean
  selected?: boolean
  dense?: boolean
  /** The one row in its list that is a tab stop. Marked text in the prose is
   *  deliberately not focusable — forty marks would be forty stops inside a
   *  paragraph — so the list is the keyboard's way in, and a list nothing can
   *  focus is no way in at all. */
  focusable?: boolean
}>(), { researchSlug: '' })

defineEmits<{ open: []; 'toggle-select': [] }>()
</script>

<style scoped>
.ann-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  cursor: pointer;
}

/* Selection is carried by the border and the checkbox, not by a fill: a tinted
   row and a warning row would otherwise be one glance apart. */
.ann-row.is-selected { border-color: var(--color-border-strong); }

.ann-row__check { flex: none; }

.ann-row__quote {
  flex: 1;
  min-width: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.ann-row__quote-text {
  font-size: var(--type-sm);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ann-row__note {
  font-size: var(--type-2xs);
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ann-row__meta {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex: none;
}

.ann-row__attempts {
  font-size: var(--type-3xs);
  color: var(--color-warning);
}

.ann-row--dense .ann-row__note { display: none; }

@media (max-width: 768px) {
  .ann-row { flex-wrap: wrap; }
  .ann-row__quote { flex-basis: 100%; }
}
</style>
