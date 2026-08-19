import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import type { ImportPreview } from '~/composables/useSectionImport'
import ImportPreviewDialog from './ImportPreviewDialog.vue'
import {
  mockCleanPreview,
  mockFilenameTitlePreview,
  mockLongBodyPreview,
  mockNoisyPreview,
} from '../../__mocks__/import'

/**
 * What we made of the file, before anything is written.
 *
 * Every state here is a prop — `preview`, `committing`, `error`, `staleSpec` —
 * with one exception: the title draft, which the dialog seeds from
 * `preview.title` on open and then owns. So every screen below is reachable
 * declaratively, and the only thing a story cannot stage is a title the person
 * has typed over.
 *
 * The ledger under "What we did with the front matter" is derived, not passed:
 * the dialog folds `metadata_report.invalid_values`, `missing_required`,
 * `unknown_keys`, the non-applied entries of `fields`, `unresolved_refs`,
 * `refused`, `ignored` and the *applied-with-a-caveat* entries of `fields` into
 * at most five groups, ordered attention → unknown → references → refused →
 * ignored. Which groups appear is therefore a property of the fixture, and the
 * stories below are named for the groups they produce.
 *
 * There is no `v-html` in this dialog. The body is shown as source, so a
 * `[[E44]]` that resolves to nothing stays inert text on the one screen whose
 * job is to say it points at nothing.
 *
 * It renders through `ModalOverlay`, which is a `<Teleport to="body">`;
 * `.storybook/preview.ts` stubs `Teleport` to a passthrough slot, so the dialog
 * appears inline in the story canvas rather than at the end of `<body>`.
 */
const meta: Meta<typeof ImportPreviewDialog> = {
  title: 'Research/ImportPreviewDialog',
  component: ImportPreviewDialog,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
  argTypes: {
    visible: { control: 'boolean' },
    fileBytes: { control: 'number' },
    sectionName: { control: 'text' },
    preview: { control: 'object' },
    committing: { control: 'boolean' },
    error: { control: 'text' },
    staleSpec: { control: 'boolean' },
    onCommit: { action: 'commit' },
    onReread: { action: 'reread' },
    onClose: { action: 'close' },
  },
}
export default meta
type Story = StoryObj<typeof ImportPreviewDialog>

const base = {
  visible: true,
  fileBytes: 4_812,
  sectionName: 'Findings',
  committing: false,
  error: '',
  staleSpec: false,
}

/** A clean file: front matter read, every key taken, no ledger at all. This is
 *  the shape the dialog must be fastest at — two facts, a title to accept and a
 *  button. If this screen ever needs reading rather than glancing at, the
 *  ledger has leaked into it. */
export const Clean: Story = {
  args: { ...base, preview: mockCleanPreview },
}

/** One of every note category at once, which is also every group the ledger can
 *  draw: two refused values plus a missing required field plus a replaced status
 *  (attention, open), two undeclared keys, two unresolved references, three
 *  refused keys, and — last and quietest — the housekeeping keys read and not
 *  used. Plus a decoding warning above the ledger. */
export const EveryNoteCategory: Story = {
  args: { ...base, fileBytes: 11_930, preview: mockNoisyPreview },
}

/** Only the loud group. A file whose front matter was fine except for one value
 *  the section refuses — the common real case, and the one where an unexpanded
 *  group would be a refusal nobody read. */
export const OnlyAttention: Story = {
  args: {
    ...base,
    preview: {
      ...mockCleanPreview,
      metadata_report: {
        ...mockCleanPreview.metadata_report,
        invalid_values: [
          { key: 'confidence', value: 'fairly high', reason: 'This field takes one of: low, medium, high.' },
        ],
        missing_required: ['owner'],
      },
    } satisfies ImportPreview,
  },
}

/** Only quiet groups. Every file this product downloaded itself carries `code`,
 *  `research`, `section`, `created` and `updated`, so this is what re-importing
 *  an export looks like: nothing to decide, one grey row that stays shut. */
export const OnlyIgnored: Story = {
  args: {
    ...base,
    preview: {
      ...mockCleanPreview,
      ignored: mockNoisyPreview.ignored,
    } satisfies ImportPreview,
  },
}

/** `title_source: 'filename'` — no front matter and no heading, so the title
 *  was made from the filename and the hint says so in as many words. The
 *  header line reads "no front matter" for the same reason. The title input is
 *  the tallest, brightest thing on the screen because this is the case where
 *  somebody is expected to change it. */
export const TitleFromFilename: Story = {
  args: { ...base, fileBytes: 940, preview: mockFilenameTitlePreview },
}

/** `title_source: 'heading'` — taken from the document's first `#`, which the
 *  hint distinguishes from front matter. Middle confidence: usually right,
 *  occasionally the author's working title. */
export const TitleFromHeading: Story = {
  args: { ...base, preview: { ...mockCleanPreview, title_source: 'heading' } satisfies ImportPreview },
}

/** A 312-line document. The well shows the first 40 and the hint underneath
 *  says "First 40 of 312 lines. All of it is saved." — the sentence exists
 *  because a truncated preview otherwise reads as a truncated import. The well
 *  scrolls sideways inside itself and never widens the dialog. */
