import type { Meta, StoryObj } from '@storybook/vue3'
import MetadataBlock from './MetadataBlock.vue'
import { fails, neverResolves } from '../../__mocks__/api'
import { markupImg } from '../../__mocks__/markup'
import { withoutShare } from '../../__mocks__/share'
import {
  metadataAllOrphaned,
  metadataCyrillic,
  metadataFilled,
  metadataInvalidStage,
  metadataNothingFilled,
  metadataPartial,
  metadataUnknownOwner,
  specAtCap,
  specCyrillic,
  specSingleField,
  specSpecifications,
} from '../../__mocks__/metadata'

/**
 * What a section says its documents record, shown at the top of the document
 * body rather than in the chrome.
 *
 * That placement is the whole point. Before this existed, agents typed the same
 * facts into the first five lines of prose — "Status: draft for review,
 * 14.08.2026. Implemented by: scanner-watchdog" — because prose is what a
 * reader sees, while the stored `status` was a chip somewhere off to the side.
 * The two then disagreed, and the reader believed the prose.
 *
 * The stories to read first are `NothingFilled` and `Incomplete`, not `Filled`.
 * A blank row standing beside filled ones is the mechanism this feature runs
 * on; the happy path is what it looks like afterwards.
 */
const meta: Meta<typeof MetadataBlock> = {
  title: 'Entry/MetadataBlock',
  component: MetadataBlock,
  tags: ['autodocs'],
  decorators: [
    // Values go through the real `renderRefs`, which asks the share module
    // whether a reference may be a link. Without this, a story renders under
    // whatever share state the last story left behind.
    withoutShare(),
    () => ({ template: '<div class="card" style="max-width: 820px; padding: 1rem"><story /></div>' }),
  ],
}
export default meta
type Story = StoryObj<typeof MetadataBlock>

const base = {
  specs: specSpecifications,
  values: metadataFilled,
  status: 'active',
  researchSlug: 'R21',
  editable: true,
  entrySpecVersion: 4,
  sectionSpecVersion: 4,
}

/**
 * Every field answered.
 *
 * `Registry` is worth looking at: it is declared `ref`, so the stored value is
 * the bare code `E47` and the block wraps it before rendering, which is what
 * makes it a link here and plain text in a `text` field holding the same
 * characters. The type is doing visible work.
 */
export const Filled: Story = { args: { ...base } }

/** The state that matters: a required field nobody answered. */
export const Incomplete: Story = {
  args: {
    ...base,
    values: metadataPartial,
    metadataStatus: { missing_required: ['owner'], complete: false, spec_version: 4 },
  },
}

/**
 * Every declared field blank — a document written before the section declared
 * anything, or by an agent that ignored the schema.
 *
 * Six rows of "Not filled" is the strongest thing this component does. It is
 * also the one arrangement where the label column carries the whole layout,
 * since no value is there to hold it open.
 */
export const NothingFilled: Story = {
  args: {
    ...base,
    values: metadataNothingFilled,
    metadataStatus: { missing_required: ['owner'], complete: false, spec_version: 4 },
  },
}

/**
 * A section that declares one field.
 *
 * Worth its own story because it is what a section looks like immediately after
 * somebody's first edit, and because the block still earns its place: system
 * `status` and a section-defined `stage` sitting one above the other is the
 * pair the whole feature is built around.
 */
export const OneDeclaredField: Story = {
  args: { ...base, specs: specSingleField, values: { stage: 'draft' } },
}

/**
 * Twelve declared fields — the cap — five of them required, which is the second
 * and separate cap. Past a dozen rows the block stops being scannable at the
 * top of a document, which is the reason the cap is twelve and not a database
 * limit.
 */
export const AtTheCap: Story = {
  args: {
    ...base,
    specs: specAtCap,
    values: { ...metadataFilled, component: 'scanner', transport: 'temporal', retries: 3 },
    metadataStatus: { missing_required: ['reviewer'], complete: false, spec_version: 7 },
    entrySpecVersion: 7,
    sectionSpecVersion: 7,
  },
}

/**
 * An explicit unknown. It answers the requirement without inventing anything,
 * which is the counterweight to the fact that a model — unlike a person — never
 * leaves a field blank when it does not know.
 *
 * Read it against `NothingFilled`: "Unknown" and "Not filled" are different
 * sentences, and only one of them means somebody looked.
 */
export const ExplicitUnknown: Story = {
  args: { ...base, values: metadataUnknownOwner },
}

/**
 * A value that no longer matches its declaration. It was kept rather than
 * discarded — dropping somebody's value to protect a type is the same mistake
 * as refusing the write — and it is re-checked on every read, so the person who
 * can fix it is the one who is told.
 */
export const InvalidValue: Story = {
  args: {
    ...base,
    values: metadataInvalidStage,
    metadataStatus: {
      issues: [{ key: 'stage', reason: '"sent for review" is not one of: draft, in-review, agreed, superseded' }],
      complete: true,
      spec_version: 4,
    },
  },
}

/**
 * Values under keys the section has stopped declaring. Removing a field decides
 * what gets collected next; it is not a verdict on what was already recorded.
 */
export const WithOrphanedValues: Story = {
  args: {
    ...base,
    specs: specSpecifications.slice(0, 3),
    values: {
      stage: 'agreed',
      produces: ['scanner-watchdog'],
      related: 'SPEC-02, SPEC-03',
      owner: 'platform',
    },
    metadataStatus: { orphaned: ['owner', 'related'], complete: true, spec_version: 5 },
  },
}

