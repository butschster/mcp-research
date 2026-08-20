import type { Meta, StoryObj } from '@storybook/vue3'
import KindChip from './KindChip.vue'
import { KIND_META, type AnnotationKind } from '../../composables/useAnnotations'

/**
 * What a mark asks for: verify, dig, or disagree.
 *
 * A third vocabulary next to status and anchor state, and the stories below are
 * mostly here to keep it that way — `verify` and `open` answer different
 * questions about the same mark, so a reviewer comparing this catalogue page
 * against StatusBadge should see two dictionaries, not one with duplicates.
 *
 * The glyph ships with the label in every variant, including `iconOnly`, where
 * the label moves to a `.sr-only` span rather than disappearing. Colour is never
 * the only thing telling the three apart — check that by reading the row in
 * `AllKinds` with the colours mentally removed.
 */
const meta: Meta<typeof KindChip> = {
  title: 'Annotations/KindChip',
  component: KindChip,
  tags: ['autodocs'],
  argTypes: {
    kind: { control: 'select', options: ['verify', 'dig', 'disagree'] },
    size: { control: 'inline-radio', options: ['sm', 'md'] },
    iconOnly: { control: 'boolean' },
  },
  args: { kind: 'verify', size: 'md', iconOnly: false },
}
export default meta
type Story = StoryObj<typeof KindChip>

/** Find a source, or say plainly it could not be confirmed. */
export const Verify: Story = { args: { kind: 'verify' } }

/** Write a child document and link it from here. */
export const Dig: Story = { args: { kind: 'dig' } }

/** Record both positions — do not rewrite the text. */
export const Disagree: Story = { args: { kind: 'disagree' } }

const KINDS = Object.keys(KIND_META) as AnnotationKind[]

/** The three side by side, which is the only way to check they are legible as a set. */
export const AllKinds: Story = {
  render: () => ({
    components: { KindChip },
    setup: () => ({ kinds: KINDS }),
    template: `
      <div style="display: flex; flex-wrap: wrap; gap: 0.75rem; align-items: center;">
        <KindChip v-for="k in kinds" :key="k" :kind="k" />
      </div>
    `,
  }),
}

/**
 * `sm` is what the row and the thread header use — a chip beside a ShortCode
 * should not be taller than it.
 */
export const Small: Story = { args: { kind: 'dig', size: 'sm' } }

/** Both sizes, so the difference is a comparison rather than a memory. */
export const AllSizes: Story = {
  render: () => ({
    components: { KindChip },
    setup: () => ({ kinds: KINDS }),
    template: `
      <div style="display: grid; gap: 0.75rem;">
        <div style="display: flex; gap: 0.75rem; align-items: center;">
          <span style="width: 3rem; font-size: 0.75rem; color: var(--color-text-faint);">md</span>
          <KindChip v-for="k in kinds" :key="k" :kind="k" size="md" />
        </div>
        <div style="display: flex; gap: 0.75rem; align-items: center;">
          <span style="width: 3rem; font-size: 0.75rem; color: var(--color-text-faint);">sm</span>
          <KindChip v-for="k in kinds" :key="k" :kind="k" size="sm" />
        </div>
      </div>
    `,
  }),
}

/**
 * `iconOnly` — for the gutter and anywhere the label would not fit.
 *
 * The word is still in the DOM, in a `.sr-only` span, so a screen reader gets
 * "Dig" where a sighted reader gets "↓". Dropping it instead would leave the
 * glyph as the only carrier, and a glyph is not a word.
 */
export const IconOnly: Story = { args: { kind: 'dig', iconOnly: true } }

/** Every combination the product can reach: three kinds × two sizes × icon-only. */
export const AllVariants: Story = {
  render: () => ({
    components: { KindChip },
    setup: () => ({ kinds: KINDS }),
    template: `
      <table style="border-collapse: separate; border-spacing: 0.75rem;">
        <thead>
          <tr style="font-size: 0.75rem; color: var(--color-text-faint); text-align: left;">
            <th></th><th>md</th><th>sm</th><th>md icon</th><th>sm icon</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="k in kinds" :key="k">
            <td style="font-size: 0.75rem; color: var(--color-text-faint);">{{ k }}</td>
            <td><KindChip :kind="k" size="md" /></td>
            <td><KindChip :kind="k" size="sm" /></td>
            <td><KindChip :kind="k" size="md" icon-only /></td>
            <td><KindChip :kind="k" size="sm" icon-only /></td>
          </tr>
        </tbody>
      </table>
    `,
  }),
}

/**
 * A kind the component has never heard of — an older client reading a mark made
 * by a newer server.
 *
 * `KIND_META` misses, so the chip falls back to a bullet and prints the raw
 * value as its label. That is deliberate: an unstyled chip saying `escalate` is
 * recoverable, an empty span is not.
 */
export const UnknownKind: Story = {
  args: { kind: 'escalate' as AnnotationKind },
}
