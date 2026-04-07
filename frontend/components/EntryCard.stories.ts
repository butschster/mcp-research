import type { Meta, StoryObj } from '@storybook/vue3'
import EntryCard from './EntryCard.vue'
import { mockEntry, mockEntryDraft, mockEntryNoTags } from '../__mocks__/entry'

const meta: Meta<typeof EntryCard> = {
  title: 'Cards/EntryCard',
  component: EntryCard,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof EntryCard>

export const FullEntry: Story = {
  args: {
    entry: mockEntry,
    researchSlug: 'R1',
  },
}

export const DraftStatus: Story = {
  args: {
    entry: mockEntryDraft,
    researchSlug: 'R1',
  },
}

export const NoDescription: Story = {
  args: {
    entry: { ...mockEntry, description: undefined },
    researchSlug: 'R1',
  },
}

export const NoTags: Story = {
  args: {
    entry: mockEntryNoTags,
    researchSlug: 'R1',
  },
}

export const ActiveStatus: Story = {
  args: {
    entry: { ...mockEntry, id: 'ent_004', code: 'E4', title: 'Slots and Render Functions', status: 'active', tags: ['vue', 'slots'] },
    researchSlug: 'R1',
  },
}

export const PendingStatus: Story = {
  args: {
    entry: { ...mockEntry, id: 'ent_005', code: 'E5', title: 'Performance Optimization', status: 'pending', tags: ['vue', 'performance'] },
    researchSlug: 'R1',
  },
}

export const AllStatuses: Story = {
  render: () => ({
    components: { EntryCard },
    template: `
      <div style="display: flex; flex-direction: column; gap: 1rem; max-width: 600px;">
        <EntryCard
          v-for="entry in entries"
          :key="entry.id"
          :entry="entry"
          researchSlug="R1"
        />
      </div>
    `,
    setup() {
      const statuses = ['active', 'completed', 'draft', 'pending', 'in_progress']
      const entries = statuses.map((status, i) => ({
        id: `ent_${i}`,
        code: `E${i + 1}`,
        title: `Entry with ${status} status`,
        description: `This is an entry in the ${status} state.`,
        status,
        tags: ['vue', status],
      }))
      return { entries }
    },
  }),
}
