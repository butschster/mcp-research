import type { Meta, StoryObj } from '@storybook/vue3'
import SkillRowList from './SkillRowList.vue'
import {
  mockAmbientSkills,
  mockChosenSkills,
  mockLibrarySkills,
  mockMigratedSkill,
  mockSkillWithoutTrigger,
} from '../../../__mocks__/skill'

/**
 * The list a human reads to answer "what is the agent following, and where did
 * each of these come from". The origin badge carries the tier, which is also
 * the precedence order — so it is text, never colour alone.
 */
const meta: Meta<typeof SkillRowList> = {
  title: 'Research/Settings/SkillRowList',
  component: SkillRowList,
}
export default meta
type Story = StoryObj<typeof SkillRowList>

export const Chosen: Story = {
  args: {
    skills: mockChosenSkills,
    canWrite: true,
    heading: 'Chosen for this research',
    note: '3 of 6',
    blurb: 'Methodology somebody picked. These are what the six-skill budget counts.',
  },
}

/** Always-on product skills: readable, and no detach control to press. */
export const AlwaysOn: Story = {
  args: {
    skills: mockAmbientSkills,
    canWrite: true,
    heading: 'Always on',
    blurb: 'Skills about the product itself. They ship with the app and are outside the budget.',
  },
}

/** A viewer sees everything and can change nothing — controls are absent, not disabled. */
export const ReadOnly: Story = {
  args: { skills: mockChosenSkills, canWrite: false, heading: 'Chosen for this research' },
}

/** The library: attach buttons instead of detach, and rows already on are inert. */
export const Library: Story = {
  args: {
    skills: [...mockLibrarySkills, { ...mockLibrarySkills[0]!, id: 'l3', slug: 'already', name: 'Already attached', attached: true }],
    actions: false,
    canWrite: true,
    heading: 'Available to attach',
  },
}

/** What the instruction migration produces: a description a machine wrote. */
export const NeedsATrigger: Story = {
  args: { skills: [mockMigratedSkill], canWrite: true, heading: 'Chosen for this research' },
}

/** No trigger line at all — the row says so, because the agent will never open it. */
export const NoTriggerLine: Story = {
  args: { skills: [mockSkillWithoutTrigger], canWrite: true },
}

export const Busy: Story = {
  args: { skills: mockChosenSkills, canWrite: true, busySlug: 'evidence-grading' },
}

export const Empty: Story = {
  args: {
    skills: [],
    canWrite: true,
    heading: 'Chosen for this research',
    emptyText: 'Nothing chosen yet — the agent works from the built-in skills alone.',
  },
}
