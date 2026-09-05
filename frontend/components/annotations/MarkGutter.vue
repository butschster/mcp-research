<template>
  <div class="mark-gutter" aria-hidden="true">
    <div
      v-for="row in rows"
      :key="row.key"
      class="mark-gutter__row"
      :style="{ top: `${row.top}px` }"
    >
      <button
        v-for="pin in row.pins"
        :key="pin.id"
        type="button"
        tabindex="-1"
        :class="['mark-gutter__pin', `mark-gutter__pin--${pin.kind}`, activeId === pin.id && 'is-active']"
        :title="pin.title"
        @click="$emit('select', pin.id)"
      >
        {{ pin.glyph }}
      </button>

      <!-- Past what fits, a count — in addition to the pins, never instead of
           them. -->
      <span v-if="row.overflow" class="mark-gutter__more" :title="`${row.overflow} more on this line`">
        +{{ row.overflow }}
      </span>

      <!-- The code only when the row is one mark. Printing it under a row of
           three would claim the row belongs to that one. -->
      <span v-if="row.pins.length === 1 && !row.overflow" class="mark-gutter__label">
        {{ row.pins[0]!.code }}
      </span>
    </div>
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
 * **Marks on one line sit side by side.** They are different marks — "find me a
 * source for this" and "and I do not believe it" is an ordinary pair to leave
 * on one sentence — and collapsing them into a single dot said the opposite of
 * what had happened. Only a line that overflows the column gets a count, and
 * the count is extra rather than a replacement.
 *
 * `aria-hidden` and `tabindex="-1"` throughout, and this is the compromise
 * worth naming: forty marks would otherwise be forty tab stops inside prose,
 * which breaks reading by keyboard far more than the markers help. The keyboard
 * route in is the counter button in the header and the skip link at the top of
 * the card; each mark also carries its own screen-reader label in the text.
 */
import { KIND_META, type Annotation } from '~/composables/useAnnotations'

const props = defineProps<{
  annotations: Annotation[]
  positions: Array<{ id: string; code: string; top: number }>
  activeId?: string | null
}>()

defineEmits<{ select: [id: string] }>()

/** Marks within this many pixels of each other are on one line of prose. */
const SAME_LINE = 10
/** How many pins the column holds before it starts counting. */
const PER_ROW = 3

interface Pin {
  id: string
  code: string
  kind: string
  glyph: string
  title: string
}

const rows = computed(() => {
  const byId = new Map(props.annotations.map((a) => [a.id, a]))
  const out: Array<{ key: string; top: number; pins: Pin[]; overflow: number }> = []

  for (const pos of props.positions) {
    const a = byId.get(pos.id)
    if (!a) continue
    const pin: Pin = {
      id: a.id,
      code: a.code,
      kind: a.kind,
      glyph: KIND_META[a.kind]?.glyph ?? '•',
      title: `${a.code} · ${KIND_META[a.kind]?.label ?? a.kind}${a.body ? ` — ${a.body}` : ''}`,
    }

    const last = out[out.length - 1]
    if (last && Math.abs(pos.top - last.top) < SAME_LINE) {
      if (last.pins.length < PER_ROW) last.pins.push(pin)
      else last.overflow += 1
      continue
    }
    out.push({ key: a.id, top: pos.top, pins: [pin], overflow: 0 })
  }
  return out
})
</script>

<style scoped>
.mark-gutter {
  position: absolute;
  top: 0;
  /* At the card's edge rather than against the first character of the prose.
     The column used to start at --entry-pad, which put the circles hard up
     against the text and made them read as part of it. */
  left: var(--space-3);
  width: var(--ann-gutter);
  height: 100%;
  pointer-events: none;
}

.mark-gutter__row {
  position: absolute;
  left: 0;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 2px;
  width: var(--ann-gutter);
}

.mark-gutter__pin {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: none;
  width: var(--ann-pin);
  height: var(--ann-pin);
  padding: 0;
  border: 0;
  border-radius: 50%;
  font-size: var(--type-3xs);
  font-weight: 700;
  line-height: 1;
  cursor: pointer;
  pointer-events: auto;
}

.mark-gutter__more {
  flex: none;
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
  line-height: 1;
}

.mark-gutter__label {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 1px;
  max-width: var(--ann-gutter);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
  line-height: 1;
}

.mark-gutter__pin--verify   { background: var(--ann-verify-wash);   color: var(--ann-verify); }
.mark-gutter__pin--dig      { background: var(--ann-dig-wash);      color: var(--ann-dig); }
.mark-gutter__pin--disagree { background: var(--ann-disagree-wash); color: var(--ann-disagree); }

.mark-gutter__pin.is-active {
  outline: 2px solid var(--color-border-strong);
  outline-offset: 1px;
}
.mark-gutter__pin:focus-visible { outline-color: var(--color-primary); }

/* No column at this width — the card's padding is down to a few pixels. The
   marks stay underlined in the prose and are reached from the header counter. */
@media (max-width: 768px) {
  .mark-gutter { display: none; }
}
</style>
