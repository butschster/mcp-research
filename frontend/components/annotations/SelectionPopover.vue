<template>
  <div
    v-if="visible"
    ref="popover"
    class="sel-pop"
    :style="style"
    role="dialog"
    aria-labelledby="sel-pop-title"
    @keydown.esc.stop.prevent="$emit('cancel')"
    @keydown.meta.enter.prevent="submit"
    @keydown.ctrl.enter.prevent="submit"
  >
    <p id="sel-pop-title" class="sel-pop__title">Mark this text</p>

    <div class="sel-pop__kinds" role="radiogroup" aria-label="What is being asked for">
      <button
        v-for="option in kinds"
        :key="option.value"
        ref="kindButtons"
        type="button"
        role="radio"
        :aria-checked="kind === option.value"
        :class="['sel-pop__kind', `sel-pop__kind--${option.value}`, kind === option.value && 'is-active']"
        :title="option.hint"
        @click="kind = option.value"
      >
        <span class="sel-pop__glyph" aria-hidden="true">{{ option.glyph }}</span>
        {{ option.label }}
      </button>
    </div>

    <textarea
      ref="note"
      v-model="body"
      class="form-textarea sel-pop__note"
      rows="2"
      placeholder="What is wrong with it?"
      aria-label="Note"
    />

    <p class="sel-pop__quote">
      <span class="sel-pop__quote-text">“{{ shortQuote }}”</span>
      <span class="sel-pop__count">{{ quote.length }} ch</span>
    </p>

    <!-- A markdown document has no blocks. Saying so here is cheaper than a
         person discovering later that the mark drifts more easily. -->
    <p v-if="entryType === 'markdown'" class="sel-pop__hint">
      This document has no blocks, so the mark is found by its text alone.
    </p>

    <p v-if="error" class="sel-pop__error" role="alert">{{ error }}</p>

    <div class="sel-pop__actions">
      <button type="button" class="btn btn-sm" :disabled="saving" @click="$emit('cancel')">Cancel</button>
      <button type="button" class="btn btn-sm btn-primary" :disabled="saving" @click="submit">
        {{ saving ? 'Marking…' : 'Mark' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * The thing that appears when a person selects a sentence.
 *
 * Neither ActionMenu nor ModalOverlay can do this job: one is a menu anchored to
 * a button, the other is a modal that takes focus — and taking focus is exactly
 * what destroys a live selection. By the time this renders the selection has
 * already been captured and painted by the caller, which is why it can afford a
 * textarea at all.
 *
 * On a narrow screen it stops being a popover and becomes a bar pinned to the
 * bottom. That is not a layout preference: the system Copy/Look-Up menu appears
 * right at the selection on iOS, and anything drawn there fights it.
 */
import { KIND_META, type AnnotationKind } from '~/composables/useAnnotations'

const props = defineProps<{
  visible: boolean
  rect: DOMRect | null
  quote: string
  entryType?: string
  saving?: boolean
  error?: string | null
}>()

const emit = defineEmits<{
  create: [payload: { kind: AnnotationKind; body: string }]
  cancel: []
}>()

const kind = ref<AnnotationKind>('verify')
const body = ref('')
const popover = ref<HTMLElement | null>(null)
const kindButtons = ref<HTMLElement[]>([])
const note = ref<HTMLTextAreaElement | null>(null)

const kinds = computed(() =>
  (Object.keys(KIND_META) as AnnotationKind[]).map((value) => ({ value, ...KIND_META[value] })))

const shortQuote = computed(() =>
  props.quote.length > 90 ? `${props.quote.slice(0, 90)}…` : props.quote)

/**
 * Below the selection, or above it when there is no room.
 *
 * Fixed positioning against the viewport rect the range reported. No animated
 * movement — a box that slides toward the words being read is harder to aim at,
 * not friendlier.
 */
const style = computed(() => {
  const r = props.rect
  if (!r) return {}
  const width = 320
  const estimatedHeight = 220
  const below = r.bottom + 8
  const flip = below + estimatedHeight > window.innerHeight
  const left = Math.min(
    Math.max(8, r.left + r.width / 2 - width / 2),
    Math.max(8, window.innerWidth - width - 8),
  )
  return {
    top: flip ? `${Math.max(8, r.top - estimatedHeight - 8)}px` : `${below}px`,
    left: `${left}px`,
  }
})

function submit() {
  if (props.saving) return
  emit('create', { kind: kind.value, body: body.value.trim() })
}

// Reset between selections: a note typed for one sentence must never be
// attached to the next one.
watch(() => props.visible, (open) => {
  if (open) {
    kind.value = 'verify'
    body.value = ''
  }
})

/** Focus moves here only when asked — see the keyboard path. Opening a popover
 *  that steals focus would make the mouse path feel like it lost the page. */
function focusFirst() {
  kindButtons.value[0]?.focus()
}

defineExpose({ focusFirst, focusNote: () => note.value?.focus() })
</script>

<style scoped>
.sel-pop {
  position: fixed;
  z-index: var(--z-overlay);
  width: min(320px, 92vw);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding: var(--space-3);
  background: var(--color-surface-raised);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-2);
}

.sel-pop__title {
  margin: 0;
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}

.sel-pop__kinds {
  display: flex;
  gap: var(--space-1);
}

.sel-pop__kind {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
  min-height: var(--control-h-sm);
  padding: 0 var(--space-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-muted);
  font-size: var(--type-2xs);
  cursor: pointer;
}

.sel-pop__kind:hover { border-color: var(--color-border-strong); }

.sel-pop__glyph { font-weight: 700; line-height: 1; }

/* The active type is carried by border, background AND the glyph already in the
   label — never by colour alone. */
.sel-pop__kind--verify.is-active   { background: var(--ann-verify-wash);   color: var(--ann-verify);   border-color: var(--ann-verify); }
.sel-pop__kind--dig.is-active      { background: var(--ann-dig-wash);      color: var(--ann-dig);      border-color: var(--ann-dig); }
.sel-pop__kind--disagree.is-active { background: var(--ann-disagree-wash); color: var(--ann-disagree); border-color: var(--ann-disagree); }

.sel-pop__note { resize: vertical; }

.sel-pop__quote {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-2);
  margin: 0;
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
}

.sel-pop__quote-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sel-pop__count { flex: none; }

.sel-pop__hint {
  margin: 0;
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
}

.sel-pop__error {
  margin: 0;
  font-size: var(--type-2xs);
  color: var(--color-error);
}

.sel-pop__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}

/* Narrow screens: a bar at the bottom, clear of the system selection menu. */
@media (max-width: 768px) {
  .sel-pop {
    top: auto !important;
    left: 0 !important;
    right: 0;
    bottom: 0;
    width: 100%;
    border-radius: var(--radius) var(--radius) 0 0;
    padding-bottom: calc(var(--space-3) + env(safe-area-inset-bottom));
  }

  .sel-pop__kind { min-height: 36px; }
}
</style>
