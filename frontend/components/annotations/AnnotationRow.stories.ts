import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import AnnotationRow from './AnnotationRow.vue'
import type { Annotation } from '../../composables/useAnnotations'
import {
  makeAnnotation,
  mockAnnotation,
  mockAnnotationAnswered,
  mockAnnotationRejected,
  mockAnnotationOrphaned,
  mockAnnotationMoved,
  mockAnnotationQueue,
} from '../../__mocks__/annotation'

/**
 * One mark, as a line.
 *
 * A component rather than the `.data-row` classes on their own, because three
 * places draw this line — the document panel, the research queue and the pass
 * review — and a mark that looks different in the queue than it does in the
 * review is a mark somebody accepts twice. That is what `AllStates` is for: it
 * is the same component under every combination those three callers produce.
 *
 * The row is a `role="option"`; its keyboard handling lives in the parent
 * listbox (see AnnotationList), so a row on its own has no arrow-key behaviour
 * and is not a tab stop.
 */
const meta: Meta<typeof AnnotationRow> = {
  title: 'Annotations/AnnotationRow',
  component: AnnotationRow,
  tags: ['autodocs'],
  argTypes: {
    selectable: { control: 'boolean' },
    selected: { control: 'boolean' },
    dense: { control: 'boolean' },
    annotation: { control: false },
    onOpen: { action: 'open' },
    'onToggle-select': { action: 'toggle-select' },
  },
  args: { selectable: false, selected: false, dense: false },
  decorators: [
    (story) => ({
      components: { story },
      template: '<div class="card card--list"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof AnnotationRow>

/** Open, anchored, with a note under the quote. */
export const Default: Story = { args: { annotation: mockAnnotation } }

/** Answered and waiting for acceptance. */
export const Answered: Story = { args: { annotation: mockAnnotationAnswered } }

/**
 * Selected. The state is carried by the border, not a fill — a tinted row and a
 * warning row would otherwise be one glance apart, and the warning is the one
 * that matters.
 */
export const Selected: Story = { args: { annotation: mockAnnotation, selected: true } }

/** With a checkbox, for the batch-decision surfaces. Unchecked. */
export const Selectable: Story = { args: { annotation: mockAnnotation, selectable: true } }

/** Checkbox and border together — what a picked row looks like in the pass review. */
export const SelectableSelected: Story = {
  args: { annotation: mockAnnotation, selectable: true, selected: true },
}

/**
 * `dense` hides the note and keeps the quote, for the narrow rail beside a
 * detail pane where the note would be read in full anyway.
 */
export const Dense: Story = {
  args: { annotation: mockAnnotation, dense: true, selectable: true },
}

/**
 * Sent back twice: `↩2` beside the badges, in the warning colour, with the count
 * in its `title`. Two attempts on one mark is the signal that a pass is looping.
 */
export const WithAttempts: Story = { args: { annotation: mockAnnotationRejected } }

/** A recovered placement — the anchor badge prints the confidence. */
export const MovedAnchor: Story = { args: { annotation: mockAnnotationMoved } }

/** The marked text is gone; the row still shows what it said. */
export const Orphaned: Story = { args: { annotation: mockAnnotationOrphaned } }

/** No note — just the quote, with nothing beneath it. */
export const NoNote: Story = {
  args: { annotation: makeAnnotation({ code: 'A19', body: undefined }) },
}

/**
 * A paragraph-length quote and a long note in a fixed-width row: both ellipsise
 * on one line each. The row's job is to be scanned; the thread is where the full
 * text is read.
 */
export const LongQuote: Story = {
  args: {
    annotation: makeAnnotation({
      code: 'A20',
      quote: {
        exact:
          'Across the ecosystem, composables replaced mixins outright by the end of 2023, and every team that ' +
          'stayed on the Options API reports that the migration cost was carried almost entirely by their ' +
          'largest components rather than being spread evenly across the codebase.',
      },
      body:
        'Three claims in one sentence and none of them is sourced — the replacement being outright, the date, ' +
        'and where the migration cost landed.',
    }),
  },
}

/** A mark whose anchor the server did not report at all: the row assumes anchored. */
export const NoAnchorPayload: Story = {
  args: { annotation: makeAnnotation({ code: 'A21', anchor: undefined }) },
}

/** Every state the three callers can produce, stacked, so they can be compared at once. */
export const AllStates: Story = {
  render: () => ({
    components: { AnnotationRow },
    setup: () => ({ rows: mockAnnotationQueue }),
    template: `
      <div>
        <AnnotationRow v-for="a in rows" :key="a.id" :annotation="a" />
      </div>
    `,
  }),
}

/**
 * The selection actually working: click a checkbox and the row's border follows.
 *
 * Selection is the parent's state, not the row's — the row only emits — and this
 * is the story that proves the round trip rather than posing it.
 */
export const InteractiveSelection: Story = {
  render: () => ({
    components: { AnnotationRow },
    setup() {
      const rows = mockAnnotationQueue.slice(0, 4)
      const selected = ref<string[]>([rows[1]!.id])
      const toggle = (a: Annotation) => {
        selected.value = selected.value.includes(a.id)
          ? selected.value.filter((id) => id !== a.id)
          : [...selected.value, a.id]
      }
      return { rows, selected, toggle }
    },
    template: `
      <div>
        <AnnotationRow
          v-for="a in rows"
          :key="a.id"
          :annotation="a"
          selectable
          :selected="selected.includes(a.id)"
          @toggle-select="toggle(a)"
        />
      </div>
    `,
  }),
}
