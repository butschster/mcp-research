import type { Meta, StoryObj } from '@storybook/vue3'
import TemplateOriginBadge from './TemplateOriginBadge.vue'

/**
 * Ownership, in words. A template's tier is not a precedence — two never apply
 * at once — which is why this is not SkillOriginBadge.
 *
 * There are three origins, not two: the global tier holds both what ships in
 * the binary and what the operator of this instance added through the API, and
 * only the first is rewritten on upgrade. `tier` alone cannot tell them apart,
 * which is what `source` is for.
 */
const meta: Meta<typeof TemplateOriginBadge> = {
  title: 'Templates/TemplateOriginBadge',
  component: TemplateOriginBadge,
}
export default meta
type Story = StoryObj<typeof TemplateOriginBadge>

export const Shipped: Story = { args: { tier: 'global', source: 'builtin' } }

/** Global too, but written on this instance — an upgrade will not touch it. */
export const AddedOnThisServer: Story = { args: { tier: 'global', source: 'user' } }
export const Team: Story = { args: { tier: 'team', source: 'user', teamName: 'Platform' } }
export const TeamFork: Story = { args: { tier: 'team', source: 'user', teamName: 'Platform', forkedFrom: 'literature-review' } }

/** A payload without `source` at all. The badge reads it as shipped, so an
 *  operator-written global that lost the field would claim we wrote it — worth
 *  seeing in the catalogue, since both callers currently do pass it. */
export const SourceMissing: Story = { args: { tier: 'global' } }

/** All four readings together: the only way to see that the three origins are
 *  distinguishable by their words and not merely by their colour. */
export const AllOfThem: Story = {
  render: () => ({
    components: { TemplateOriginBadge },
    template: `
      <div style="display:flex;gap:.5rem;flex-wrap:wrap">
        <TemplateOriginBadge tier="global" source="builtin" />
        <TemplateOriginBadge tier="global" source="user" />
        <TemplateOriginBadge tier="team" source="user" team-name="Platform" />
        <TemplateOriginBadge tier="team" source="user" team-name="Platform" forked-from="literature-review" />
      </div>`,
  }),
}
