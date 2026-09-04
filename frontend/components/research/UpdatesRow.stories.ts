import type { Meta, StoryObj } from '@storybook/vue3'
import UpdatesRow from './UpdatesRow.vue'
import type { EntryUpdate } from '../../composables/useEntryUpdates'

const changed: EntryUpdate = {
  entry_id: 'entry-5',
  entry_code: 'E5',
  research_id: 'research-7',
  section_id: 'section-interviews',
  title: 'Enterprise interview constraints',
  entry_type: 'markdown',
  status: 'active',
  current_revision: 7,
  seen_revision: 4,
  unseen_revisions: 3,
  kind: 'changed',
  updated_at: '2026-09-04T12:00:00Z',
}

const meta: Meta<typeof UpdatesRow> = {
  title: 'Research/UpdatesRow',
  component: UpdatesRow,
  tags: ['autodocs'],
  args: {
    update: changed,
    researchSlug: 'R7',
    sectionName: 'Customer Interviews',
  },
}
export default meta
type Story = StoryObj<typeof UpdatesRow>

export const Changed: Story = {}

export const NewDocument: Story = {
  args: {
    update: {
      ...changed,
      entry_id: 'entry-12',
      entry_code: 'E12',
      title: 'Pricing evidence',
      current_revision: 1,
      seen_revision: 0,
      unseen_revisions: 1,
      kind: 'new',
    },
  },
}

export const WithoutShortCode: Story = {
  args: {
    update: { ...changed, entry_code: undefined },
  },
}

export const LongTitleOnMobile: Story = {
  parameters: { viewport: { defaultViewport: 'mobile' } },
  args: {
    update: {
      ...changed,
      title: 'Ограничения корпоративных клиентов и согласование изменений в нескольких подразделениях',
    },
    sectionName: 'Интервью с корпоративными заказчиками',
  },
}
