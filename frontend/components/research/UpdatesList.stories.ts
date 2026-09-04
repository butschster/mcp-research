import type { Meta, StoryObj } from '@storybook/vue3'
import UpdatesList from './UpdatesList.vue'
import type { EntryUpdate } from '../../composables/useEntryUpdates'

const sections = [
  { id: 's-findings', name: 'findings', display_name: 'Findings' },
  { id: 's-interviews', name: 'interviews', display_name: 'Customer Interviews' },
]

const mixed: EntryUpdate[] = [
  {
    entry_id: 'e-new', entry_code: 'E12', research_id: 'r1', section_id: 's-findings',
    title: 'Pricing evidence', description: 'New evidence', entry_type: 'markdown', status: 'active',
    current_revision: 1, seen_revision: 0, unseen_revisions: 1, kind: 'new', updated_at: new Date(Date.now() - 7_200_000).toISOString(),
  },
  {
    entry_id: 'e-changed', entry_code: 'E5', research_id: 'r1', section_id: 's-interviews',
    title: 'Ограничения корпоративных клиентов и очень длинный заголовок документа', entry_type: 'markdown', status: 'draft',
    current_revision: 7, seen_revision: 4, unseen_revisions: 3, kind: 'changed', updated_at: new Date(Date.now() - 3_600_000).toISOString(),
  },
]

const meta: Meta<typeof UpdatesList> = {
  title: 'Research/UpdatesList',
  component: UpdatesList,
  tags: ['autodocs'],
  args: { sections, researchSlug: 'R7', updates: mixed },
  argTypes: { onRetry: { action: 'retry' } },
}
export default meta
type Story = StoryObj<typeof UpdatesList>

export const Mixed: Story = {}
export const OnlyNew: Story = { args: { updates: mixed.filter((entry) => entry.kind === 'new') } }
export const OnlyChanged: Story = { args: { updates: mixed.filter((entry) => entry.kind === 'changed') } }
export const Empty: Story = { args: { updates: [] } }
export const Loading: Story = { args: { updates: [], loading: true } }
export const Refreshing: Story = { args: { refreshing: true } }
export const Failed: Story = { args: { updates: [], error: 'Could not load updates' } }
export const UnknownSection: Story = {
  args: { updates: [{ ...mixed[0]!, section_id: 'section-removed' }] },
}
export const Overloaded: Story = {
  args: {
    updates: Array.from({ length: 105 }, (_, index): EntryUpdate => ({
      ...mixed[index % mixed.length]!,
      entry_id: `entry-${index}`,
      entry_code: `E${index + 1}`,
      title: `Document ${index + 1}`,
    })),
  },
}
