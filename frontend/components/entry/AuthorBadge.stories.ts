import type { Meta, StoryObj } from '@storybook/vue3'
import AuthorBadge from './AuthorBadge.vue'

/**
 * Who wrote a revision. Only `human` and `restore` carry colour — in a product
 * where an agent writes nearly everything, those are the two rows worth
 * spotting, and colouring all four would make none of them stand out.
 */
const meta: Meta<typeof AuthorBadge> = {
  title: 'Entry/AuthorBadge',
  component: AuthorBadge,
  tags: ['autodocs'],
  argTypes: {
    kind: { control: 'select', options: ['agent', 'human', 'import', 'restore'] },
    variant: { control: 'inline-radio', options: ['inline', 'glyph'] },
  },
}
export default meta
type Story = StoryObj<typeof AuthorBadge>

export const Agent: Story = { args: { kind: 'agent' } }

/** The exception worth spotting: a person edited this. */
export const Human: Story = { args: { kind: 'human' } }

export const Import: Story = { args: { kind: 'import' } }

/** An event rather than an author — it explains a paragraph coming back. */
export const Restore: Story = { args: { kind: 'restore' } }

/** All four together, as they appear down a revision rail. */
export const AllKinds: Story = {
  render: () => ({
    components: { AuthorBadge },
    template: `
      <div style="display:flex;flex-direction:column;gap:var(--space-2);font-size:var(--type-xs);">
        <AuthorBadge kind="agent" />
        <AuthorBadge kind="human" />
        <AuthorBadge kind="restore" />
        <AuthorBadge kind="import" />
      </div>
    `,
  }),
}

/** Glyph only, for a row too tight for the word. */
export const GlyphOnly: Story = {
  render: () => ({
    components: { AuthorBadge },
    template: `
      <div style="display:flex;gap:var(--space-3);font-size:var(--type-sm);">
        <AuthorBadge kind="agent" variant="glyph" />
        <AuthorBadge kind="human" variant="glyph" />
        <AuthorBadge kind="restore" variant="glyph" />
        <AuthorBadge kind="import" variant="glyph" />
      </div>
    `,
  }),
}

/** An unknown kind falls back to the agent glyph and prints what it was given. */
export const UnknownKind: Story = { args: { kind: 'migration' } }