/**
 * A document that is *nothing but* orphaned values: the section was rewritten
 * and every key it once declared was dropped.
 *
 * The declared list is empty, so the block renders on the strength of the
 * discarded half alone — no "Status" row, no headings above it, only "No longer
 * collected". That branch is reachable in production the moment somebody
 * replaces a spec rather than extending it, and it is the one arrangement where
 * the component's own `v-if` is carried by `orphaned` instead of `specs`.
 */
export const EntirelyOrphaned: Story = {
  args: {
    ...base,
    specs: [],
    values: metadataAllOrphaned,
    metadataStatus: { orphaned: Object.keys(metadataAllOrphaned), complete: true, spec_version: 6 },
  },
}

/** A document whose section has moved on since it was last written. */
export const BehindTheSpec: Story = {
  args: { ...base, entrySpecVersion: 4, sectionSpecVersion: 6 },
}

/** A viewer sees everything and can change nothing. */
export const ReadOnly: Story = {
  args: { ...base, editable: false },
}

/**
 * The editor, opened.
 *
 * Every control is here and nowhere else: a select for an enum, a date input, a
 * plain box for a repeated field taking commas, the `unknown` toggle beside
 * each required field, and the help line saying where the value comes from. The
 * agent reads that help text; a person editing by hand is the only one who sees
 * it rendered, so this is the story where it is checked.
 *
 * The row heights are the thing to look at. `--control-h` is stated rather than
 * derived so a select, a text box and the toggle come out level — they did not,
 * once, and the difference was visible and hard to name.
 */
export const Editing: Story = {
  args: {
    ...base,
    values: metadataPartial,
    onSave: async (values: Record<string, unknown>) => {
      // eslint-disable-next-line no-console
      console.log('[story] save', values)
    },
  },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Edit values')
  },
}

/**
 * Mid-save: the button reads "Saving..." and every control is disabled.
 *
 * Reachable only because saving is a function prop the component awaits. An
 * emit returns undefined, so an awaited emit resolves at once and this state
 * lasts a frame — which is how it was written first, and why this story exists
 * to hold the shape that replaced it.
 */
export const Saving: Story = {
  args: { ...base, values: metadataPartial, onSave: () => neverResolves() },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Edit values')
    await clickButton(canvasElement, 'Save')
  },
}

/**
 * The server refused the write — a value that does not match its declared type,
 * a key the section does not declare.
 *
 * The editor stays open with the typed values still in it. Closing it on a
 * refusal would discard the work and leave the reason on screen with nothing to
 * apply it to.
 */
export const SaveFails: Story = {
  args: {
    ...base,
    values: metadataPartial,
    onSave: () => fails('owner: "" is not one of the declared values'),
  },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickButton(canvasElement, 'Edit values')
    await clickButton(canvasElement, 'Save')
  },
}

/**
 * Long Cyrillic labels and values, which is what this product actually holds.
 *
 * `--metadata-col-max: 16rem` is the only thing keeping the first label from
 * pushing the value column off the page, and the labels here are long enough to
 * reach it. Check this one at the mobile viewport too: below 768px the grid
 * collapses to a single column and the cap stops mattering.
 */
export const LongCyrillic: Story = {
  args: {
    ...base,
    specs: specCyrillic,
    values: metadataCyrillic,
    metadataStatus: { complete: true, spec_version: 2 },
    entrySpecVersion: 2,
    sectionSpecVersion: 2,
  },
}

/**
 * Markup inside a value.
 *
 * Field values are agent-authored text and go to `v-html` through `renderRefs`,
 * in both halves of this component — the declared rows and the orphaned ones.
 * That is the same sink as an entry description, so it gets the same story:
 * `<b>bold</b>` must read as text, and `[[E47]]` beside it must still become a
 * link. An executed payload prints `XSS EXECUTED` where the image tag was.
 *
 * This story is honest here, and would not be everywhere: `renderRefs` reaches
 * this component through the real composable (an explicit
 * `~/composables/useCrossRefs` import, and the auto-import in `.storybook/main.ts`
 * points at the real file too). The identity stub in `.storybook/stubs/imports.ts`
 * is only reachable via `#imports`, which nothing in `components/` uses. If a
 * component ever does, its escaping stories stop meaning anything and that stub
 * should be deleted rather than trusted.
 */
export const MarkupInValues: Story = {
  args: {
    ...base,
    specs: specSpecifications.slice(0, 4),
    values: {
      stage: `<b>draft</b> ${markupImg}`,
      produces: [`scanner-watchdog ${markupImg}`, '<script>alert(1)</script>'],
      owner: 'platform — see [[E47]]',
      // An orphaned value reaches v-html down a second, separately written
      // path. A fix applied to one of the two leaves the other.
      implemented_by: `<b>scanner-orchestrator</b> ${markupImg} — see [[E50]]`,
    },
  },
}

/**
 * Clicks the first button whose label matches, once it exists — the editor's
 * controls only appear after the previous click. There is no `@storybook/test`
 * in this project, so the catalogue polls, as `HistoryPanel.stories.ts` does.
 */
async function clickButton(root: HTMLElement, label: string): Promise<void> {
  for (let i = 0; i < 50; i++) {
    const button = Array.from(root.querySelectorAll('button'))
      .find(b => b.textContent?.trim() === label) as HTMLElement | undefined
    if (button) {
      button.click()
      return
    }
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
}
