import type { Meta, StoryObj } from '@storybook/vue3'
import PassReviewModal from './PassReviewModal.vue'
import {
  makeAnnotation,
  mockAnnotationsAnswered,
  mockAnnotationAnswered,
} from '../../__mocks__/annotation'

/**
 * Accepting a pass: the rail of answered marks on the left, the answer to the
 * selected one on the right, one decision at the bottom.
 *
 * The point of the screen is that it is not fifty clicks. It also nests three
 * levels of the annotation components — AnnotationList inside the rail,
 * AnnotationRow inside that, KindChip and AnchorBadge inside that — all resolved
 * by their Nuxt-derived names, which Storybook only knows because
 * `.storybook/preview.ts` registers them. Unregistered, this modal renders with
 * an empty rail and looks like a data problem.
 *
 * Two behaviours worth watching rather than reading: the selection resets every
 * time the modal opens, because a fresh batch is a fresh decision and a carried
 * selection accepts rows nobody looked at; and "Send back" goes through a
 * `SendBackModal`, so cancelling it emits nothing.
 */
const meta: Meta<typeof PassReviewModal> = {
  title: 'Annotations/PassReviewModal',
  component: PassReviewModal,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
  argTypes: {
    visible: { control: 'boolean' },
    researchSlug: { control: 'text' },
    busy: { control: 'boolean' },
    result: { control: 'text' },
    annotations: { control: false },
    onAccept: { action: 'accept' },
    'onSend-back': { action: 'send-back' },
    onClose: { action: 'close' },
  },
  args: {
    visible: true,
    researchSlug: 'R1',
    annotations: mockAnnotationsAnswered,
    busy: false,
    result: null,
  },
}
export default meta
type Story = StoryObj<typeof PassReviewModal>

/**
 * Four answered marks across two documents. Nothing is selected yet, so both
 * decision buttons are disabled — the first row is only *shown* in the detail
 * pane, which is not the same as being picked.
 */
export const AnsweredBatch: Story = {}

/**
 * One mark. The lead line is the reason this story exists: "1 mark was answered
 * and is waiting for you", singular throughout, because "1 marks were" is the
 * kind of thing that makes a person distrust the count next to it.
 */
export const SingleMark: Story = {
  args: { annotations: [mockAnnotationAnswered] },
}

/**
 * A batch is being applied. Both decisions disable and Accept reads "Working…",
 * so the same fourteen marks cannot be accepted twice while the request is out.
 */
export const Busy: Story = {
  args: { busy: true },
}

/**
 * A partial failure — its own outcome, and one the server reports per row so it
 * can be said without asking again.
 *
 * Twelve landed, two did not, and the two that did not are still in the rail to
 * be retried. A batch call that could only say "failed" would have made this
 * screen unusable at exactly the size it is meant for.
 */
export const PartialFailure: Story = {
  args: {
    result: '12 of 14 accepted. A18 and A22 were changed by the agent while you were reviewing — reload and check those two.',
  },
}

/** Everything applied, said plainly. */
export const AllAccepted: Story = {
  args: { result: '14 of 14 accepted.' },
}

/**
 * Fourteen marks across four documents — the size the screen is built for.
 *
 * The rail scrolls inside the modal; the detail pane does not, so the answer
 * stays put while the list moves under it.
 */
export const LargeBatch: Story = {
  args: {
    annotations: (() => {
      const entries = [
        { id: 'ent_001', code: 'E1', title: 'Component Composition Patterns' },
        { id: 'ent_002', code: 'E2', title: 'Reactive State Management' },
        { id: 'ent_003', code: 'E3', title: 'Template Syntax Deep Dive' },
        { id: 'ent_004', code: 'E4', title: 'Slots and Render Functions' },
      ]
      return Array.from({ length: 14 }, (_, i) => {
        const entry = entries[i % entries.length]!
        return makeAnnotation({
          code: `A${200 + i}`,
          entry_id: entry.id,
          entry_code: entry.code,
          entry_title: entry.title,
          status: 'answered',
          kind: (['verify', 'dig', 'disagree'] as const)[i % 3],
          attempts: i % 4 === 0 ? 1 : 0,
          resolution: `Confirmed against the source and recorded in [[E${(i % 6) + 1}]].`,
          resolved_revision: 6 + i,
          answered_at: '2025-03-20T09:00:00Z',
          anchor: {
            state: (['anchored', 'moved', 'drifted', 'anchored'] as const)[i % 4]!,
            strategy: 'block+quote',
            confidence: i % 4 === 1 ? 0.68 : 1,
            block_index: i,
          },
        })
      })
    })(),
  },
}

/**
 * An answer with no note behind it and a cross-reference in the resolution.
 *
 * The answer goes through `renderRefs`, so `[[E3]]` is a link and anything
 * angle-bracketed in it is text — the mark's body is written by a person and
 * reaches `v-html` nowhere, but the resolution is written by an agent and does.
 */
export const AnswerWithCrossRefs: Story = {
  args: {
    annotations: [
      makeAnnotation({
        code: 'A30',
        status: 'answered',
        body: undefined,
        resolution:
          'Could not be confirmed. The claim traces back to a single blog post, which is now cited in [[E3]] ' +
          'as an opinion rather than a finding; the survey it misquotes is [[R2:E5]].',
        resolved_revision: 11,
        answered_at: '2025-03-20T12:00:00Z',
      }),
    ],
  },
}

/**
 * A mark answered with nothing at all — the pane prints an em dash rather than
 * an empty paragraph, so "the agent said nothing" is visible instead of looking
 * like a rendering failure.
 */
export const AnswerMissing: Story = {
  args: {
    annotations: [makeAnnotation({ code: 'A31', status: 'answered', resolution: undefined })],
  },
}

/**
 * Nothing to review.
 *
 * Reachable when the last mark of a batch is accepted in another tab: the rail
 * shows its empty state, the detail pane asks for a pick that cannot be made,
 * and every button is disabled. The lead line above still reads "0 marks were
 * answered", which is the honest sentence for it.
 */
export const NothingToReview: Story = {
  args: { annotations: [] },
}

/** Closed — the modal is mounted and draws nothing. */
export const Hidden: Story = {
  args: { visible: false },
}
