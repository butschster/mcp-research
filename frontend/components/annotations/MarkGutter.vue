<template>
  <div class="mark-gutter" aria-hidden="true">
    <button
      v-for="group in groups"
      :key="group.id"
      type="button"
      tabindex="-1"
      :class="['mark-gutter__pin', `mark-gutter__pin--${group.kind}`, activeId === group.id && 'is-active']"
      :style="{ top: `${group.top}px` }"
      @click="$emit('select', group.id)"
    >
      <span class="mark-gutter__glyph">{{ group.glyph }}</span>
      <span class="mark-gutter__label">{{ group.label }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
/**
 * The column of markers beside the prose.
 *
 * Positioned from measurements of somebody else's DOM — the document is
 * rendered through `v-html`, so there is no component tree to hang these off,
 * only rectangles. Nothing in the catalog does that.
 *
 * `aria-hidden` and `tabindex="-1"` throughout, and this is the compromise
 * worth naming: forty marks would otherwise be forty tab stops inside prose,
 * which breaks reading by keyboard far more than the markers help. The keyboard
 * route in is the counter button in the header and the skip link at the top of
 * the card; each mark also carries its own screen-reader tail in the text.
 */
import { KIND_META, type Annotation } from '~/composables/useAnnotations'

const props = defineProps<{
  annotations: Annotation[]
  positions: Array<{ id: string; code: string; top: number }>
  activeId?: string | null
}>()

defineEmits<{ select: [id: string] }>()

/** Marks closer together than this share one pin — two circles overlapping is
 *  worse than a count. */
const CLUSTER = 24

const groups = computed(() => {
  const byId = new Map(props.annotations.map((a) => [a.id, a]))
  const out: Array<{ id: string; top: number; kind: string; glyph: string; label: string; count: number }> = []

  for (const pos of props.positions) {
    const a = byId.get(pos.id)
    if (!a) continue
    const last = out[out.length - 1]
    if (last && pos.top - last.top < CLUSTER) {
      // A cluster stops naming one mark and starts counting them: showing the
      // first code would claim the pin belongs to it. The count is its own
      // field — reading it back out of the label worked only while every code
      // began with a letter.
      last.count += 1
      last.label = String(last.count)
      last.glyph = '•'
      continue
    }
    out.push({
      id: a.id,
      top: pos.top,
      kind: a.kind,
      glyph: KIND_META[a.kind]?.glyph ?? '•',
      label: a.code,
      count: 1,
    })
  }
  return out
})
</script>

<style scoped>
.mark-gutter {
  position: absolute;
  top: 0;
  left: var(--entry-pad, var(--space-5));
  width: var(--ann-gutter);
  height: 100%;
  pointer-events: none;
}

.mark-gutter__pin {
  position: absolute;
  left: 0;
  /* Inside the column, not 2px over it: the overhang sat on the first
     characters of every line it stood beside and ate their clicks. */
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1px;
  width: var(--ann-gutter);
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
  pointer-events: auto;
}

.mark-gutter__glyph {
  display: flex;
  align-items: center;
  justify-content: center;
  width: var(--ann-gutter);
  height: var(--ann-gutter);
  border-radius: 50%;
  font-size: var(--type-2xs);
  font-weight: 700;
  line-height: 1;
}

.mark-gutter__label {
  max-width: var(--ann-gutter);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
  line-height: 1;
}

.mark-gutter__pin--verify   .mark-gutter__glyph { background: var(--ann-verify-wash);   color: var(--ann-verify); }
.mark-gutter__pin--dig      .mark-gutter__glyph { background: var(--ann-dig-wash);      color: var(--ann-dig); }
.mark-gutter__pin--disagree .mark-gutter__glyph { background: var(--ann-disagree-wash); color: var(--ann-disagree); }

.mark-gutter__pin.is-active .mark-gutter__glyph {
  outline: 2px solid var(--color-border-strong);
  outline-offset: 1px;
}

/* No column at this width — the card's padding is down to a few pixels. The
   document falls back to inline chips beside the marked text. */
@media (max-width: 768px) {
  .mark-gutter { display: none; }
}
</style>
