import type { Meta, StoryObj } from '@storybook/vue3'
import SkillOriginBadge from './SkillOriginBadge.vue'

/**
 * Where a skill comes from — and, because the tiers are also the precedence
 * order, which of two skills wins. Text, never colour alone; the title says
 * what the tier means rather than repeating its name.
 */
const meta: Meta<typeof SkillOriginBadge> = {
  title: 'Research/Settings/SkillOriginBadge',
  component: SkillOriginBadge,
}
export default meta
type Story = StoryObj<typeof SkillOriginBadge>

export const Private: Story = { args: { tier: 'private' } }
export const Team: Story = { args: { tier: 'team' } }
export const TeamFork: Story = { args: { tier: 'team', forkedFrom: 'evidence-grading' } }
export const Builtin: Story = { args: { tier: 'builtin' } }

/** A product skill reads as built in whatever tier its row is on. */
export const ProductSkill: Story = { args: { tier: 'builtin', ambient: true } }

export const AllOfThem: Story = {
  render: () => ({
    components: { SkillOriginBadge },
    template: `
      <div style="display:flex;gap:.5rem;flex-wrap:wrap">
        <SkillOriginBadge tier="private" />
        <SkillOriginBadge tier="team" />
        <SkillOriginBadge tier="team" forked-from="evidence-grading" />
        <SkillOriginBadge tier="builtin" />
        <SkillOriginBadge tier="builtin" :ambient="true" />
      </div>`,
  }),
}
