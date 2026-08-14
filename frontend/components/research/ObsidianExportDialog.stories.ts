import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import type { ObsidianExportOptions } from '~/composables/useObsidianExport'
import ObsidianExportDialog from './ObsidianExportDialog.vue'

/**
 * What goes into the vault.
 *
 * Three of the five options only ever remove content, one adds interactive HTML
 * files and one adds the revision history. Every opening starts from the
 * defaults — a remembered setting that silently drops the Sessions folder is a
 * worse bug than ticking a box again.
 */
const meta: Meta<typeof ObsidianExportDialog> = {
  title: 'Research/ObsidianExportDialog',
  component: ObsidianExportDialog,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component:
          'A `ModalOverlay size="sm"` titled through `labelledby`, so it announces ' +
          'as "Obsidian export options" rather than as an unnamed dialog. `counts` ' +
          'is per-option: a zero count leaves the row in place but disables it and ' +
          'reads "none", so the option set keeps the same shape from one research ' +
          'to the next. `download` carries a fresh copy of the toggles; `close` ' +
          'discards them.',
      },
    },
  },
  argTypes: {
    visible: { control: 'boolean' },
    counts: { control: 'object' },
  },
}
export default meta
type Story = StoryObj<typeof ObsidianExportDialog>

/** A research with everything in it: three counted options, two uncounted. */
export const Default: Story = {
  args: { visible: true, counts: { sessions: 4, tasks: 12, roadmaps: 2 } },
}

/** Sessions but no tasks and no roadmaps — the shape the export page actually
 *  reaches when a research is part-filled. The empty rows stay visible, read
 *  "none" and cannot be ticked. (All three at zero is not reachable: the export
 *  page hides the caret entirely, so the dialog never opens.) */
export const SomeOptionsEmpty: Story = {
  args: { visible: true, counts: { sessions: 3, tasks: 0, roadmaps: 0 } },
}

/** Counts unknown — the rows render without them rather than guessing, and stay
 *  tickable. This is the shape when `counts` is omitted altogether. */
export const WithoutCounts: Story = {
  args: { visible: true },
}

/** A long-running research: four-digit counts, to check the right-aligned
 *  tabular figures do not push the labels around. */
export const ManyItems: Story = {
  args: { visible: true, counts: { sessions: 37, tasks: 1284, roadmaps: 9 } },
}

/** The no-memory rule, demonstrated: untick something, download or cancel, then
 *  reopen — the toggles are back at the defaults, and the emitted options are
 *  whatever was on screen at the moment Download was pressed. */
export const ReopensWithDefaults: Story = {
  render: () => ({
    components: { ObsidianExportDialog },
    setup() {
      const visible = ref(false)
      const last = ref<ObsidianExportOptions | null>(null)
      function onDownload(opts: ObsidianExportOptions) {
        last.value = opts
        visible.value = false
      }
      return { visible, last, onDownload }
    },
    template: `
      <div style="padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-3); align-items: flex-start;">
        <button class="btn btn-sm" @click="visible = true">Obsidian export options</button>
        <p style="margin: 0; font-size: var(--type-xs); color: var(--color-text-muted);">
          Last downloaded with: <code>{{ last ? JSON.stringify(last) : '—' }}</code>
        </p>
        <ObsidianExportDialog
          :visible="visible"
          :counts="{ sessions: 4, tasks: 12, roadmaps: 2 }"
          @close="visible = false"
          @download="onDownload"
        />
      </div>
    `,
  }),
}
