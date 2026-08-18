import type { Meta, StoryObj } from '@storybook/vue3'
import CopyLine from './CopyLine.vue'

/**
 * A one-line command with a copy button. Exists because the three hand-rolled
 * copies already in this product announce nothing to a screen reader.
 */
const meta: Meta<typeof CopyLine> = { title: 'Common/CopyLine', component: CopyLine }
export default meta
type Story = StoryObj<typeof CopyLine>

export const Default: Story = {
  args: {
    text: 'Start a new research. Check template_list first and follow the methodology that fits.',
    label: 'Paste this into your AI client',
  },
}

export const NoLabel: Story = { args: { text: 'make run-sse' } }

/** Long enough to wrap rather than push the card sideways. */
export const Long: Story = {
  args: {
    text: 'Start a new research using the Technology comparison methodology (template_get slug: technology-comparison) and follow it rather than proposing a structure first.',
    label: 'Paste this into your AI client',
  },
}
