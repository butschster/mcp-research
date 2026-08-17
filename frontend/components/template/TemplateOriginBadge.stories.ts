import type { Meta, StoryObj } from '@storybook/vue3'
import TemplateOriginBadge from './TemplateOriginBadge.vue'

/**
 * Ownership, in words. A template's tier is not a precedence — two never apply
 * at once — which is why this is not SkillOriginBadge.
 */
const meta: Meta<typeof TemplateOriginBadge> = {
  title: 'Templates/TemplateOriginBadge',
  component: TemplateOriginBadge,
}
export default meta
type Story = StoryObj<typeof TemplateOriginBadge>

export const Global: Story = { args: { tier: 'global' } }
export const Team: Story = { args: { tier: 'team', teamName: 'Platform' } }
export const TeamFork: Story = { args: { tier: 'team', teamName: 'Platform', forkedFrom: 'literature-review' } }
