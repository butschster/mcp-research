import type { Meta, StoryObj } from '@storybook/vue3'
import SkillDetail from './SkillDetail.vue'
import { mockSkillWithBody, mockMigratedSkill } from '../../../__mocks__/skill'

/**
 * Reading one skill. The trigger line is set apart from the body because it is
 * the only part always in the agent's context — everything below it is read on
 * demand.
 */
const meta: Meta<typeof SkillDetail> = {
  title: 'Research/Settings/SkillDetail',
  component: SkillDetail,
}
export default meta
type Story = StoryObj<typeof SkillDetail>

export const Default: Story = { args: { visible: true, skill: mockSkillWithBody } }

export const Loading: Story = { args: { visible: true, skill: mockSkillWithBody, loading: true } }

/** The attachment is unaffected; only this page failed to fetch the text. */
export const Failed: Story = { args: { visible: true, skill: mockSkillWithBody, error: true } }

/** A migrated instruction: long, unstructured, and never reviewed by anyone. */
export const MigratedInstruction: Story = {
  args: {
    visible: true,
    skill: {
      ...mockMigratedSkill,
      body: 'The agent should keep entries short and always cite a source for cost claims and never mark a section complete while an unknown is open and prefer measured evidence over vendor documentation and ask before proposing a structure.',
    },
  },
}
