<template>
  <div class="split">
    <button
      class="btn btn-sm split-main"
      :class="{ 'is-busy': busy, 'split-solo': !withOptions }"
      :aria-disabled="busy || disabled ? 'true' : undefined"
      :aria-busy="busy ? 'true' : undefined"
      :aria-describedby="describedby"
      :title="hint"
      @click="onMain"
    >
      <span class="split-icon" aria-hidden="true">
        <span v-if="busy" class="split-spinner"></span>
        <slot v-else name="icon" />
      </span>
      {{ busy ? busyLabel : label }}
    </button>
    <button
      v-if="withOptions"
      class="btn btn-sm split-caret"
      :class="{ 'is-busy': busy }"
      :aria-label="optionsLabel"
      aria-haspopup="dialog"
      :aria-expanded="optionsOpen ? 'true' : 'false'"
      :aria-disabled="busy || disabled ? 'true' : undefined"
      @click="onOptions"
    >
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="6 9 12 15 18 9"/></svg>
    </button>
  </div>
</template>

<script setup lang="ts">
/**
 * A button whose wide half performs the default act and whose narrow half opens
 * the settings for it.
 *
 * The two halves are two tab stops on purpose: they are two different actions,
 * and folding them into one stop behind arrow keys would invent a widget the
 * rest of the product does not have.
 *
 * `aria-disabled` rather than `disabled` while busy — a disabled button drops
 * focus, and the control a keyboard user just pressed is exactly where they
 * should still be.
 */
const props = withDefaults(defineProps<{
  label: string
  busy?: boolean
  busyLabel?: string
  optionsLabel?: string
  /** Hidden when there is nothing worth configuring on this research. */
  withOptions?: boolean
  optionsOpen?: boolean
  disabled?: boolean
  describedby?: string
  /** Explains the format on hover, so the toolbar does not have to carry a
   *  paragraph that is only read once. */
  hint?: string
}>(), {
  busy: false,
  busyLabel: 'Preparing…',
  optionsLabel: 'Export options',
  withOptions: true,
  optionsOpen: false,
  disabled: false,
})

const emit = defineEmits<{ download: []; options: [] }>()

function onMain() {
  if (props.busy || props.disabled) return
  emit('download')
}
function onOptions() {
  if (props.busy || props.disabled) return
  emit('options')
}
</script>

<style scoped>
.split {
  display: inline-flex;
}
/* The seam: adjoining corners squared off so the two halves read as one
   control, with a hairline saying they do different things. */
.split-main {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}
/* With no caret beside it there is no seam, so the button is a button. */
.split-solo {
  border-top-right-radius: var(--radius-sm);
  border-bottom-right-radius: var(--radius-sm);
}
.split-caret {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  border-left-color: var(--color-border-strong);
  margin-left: -1px;
  padding-inline: var(--space-2);
}
/* A hover lift on half a control looks like a broken control. */
.split-main:hover,
.split-caret:hover,
.split-main:active,
.split-caret:active {
  transform: none;
  box-shadow: none;
}
/* The focus ring must not be painted over by the neighbouring half. */
.split-main:focus-visible,
.split-caret:focus-visible {
  position: relative;
  z-index: 1;
}
.split-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
}
.is-busy {
  opacity: 0.6;
  cursor: progress;
}
.split-spinner {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-primary);
  animation: split-pulse 1.4s ease-in-out infinite;
}
@keyframes split-pulse {
  0%, 100% { opacity: 0.3; transform: scale(0.8); }
  50% { opacity: 1; transform: scale(1); }
}
@media (prefers-reduced-motion: reduce) {
  .split-spinner { animation: none; opacity: 0.8; }
}
</style>
