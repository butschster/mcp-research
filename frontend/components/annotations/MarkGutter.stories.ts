import type { Meta, StoryObj } from '@storybook/vue3'
import MarkGutter from './MarkGutter.vue'
import type { Annotation } from '../../composables/useAnnotations'
import {
  makeAnnotation,
  mockAnnotation,
  mockAnnotationDig,
  mockAnnotationDisagree,
  mockAnnotationDrifted,
  markPositions,
} from '../../__mocks__/annotation'

/**
 * The column of markers beside the prose.
 *
 * It is positioned from measurements of somebody else's DOM: the document is
 * rendered through `v-html`, so there is no component tree to hang pins off,
 * only rectangles. `positions` is exactly what `useAnnotationOverlay.positions()`
 * returns — `top` in pixels from the top of the entry card — and the stories
 * below hand it fixed numbers so the pins land beside the fake paragraphs drawn
 * behind them.
 *
 * Two things the catalogue cannot show you and should therefore say: the whole
 * column is `aria-hidden` with `tabindex="-1"` throughout, because forty marks
 * would otherwise be forty tab stops inside prose; and it disappears entirely
 * below 768px, where the card's padding is a few pixels and the document falls
 * back to inline chips. Switch to the Mobile viewport on any story to see the
 * second one.
 */
const meta: Meta<typeof MarkGutter> = {
  title: 'Annotations/MarkGutter',
  component: MarkGutter,
  tags: ['autodocs'],
  argTypes: {
    activeId: { control: 'text' },
    annotations: { control: false },
    positions: { control: false },
    onSelect: { action: 'select' },
  },
}
export default meta
type Story = StoryObj<typeof MarkGutter>

const LINES = [
  'Across the ecosystem, composables replaced mixins outright by the end of 2023.',
  'Teams that stayed on the Options API report the migration cost was carried by their largest components.',
  'Provide/inject is the escape hatch for deep trees.',
  'Keeping components under 200 lines is a hard rule.',
  'Pinia is a drop-in replacement for Vuex in every case.',
  'Reactivity is built on Proxy, which cannot observe property addition on arrays.',
]

/** The entry card the gutter measures itself against: relative, padded, tall. */
function scene(annotations: Annotation[], tops: number[], activeId?: string): Story['render'] {
  return (args: any) => ({
    components: { MarkGutter },
    setup: () => ({
      args,
      annotations,
      positions: markPositions(annotations, tops),
      activeId: activeId ?? null,
      lines: LINES,
    }),
    template: `
      <div style="position: relative; min-height: 420px; padding: 1rem 1rem 1rem 5rem;
                  background: var(--color-surface-raised); border: 1px solid var(--color-border);
                  border-radius: var(--radius); --entry-pad: 1rem;">
        <MarkGutter
          :annotations="annotations"
          :positions="positions"
          :active-id="activeId"
          @select="args.onSelect"
        />
        <p v-for="(line, i) in lines" :key="i"
           style="margin: 0 0 1.25rem; line-height: 1.7; max-width: 40rem; color: var(--color-text-muted);">
          {{ line }}
        </p>
      </div>
    `,
  })
}

/** A single mark against the opening sentence. */
export const OneMark: Story = { render: scene([mockAnnotation], [8]) }

/**
 * Four marks down the page, across all three kinds. The pin carries the kind's
 * glyph and the mark's code, so a person can find A5 in the queue and know where
 * on the page it lives.
 */
export const SeveralMarks: Story = {
  render: scene(
    [mockAnnotation, mockAnnotationDig, mockAnnotationDisagree, mockAnnotationDrifted],
    [8, 110, 210, 320],
  ),
}

/**
 * Two marks on one line — 12px apart, inside the 10px same-line window's reach
 * once rounding is allowed for — sit side by side.
 *
 * This is the ordinary case, not an edge one: "find me a source for this" and
 * "and I do not believe it" belong on the same sentence. An earlier version
 * collapsed them into a single counting dot, which said one thing had happened
 * where two had. The code disappears when a row holds more than one mark —
 * printing it would claim the row belongs to that one.
 */
