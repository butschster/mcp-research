<template>
  <div class="role-select">
    <select
      :id="inputId"
      :value="modelValue"
      :disabled="disabled"
      :aria-busy="busy || undefined"
      :title="disabled ? disabledReason : undefined"
      :aria-describedby="disabled && disabledReason ? reasonId : (describe ? descriptionId : undefined)"
      :aria-label="label"
      @change="onChange"
    >
      <option v-for="role in roles" :key="role" :value="role">{{ ROLE_LABELS[role] }}</option>
    </select>
    <p v-if="describe" :id="descriptionId" class="role-description">
      {{ ROLE_DESCRIPTIONS[modelValue] }}
    </p>
    <p v-if="disabled && disabledReason" :id="reasonId" class="sr-only">{{ disabledReason }}</p>
  </div>
</template>

<script setup lang="ts">
import { ROLE_DESCRIPTIONS, ROLE_LABELS, type TeamRole } from '~/composables/useTeams'

/**
 * Picks one of the three roles.
 *
 * `busy` dims the control rather than disabling it: the browser blurs a
 * disabled element, so disabling the select someone has just used sends their
 * next Tab back to the top of the page.
 *
 * A native `<select>` on purpose: role is a three-value enum, which is the
 * exact case the element exists for, and it brings keyboard, screen-reader and
 * mobile behaviour that a custom menu would have to reimplement.
 *
 * It also owns the "last owner" refusal, because the reason has to reach both
 * the pointer (`title`) and the screen reader (`aria-describedby`), and a
 * caller that only remembers one of those ships a control that is disabled for
 * no stated reason.
 */
const props = withDefaults(
  defineProps<{
    modelValue: TeamRole
    disabled?: boolean
    disabledReason?: string
    busy?: boolean
    describe?: boolean
    label?: string
    /** Id for the select itself, so a caller's `label for=` reaches it rather
     *  than landing on this component's wrapper. */
    inputId?: string
  }>(),
  { label: 'Role' },
)

const emit = defineEmits<{ 'update:modelValue': [role: TeamRole] }>()

const roles: TeamRole[] = ['viewer', 'editor', 'owner']
const uid = useId()
const descriptionId = `role-desc-${uid}`
const reasonId = `role-reason-${uid}`

function onChange(event: Event) {
  emit('update:modelValue', (event.target as HTMLSelectElement).value as TeamRole)
}
</script>

<style scoped>
.role-select { display: flex; flex-direction: column; gap: var(--space-1); min-width: 0; }
select[aria-busy='true'] { opacity: 0.6; }
select {
  padding: 0.4rem 0.6rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-size: var(--type-xs);
  font-family: inherit;
  cursor: pointer;
  transition: all var(--transition-fast);
}
select:hover:not(:disabled) { border-color: rgba(108, 197, 224, 0.25); }
select:disabled { cursor: not-allowed; opacity: 0.55; }
.role-description {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  margin: 0;
}
</style>
