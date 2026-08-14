import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import RoleSelect from './RoleSelect.vue'
import type { TeamRole } from '~/composables/useTeams'

/**
 * Picks one of the three roles.
 *
 * A native select on purpose: role is a three-value enum, which is the exact
 * case the element exists for, and it brings keyboard, screen-reader and mobile
 * behaviour a custom menu would have to reimplement.
 *
 * It also owns the last-owner refusal. The reason has to reach the pointer
 * (`title`) and the screen reader (`aria-describedby`) both, and a caller that
 * remembers only one of them ships a control that is disabled for no stated
 * reason — so the control carries it, not the caller.
 *
 * Two consumers: the member rows, where it is bare and labelled after the
 * person, and the invite dialog, where `describe` turns on because the reader
 * is choosing for someone who is not in the room yet.
 */
const meta: Meta<typeof RoleSelect> = {
  title: 'Team/RoleSelect',
  component: RoleSelect,
  tags: ['autodocs'],
  argTypes: {
    modelValue: { control: 'select', options: ['viewer', 'editor', 'owner'] },
    disabled: { control: 'boolean' },
    disabledReason: { control: 'text' },
    busy: { control: 'boolean' },
    describe: { control: 'boolean' },
    label: { control: 'text' },
    'onUpdate:modelValue': { action: 'update:modelValue' },
  },
}
export default meta
type Story = StoryObj<typeof RoleSelect>

const LAST_OWNER = 'A team needs at least one owner. Make someone else an owner first.'

/** The default a new invitation starts from: an under-permissioned colleague
 *  says so in thirty seconds, an over-permissioned one never does. */
export const Viewer: Story = {
  args: { modelValue: 'viewer' },
}

export const Editor: Story = {
  args: { modelValue: 'editor' },
}

export const Owner: Story = {
  args: { modelValue: 'owner' },
}

/** `describe` on: the one line under the control saying what the role may do.
 *  The invite dialog sets it; the member rows do not, because five stacked
 *  descriptions in a table is a document, not a list. */
export const WithDescription: Story = {
  args: { modelValue: 'editor', describe: true },
}

/** All three, described, side by side — the whole permission model on one
 *  screen. */
export const AllRoles: Story = {
  render: () => ({
    components: { RoleSelect },
    setup: () => ({ roles: ['viewer', 'editor', 'owner'] as TeamRole[] }),
    template: `
      <div style="display: flex; gap: var(--space-6); flex-wrap: wrap;">
        <div v-for="role in roles" :key="role" style="max-width: 220px;">
          <RoleSelect :model-value="role" describe :label="'Role: ' + role" />
        </div>
      </div>
    `,
  }),
}

/** The last owner. The team would be left with nobody who can manage it, so the
 *  control is not offered in a state the server would refuse — and it says why,
 *  on hover and to a screen reader. */
export const LastOwnerDisabled: Story = {
  args: { modelValue: 'owner', disabled: true, disabledReason: LAST_OWNER },
}

/** Disabled with the reason left out — what a caller ships by forgetting it.
 *  Kept in the catalogue as the thing not to do: the control refuses and
 *  explains nothing. */
export const DisabledWithoutReason: Story = {
  args: { modelValue: 'owner', disabled: true },
}

/** A change is in flight. The select locks rather than queueing a second
 *  request against the row that is still settling. */
export const Busy: Story = {
  args: { modelValue: 'editor', busy: true },
}

/** In a member row the accessible name carries the person, so a screen reader
 *  hears whose role is being changed rather than "Role, combo box" five times. */
export const LabelledAfterAPerson: Story = {
  args: { modelValue: 'editor', label: 'Role of Екатерина Ковалёва' },
}

/** Wired to a `ref`, which is how both callers use it: pick a role and the
 *  description under it follows the selection. */
export const Interactive: Story = {
  render: () => ({
    components: { RoleSelect },
    setup() {
      const role = ref<TeamRole>('viewer')
      return { role }
    },
    template: `
      <div style="max-width: 260px; display: flex; flex-direction: column; gap: var(--space-3);">
        <RoleSelect v-model="role" describe label="Role for the invited person" />
        <p style="margin: 0; font-size: var(--type-xs); color: var(--color-text-muted);">
          Emitted: <code>{{ role }}</code>
        </p>
      </div>
    `,
  }),
}
