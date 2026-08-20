import type { Meta, StoryObj } from '@storybook/vue3'
import ThreadCard from './ThreadCard.vue'
import {
  makeAnnotation,
  mockAnnotation,
  mockAnnotationAnswered,
  mockAnnotationRejected,
  mockAnnotationDrifted,
  mockAnnotationOrphaned,
  mockAnnotationClosed,
  mockAnnotationMarkdown,
} from '../../__mocks__/annotation'

/**
 * One mark, opened — rendered inline under the block it belongs to rather than
 * in a modal, because a modal would cover the sentence the whole conversation is
 * about.
 *
 * The card is assembled from four other components (ShortCode, KindChip,
 * StatusBadge, AnchorBadge) plus EditableField for the note, all of which Nuxt
 * auto-registers and Storybook does not; they are registered by their
 * Nuxt-derived names in `.storybook/preview.ts`. The resolution goes through
 * `renderRefs`, so `[[E3]]` in an answer is a link here, escaped first.
 *
 * The footer changes with status and that is the point of most of these stories:
 * an answered mark offers Accept and Send back, an open one offers Dismiss, a
 * settled one offers Reopen, and a viewer gets none of them.
 */
const meta: Meta<typeof ThreadCard> = {
  title: 'Annotations/ThreadCard',
  component: ThreadCard,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
    canWrite: { control: 'boolean' },
    busy: { control: 'boolean' },
    annotation: { control: false },
    onAccept: { action: 'accept' },
    onDismiss: { action: 'dismiss' },
    onReopen: { action: 'reopen' },
    onClose: { action: 'close' },
    'onUpdate-body': { action: 'update-body' },
  },
  args: { researchSlug: 'R1', canWrite: true, busy: false },
}
export default meta
type Story = StoryObj<typeof ThreadCard>

/** Open, anchored, waiting for the agent. The only action a person has is Dismiss. */
export const Open: Story = { args: { annotation: mockAnnotation } }

/**
 * The agent answered and the mark is waiting on a person: Accept, or Send back
 * with a reason. "Send back" opens `SendBackModal` — cancelling it emits
 * nothing, which is the behaviour to check here.
 */
export const Answered: Story = { args: { annotation: mockAnnotationAnswered } }

/**
 * Two previous answers were sent back, and both reasons are shown in full.
 *
 * Not collapsed and not summarised: an agent repeating a rejected answer is the
 * commonest way a pass wastes itself, and these are the reasons it is required
 * to read before trying a third time.
 */
export const WithRejections: Story = { args: { annotation: mockAnnotationRejected } }

/**
 * The sentence under the mark changed after the mark was made.
 *
 * The warning carries a link into the diff from the anchored revision to now,
 * because "was my doubt addressed or quietly buried" is the only question a
 * drifted mark raises, and nothing narrower answers it.
 */
export const Drifted: Story = { args: { annotation: mockAnnotationDrifted } }

/** The marked text is gone from the document entirely — a different sentence, same link. */
export const Orphaned: Story = { args: { annotation: mockAnnotationOrphaned } }

/**
 * Orphaned on a mark that was never anchored to a revision, so there is nothing
 * to diff against and the link is omitted rather than pointing at nothing.
 */
export const OrphanedWithoutDiff: Story = {
  args: {
    annotation: makeAnnotation({
      code: 'A15',
      anchored_revision: 0,
      anchor: { state: 'orphaned', strategy: 'none', confidence: 0, block_index: -1 },
    }),
  },
}

/**
 * A viewer's card: no pencil on the note, no footer at all.
 *
 * `canWrite` is the whole difference — a viewer can read a mark and its answer,
 * and can do nothing to either.
 */
export const ReadOnly: Story = {
  args: { annotation: mockAnnotationAnswered, canWrite: false },
}

/**
 * A decision is in flight. The card dims and stops taking pointer events, so the
 * same Accept cannot be pressed twice while the request is out.
 */
export const Busy: Story = {
  args: { annotation: mockAnnotationAnswered, busy: true },
}

/** Settled and closed. The only thing left to do is reopen it. */
export const Closed: Story = { args: { annotation: mockAnnotationClosed } }

/** Dismissed — the mark was withdrawn rather than answered, and can come back. */
export const Dismissed: Story = {
  args: {
    annotation: makeAnnotation({
      code: 'A16',
      status: 'dismissed',
      body: 'Withdrawn — I had misread the paragraph above it.',
      closed_at: '2025-03-19T09:00:00Z',
    }),
  },
}

/** A mark with no note: the body field shows its placeholder, not an empty box. */
export const NoNote: Story = {
  args: { annotation: makeAnnotation({ code: 'A17', body: undefined }) },
}

/**
 * A markdown document — no blocks, so the anchor badge's hint says the mark is
 * found by its text alone. Everything else about the card is unchanged.
 */
export const OnMarkdownDocument: Story = { args: { annotation: mockAnnotationMarkdown } }

/**
 * A paragraph-length quote and a long answer with a cross-reference in it.
 *
 * The quote block wraps rather than truncating — in a thread the exact wording is
 * the evidence, and a card that hides half of it makes the reader open the
 * document to check what was actually marked.
 */
export const LongContent: Story = {
  args: {
    annotation: makeAnnotation({
      code: 'A18',
      status: 'answered',
      quote: {
        exact:
          'Across the ecosystem, composables replaced mixins outright by the end of 2023, and every team ' +
          'that stayed on the Options API reports that the migration cost was carried almost entirely by ' +
          'their largest components rather than being spread evenly across the codebase.',
      },
      body:
        'Three claims in one sentence and none of them is sourced: that the replacement was outright, that it ' +
        'finished in 2023, and that the cost landed on the largest components. The third is the one I doubt most.',
      resolution:
        'Only the third could be confirmed. The 2023 State of Vue survey reports 61% composable adoption, so ' +
        '"outright" is now "largely", and the date is now "by 2023" rather than "by the end of 2023". The cost ' +
        'claim holds and is sourced in [[E3]]; the migration notes are in [[R2:E5]].',
      resolved_revision: 9,
      attempts: 1,
      rejections: [
        { reason: 'The first answer changed the sentence without citing anything.', revision: 7, at: '2025-03-19T20:00:00Z' },
      ],
    }),
  },
}
