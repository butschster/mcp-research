import type { Meta, StoryObj } from '@storybook/vue3'
import EntryNavigation from './EntryNavigation.vue'

const meta: Meta<typeof EntryNavigation> = {
  title: 'Entry/EntryNavigation',
  component: EntryNavigation,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof EntryNavigation>

export const BothPrevAndNext: Story = {
  args: {
    prev: { id: 'ent_001', code: 'E1', title: 'Component Composition Patterns' },
    next: { id: 'ent_003', code: 'E3', title: 'Template Syntax Deep Dive' },
    researchSlug: 'R1',
  },
}

export const OnlyPrev: Story = {
  args: {
    prev: { id: 'ent_001', code: 'E1', title: 'Component Composition Patterns' },
    next: undefined,
    researchSlug: 'R1',
  },
}

export const OnlyNext: Story = {
  args: {
    prev: undefined,
    next: { id: 'ent_003', code: 'E3', title: 'Template Syntax Deep Dive' },
    researchSlug: 'R1',
  },
}

export const Neither: Story = {
  args: {
    prev: undefined,
    next: undefined,
    researchSlug: 'R1',
  },
}
