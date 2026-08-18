import type { Meta, StoryObj } from '@storybook/vue3'
import ProvenanceMenuHeader from './ProvenanceMenuHeader.vue'

/**
 * Who wrote this document and when, at the top of its actions panel.
 *
 * It was a button first — the whole block clickable, so a list of verbs would
 * not hold a dead row. That produced the opposite complaint: a link-shaped
 * `View history →` sitting among buttons, and three lines of near-identical
 * text that read as pasted in. So the fact is now only a fact, and it carries
 * no hover and no pointer at all: inertness has to be provable by the thing a
 * person tests it with.
 *
 * The tile is recessed rather than lightened because hover in this product goes
 * lighter, and a lighter tile reads as already-hovered.
 */
const meta: Meta<typeof ProvenanceMenuHeader> = {
  title: 'Entry/ProvenanceMenuHeader',
  component: ProvenanceMenuHeader,
  tags: ['autodocs'],
  decorators: [
    // Inside the panel it lives in: the tile's inset, its recess and the icon
    // column it shares with the verbs are all ActionMenu's, so rendered bare it
    // would show a version of itself that exists nowhere.
    () => ({
      template: `
        <div class="action-menu-list action-menu-list--wide action-menu-list--right"
             style="position: static">
          <story />
          <button class="action-menu-item">Revision history</button>
          <button class="action-menu-item">Download .md</button>
        </div>`,
    }),
  ],
}
export default meta
type Story = StoryObj<typeof ProvenanceMenuHeader>

const hoursAgo = (n: number) => new Date(Date.now() - n * 3600_000).toISOString()

/** The common case in this product: a model wrote it, so there is no name. */
export const WrittenByAnAgent: Story = {
  args: { revision: 7, authorKind: 'agent', revisedAt: hoursAgo(2) },
}

/** A person, named — the half of the answer the old strip could never fit. */
export const WrittenByAPerson: Story = {
  args: { revision: 12, authorKind: 'human', revisedAt: hoursAgo(0.3), authorName: 'Павел Бучнев' },
}

/** With `auth_enabled: false` there are no users, so the kind stands alone. */
export const NoName: Story = {
  args: { revision: 3, authorKind: 'human', revisedAt: hoursAgo(26) },
}

/**
 * No revision yet.
 *
 * The block still renders, from the entry's own timestamp: a document nobody
 * has revised still has an author kind and a time, and one menu that degrades
 * is better than two menus that differ.
 */
export const NoRevisionYet: Story = {
  args: { authorKind: 'agent', revisedAt: hoursAgo(0.1) },
}

/** A restore is its own kind of authorship and says so. */
export const Restored: Story = {
  args: { revision: 9, authorKind: 'restore', revisedAt: hoursAgo(0.05) },
}

/** An import carries no person at all. */
export const Imported: Story = {
  args: { revision: 1, authorKind: 'import', revisedAt: hoursAgo(700) },
}

/** Three digits, which is where the tabular figures earn their keep. */
export const ManyRevisions: Story = {
  args: { revision: 342, authorKind: 'agent', revisedAt: hoursAgo(1) },
}

/**
 * A long Russian name against the panel's 232px. The width was chosen so a full
 * name survives; this is the story that proves or disproves it.
 */
export const LongName: Story = {
  args: {
    revision: 148,
    authorKind: 'human',
    revisedAt: hoursAgo(5),
    authorName: 'Александра Константинопольская-Верещагина',
  },
}
