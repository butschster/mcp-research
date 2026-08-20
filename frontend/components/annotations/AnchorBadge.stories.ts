import type { Meta, StoryObj } from '@storybook/vue3'
import AnchorBadge from './AnchorBadge.vue'
import { ANCHOR_META, type AnchorState } from '../../composables/useAnnotations'

/**
 * Where the marked text is now — a third axis beside kind and status.
 *
 * A mark can be open and drifted, or answered and orphaned, so this cannot be
 * folded into StatusBadge without one of the two facts being dropped. The
 * stories are grouped so the four states can be read as a set: anchored is
 * quiet, drifted and orphaned carry colour because they are the two a person
 * has to act on.
 *
 * The hint is a `title`, so hover a badge in the canvas to read it.
 */
const meta: Meta<typeof AnchorBadge> = {
  title: 'Annotations/AnchorBadge',
  component: AnchorBadge,
  tags: ['autodocs'],
  argTypes: {
    state: { control: 'select', options: ['anchored', 'drifted', 'moved', 'orphaned'] },
    confidence: { control: { type: 'range', min: 0, max: 1, step: 0.01 } },
    entryType: { control: 'inline-radio', options: ['blocks', 'markdown'] },
  },
  args: { state: 'anchored', entryType: 'blocks' },
}
export default meta
type Story = StoryObj<typeof AnchorBadge>

/** The marked text is where it was. Nothing to do. */
export const Anchored: Story = { args: { state: 'anchored', confidence: 1 } }

/**
 * The block is still here, but the sentence under the mark changed.
 *
 * ThreadCard pairs this with a link into the diff, because "was my doubt
 * addressed or quietly buried" is the only question a drifted mark raises.
 */
export const Drifted: Story = { args: { state: 'drifted', confidence: 0.72 } }

/**
 * The sentence turned up elsewhere in the document, and the placement was
 * recovered rather than exact — so the percentage prints.
 */
export const Moved: Story = { args: { state: 'moved', confidence: 0.64 } }

/**
 * A recovery the server is barely willing to claim. The number is the whole
 * point of the state: 60% means go and look.
 */
export const MovedLowConfidence: Story = { args: { state: 'moved', confidence: 0.41 } }

/**
 * Moved, but matched exactly — `confidence` is 1, so no percentage is printed.
 *
 * This is the rule the component exists to enforce: showing "100%" beside every
 * ordinary mark trains people to stop reading the number, which is the number
 * that matters on the one row where it is 41.
 */
export const MovedFullConfidence: Story = { args: { state: 'moved', confidence: 1 } }

/**
 * `moved` with no confidence at all — an older payload, or a strategy that does
 * not report one. The badge says "Moved" and stays silent about how sure it is,
 * rather than printing 0%.
 */
export const MovedNoConfidence: Story = { args: { state: 'moved', confidence: undefined } }

/** The marked text is gone from the document. */
export const Orphaned: Story = { args: { state: 'orphaned', confidence: 0 } }

const STATES = Object.keys(ANCHOR_META) as AnchorState[]

/** All four, with the confidence each one is realistically carrying. */
export const AllStates: Story = {
  render: () => ({
    components: { AnchorBadge },
    setup: () => ({
      rows: [
        { state: 'anchored' as AnchorState, confidence: 1 },
        { state: 'drifted' as AnchorState, confidence: 0.72 },
        { state: 'moved' as AnchorState, confidence: 0.64 },
        { state: 'orphaned' as AnchorState, confidence: 0 },
      ],
    }),
    template: `
      <div style="display: flex; flex-wrap: wrap; gap: 0.75rem; align-items: center;">
        <AnchorBadge v-for="r in rows" :key="r.state" :state="r.state" :confidence="r.confidence" />
      </div>
    `,
  }),
}

/**
 * The same four on a markdown document.
 *
 * Visually identical — the difference is in the `title`, which adds that the
 * document has no blocks so the mark is found by its text alone. Saying that
 * out loud beats implying a precision the document cannot give, and it is why
 * `entryType` is a prop rather than something the badge could infer.
 */
export const OnMarkdownDocument: Story = {
  render: () => ({
    components: { AnchorBadge },
    setup: () => ({ states: STATES }),
    template: `
      <div>
        <p style="margin: 0 0 0.75rem; font-size: 0.75rem; color: var(--color-text-faint);">
          Hover each badge — the hint carries the markdown caveat.
        </p>
        <div style="display: flex; flex-wrap: wrap; gap: 0.75rem; align-items: center;">
          <AnchorBadge v-for="s in states" :key="s" :state="s" :confidence="0.81" entry-type="markdown" />
        </div>
      </div>
    `,
  }),
}

/**
 * A state the component has never heard of. `ANCHOR_META` misses and it falls
 * back to the anchored wording with a bullet — wrong, but readable, which beats
 * an empty badge sitting where a warning should be.
 */
export const UnknownState: Story = { args: { state: 'pinned' as AnchorState } }
