import type { Meta, StoryObj } from '@storybook/vue3'
import ExportDocument from './ExportDocument.vue'
import {
  mockExportData,
  mockExportDataBlocks,
  mockExportDataEmpty,
  mockExportDataShared,
} from '../../__mocks__/export'
import { withShare, withoutShare } from '../../__mocks__/share'

/**
 * The printable research document.
 *
 * Extracted from the export page so a shared view can render the same thing.
 * Two copies would have drifted the first time an entry type was added, and the
 * one that drifted would have been the copy nobody was looking at — the one a
 * client sees.
 *
 * It renders whatever it is handed and asks no questions about permission. The
 * server decides what is in the payload: under a share it has already dropped
 * `instruction`, `memory`, and any session or task the link does not include.
 * There is no `include` prop here on purpose — a client-side filter over a
 * payload that should never have contained the data is not a defence.
 *
 * Every story below is outside a share by default, so `[[E3]]` renders as a
 * link into `/research/…`; `SharedExport` shows the same document under a link,
 * where the paths change and a reference out of the research goes plain.
 */
const meta: Meta<typeof ExportDocument> = {
  title: 'Research/ExportDocument',
  component: ExportDocument,
  tags: ['autodocs'],
  decorators: [withoutShare(), () => ({ template: '<div style="max-width: 820px;"><story /></div>' })],
  argTypes: {
    data: { control: 'object' },
    researchSlug: { control: 'text' },
  },
  args: { researchSlug: 'R7' },
}
export default meta
type Story = StoryObj<typeof ExportDocument>

/** The whole document: cover, contents, two sections of entries, a session with
 *  its questions, and the task list. This is what goes to the printer. */
export const FullDocument: Story = {
  args: { data: mockExportData },
}

/** A research with sections and nothing in them yet. Each section says "No
 *  entries" rather than vanishing — a contents list that skips the empty ones
 *  makes the document look complete when it is not. */
export const EmptyResearch: Story = {
  args: { data: mockExportDataEmpty },
}

/** An entry with a mermaid diagram. It is drawn, not printed as source: this
 *  page is what becomes the PDF, and a flattened diagram is the one thing a
 *  reader cannot reconstruct. Markdown entries carry theirs as fences; this one
 *  is a blocks entry, which renders through the block renderer. */
export const WithDiagram: Story = {
  args: { data: mockExportDataBlocks },
}

/**
 * What a share export looks like: sessions and tasks are absent from the
 * payload, not filtered out of it, so the contents list has no entries for them
 * and the two sections at the end simply do not exist.
 *
 * The banner is not part of this component — the print stylesheet hides it — but
 * the share context is set, so the cross-references in the entries resolve to
 * `/s/{token}/entry/…`.
 */
export const SharedExport: Story = {
  decorators: [withShare()],
  args: { data: mockExportDataShared },
}

/** A payload with no sections at all — a research created a minute ago. The
 *  cover and an empty contents list, which is the honest rendering of nothing. */
export const NoSections: Story = {
  args: {
    data: { research: mockExportDataEmpty.research, sections: [], sessions: [], tasks: [] },
  },
}

/** A research name and goal long enough to run over several lines, with tags
 *  that wrap. The cover grows; nothing truncates, because this is a document. */
export const LongCover: Story = {
  args: {
    data: {
      ...mockExportData,
      research: {
        ...mockExportData.research,
        name: 'Ценообразование конкурентов в сегменте корпоративных подписок с посадочными местами, третий квартал',
        goal: 'Understand how three competitors price seat-based tiers, where our own tiering sits against them, which of the differences customers actually mention in conversation, and whether the bands we publish survive the September release Kestrel has pre-announced.',
        tags: ['pricing', 'competitive', 'q3', 'kestrel', 'northlight', 'verge', 'seats', 'procurement'],
      },
    },
  },
}
