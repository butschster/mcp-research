import type { Meta, StoryObj } from '@storybook/vue3'
import KanbanCard from './KanbanCard.vue'
import { markupTaskTitle } from '../../__mocks__/markup'
import { mockTask, mockTaskHigh } from '../../__mocks__/task'

const meta: Meta<typeof KanbanCard> = {
  title: 'Tasks/KanbanCard',
  component: KanbanCard,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 280px"><story /></div>',
    }),
  ],
  argTypes: {
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof KanbanCard>

export const Default: Story = {
  args: {
    task: mockTask,
    researchSlug: 'R1',
  },
}

export const HighPriority: Story = {
  args: {
    task: mockTaskHigh,
    researchSlug: 'R1',
  },
}

/**
 * A title with markup in it.
 *
 * The card's title is the whole of its content and it goes to `v-html` through
 * `renderRefs`. A task title is written by an agent from whatever it was
 * reading, so `<script>` in one is a Tuesday, not an attack — which is exactly
 * why it has to render as the text it is, with `[[E3]]` still linking. An
 * executed payload prints `XSS EXECUTED` in the middle of the card.
 */
export const MarkupInTitle: Story = {
  args: {
    task: { ...mockTask, code: 'T7', title: markupTaskTitle },
    researchSlug: 'R1',
  },
}

export const LongTitle: Story = {
  args: {
    task: {
      ...mockTask,
      code: 'T99',
      title: 'Investigate the long-term implications of migrating from Options API to Composition API across all enterprise-level Vue components in the monorepo',
    },
    researchSlug: 'R1',
  },
}
