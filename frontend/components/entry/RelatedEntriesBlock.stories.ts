import type { Meta, StoryObj } from '@storybook/vue3'
import RelatedEntriesBlock from './RelatedEntriesBlock.vue'
import { withShare, withoutShare } from '../../__mocks__/share'

/**
 * Other entries that share tags with this one.
 *
 * The list crosses researches on the owner side, which is the point of it: a
 * tag is how two investigations turn out to be about the same thing. Under a
 * share it must not — the query is scoped server-side — and a row from outside
 * renders inert if one ever arrives anyway. An endpoint that starts returning
 * more than it should must not quietly turn into navigation here.
 */
const meta: Meta<typeof RelatedEntriesBlock> = {
  title: 'Entry/RelatedEntriesBlock',
  component: RelatedEntriesBlock,
  tags: ['autodocs'],
  // Share state is module state; this gives the ordinary stories a known
  // starting point rather than whatever the last story left behind. The
  // trade-offs are in __mocks__/share.ts.
  decorators: [withoutShare()],
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

/** Inside a share link. Same-research rows point at `/s/{token}/entry/…`, which
 *  is the only entry route a visitor has. */
export const InsideAShare: Story = {
  decorators: [withShare()],
  args: {
    entries: relatedEntries,
    currentTags: ['vue', 'composables'],
    researchSlug: 'R7',
    researchId: 'res_001',
  },
}

/**
 * The branch that should be unreachable: a related entry from another research,
 * served to a share. It renders inert — no href, no hover — rather than
 * offering a route out of the shared research.
 *
 * It is still a leak, and not one this component can close: the row shows the
 * foreign entry's title. The fix is server-side, and this story is here so the
 * consequence of skipping it is visible.
 */
export const SharedWithForeignRow: Story = {
  decorators: [withShare()],
  args: {
    entries: [
      ...relatedEntries,
      { id: 'ent_r2_001', code: 'E1', title: 'Pinia Store Patterns', tags: ['vue', 'pinia'], research_id: 'res_002' },
    ],
    currentTags: ['vue'],
    researchSlug: 'R7',
    researchId: 'res_001',
  },
}
