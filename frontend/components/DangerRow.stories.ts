import type { Meta, StoryObj } from '@storybook/vue3'
import DangerRow from './DangerRow.vue'
import DangerZone from './DangerZone.vue'

/**
 * One irreversible action, with the reason it might be refused.
 *
 * The label, the consequence and the disabled-reason were written by hand twice
 * on the team page and zero times on the settings page — which is why revoking
 * an API key there is still a bare text link with no confirmation at all.
 *
 * The row is a `<div>` holding a button rather than one big button, because a
 * refusal has to be able to carry its own escape hatch: "you are the only
 * owner, choose a new one" is a link, and a link inside a button is not markup
 * a browser keeps. The reason is rendered rather than hung on `title`, since a
 * disabled control's tooltip reaches neither a keyboard nor a screen reader —
 * and the reason is the entire point of refusing.
 */
const meta: Meta<typeof DangerRow> = {
  title: 'Layout/DangerRow',
  component: DangerRow,
  tags: ['autodocs'],
  args: { label: 'Leave team', note: 'Removes your access to 18 researches.', actionLabel: 'Leave' },
  argTypes: {
    label: { control: 'text' },
    note: { control: 'text' },
    actionLabel: { control: 'text' },
    busyLabel: { control: 'text' },
    busy: { control: 'boolean' },
    disabled: { control: 'boolean' },
    disabledReason: { control: 'text' },
  },
  decorators: [
    () => ({ components: { DangerZone }, template: '<DangerZone><story /></DangerZone>' }),
  ],
}
export default meta
type Story = StoryObj<typeof DangerRow>

/** Available. The button is outlined red, not filled: reachable, not tempting. */
export const Default: Story = {}

/** Refused, with the reason visible and wired to the button through
 *  `aria-describedby`. */
export const Refused: Story = {
  args: {
    disabled: true,
    disabledReason: 'You are the only owner. Make someone else an owner before leaving.',
  },
}

/**
 * Refused, with a way out.
 *
 * This is why the row is not itself a button. "Choose a new owner" scrolls to
 * the member list and focuses the first role select the reader can change — a
 * refusal that states only the rule leaves them to work out that the fix is
 * elsewhere on the page, in a control they have not noticed.
 */
export const RefusedWithEscape: Story = {
  args: {
    disabled: true,
    disabledReason: 'You are the only owner. Make someone else an owner before leaving.',
  },
  render: (args) => ({
    components: { DangerRow },
    setup: () => ({ args }),
    template: `
      <DangerRow v-bind="args">
        <template #escape><button class="link-btn">Choose a new owner</button></template>
      </DangerRow>
    `,
  }),
}

/** Mid-request. The label changes rather than the button vanishing, so the row
 *  keeps its height and nothing below it moves. */
export const Busy: Story = {
  args: { busy: true, busyLabel: 'Leaving…' },
}

/** No note. Some actions have no second sentence worth writing, and an empty
 *  line under the label is worse than none. */
export const NoNote: Story = {
  args: { label: 'Revoke every API key', note: undefined, actionLabel: 'Revoke' },
}

/** A long consequence beside a long label — the text column takes the room and
 *  the button keeps its place at the end of the row. */
export const LongText: Story = {
  args: {
    label: 'Удалить команду «Аналитика рынка и конкурентная разведка»',
    note: 'Команда удаляется для всех её участников, включая тех, кто сейчас читает исследования в ней.',
    actionLabel: 'Удалить',
  },
}
