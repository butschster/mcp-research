import type { Meta, StoryObj } from '@storybook/vue3'
import EntriesView from './EntriesView.vue'
import { mockEntry, mockEntryDraft } from '../../__mocks__/entry'
import { markupDescription, markupImg } from '../../__mocks__/markup'
import { mockSections, mockSection, mockSectionCompleted } from '../../__mocks__/section'
import { withShare, withoutShare } from '../../__mocks__/share'
import { mockSpecEntries, mockSpecSection, specAtCap, specSingleField } from '../../__mocks__/metadata'

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

/* --- Sections that declare fields -------------------------------------------
 *
 * Everything below needs `sectionInfo.field_spec`. Without it the component is
 * exactly what it was before this feature: no toggle, no chips, no table — which
 * is what `SectionMode` above still shows, and is the right behaviour for a
 * section that is a topic rather than a class of document.
 *
 * The table only exists in section mode — one declaration measures one set of
 * columns. The "N missing" chip is on both grids, because the grouped one holds
 * a section per group and can measure each against its own declaration; that is
 * the surface people actually browse a research from, and the chip is worth
 * least where it cannot be seen.
 */

/**
 * Six sibling specifications in a section that declares six fields, as cards.
 *
 * The "N missing" chips are the point of this view: E52 answers no owner and
 * E54 answers nothing at all, and both say so before anyone opens them. E53
 * carries an explicit unknown, which is *not* a gap and correctly shows no chip.
 */
export const SectionWithFields: Story = {
  args: {
    entries: mockSpecEntries,
    sections: [mockSpecSection],
    researchSlug: 'R21',
    loading: false,
    mode: 'section',
    sectionInfo: mockSpecSection,
    tags: [],
  },
}

/**
 * All entries, grouped, where one of the groups declares fields.
 *
 * The chips appear under "Спецификации" and nowhere else on the page: the
 * research's other sections declare nothing, so their documents cannot be
 * incomplete and are not decorated as if they were. Two sections behaving
 * differently in one list is the intended reading — most sections are topics.
 */
export const AllEntriesWithFields: Story = {
  args: {
    entries: [...mockSpecEntries.slice(0, 3), { ...mockEntry, section_id: 'sec_001' }],
    sections: [mockSpecSection, mockSection],
    researchSlug: 'R21',
    loading: false,
    mode: 'all',
    tags: [{ tag: 'spec', count: 3 }, { tag: 'vue', count: 1 }],
  },
}

/**
 * The same section, tabulated. **This is the story to read.**
 *
 * The table is not the payoff of declaring fields, it is the enforcement
 * mechanism — a blank cell beside filled ones is the strongest force there is on
 * whether an optional field ever gets answered, and nothing else in the product
 * puts eighteen documents' worth of blanks in one place.
 *
 * `Registry` holds a bare `E47` and comes out a link, because the field is
 * declared `ref` and the table runs values through the same renderer the
 * document's own block does. A table that printed `[[E47]]` while the page
 * linked it would be one string rendered two ways.
 *
 * Three things it says at a glance. `Reviewed` is empty for every row: a column
 * of dashes is the feature's own output, and the answer to whether that field
 * should have been declared at all. `Owner` on E53 reads `unknown`, not a dash —
 * somebody looked, which is a different fact from silence. E54's row is empty
 * end to end.
 */
export const SectionTable: Story = {
  args: { ...SectionWithFields.args },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Table')
  },
}

/**
 * A section where every document leaves every declared field blank.
 *
 * A table of nothing but dashes, which is a legitimate production state — the
 * fields were declared and no agent has written one since. It reads as an
 * indictment of the declaration rather than of the documents, which is the
 * correct reading.
 */
export const SectionTableAllEmpty: Story = {
  args: {
    ...SectionWithFields.args,
    entries: mockSpecEntries.map(e => ({ ...e, metadata: {} })),
  },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Table')
  },
}

/**
 * One declared field: two fixed columns and one of its own.
 *
 * Worth keeping in the catalogue because the toggle is offered here — the
 * component asks only whether the section declares anything, not whether it
 * declares enough to be worth tabulating.
 */
export const SectionTableSingleField: Story = {
  args: {
    ...SectionWithFields.args,
    sectionInfo: { ...mockSpecSection, field_spec: specSingleField },
  },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Table')
  },
}

/**
 * Twelve columns — the cap — with Cyrillic values in them.
 *
 * The frame scrolls sideways, the page body must not. That is the only reason
 * `.meta-table-wrap` exists, and a wide table with long values is the only way
 * to find out whether it still holds. Drag the table horizontally; the story
 * container around it should not move.
 */
export const SectionTableWide: Story = {
  args: {
    ...SectionWithFields.args,
    sectionInfo: { ...mockSpecSection, field_spec: specAtCap },
    entries: mockSpecEntries.map(e => ({
      ...e,
      metadata: {
        ...(e.metadata as Record<string, unknown>),
        component: 'сканер площадок',
        transport: 'temporal',
        reviewer: 'команда платформы наблюдаемости площадок',
        schema_url: 'https://example.invalid/schemas/site-payload.json',
        retries: 3,
        supersedes: ['E48', 'E49'],
      },
    })),
  },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Table')
  },
}

/**
 * The declaring section inside a share link.
 *
 * A visitor gets the chips and the table like anyone else — the values are part
 * of the document, not working process. Every row still links through
 * `entryPath()` to `/s/{token}/entry/…`.
 */
export const SectionTableInsideAShare: Story = {
  decorators: [withShare()],
  args: { ...SectionWithFields.args, researchSlug: 'R7' },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Table')
  },
}

/**
 * Markup inside a table cell.
 *
 * The cell is a fourth `v-html` in this one component, written separately from
 * the three card copies above, and field values are agent-authored text like
 * every other string here. `<b>bold</b>` must read as text and `[[E47]]` must
 * still be a link; an executed payload prints `XSS EXECUTED` where the image
 * tag was.
 *
 * Cells are `white-space: nowrap`, so a value long enough to matter widens the
 * column and the frame scrolls. That is the intended behaviour and this is
 * where to check it has not become the page scrolling.
 */
export const SectionTableMarkup: Story = {
  args: {
    ...SectionWithFields.args,
    entries: [
      {
        ...mockSpecEntries[0],
        metadata: {
          stage: `<b>draft</b> ${markupImg}`,
          produces: ['<script>alert(1)</script>', 'scanner-watchdog'],
          owner: 'platform — см. [[E47]]',
        },
      },
      ...mockSpecEntries.slice(1, 3),
    ],
  },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Table')
  },
}

/**
 * Clicks the first button whose label matches, once it exists. No
 * `@storybook/test` in this project, so the catalogue polls — same helper shape
 * as `HistoryPanel.stories.ts`.
 */
async function clickButton(root: HTMLElement, label: string): Promise<void> {
  for (let i = 0; i < 50; i++) {
    const button = Array.from(root.querySelectorAll('button'))
      .find(b => b.textContent?.trim() === label) as HTMLElement | undefined
    if (button) {
      button.click()
      return
    }
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
}
