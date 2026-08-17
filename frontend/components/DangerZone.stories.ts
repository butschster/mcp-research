import type { Meta, StoryObj } from '@storybook/vue3'
import DangerZone from './DangerZone.vue'
import DangerRow from './DangerRow.vue'

/**
 * The box at the bottom of a page holding the things that cannot be undone.
 *
 * It was two classes in the global stylesheet — `.danger-zone` and
 * `.danger-row`, extracted from the team page, which then went on declaring its
 * own copies of both. Scoped selectors carry a `[data-v]` attribute and win, so
 * one class name meant a red-bordered box in the primitive and a borderless
 * column on the only page using it, and the CSS discipline check counted the
 * page overriding the rule as its consumer.
 *
 * A component cannot be shadowed that way, which is the whole reason it is one.
 */
const meta: Meta<typeof DangerZone> = {
  title: 'Layout/DangerZone',
  component: DangerZone,
  tags: ['autodocs'],
  argTypes: { title: { control: 'text' }, lead: { control: 'text' } },
}
export default meta
type Story = StoryObj<typeof DangerZone>

const withRows = (rows: string) => ({
  components: { DangerZone, DangerRow },
  template: `<DangerZone v-bind="args">${rows}</DangerZone>`,
})

/** The team page's pair: leaving, and deleting. */
export const TeamPage: Story = {
  render: (args) => ({
    setup: () => ({ args }),
    ...withRows(`
      <DangerRow label="Leave team" note="Removes your access to 18 researches." action-label="Leave" />
      <DangerRow label="Delete team" note="The team is empty and can be deleted." action-label="Delete" />
    `),
  }),
}

/** One row. The divider only appears between rows, so a single action does not
 *  sit above a rule leading nowhere. */
export const SingleRow: Story = {
  render: (args) => ({
    setup: () => ({ args }),
    ...withRows(`<DangerRow label="Revoke API key" note="Anything using it stops working immediately." action-label="Revoke" />`),
  }),
}

/** With a lead, for a section whose rows need framing before they are read. */
export const WithLead: Story = {
  args: { lead: 'These cannot be undone, and they affect everyone in the team.' },
  render: (args) => ({
    setup: () => ({ args }),
    ...withRows(`
      <DangerRow label="Transfer ownership" note="You become an editor." action-label="Transfer" />
      <DangerRow label="Delete team" note="Deleted for everyone in it." action-label="Delete" />
    `),
  }),
}

/** A custom title. The heading is the section's accessible name, so it is a
 *  real `<h2>` rather than styled text. */
export const CustomTitle: Story = {
  args: { title: 'Irreversible' },
  render: (args) => ({
    setup: () => ({ args }),
    ...withRows(`<DangerRow label="Delete account" note="Every research you own goes with it." action-label="Delete" />`),
  }),
}
