import type { Meta, StoryObj } from '@storybook/vue3'
import TemplateCriteria from './TemplateCriteria.vue'
import { mockGlobalTemplates, mockTemplateLongCriteria } from '../../__mocks__/template'

/**
 * The pair of matching criteria — what an agent chooses on, and what a reader
 * compares across methodologies. Deliberately not muted: `.data-row` has a hover
 * state, and muted at this size fails contrast against it.
 */
const meta: Meta<typeof TemplateCriteria> = {
  title: 'Templates/TemplateCriteria',
  component: TemplateCriteria,
}
export default meta
type Story = StoryObj<typeof TemplateCriteria>

const tp = mockGlobalTemplates[0]!

export const Default: Story = {
  args: { whenToUse: tp.when_to_use, whenNotToUse: tp.when_not_to_use },
}

/** In a list row: tighter, smaller, same shape. */
export const Dense: Story = {
  args: { whenToUse: tp.when_to_use, whenNotToUse: tp.when_not_to_use, dense: true },
}

/** Only the positive form — legal, and worth less. */
export const NoNegative: Story = { args: { whenToUse: tp.when_to_use } }

/** Saved with nothing to match on: in the list, and unchoosable. */
export const Missing: Story = { args: { whenToUse: '' } }

/** What the service still accepts: two criteria near the 240-rune cap, one of
 *  them carrying an unbroken product name. This is what `overflow-wrap: anywhere`
 *  on the text is for, and the only story that proves the label column holds. */
export const AtTheLimit: Story = {
  args: {
    whenToUse: mockTemplateLongCriteria.when_to_use,
    whenNotToUse: mockTemplateLongCriteria.when_not_to_use,
  },
}

/** The same pair in a row: smaller and tighter, and still two columns. */
export const AtTheLimitDense: Story = {
  args: {
    whenToUse: mockTemplateLongCriteria.when_to_use,
    whenNotToUse: mockTemplateLongCriteria.when_not_to_use,
    dense: true,
  },
}
