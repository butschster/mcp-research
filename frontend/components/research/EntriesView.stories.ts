import type { Meta, StoryObj } from '@storybook/vue3'
import EntriesView from './EntriesView.vue'
import { mockEntry, mockEntryDraft } from '../../__mocks__/entry'
import { markupDescription } from '../../__mocks__/markup'
import { mockSections, mockSection, mockSectionCompleted } from '../../__mocks__/section'
import { withShare, withoutShare } from '../../__mocks__/share'

/**
 * The entry grid, either grouped by section or filtered to one.
 *
 * The cards are built inline here rather than reusing `EntryCard`, twice — so
 * the path each of them points at is written in three places in this component
 * alone. All three now ask `entryPath()`, which is the only reason the same grid
 * can be rendered under a share link without dropping an anonymous visitor onto
 * the login wall.
 */
const meta: Meta<typeof EntriesView> = {
  title: 'Research/EntriesView',
  component: EntriesView,
  tags: ['autodocs'],
  decorators: [
    // Share state is module state; this gives the ordinary stories a known
    // starting point rather than whatever the last story left behind. The
    // trade-offs are in __mocks__/share.ts.
    withoutShare(),
    () => ({
      template: '<div style="max-width: 800px"><story /></div>',
    }),
  ],
  argTypes: {
    mode: { control: 'select', options: ['all', 'section'] },
    loading: { control: 'boolean' },
  },
}
export default meta
type Story = StoryObj<typeof EntriesView>

const entriesWithSections = [
  { ...mockEntry, section_id: 'sec_001' },
  { ...mockEntryDraft, section_id: 'sec_001' },
  { ...mockEntry, id: 'ent_004', code: 'E4', title: 'Slots and Render Functions', tags: ['vue', 'slots'], status: 'active', section_id: 'sec_002' },
  { ...mockEntry, id: 'ent_005', code: 'E5', title: 'Performance Optimization', tags: ['vue', 'performance'], status: 'pending', section_id: 'sec_002' },
]

const tags = [
  { tag: 'vue', count: 4 },
  { tag: 'composables', count: 2 },
  { tag: 'slots', count: 1 },
  { tag: 'performance', count: 1 },
  { tag: 'reactivity', count: 1 },
]

export const AllEntriesGrouped: Story = {
  args: {
    entries: entriesWithSections,
    sections: [mockSection, mockSectionCompleted],
    researchSlug: 'R1',
    loading: false,
    mode: 'all',
    tags,
  },
}

export const Loading: Story = {
  args: {
    entries: [],
    sections: mockSections,
    researchSlug: 'R1',
    loading: true,
    mode: 'all',
    tags: [],
  },
}

export const SectionMode: Story = {
  args: {
    entries: [
      { ...mockEntry, section_id: 'sec_001' },
      { ...mockEntryDraft, section_id: 'sec_001' },
    ],
    sections: mockSections,
    researchSlug: 'R1',
    loading: false,
    mode: 'section',
    sectionInfo: mockSection,
    tags: [],
  },
}

export const TagFilterActive: Story = {
  args: {
    entries: entriesWithSections,
    sections: [mockSection, mockSectionCompleted],
    researchSlug: 'R1',
    loading: false,
    mode: 'all',
    tags,
  },
}

export const Empty: Story = {
  args: {
    entries: [],
    sections: mockSections,
    researchSlug: 'R1',
    loading: false,
    mode: 'all',
    tags: [],
  },
}

const markupEntry = {
  ...mockEntry,
  id: 'ent_markup',
  code: 'E9',
  title: 'Author-supplied HTML in a description',
  description: markupDescription,
  section_id: 'sec_001',
}

/**
 * A description with markup in it, in the **grouped** branch.
 *
 * This component writes the entry card three times — twice in its own template
 * and once again in `EntryCard` — and each copy calls `renderRefs` into
 * `v-html`. Duplicated markup is duplicated exposure: a fix applied to one copy
 * leaves the other two, and nothing in the catalogue would have shown which was
 * which. Hence a story per branch rather than one for the component.
 *
 * The markup must read as text and `[[E3]]` must still be a link. An executed
 * payload prints `XSS EXECUTED` in place of the image tag.
 */
export const MarkupInDescription: Story = {
  args: {
    entries: [markupEntry, ...entriesWithSections],
    sections: [mockSection, mockSectionCompleted],
    researchSlug: 'R1',
    loading: false,
    mode: 'all',
    tags,
  },
}

/** The same description in the **section** branch — the second copy of the card,
 *  reached by a different `v-if` and fixed separately from the one above. */
export const MarkupInDescriptionSectionMode: Story = {
  args: {
    entries: [markupEntry, { ...mockEntryDraft, section_id: 'sec_001' }],
    sections: mockSections,
    researchSlug: 'R1',
    loading: false,
    mode: 'section',
    sectionInfo: mockSection,
    tags: [],
  },
}

/** The grouped grid inside a share link: every card points at
 *  `/s/{token}/entry/…`, in both the grouped branch and the flat one. */
export const InsideAShare: Story = {
  decorators: [withShare()],
  args: {
    entries: entriesWithSections,
    sections: [mockSection, mockSectionCompleted],
    researchSlug: 'R7',
    loading: false,
    mode: 'all',
    tags,
  },
}

/**
 * An empty section inside a share link.
 *
 * The copy changes: the owner is told "Claude will populate this section with
 * research entries", which is true and useful to them and meaningless to a
 * client — it describes a tool they have never heard of doing work they have no
 * part in. A visitor gets "No entries in this section / There's nothing here
 * yet". Compare with `Empty` above.
 */
export const InsideAShareEmptySection: Story = {
  decorators: [withShare()],
  args: {
    entries: [],
    sections: mockSections,
    researchSlug: 'R7',
    loading: false,
    mode: 'section',
    sectionInfo: mockSection,
    tags: [],
  },
}

/** One section, inside a share link — the visitor's default view, since the
 *  shared overview opens on the first section rather than on "all entries". */
export const InsideAShareSectionMode: Story = {
  decorators: [withShare()],
  args: {
    entries: [
      { ...mockEntry, section_id: 'sec_001' },
      { ...mockEntryDraft, section_id: 'sec_001' },
    ],
    sections: mockSections,
    researchSlug: 'R7',
    loading: false,
    mode: 'section',
    sectionInfo: mockSection,
    tags: [],
  },
}
