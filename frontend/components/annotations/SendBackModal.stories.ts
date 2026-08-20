import type { Meta, StoryObj } from '@storybook/vue3'
import SendBackModal from './SendBackModal.vue'

/**
 * Collecting the reason an answer was refused.
 *
 * This exists because the same job was being done by `window.prompt`, and the
 * text it collects is the most load-bearing free text in the feature: the agent
 * is required to read it before its next attempt. A native prompt blocks the
 * thread, cannot be styled, is suppressible by the browser, and — called from
 * inside the pass review, which is itself a modal with a focus trap — fought
 * the dialog it was opened from. None of which a catalogue page can show, which
 * was the other half of the problem: the two send-back paths were the only
 * behaviour in this feature Storybook could not demonstrate at all.
 *
 * The copy changes with the count, because "these marks go back" reads wrong
 * for one.
 */
const meta: Meta<typeof SendBackModal> = {
  title: 'Annotations/SendBackModal',
  component: SendBackModal,
  tags: ['autodocs'],
  argTypes: {
    visible: { control: 'boolean' },
    count: { control: { type: 'number', min: 1 } },
    busy: { control: 'boolean' },
  },
  args: { visible: true, count: 1, busy: false },
}
export default meta
type Story = StoryObj<typeof SendBackModal>

/** One mark, sent back from its thread on the document page. */
export const SingleMark: Story = {}

/** A batch, sent back from the pass review. */
export const Batch: Story = {
  args: { count: 6 },
}

/** In flight. The button says so and refuses a second press. */
export const Sending: Story = {
  args: { count: 3, busy: true },
}

/**
 * Closed. Kept so the catalogue shows that nothing renders — the reason field
 * is cleared on every open, so a reason typed for one batch can never travel to
 * the next.
 */
export const Hidden: Story = {
  args: { visible: false },
}
