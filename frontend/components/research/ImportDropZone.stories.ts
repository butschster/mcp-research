import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import ImportDropZone from './ImportDropZone.vue'

/**
 * The drop target for one markdown file, wrapped around a section's entry list.
 *
 * At rest it is invisible: it renders its slot and nothing else. A dashed box
 * standing there permanently turns a section into a file-upload screen, and
 * this is a section. The overlay only exists while a file is genuinely over the
 * pane — `DataTransfer.types` containing `'Files'`, not any drag at all, so
 * dragging a Kanban card across an entry list does not arm it.
 *
 * That is what makes this component awkward to stage: its interesting state is
 * a function of a live browser drag, which no prop reaches. The `ArmedByFileDrag`
 * story synthesises one — a real `DragEvent` carrying a real `DataTransfer` with
 * a file in it, dispatched from a `play` function. That works in Chromium and
 * Firefox; if a browser rejects the `dataTransfer` init the overlay simply will
 * not appear, and the story is then a resting one. Nothing was added to the
 * component to make this easier: a `forceArmed` prop would be a state the
 * product cannot reach, documented as if it could.
 *
 * Refusals are emitted, never rendered here — `EntriesView` shows them above
 * this component emits are rendered by `EntriesView`, above the list. The
 * `Refusals` story shows what it emits rather than what it draws, for that
 * reason.
 */
const meta: Meta<typeof ImportDropZone> = {
  title: 'Research/ImportDropZone',
  component: ImportDropZone,
  tags: ['autodocs'],
  argTypes: {
    enabled: { control: 'boolean' },
    sectionName: { control: 'text' },
    maxBytes: { control: 'number' },
    reading: { control: 'boolean' },
    onFile: { action: 'file' },
    onRefuse: { action: 'refuse' },
  },
}
export default meta
type Story = StoryObj<typeof ImportDropZone>

const EXTENSIONS = ['.md', '.markdown']
const MAX_BYTES = 256 * 1024

/** Three entry cards, so the overlay has something the size of a real pane to
 *  sit over — its `inset` is negative, and against an empty div that reads as a
 *  box rather than as the column. */
const SLOT = `
  <div style="display: grid; gap: var(--space-3);">
    <div v-for="e in entries" :key="e.code" class="card" style="padding: var(--space-4); display: grid; gap: var(--space-2);">
      <div style="display: flex; align-items: center; gap: var(--space-2);">
        <code style="font-family: 'JetBrains Mono', monospace; font-size: var(--type-xs); color: var(--color-text-faint);">{{ e.code }}</code>
        <strong style="font-size: var(--type-base);">{{ e.title }}</strong>
      </div>
      <p style="margin: 0; font-size: var(--type-sm); color: var(--color-text-muted);">{{ e.description }}</p>
    </div>
  </div>
`

const entries = [
  { code: 'E1', title: 'What counts as a seat', description: 'The definition three vendors disagree on, and what it costs.' },
  { code: 'E2', title: 'Northwind list pricing', description: 'Public tiers as of the third week of the quarter.' },
  { code: 'E3', title: 'The reconstructed Cascade invoice', description: 'Dormant accounts, billed.' },
]

/** Builds a `DragEvent` that looks to the component exactly like a file drag:
 *  `dataTransfer.types` contains `'Files'` because a `File` was added to it. */
function fileDragEvent(type: string): Event {
  const dt = new DataTransfer()
  dt.items.add(new File(['# Notes\n'], 'notes.md', { type: 'text/markdown' }))
  try {
    return new DragEvent(type, { dataTransfer: dt, bubbles: true, cancelable: true })
  } catch {
    // Some engines refuse `dataTransfer` in the init dict. Attaching it after
    // the fact still satisfies the handler, which only reads `.types`.
    const e = new Event(type, { bubbles: true, cancelable: true })
    Object.defineProperty(e, 'dataTransfer', { value: dt })
    return e
  }
}

/** At rest, and this is the point: identical to no drop zone at all. The only
 *  thing it contributes to the DOM is a visually hidden `<input type="file">`. */
export const Resting: Story = {
  args: { enabled: true, sectionName: 'Findings', maxBytes: MAX_BYTES, extensions: EXTENSIONS, reading: false },
  render: (args) => ({
    components: { ImportDropZone },
    setup: () => ({ args, entries }),
    template: `<ImportDropZone v-bind="args">${SLOT}</ImportDropZone>`,
  }),
}

/** The armed overlay, staged by dispatching a synthetic file `dragenter`. The
 *  border is solid rather than dashed: `armed` is `over && enabled`, so `over`
 *  is true whenever the overlay renders and the `is-over` modifier is always
 *  applied. The list stays visible and stays untouchable — `pointer-events:
 *  none`, because a target that swallows the pointer eats the drop it was drawn
 *  for. */
