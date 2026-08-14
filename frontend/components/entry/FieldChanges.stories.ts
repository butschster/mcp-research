import type { Meta, StoryObj } from '@storybook/vue3'
import FieldChanges from './FieldChanges.vue'

/**
 * What a revision changed outside the body. This is the answer to the case that
 * used to render as "Nothing changed": a status flip is a real write with an
 * empty content diff.
 */
const meta: Meta<typeof FieldChanges> = {
  title: 'Entry/FieldChanges',
  component: FieldChanges,
  tags: ['autodocs'],
}
export default meta
type Story = StoryObj<typeof FieldChanges>

/** The most common one — triage, straight from the entry page's status dropdown. */
export const StatusOnly: Story = {
  args: { fields: [{ field: 'status', before: 'draft', after: 'active' }] },
}

/** A retitle plus tags, rendered with the same components the entry page uses. */
export const TitleAndTags: Story = {
  args: {
    fields: [
      { field: 'title', before: 'Pricing model', after: 'Pricing and packaging' },
      { field: 'tags', before: '', after: 'pricing, commercial' },
    ],
  },
}

/** Tags removed entirely — an empty side renders as an em dash. */
export const TagsCleared: Story = {
  args: { fields: [{ field: 'tags', before: 'pricing, commercial', after: '' }] },
}

/** Everything at once, the shape a big edit produces. */
export const AllFields: Story = {
  args: {
    fields: [
      { field: 'title', before: 'Pricing model', after: 'Pricing and packaging' },
      { field: 'description', before: 'Seat based above 50 users.', after: 'Seat based above 100 users, billed annually or monthly at a premium.' },
      { field: 'status', before: 'draft', after: 'completed' },
      { field: 'tags', before: 'pricing', after: 'pricing, commercial, enterprise' },
    ],
  },
}

/** Cyrillic, at the length these fields really reach. */
export const Cyrillic: Story = {
  args: {
    fields: [
      { field: 'title', before: 'Отчёт по задержкам', after: 'Отчёт по задержкам и потерям пакетов' },
      { field: 'status', before: 'draft', after: 'active' },
    ],
  },
}

/** Nothing changed outside the body — the component renders nothing at all. */
export const Empty: Story = {
  args: { fields: [] },
}
