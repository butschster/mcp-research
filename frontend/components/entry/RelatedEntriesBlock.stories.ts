import type { Meta, StoryObj } from '@storybook/vue3'
import RelatedEntriesBlock from './RelatedEntriesBlock.vue'

const meta: Meta<typeof RelatedEntriesBlock> = {
  title: 'Entry/RelatedEntriesBlock',
  component: RelatedEntriesBlock,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
    researchId: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof RelatedEntriesBlock>

const relatedEntries = [
  { id: 'ent_002', code: 'E2', title: 'Reactive State Management', tags: ['vue', 'reactivity'], research_id: 'res_001' },
  { id: 'ent_004', code: 'E4', title: 'Slots and Render Functions', tags: ['vue', 'slots'], research_id: 'res_001' },
  { id: 'ent_005', code: 'E5', title: 'Performance Optimization', tags: ['vue', 'performance'], research_id: 'res_001' },
]

export const WithSharedTags: Story = {
  args: {
    entries: relatedEntries,
    currentTags: ['vue', 'composables'],
    researchSlug: 'R1',
    researchId: 'res_001',
  },
}

export const CrossResearch: Story = {
  args: {
    entries: [
      ...relatedEntries,
      { id: 'ent_r2_001', code: 'E1', title: 'Pinia Store Patterns', tags: ['vue', 'pinia'], research_id: 'res_002' },
    ],
    currentTags: ['vue'],
    researchSlug: 'R1',
    researchId: 'res_001',
  },
}

export const Empty: Story = {
  args: {
    entries: [],
    currentTags: ['vue'],
    researchSlug: 'R1',
    researchId: 'res_001',
  },
}