export const ArmedByFileDrag: Story = {
  ...Resting,
  args: { ...Resting.args! },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    const zone = canvasElement.querySelector('.drop-zone')
    zone?.dispatchEvent(fileDragEvent('dragenter'))
  },
}

/** A viewer, or the all-entries mode where there is no single target section.
 *  The same file drag arms nothing, because `armed` requires `enabled`. The
 *  `dragover` handler still runs and still calls `preventDefault` — without it
 *  the browser navigates to the dropped file and replaces the whole app — so a
 *  drop here is answered with a refusal rather than by losing the page. Check
 *  the Actions panel: dropping emits `refuse` with `retryable: false`. */
export const ReadOnly: Story = {
  ...Resting,
  args: { ...Resting.args!, enabled: false },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    const zone = canvasElement.querySelector('.drop-zone')
    zone?.dispatchEvent(fileDragEvent('dragenter'))
    zone?.dispatchEvent(fileDragEvent('drop'))
  },
}

/** The three client-side refusals and the one acceptance, driven through the
 *  picker rather than through a drag — same `accept()` on the other side of
 *  both. They are courtesy, not enforcement: the server refuses all three
 *  again. They exist so a 40 MB mis-drop is answered instantly instead of after
 *  a minute of upload.
 *
 *  The last button re-picks the same file twice, which only fires `change` the
 *  second time because the handler clears `input.value` — exactly what somebody
 *  does after fixing the file in their editor. */
export const Refusals: Story = {
  render: () => ({
    components: { ImportDropZone },
    setup() {
      const zone = ref<any>(null)
      const log = ref<string[]>([])
      const reading = ref(false)

      function push(line: string) {
        log.value = [line, ...log.value].slice(0, 6)
      }

      /** Puts real `File`s on this zone's hidden input and fires `change`,
       *  which is the same entry point the picker uses. Scoped through the
       *  component's own root element, because in docs mode several stories
       *  share one page and `document.querySelector` would find the first. */
      function pick(files: File[]) {
        const root = zone.value?.$el as HTMLElement | undefined
        const input = root?.querySelector<HTMLInputElement>('input[type=file]')
        if (!input) return
        const dt = new DataTransfer()
        files.forEach((f) => dt.items.add(f))
        input.files = dt.files
        input.dispatchEvent(new Event('change', { bubbles: true }))
      }

      const good = () => new File(['# Notes\n'], 'meeting-notes.md', { type: 'text/markdown' })
      const wrongType = () => new File(['%PDF-1.7'], 'contract.pdf', { type: 'application/pdf' })
      const tooBig = () => new File([new Uint8Array(320 * 1024)], 'whole-vault.md', { type: 'text/markdown' })

      return {
        zone,
        log,
        reading,
        entries,
        push,
        pick,
        good,
        wrongType,
        tooBig,
        onFile: (f: File) => push(`file → ${f.name} (${Math.round(f.size / 1024)} KB)`),
        onRefuse: (r: { message: string; retryable: boolean }) =>
          push(`refuse → ${r.message} (retryable: ${r.retryable})`),
      }
    },
    template: `
      <div style="display: grid; gap: var(--space-4);">
        <div style="display: flex; flex-wrap: wrap; gap: var(--space-2);">
          <button class="btn btn-sm" @click="zone?.open()">Open the picker</button>
          <button class="btn btn-sm" @click="pick([good()])">Pick a good .md</button>
          <button class="btn btn-sm" @click="pick([wrongType()])">Pick a .pdf</button>
          <button class="btn btn-sm" @click="pick([tooBig()])">Pick a 320 KB .md</button>
          <button class="btn btn-sm" @click="pick([good(), good()])">Pick two files</button>
          <button class="btn btn-sm" @click="reading = !reading">reading: {{ reading }}</button>
        </div>

        <pre style="margin: 0; padding: var(--space-3); border: 1px solid var(--color-border); border-radius: var(--radius-sm); background: var(--color-bg); font-family: 'JetBrains Mono', monospace; font-size: var(--type-xs); min-height: 6rem; white-space: pre-wrap;">{{ log.join('\\n') || '(nothing emitted yet)' }}</pre>

        <ImportDropZone
          ref="zone"
          :enabled="true"
          section-name="Findings"
          :max-bytes="${MAX_BYTES}"
          :extensions="EXTENSIONS"
          :reading="reading"
          @file="onFile"
          @refuse="onRefuse"
        >${SLOT}</ImportDropZone>
      </div>
    `,
  }),
}

/** Mid-read. `accept()` returns early while `reading` is true, so a second file
 *  arriving during the first one's round trip is dropped on the floor —
 *  silently, with no refusal emitted. Flip `reading` in the `Refusals` story
 *  and try to pick a good file to see it: nothing happens at all. Worth naming,
 *  because "nothing happens" is indistinguishable from a broken picker. */
export const WhileReading: Story = {
  ...Resting,
  args: { ...Resting.args!, reading: true },
}
