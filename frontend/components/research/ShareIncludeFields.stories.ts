import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import ShareIncludeFields from './ShareIncludeFields.vue'
import type { ShareInclude } from '~/composables/useShare'
import { includeDefault, includeEntriesOnly, includeEverything } from '../../__mocks__/share'

/**
 * The four include flags of a share link, as checkboxes.
 *
 * One component for the create form and the edit form. Two hand-maintained
 * copies of four checkboxes is how one of them ends up describing a flag the
 * other has renamed — and these words are a promise to the owner about what a
 * stranger will see, so they are the last words in the product that should
 * drift.
 *
 * **Documents are not a checkbox.** Sections, documents and cross-references are
 * what a share link is; the legend says so as a fact above the boxes rather than
 * as a ticked and disabled row, which would invite the reader to try to untick
 * it and then wonder what is wrong.
 *
 * The default has roadmaps and downloads on, sessions and tasks off. Interview
 * transcripts and an internal todo list are the two things an owner is most
 * likely to hand over by accident, so they start closed.
 */
const meta: Meta<typeof ShareIncludeFields> = {
  title: 'Research/ShareIncludeFields',
  component: ShareIncludeFields,
  tags: ['autodocs'],
  decorators: [() => ({ template: '<div style="max-width: 480px;"><story /></div>' })],
  argTypes: {
    modelValue: { control: 'object' },
    'onUpdate:modelValue': { action: 'update:modelValue' },
  },
}
export default meta
type Story = StoryObj<typeof ShareIncludeFields>

/** What the create form proposes: roadmaps and downloading, nothing personal. */
export const Default: Story = {
  args: { modelValue: includeDefault },
}

/** Everything an owner can include. The link still shows less than the owner's
 *  own view — instructions, memory, provenance and revision history are never
 *  shared and have no flag here. */
export const EverythingOn: Story = {
  args: { modelValue: includeEverything },
}

/**
 * All four off — the narrowest link there is, and still not an empty one: it
 * carries the sections, the documents and the references between them.
 */
export const EntriesOnly: Story = {
  args: { modelValue: includeEntriesOnly },
}

/** Sessions on, everything else off. The longest label, on its own, is where the
 *  row height and the wrap show. */
export const SessionsOnly: Story = {
  args: { modelValue: { sessions: true, tasks: false, roadmaps: false, export: false } },
}

/**
 * Wired to a `v-model`, which is the only way to see that it emits a whole new
 * object rather than mutating the one it was given.
 *
 * That matters upstream: the edit form compares what is in the boxes against
 * what the link already carries to work out which flags are *widening*, and a
 * mutated prop would make the two identical and the warning never appear.
 */
export const Interactive: Story = {
  render: () => ({
    components: { ShareIncludeFields },
    setup() {
      const include = ref<ShareInclude>({ ...includeDefault })
      return { include }
    },
    template: `
      <div style="display: flex; flex-direction: column; gap: var(--space-3);">
        <ShareIncludeFields v-model="include" />
        <p style="margin: 0; font-size: var(--type-xs); color: var(--color-text-muted); overflow-wrap: anywhere;">
          <code>{{ JSON.stringify(include) }}</code>
        </p>
      </div>
    `,
  }),
}

/** ≤375px: four rows of one line each, the longest wrapping under its box
 *  rather than beside it. Each row keeps a `--control-h-sm` target. */
export const Mobile: Story = {
  parameters: { viewport: { defaultViewport: 'mobile' } },
  args: { modelValue: includeEverything },
}