export const LongBody: Story = {
  args: { ...base, fileBytes: 38_400, preview: mockLongBodyPreview },
}

/** A document shorter than the clamp: the hint drops the "first N of" phrasing
 *  and just states the count, so the clamp is never mentioned when it did not
 *  happen. */
export const ShortBody: Story = {
  args: {
    ...base,
    fileBytes: 512,
    preview: { ...mockCleanPreview, body: 'One line, and that is the whole file.\n', body_lines: 2 } satisfies ImportPreview,
  },
}

/** The commit is in flight. Both footer buttons are disabled and the primary
 *  one says "Creating…"; the report stays fully readable underneath, because
 *  the request can still fail and everything the person read has to survive
 *  that. */
export const Committing: Story = {
  args: { ...base, preview: mockCleanPreview, committing: true },
}

/** The commit failed and the dialog stayed open. The server's own sentence,
 *  in an `role="alert"` at the bottom of the body, with the title they fixed
 *  and the report they read still on screen — closing on failure would make
 *  them do the whole thing again. */
export const CommitFailed: Story = {
  args: {
    ...base,
    preview: mockCleanPreview,
    error: 'An entry titled “What counts as a seat” already exists in this section.',
  },
}

/** A failure both loud and long, on a file that already had a ledger: the alert
 *  sits below the document well, so on a noisy report it is below the fold and
 *  the dialog scrolls to nothing new. Worth staging precisely because that is
 *  the arrangement, not to endorse it. */
export const CommitFailedWithLedger: Story = {
  args: {
    ...base,
    fileBytes: 11_930,
    preview: mockNoisyPreview,
    error:
      'The field “owner” is required by this section and no value was supplied. Set it in the file’s front matter, or add it after the document exists.',
  },
}

/** `staleSpec` — a `section.updated` arrived while this was open, so the fields
 *  the preview was checked against are no longer the ones the commit will be
 *  checked by. The banner is at the top, above the title, and offers the only
 *  useful action: read the file again. Committing is still allowed; refusing
 *  outright would strand somebody whose colleague renamed an unrelated field. */
export const StaleSpec: Story = {
  args: { ...base, preview: mockNoisyPreview, staleSpec: true, fileBytes: 11_930 },
}

/** Long Cyrillic in the title input and in a quoted rejected value, at once.
 *  The input scrolls rather than wraps, so a 140-character title is editable
 *  but not fully visible — the ledger's value quote, which does wrap, is where
 *  a long string is actually readable. */
export const CyrillicOverflow: Story = {
  args: {
    ...base,
    sectionName: 'Выводы',
    fileBytes: 11_930,
    preview: mockNoisyPreview,
  },
}

/** `preview: null` with `visible: true`. Not a state `EntriesView` reaches —
 *  it only sets `visible` for the `previewing` and `committing` phases, both of
 *  which have a preview — but the prop is nullable, so this is what the guard
 *  renders: the chrome and the footer, with the primary button disabled because
 *  the title draft is empty. */
export const NoPreview: Story = {
  args: { ...base, preview: null },
}

/** Hidden. `ModalOverlay` renders nothing at all when `visible` is false, so
 *  the canvas is empty — this story exists to prove that, since a dialog that
 *  merely hides itself with CSS still traps focus. */
export const Hidden: Story = {
  args: { ...base, visible: false, preview: mockCleanPreview },
}

/** The title draft, which is the one piece of state the dialog owns. It is
 *  seeded from `preview.title` when the dialog becomes visible and reseeded
 *  whenever the preview object changes — so "Read the file again" replaces an
 *  edit, deliberately. Open, retype the title, commit, and see what was
 *  emitted; then reopen and watch the draft come back from the fixture. */
export const TitleDraftRoundTrip: Story = {
  render: () => ({
    components: { ImportPreviewDialog },
    setup() {
      const visible = ref(false)
      const which = ref<ImportPreview>(mockFilenameTitlePreview)
      const committed = ref<string | null>(null)
      function onCommit(payload: { title: string }) {
        committed.value = payload.title
        visible.value = false
      }
      function reread() {
        // Same object identity would not retrigger the watcher; the composable
        // assigns a fresh preview after a re-read, so the story does too.
        which.value = { ...which.value, title: `${which.value.title} (re-read)` }
      }
      return { visible, which, committed, onCommit, reread }
    },
    template: `
      <div style="padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-3); align-items: flex-start;">
        <button class="btn btn-sm" @click="visible = true">Drop a file</button>
        <p style="margin: 0; font-size: var(--type-xs); color: var(--color-text-muted);">
          Committed as: <code>{{ committed ?? '—' }}</code>
        </p>
        <ImportPreviewDialog
          :visible="visible"
          :file-bytes="940"
          section-name="Findings"
          :preview="which"
          :committing="false"
          error=""
          :stale-spec="true"
          @commit="onCommit"
          @reread="reread"
          @close="visible = false"
        />
      </div>
    `,
  }),
}
