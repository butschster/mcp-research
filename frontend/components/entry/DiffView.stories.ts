import type { Meta, StoryObj } from '@storybook/vue3'
import DiffView from './DiffView.vue'
import {
  mockDiffWordLevel,
  mockDiffWithGaps,
  mockDiffFirstVersion,
  mockDiffNoChanges,
  mockDiffEmpty,
  mockDiffTruncated,
  mockDiffLongLines,
} from '../../__mocks__/revision'

/**
 * Renders a `DiffResult` from `GET /api/entries/{id}/diff`. Every state below
 * is one the API actually produces:
 *
 * - a replaced line carries a word-level breakdown on both sides
 * - unchanged runs outside the context window collapse behind a click
 * - `truncated` means the documents were too large to align
 * - an empty `lines` array means there was nothing to compare
 */
const meta: Meta<typeof DiffView> = {
  title: 'Entry/DiffView',
  component: DiffView,
  tags: ['autodocs'],
  argTypes: {
    diff: { control: 'object' },
    context: {
      control: { type: 'number', min: 0, max: 20 },
      description: 'Unchanged lines kept around each change before the rest collapses.',
    },
  },
}
export default meta
type Story = StoryObj<typeof DiffView>

/** A sentence rewritten: only the two changed tokens are highlighted. */
export const WordLevelReplacement: Story = {
  args: { diff: mockDiffWordLevel },
}

/** Revision 1 of an entry — every line is an addition, nothing collapses. */
export const AdditionsOnly: Story = {
  args: { diff: mockDiffFirstVersion },
}

/** A long document with two edits far apart: two gaps, each expandable. */
export const CollapsedContext: Story = {
  args: { diff: mockDiffWithGaps },
}

/** The same diff with the gap expanded, as a reader sees it after one click. */
export const ExpandedContext: Story = {
  args: { diff: mockDiffWithGaps },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    const gap = canvasElement.querySelector('.diff-gap-btn') as HTMLElement | null
    gap?.click()
  },
}

/** `context: 0` — changes only, the tightest reading of a long document. */
export const NoContext: Story = {
  args: { diff: mockDiffWithGaps, context: 0 },
}

/** A wide context window, for readers who want the surrounding paragraphs. */
export const WideContext: Story = {
  args: { diff: mockDiffWithGaps, context: 10 },
}

/**
 * Two revisions compared that turn out to be identical (`?from=3&to=3`).
 * Nothing collapses: a document made entirely of unchanged lines is shown as
 * it is rather than as one giant gap.
 */
export const IdenticalRevisions: Story = {
  args: { diff: mockDiffNoChanges },
}

/** Nothing to compare at all — both revisions had an empty body. */
export const NothingChanged: Story = {
  args: { diff: mockDiffEmpty },
}

/**
 * Past `maxDiffLines` the backend stops aligning and reports the whole
 * document replaced. The note explains that; the lines below it are the raw
 * remove/add pairs, without word-level detail.
 */
export const Truncated: Story = {
  args: { diff: mockDiffTruncated },
}

/** Long URLs and wide table rows wrap instead of clipping. */
export const LongLines: Story = {
  args: { diff: mockDiffLongLines },
}

/** Narrow container, the width the panel gives it on a phone. */
export const InNarrowColumn: Story = {
  args: { diff: mockDiffWordLevel },
  render: (args: any) => ({
    components: { DiffView },
    setup: () => ({ args }),
    template: `
      <div style="max-width: 340px;">
        <DiffView v-bind="args" />
      </div>
    `,
  }),
}
