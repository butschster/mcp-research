import type { Meta, StoryObj } from '@storybook/vue3'
import ActionMenu from '../ActionMenu.vue'
import ProvenanceMenuHeader from './ProvenanceMenuHeader.vue'

const clockIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v5h5"/><path d="M3.05 13A9 9 0 1 0 6 5.3L3 8"/><path d="M12 7v5l4 2"/></svg>`
const downloadIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>`
const trashIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>`

// The panel is absolutely positioned; the story needs room under the trigger.
const roomBelow = (story: string) => `<div style="min-height: 260px;">${story}</div>`

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
 *
 * Every story renders it inside a real, open `ActionMenu` with the verbs that
 * ship beneath it. Nothing about this block is its own: the inset, the recess,
 * the raised surface behind it and the `--menu-icon` column its glyph sits in
 * are all `ActionMenu`'s, declared in that component's scoped stylesheet. A
 * hand-written `<div class="action-menu-list">` in a decorator carries none of
 * the scope hashes and would show a version of this component that exists
 * nowhere.
 */
const meta: Meta<typeof ProvenanceMenuHeader> = {
  title: 'Entry/ProvenanceMenuHeader',
  component: ProvenanceMenuHeader,
  tags: ['autodocs'],
  parameters: {
    // The panel opens from a play function, which does not run on the docs page
    // unless asked. Without this, autodocs is a column of closed ⋯ buttons.
    docs: { story: { autoplay: true } },
  },
  render: (args) => ({
    components: { ActionMenu, ProvenanceMenuHeader },
    setup: () => ({ args }),
    template: roomBelow(`
      <ActionMenu title="Document actions" width="wide">
        <ProvenanceMenuHeader v-bind="args" />
        <button class="action-menu-item">${clockIcon} Revision history</button>
        <button class="action-menu-item">${downloadIcon} Download .md</button>
      </ActionMenu>
    `),
  }),
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await openMenu(canvasElement)
  },
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
 * is better than two menus that differ. The `r7 ·` pair disappears together —
 * a lone separator would be the visible bug here.
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
 * name survives; this is the story that proves or disproves it. The name
 * ellipsises on one line rather than wrapping — the tile is a fact, and a fact
 * that grows to three lines pushes the verbs off the reader's first glance.
 */
export const LongName: Story = {
  args: {
    revision: 148,
    authorKind: 'human',
    revisedAt: hoursAgo(5),
    authorName: 'Александра Константинопольская-Верещагина',
  },
}

/**
 * The document menu as it ships to someone who can write.
 *
 * The whole panel, in the order the entry page renders it: the fact, the two
 * reads, then a rule and the one destructive verb. This is the arrangement
 * being judged — the tile only makes sense as the thing the verbs are indented
 * to, and the rule only reads as a rule when there is something above and below
 * it.
 */
export const TheDocumentMenu: Story = {
  args: { revision: 7, authorKind: 'agent', revisedAt: hoursAgo(2) },
  render: (args) => ({
    components: { ActionMenu, ProvenanceMenuHeader },
    setup: () => ({ args }),
    template: roomBelow(`
      <ActionMenu title="Document actions" width="wide">
        <ProvenanceMenuHeader v-bind="args" />
        <button class="action-menu-item">${clockIcon} Revision history</button>
        <button class="action-menu-item">${downloadIcon} Download .md</button>
        <div class="action-menu-divider" role="separator"></div>
        <button class="action-menu-item action-menu-item--danger">${trashIcon} Delete document</button>
      </ActionMenu>
    `),
  }),
}

/**
 * The same menu for a viewer, or behind a share link: no divider, no Delete.
 *
 * Two items and a fact is the case that tests whether the panel is worth
 * having at all. It is: both verbs were icon-only buttons in the header row
 * before, and the fact they now sit under was a strip of grey text under the
 * title that no one read.
 */
export const TheViewersMenu: Story = {
  args: { revision: 24, authorKind: 'human', revisedAt: hoursAgo(9), authorName: 'Павел Бучнев' },
}

/**
 * Clicks the nth `⋯` once it is mounted.
 *
 * There is no `@storybook/test` in this project, so the catalogue polls — same
 * helper shape as `HistoryPanel.stories.ts`. The trigger carries no text, so it
 * is found by `aria-haspopup` rather than by label.
 */
async function openMenu(root: HTMLElement, index = 0): Promise<void> {
  for (let i = 0; i < 50; i++) {
    const triggers = Array.from(
      root.querySelectorAll<HTMLElement>('button[aria-haspopup="true"]')
    )
    const trigger = triggers[index]
    if (trigger) {
      trigger.click()
      return
    }
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
  throw new Error(`ActionMenu trigger #${index} never appeared`)
}