export const TwoOnOneLine: Story = {
  render: scene([mockAnnotation, mockAnnotationDig, mockAnnotationDisagree], [8, 12, 300]),
}

/**
 * Six marks on one line. Three fit in the column; the rest are counted after
 * them.
 *
 * The count is *in addition to* the pins, never instead: the reader still sees
 * three kinds and can still click them. `+3` says only that the line carries
 * more than the column can draw.
 */
export const MoreThanFits: Story = {
  render: scene(
    [
      mockAnnotation,
      mockAnnotationDig,
      mockAnnotationDisagree,
      makeAnnotation({ code: 'A20', kind: 'verify' }),
      makeAnnotation({ code: 'A21', kind: 'dig' }),
      makeAnnotation({ code: 'A22', kind: 'disagree' }),
    ],
    [40, 42, 44, 46, 48, 50],
  ),
}

/**
 * Six marks 20px apart — each on its own line of prose.
 *
 * Nothing groups: 20px is outside the same-line window, so this is six rows of
 * one, each keeping its code. The shape a heavily marked paragraph takes when
 * the marks are on different lines rather than the same one.
 */
export const OnePerLine: Story = {
  render: scene(
    [
      mockAnnotation,
      mockAnnotationDig,
      mockAnnotationDisagree,
      makeAnnotation({ code: 'A23', kind: 'verify' }),
      makeAnnotation({ code: 'A24', kind: 'dig' }),
      makeAnnotation({ code: 'A25', kind: 'disagree' }),
    ],
    [20, 40, 60, 80, 100, 120],
  ),
}

/**
 * One mark selected — the pin takes an outline, not a fill, so an active pin and
 * a warning pin stay a glance apart.
 */
export const ActiveMark: Story = {
  render: scene(
    [mockAnnotation, mockAnnotationDig, mockAnnotationDisagree],
    [8, 110, 210],
    mockAnnotationDig.id,
  ),
}

/** A long document with marks all the way down, to check the column scrolls with it. */
export const ManyMarks: Story = {
  render: (args: any) => ({
    components: { MarkGutter },
    setup() {
      const annotations = Array.from({ length: 18 }, (_, i) =>
        makeAnnotation({
          code: `A${i + 30}`,
          kind: (['verify', 'dig', 'disagree'] as const)[i % 3],
        }))
      const positions = markPositions(annotations, annotations.map((_, i) => 8 + i * 46))
      return { args, annotations, positions }
    },
    template: `
      <div style="position: relative; min-height: 900px; padding: 1rem 1rem 1rem 5rem;
                  background: var(--color-surface-raised); border: 1px solid var(--color-border);
                  border-radius: var(--radius); --entry-pad: 1rem;">
        <MarkGutter :annotations="annotations" :positions="positions" @select="args.onSelect" />
        <p style="margin: 0; max-width: 40rem; color: var(--color-text-faint); font-size: 0.875rem;">
          Eighteen marks, 46px apart — one per line, and none of them a tab stop.
        </p>
      </div>
    `,
  }),
}

/**
 * A position whose annotation is not in the list — a mark deleted between the
 * paint and the measure. The pin is skipped rather than drawn empty.
 */
export const PositionWithoutAnnotation: Story = {
  render: (args: any) => ({
    components: { MarkGutter },
    setup: () => ({
      args,
      annotations: [mockAnnotation],
      positions: [
        { id: mockAnnotation.id, code: mockAnnotation.code, top: 8 },
        { id: 'ann_gone', code: 'A99', top: 120 },
      ],
    }),
    template: `
      <div style="position: relative; min-height: 260px; padding: 1rem 1rem 1rem 5rem;
                  background: var(--color-surface-raised); border: 1px solid var(--color-border);
                  border-radius: var(--radius); --entry-pad: 1rem;">
        <MarkGutter :annotations="annotations" :positions="positions" @select="args.onSelect" />
        <p style="margin: 0; max-width: 40rem; color: var(--color-text-faint); font-size: 0.875rem;">
          Two positions, one annotation — only A1 gets a pin.
        </p>
      </div>
    `,
  }),
}

/** Nothing marked: the column renders and draws nothing, taking no space from the prose. */
export const Empty: Story = { render: scene([], []) }
