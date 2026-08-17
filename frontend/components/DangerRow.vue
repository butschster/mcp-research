<template>
  <div class="danger-row">
    <div class="danger-text">
      <span class="danger-label">{{ label }}</span>
      <span v-if="note" class="danger-note">{{ note }}</span>
      <!-- A refusal that names no way out is a dead end. The caller puts the
           way out here: "choose a new owner", "move the researches". -->
      <p v-if="disabled && disabledReason" :id="reasonId" class="danger-reason">
        {{ disabledReason }}
        <slot name="escape" />
      </p>
    </div>

    <button
      type="button"
      class="btn btn-sm"
      :class="{ 'btn-danger': !disabled }"
      :disabled="disabled || busy"
      :aria-describedby="disabled && disabledReason ? reasonId : undefined"
      @click="emit('action')"
    >
      {{ busy ? busyLabel : actionLabel }}
    </button>
  </div>
</template>

<script setup lang="ts">
/**
 * One irreversible action, with the reason it might be refused.
 *
 * The label, the consequence and the disabled-reason were written by hand twice
 * on the team page and zero times on the settings page — which is why revoking
 * an API key there is a bare text link with no confirmation at all.
 *
 * The row is a `<div>` with a button in it rather than one big button, because
 * a refusal has to be able to carry its own escape hatch — "you are the only
 * owner, choose a new one" is a link, and a link inside a button is not markup
 * a browser will keep. The reason is visible rather than a `title`, since a
 * disabled control's tooltip reaches neither a keyboard nor a screen reader.
 */
withDefaults(
  defineProps<{
    label: string
    /** What happens if they press it. */
    note?: string
    actionLabel: string
    busyLabel?: string
    busy?: boolean
    disabled?: boolean
    /** Why it is refused. Rendered, not hidden — and wired to the button. */
    disabledReason?: string
  }>(),
  { busyLabel: 'Working…' },
)

const emit = defineEmits<{ action: [] }>()

const reasonId = `danger-reason-${useId()}`
</script>

<style scoped>
.danger-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-3) 0;
}
.danger-row + .danger-row { border-top: 1px solid var(--color-border); }
.danger-text { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.danger-label { font-size: var(--type-sm); color: var(--color-text); }
.danger-note { font-size: var(--type-xs); color: var(--color-text-muted); }
.danger-reason {
  margin: var(--space-1) 0 0;
  font-size: var(--type-xs);
  color: var(--color-warning);
}
.danger-reason :slotted(*) { margin-left: var(--space-2); }

@media (max-width: 768px) {
  .danger-row { flex-direction: column; align-items: stretch; gap: var(--space-2); }
}
</style>
