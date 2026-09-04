import type { Meta, StoryObj } from '@storybook/vue3'
import TagList from './TagList.vue'

/**
 * The tag chips, on a card and in the entries toolbar.
 *
 * Two things about it are not obvious from the props. A clickable tag renders a
 * `<button>` and an inert one a `<span>` — an inert button is its own kind of
 * lie, and the clickable chip is the primary filter on two surfaces, so it needs
 * the tab stop and the `aria-pressed` a span cannot carry. And a chip caps at
 * 12rem with an ellipsis, with the full text in `title`: tags are agent-written
 * and `kubernetes-admission-controller-webhook-v2` is a real one.
 */
const meta: Meta<typeof TagList> = {
  title: 'Base/TagList',
  component: TagList,
  tags: ['autodocs'],
  argTypes: {
    tags: { control: 'object' },
    clickable: { control: 'boolean' },
    activeTag: { control: 'text' },
    counts: { control: 'object' },
  },
}
export default meta
type Story = StoryObj<typeof TagList>

export const Empty: Story = {
  args: { tags: [] },
}

export const FewTags: Story = {
  args: { tags: ['vue', 'composables', 'typescript'] },
}

export const ManyTags: Story = {
  args: {
    tags: ['vue', 'react', 'angular', 'svelte', 'typescript', 'javascript', 'css', 'architecture', 'performance', 'testing'],
  },
}

export const Clickable: Story = {
  args: {
    tags: ['vue', 'composables', 'typescript', 'architecture'],
    clickable: true,
  },
}

export const WithActiveTag: Story = {
  args: {
    tags: ['vue', 'composables', 'typescript', 'architecture'],
    clickable: true,
    activeTag: 'composables',
  },
}

export const WithCounts: Story = {
  args: {
    tags: ['vue', 'composables', 'typescript', 'architecture'],
    clickable: true,
    counts: { vue: 12, composables: 5, typescript: 8, architecture: 3 },
  },
}

export const WithActiveAndCounts: Story = {
  args: {
    tags: ['vue', 'composables', 'typescript', 'architecture'],
    clickable: true,
    activeTag: 'vue',
    counts: { vue: 12, composables: 5, typescript: 8, architecture: 3 },
  },
}

/**
 * The truncation, which is this component's and not the toolbar's.
 *
 * Before the cap a tag this long widened the card it sat on and, in the
 * one-row entries toolbar, pushed every control after it onto a second line.
 * Each chip stops at 12rem — about twenty characters here — and keeps the whole
 * tag in `title`, so nothing is lost, only folded.
 *
 * Two details worth reading off the render: the ellipsis is on `.tag-text` and
 * not on the chip, because `.tag` is `inline-flex` and an ellipsis only draws on
 * a block box with its own overflow; and the count stays visible on a truncated
 * chip, because it is a sibling of the text with `flex: none`.
 */
export const LongTagNames: Story = {
  args: {
    tags: [
      'kubernetes-admission-controller-webhook-v2',
      'observability-platform-migration-plan-q3',
      'vue',
      'the-longest-tag-anyone-has-ever-typed-into-this-product-so-far',
    ],
    clickable: true,
    activeTag: 'observability-platform-migration-plan-q3',
    counts: {
      'kubernetes-admission-controller-webhook-v2': 12,
      'observability-platform-migration-plan-q3': 4,
      vue: 2,
      'the-longest-tag-anyone-has-ever-typed-into-this-product-so-far': 1,
    },
  },
}

/**
 * Counts supplied, and every one of them is 1 — a young research, where no two
 * documents share a tag yet.
 *
 * Nothing is printed. A `1` beside a tag says only that the tag exists, which
 * the chip already said, and a row of them is noise on the surface that has the
 * least room for it. The rule is `count > 1`, so this is the same data as
 * `WithCounts` minus everything worth reading.
 */
export const CountsOfOne: Story = {
  args: {
    tags: ['vue', 'composables', 'typescript', 'architecture'],
    clickable: true,
    counts: { vue: 1, composables: 1, typescript: 1, architecture: 1 },
  },
}
