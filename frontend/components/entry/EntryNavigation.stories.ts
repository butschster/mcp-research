import type { Meta, StoryObj } from '@storybook/vue3'
import EntryNavigation from './EntryNavigation.vue'
import { withShare, withoutShare } from '../../__mocks__/share'

/**
 * Previous and next entry in the same section.
 *
 * Both destinations come from `entryPath()`, so the pair works unchanged inside
 * a share link. Siblings are always within the shared research — they are the
 * section's own list — so there is no inert case here.
 */
const meta: Meta<typeof EntryNavigation> = {
  title: 'Entry/EntryNavigation',
  component: EntryNavigation,
  tags: ['autodocs'],
  // Share state is module state; this gives the ordinary stories a known
  // starting point rather than whatever the last story left behind. The
  // trade-offs are in __mocks__/share.ts.
  decorators: [withoutShare()],
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

/** Inside a share link: both buttons point at `/s/{token}/entry/…`. Reading
 *  through a section end to end is most of what a shared view is for, so this is
 *  the pair that matters most to get right. */
export const InsideAShare: Story = {
  decorators: [withShare()],
  args: {
    prev: { id: 'ent_701', code: 'E1', title: 'What counts as a seat' },
    next: { id: 'ent_704', code: 'E4', title: 'Where we sit' },
    researchSlug: 'R7',
  },
}
