import type { Meta, StoryObj } from '@storybook/vue3'
import ExportSplitButton from './ExportSplitButton.vue'

const vaultIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/></svg>`

/**
 * A download button with its settings folded into a second half.
 *
 * The wide half always does the default thing; the caret always opens the
 * options. Keeping the two paths distinct is what stops a remembered setting
 * from silently changing what the wide half produces.
 */
const meta: Meta<typeof ExportSplitButton> = {
  title: 'Export/SplitButton',
  component: ExportSplitButton,
  tags: ['autodocs'],
  argTypes: {
    label: { control: 'text' },
    busy: { control: 'boolean' },
    busyLabel: { control: 'text' },
    optionsLabel: { control: 'text' },
    withOptions: { control: 'boolean' },
    optionsOpen: { control: 'boolean' },
    disabled: { control: 'boolean' },
    describedby: { control: 'text' },
  },
  parameters: {
    docs: {
      description: {
        component:
          'Two buttons that read as one control: the wide half emits `download`, ' +
          'the caret emits `options`. Both halves are separate tab stops, and both ' +
          'go inert via `aria-disabled` (never `disabled`) while `busy` or ' +
          '`disabled`, so focus stays where the user left it. The icon comes from ' +
          'the `icon` slot; the busy spinner replaces it.',
      },
    },
  },
  args: {
    label: 'Obsidian .zip',
    optionsLabel: 'Obsidian export options',
  },
  render: (args) => ({
    components: { ExportSplitButton },
    setup: () => ({ args }),
    template: `
      <div style="padding: var(--space-4);">
        <ExportSplitButton v-bind="args">
          <template #icon>${vaultIcon}</template>
        </ExportSplitButton>
      </div>
    `,
  }),
}
export default meta
type Story = StoryObj<typeof ExportSplitButton>

export const Idle: Story = {}

/** Busy: the label says what is happening, the icon becomes a pulsing dot, and
 *  focus stays on the button — `aria-disabled`, never `disabled`. The caret goes
 *  inert with it, so options cannot be changed mid-download. */
export const Busy: Story = {
  args: { busy: true, busyLabel: 'Preparing…' },
}

/** The dialog is open, which the caret announces via `aria-expanded`. */
export const OptionsOpen: Story = {
  args: { optionsOpen: true },
}

/** Nothing worth configuring on this research: the caret disappears rather
 *  than opening a dialog of dead toggles. The remaining half keeps both
 *  rounded corners. */
export const WithoutOptions: Story = {
  args: { withOptions: false },
}

/** Inert for a reason other than a download in flight. Note that this state is
 *  currently indistinguishable from Idle — `disabled` sets `aria-disabled` and
 *  swallows the click, but paints nothing. */
export const Disabled: Story = {
  args: { disabled: true },
}

/** A long label, to check the seam holds when the halves are uneven. */
export const LongLabel: Story = {
  args: { label: 'Download as an Obsidian vault (.zip)' },
}

/** No `icon` slot: the 14px icon box still holds its space, so a set of split
 *  buttons with and without icons still aligns. */
export const WithoutIcon: Story = {
  render: (args: any) => ({
    components: { ExportSplitButton },
    setup: () => ({ args }),
    template: `
      <div style="padding: var(--space-4);">
        <ExportSplitButton v-bind="args" />
      </div>
    `,
  }),
}

/** How the export page wires it: `describedby` points at the one line beneath
 *  the toolbar that explains the format at rest and narrates the download, so a
 *  screen reader reads the button and its note together. */
export const InExportToolbar: Story = {
  args: { describedby: 'obsidian-note' },
  render: (args: any) => ({
    components: { ExportSplitButton },
    setup: () => ({ args }),
    template: `
      <div style="padding: var(--space-4); display: flex; flex-direction: column; align-items: flex-end; gap: var(--space-2);">
        <div style="display: flex; align-items: center; gap: var(--space-2);">
          <button class="btn btn-sm">Download .md</button>
          <ExportSplitButton v-bind="args">
            <template #icon>${vaultIcon}</template>
          </ExportSplitButton>
        </div>
        <p id="obsidian-note" style="margin: 0; max-width: 42ch; text-align: right; font-size: var(--type-xs); color: var(--color-text-muted); line-height: 1.4;">
          A folder per section, a note per entry, and every [[E3]] link still working —
          opens in Obsidian, or any editor that reads markdown.
        </p>
      </div>
    `,
  }),
}

/** Every state side by side, which is where the seam between the halves either
 *  holds or does not. */
export const AllStates: Story = {
  render: () => ({
    components: { ExportSplitButton },
    setup: () => ({ vaultIcon }),
    template: `
      <div style="display: flex; flex-direction: column; gap: var(--space-4); padding: var(--space-4);">
        <div v-for="row in [
          { caption: 'Idle', props: {} },
          { caption: 'Busy', props: { busy: true } },
          { caption: 'Options open', props: { optionsOpen: true } },
          { caption: 'No options (caret hidden)', props: { withOptions: false } },
          { caption: 'Disabled', props: { disabled: true } },
          { caption: 'Long label', props: { label: 'Download as an Obsidian vault (.zip)' } },
        ]" :key="row.caption" style="display: flex; align-items: center; gap: var(--space-4);">
          <span style="min-width: 20ch; font-size: var(--type-xs); color: var(--color-text-muted);">{{ row.caption }}</span>
          <ExportSplitButton label="Obsidian .zip" options-label="Obsidian export options" v-bind="row.props">
            <template #icon><span v-html="vaultIcon" style="display: inline-flex;" /></template>
          </ExportSplitButton>
        </div>
      </div>
    `,
  }),
}
