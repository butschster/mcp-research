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

export const NewDocument: Story = {
  args: {
    entry: mockEntry,
    researchSlug: 'R1',
    update: {
      entry_id: mockEntry.id,
      entry_code: mockEntry.code,
      research_id: 'research-1',
      section_id: 'section-1',
      title: mockEntry.title,
      entry_type: 'markdown',
      status: mockEntry.status,
      current_revision: 1,
      seen_revision: 0,
      unseen_revisions: 1,
      kind: 'new',
      updated_at: new Date().toISOString(),
    },
  },
}

export const ChangedDocument: Story = {
  args: {
    entry: mockEntry,
    researchSlug: 'R1',
    update: {
      entry_id: mockEntry.id,
      entry_code: mockEntry.code,
      research_id: 'research-1',
      section_id: 'section-1',
      title: mockEntry.title,
      entry_type: 'markdown',
      status: mockEntry.status,
      current_revision: 8,
      seen_revision: 4,
      unseen_revisions: 4,
      kind: 'changed',
      updated_at: new Date().toISOString(),
    },
  },
}

/** Every header signal at the narrow width used by a small phone. */
export const NarrowChangedWithFlags: Story = {
  decorators: [() => ({ template: '<div style="width: 320px"><story /></div>' })],
  args: {
    entry: {
      ...mockEntry,
      title: 'Очень длинное название документа без короткого варианта для карточки',
      status: 'draft',
    },
    researchSlug: 'R1',
    missingRequired: 3,
    update: {
      entry_id: mockEntry.id,
      entry_code: mockEntry.code,
      research_id: 'research-1',
      section_id: 'section-1',
      title: mockEntry.title,
      entry_type: 'markdown',
      status: 'draft',
      current_revision: 14,
      seen_revision: 8,
      unseen_revisions: 6,
      kind: 'changed',
      updated_at: new Date().toISOString(),
    },
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

/**
 * A document that leaves a required field unanswered.
 *
 * The chip is on the card and not only on the document, because that is the
 * whole mechanism: a blank next to filled ones is the strongest force there is
 * on whether an optional field ever gets answered, and a gap visible only after
 * you open the document is a gap nobody meets. The count comes from the
 * section's declaration, so a card in a section that declares nothing can never
 * show it.
 */
export const MissingRequiredField: Story = {
  args: {
    entry: { ...mockEntry, title: 'SPEC-03 · Сервис watchdog', status: 'active' },
    researchSlug: 'R21',
    missingRequired: 1,
  },
}

/** Three of them. The title text is pluralised — hover the chip. */
export const SeveralMissing: Story = {
  args: {
    entry: { ...mockEntry, title: 'SPEC-05 · Инциденты без права решения', status: 'draft' },
    researchSlug: 'R21',
    missingRequired: 3,
  },
}

/**
 * Zero is not "0 missing", it is nothing at all — a complete document says so by
 * looking exactly as it did before the feature existed. Compare with
 * `MissingRequiredField`: the two differ by one prop.
 */
export const NothingMissing: Story = {
  args: {
    entry: { ...mockEntry, title: 'SPEC-01 · Payload состояния площадки' },
    researchSlug: 'R21',
    missingRequired: 0,
  },
}

/**
 * The chip against a title long enough to wrap, in Cyrillic.
 *
 * The header is a flex row and the chip does not shrink, so this is where the
 * title has to give way rather than the count. Russian titles are what this
 * product actually holds, and they are the ones long enough to find out.
 */
export const MissingWithLongTitle: Story = {
  args: {
    entry: {
      ...mockEntry,
      title: 'SPEC-18 · temporal-watchdog: эмиттер инцидентов без права принимать решения о площадке',
      description: 'Границы ответственности сторожа и то, чего он не решает. См. [[E47]].',
      tags: ['spec', 'watchdog', 'temporal'],
      status: 'active',
    },
    researchSlug: 'R21',
    missingRequired: 2,
  },
}
