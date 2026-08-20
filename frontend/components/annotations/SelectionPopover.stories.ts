import type { Meta, StoryObj } from '@storybook/vue3'
import SelectionPopover from './SelectionPopover.vue'

/**
 * What appears when a person selects a sentence they do not believe.
 *
 * The popover is `position: fixed` and placed from the `DOMRect` the live
 * `Range` reported, so a story has to supply one. These build it with
 * `new DOMRect(...)` against coordinates that put the box under a paragraph
 * drawn behind it — the selection itself is painted by the caller before this
 * ever renders, which is why there is no real selection here and does not need
 * to be.
 *
 * Two behaviours are worth watching in the canvas rather than reading about:
 * the box flips above the selection when there is no room below it (see
 * `NearViewportBottom`), and at 768px it stops being a popover and becomes a bar
 * pinned to the bottom edge — switch to the Mobile viewport on any story. That
 * is not a layout preference: iOS draws its own Copy/Look-Up menu right at the
 * selection, and anything placed there fights it.
 */
const meta: Meta<typeof SelectionPopover> = {
  title: 'Annotations/SelectionPopover',
  component: SelectionPopover,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
  argTypes: {
    visible: { control: 'boolean' },
    quote: { control: 'text' },
    entryType: { control: 'inline-radio', options: ['blocks', 'markdown'] },
    saving: { control: 'boolean' },
    error: { control: 'text' },
    rect: { control: false },
    onCreate: { action: 'create' },
    onCancel: { action: 'cancel' },
  },
}
export default meta
type Story = StoryObj<typeof SelectionPopover>

const QUOTE = 'Composables replaced mixins outright by the end of 2023.'

/** The rectangle a `Range` over the highlighted sentence below would report. */
function rectAt(top: number, left = 120, width = 420, height = 22): DOMRect {
  return new DOMRect(left, top, width, height)
}

type PopoverState = {
  quote?: string
  rect?: DOMRect
  entryType?: string
  saving?: boolean
  error?: string | null
}

/**
 * Draws the paragraph the mark is being made on, with the quoted sentence
 * painted the way the caller paints it, and the popover over the top.
 */
function scene(state: PopoverState): Story['render'] {
  return (args: any) => ({
    components: { SelectionPopover },
    setup: () => ({
      args,
      quote: state.quote ?? QUOTE,
      rect: state.rect ?? rectAt(150),
      entryType: state.entryType ?? 'blocks',
      saving: !!state.saving,
      error: state.error ?? null,
    }),
    template: `
      <div style="min-height: 100vh; padding: 3rem 7.5rem; background: var(--color-bg); color: var(--color-text);">
        <p style="max-width: 46rem; line-height: 1.7;">
          Across the ecosystem, <mark style="background: var(--ann-verify-wash); color: inherit;">{{ quote }}</mark>
          Teams that stayed on the Options API report the migration cost was carried
          almost entirely by the largest components.
        </p>
        <SelectionPopover
          :visible="true"
          :rect="rect"
          :quote="quote"
          :entry-type="entryType"
          :saving="saving"
          :error="error"
          @create="args.onCreate"
          @cancel="args.onCancel"
        />
      </div>
    `,
  })
}

/** A sentence selected mid-document. `verify` is preselected; the note is empty. */
export const Default: Story = { render: scene({}) }

/**
 * A selection near the bottom of the window: the estimated height would run off
 * the screen, so the box flips above the sentence instead. Same props — only the
 * rect moved.
 */
export const NearViewportBottom: Story = {
  render: scene({ rect: rectAt(Math.max(200, window.innerHeight - 90)) }),
}

/** A selection hard against the left edge — the box clamps to an 8px margin. */
export const NearViewportEdge: Story = {
  render: scene({ rect: rectAt(150, 4, 90) }),
}

/** The request is in flight: both buttons disable and Mark becomes "Marking…". */
export const Saving: Story = { render: scene({ saving: true }) }

/**
 * The create failed.
 *
 * The message lands inside the popover, above the buttons, and the typed note
 * survives — a toast on the page underneath would take the note with it when the
 * selection collapsed.
 */
export const WithError: Story = {
  render: scene({ error: 'Could not place the mark — the document changed while you were typing.' }),
}

/**
 * A markdown document has no blocks, so the mark is found by its quote alone and
 * the popover says so before the mark exists rather than after it drifts.
 */
export const OnMarkdownDocument: Story = { render: scene({ entryType: 'markdown' }) }

/**
 * A long selection: the quote line truncates at 90 characters with an ellipsis
 * and the character count beside it keeps counting the whole thing, so a person
 * can tell a paragraph-length mark from a sentence-length one.
 */
export const LongQuote: Story = {
  render: scene({
    quote:
      'Across the ecosystem, composables replaced mixins outright by the end of 2023, and the teams that ' +
      'stayed on the Options API report that the migration cost was carried almost entirely by their largest ' +
      'components rather than being spread evenly across the codebase.',
  }),
}

/** A two-word selection — the shortest thing worth marking. */
export const ShortQuote: Story = { render: scene({ quote: 'merely discouraged' }) }

/** No rect at all: the style computes to nothing and the box lands at the origin. */
export const NoRect: Story = {
  render: (args: any) => ({
    components: { SelectionPopover },
    setup: () => ({ args, quote: QUOTE }),
    template: `
      <div style="min-height: 100vh; padding: 3rem; background: var(--color-bg); color: var(--color-text);">
        <p style="font-size: 0.875rem; color: var(--color-text-muted);">
          <code>rect</code> is null — the caller has a selection but no measurable range.
        </p>
        <SelectionPopover
          :visible="true"
          :rect="null"
          :quote="quote"
          @create="args.onCreate"
          @cancel="args.onCancel"
        />
      </div>
    `,
  }),
}

/** Not visible — nothing renders, and no leftover note is held for the next selection. */
export const Hidden: Story = {
  render: (args: any) => ({
    components: { SelectionPopover },
    setup: () => ({ args, rect: rectAt(150), quote: QUOTE }),
    template: `
      <div style="padding: 3rem; color: var(--color-text-muted); font-size: 0.875rem;">
        <code>visible: false</code> — the popover is mounted and draws nothing.
        <SelectionPopover
          :visible="false"
          :rect="rect"
          :quote="quote"
          @create="args.onCreate"
          @cancel="args.onCancel"
        />
      </div>
    `,
  }),
}
