import type { Meta, StoryObj } from '@storybook/vue3'
import EntryCard from './EntryCard.vue'
import { mockEntry, mockEntryDraft, mockEntryNoTags } from '../__mocks__/entry'
import { markupDescription } from '../__mocks__/markup'
import { withShare, withoutShare } from '../__mocks__/share'

/**
 * One entry, as a link.
 *
 * Where it points is no longer written into the template: it asks `entryPath()`,
 * which answers `/research/{slug}/entry/{code}` normally and
 * `/s/{token}/entry/{code}` under a share link. The card itself does not know
 * which it is in, and should not — a component that has to be told where it is
 * gets it wrong in the one place nobody checks.
 */
const meta: Meta<typeof EntryCard> = {
  title: 'Cards/EntryCard',
  component: EntryCard,
  tags: ['autodocs'],
  // Share state is module state; this gives the ordinary stories a known
  // starting point rather than whatever the last story left behind. The
  // trade-offs are in __mocks__/share.ts.
  decorators: [withoutShare()],
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

/**
 * The same card inside a share link. Nothing about it looks different — that is
 * the point — but the href is now `/s/{token}/entry/E1`, which is the only route
 * an anonymous visitor can follow. Hover the title to see it.
 */
export const InsideAShare: Story = {
  decorators: [withShare()],
  args: {
    entry: mockEntry,
    researchSlug: 'R7',
  },
}

/**
 * A description with markup in it.
 *
 * The description is the one field on this card that goes to `v-html`, through
 * `renderRefs`. Everything here must read as text — `<b>bold</b>` in angle
 * brackets, the script tag spelled out — while `[[E3]]` beside it is still a
 * link. If either half stops being true the card says so on sight: an executed
 * payload prints `XSS EXECUTED` where the image tag was.
 *
 * The title above it is `{{ }}` interpolation and was never at risk. It is
 * worth knowing which of the two fields is which, because they look identical
 * in the template.
 */
export const MarkupInDescription: Story = {
  args: {
    entry: { ...mockEntry, description: markupDescription },
    researchSlug: 'R1',
  },
}

export const ArtifactEntry: Story = {
  args: {
    entry: {
      ...mockEntry,
      entry_type: 'artifact',
      title: 'Local LLM benchmark: speed and memory',
      description: 'Throughput and VRAM for four local models on one card.',
    },
    researchSlug: 'R1',
  },
}
