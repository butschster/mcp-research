import type { Meta, StoryObj } from '@storybook/vue3'
import TeamChip from './TeamChip.vue'

/**
 * Names the team a research belongs to.
 *
 * One neutral chip rather than a tag: tags take their colour from a hash of
 * their text, so running team names through that machinery would paint them six
 * arbitrary colours and imply a taxonomy that does not exist.
 *
 * It only ever appears on a shared team. A card that says "Personal" on every
 * row is noise, so the caller — `ResearchCard`, the research header — decides,
 * and the chip renders whatever name it is handed.
 */
const meta: Meta<typeof TeamChip> = {
  title: 'Team/TeamChip',
  component: TeamChip,
  tags: ['autodocs'],
  argTypes: {
    name: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof TeamChip>

export const Default: Story = {
  args: { name: 'Product Research' },
}

/** The normal case for this product: a Russian department name. */
export const Cyrillic: Story = {
  args: { name: 'Отдел интеграций' },
}

/** A team named after the thing it works on rather than after itself. The chip
 *  wraps inside the card instead of pushing the timestamp off the row. */
export const LongName: Story = {
  args: { name: 'Аналитика рынка и конкурентная разведка (второй квартал)' },
  render: (args: any) => ({
    components: { TeamChip },
    setup: () => ({ args }),
    template: `
      <div style="width: 320px; border: 1px solid var(--color-border); border-radius: var(--radius-md); padding: var(--space-4);">
        <TeamChip v-bind="args" />
      </div>
    `,
  }),
}

/** Where it actually renders: the footer of a research card, ahead of the tags
 *  and of the read-only badge, with the timestamp holding the right edge. */
export const InCardFooter: Story = {
  render: () => ({
    components: { TeamChip },
    template: `
      <div style="width: 360px; border: 1px solid var(--color-border); border-radius: var(--radius-md); padding: var(--space-4); display: flex; flex-direction: column; gap: var(--space-3);">
        <div style="display:flex;align-items:center;gap:var(--space-2);">
          <span class="short-code">R7</span>
          <h3 class="card-title" style="margin:0;">Ценообразование в SaaS</h3>
        </div>
        <div style="display:flex;justify-content:space-between;align-items:center;gap:var(--space-2);">
          <div style="display:flex;gap:var(--space-2);flex-wrap:wrap;">
            <TeamChip name="Аналитика рынка" />
            <span class="readonly-badge">Read-only</span>
            <span class="tag tag-hue-2">pricing</span>
          </div>
          <span class="card-meta" style="white-space:nowrap;">2h ago</span>
        </div>
      </div>
    `,
  }),
}

/** Several at once, as the research list shows them — the point of the neutral
 *  treatment is that four teams read as four teams, not as four categories. */
export const AcrossTeams: Story = {
  render: () => ({
    components: { TeamChip },
    template: `
      <div style="display: flex; flex-wrap: wrap; gap: var(--space-2);">
        <TeamChip name="Product Research" />
        <TeamChip name="Growth" />
        <TeamChip name="Аналитика рынка" />
        <TeamChip name="Отдел интеграций" />
        <TeamChip name="Legal &amp; Compliance" />
      </div>
    `,
  }),
}
