import type { Meta, StoryObj } from '@storybook/vue3'
import RevisionRow from './RevisionRow.vue'

/**
 * One row of the revision rail. Two lines: the summary leads because it is what
 * a reader scans for, author and time recede, and the number sits in a
 * fixed-width spine on the right so six rows read as a column rather than a
 * wall.
 *
 * Rendered inside a `<ul>` at rail width, which is where it lives.
 */
const meta: Meta<typeof RevisionRow> = {
  title: 'Entry/RevisionRow',
  component: RevisionRow,
  tags: ['autodocs'],
  render: (args: any) => ({
    components: { RevisionRow },
    setup: () => ({ args }),
    template: `
      <ul style="list-style:none;margin:0;padding:0;width:300px;border:1px solid var(--color-border);border-radius:var(--radius-sm);">
        <RevisionRow v-bind="args" />
      </ul>
    `,
  }),
}
export default meta
type Story = StoryObj<typeof RevisionRow>

const base = {
  revision: 4,
  summary: 'Updated content',
  author_kind: 'agent',
  created_at: new Date(Date.now() - 3 * 3600_000).toISOString(),
  session_code: 'SS1',
}

export const Agent: Story = { args: { revision: base, selected: false } }

/** Selected: the primary-coloured spine, matching `.sidebar-item.active`. */
export const Selected: Story = { args: { revision: base, selected: true } }

/** A person's edit — the only warm row in a grey column. */
export const Human: Story = {
  args: { revision: { ...base, revision: 5, author_kind: 'human', summary: 'Updated content, tags' }, selected: false },
}

/** A restore explains a paragraph that came back. */
export const Restore: Story = {
  args: { revision: { ...base, revision: 6, author_kind: 'restore', summary: 'Restored revision 2' }, selected: false },
}

/** The newest revision: no restore affordance, marked as current. */
export const Current: Story = {
  args: { revision: { ...base, revision: 6 }, selected: true, isCurrent: true },
}

/** The far end of a shift-click comparison — dashed spine. */
export const ComparisonBase: Story = {
  args: { revision: { ...base, revision: 2 }, selected: false, isBase: true },
}

/** Revision 1 has no summary in the payload; the row says what it was. */
export const FirstRevision: Story = {
  args: {
    revision: { ...base, revision: 1, summary: '', created_at: new Date(Date.now() - 40 * 86400_000).toISOString() },
    selected: false,
  },
}

/** A long summary clamps at two lines and keeps the full text in a tooltip. */
export const LongSummary: Story = {
  args: {
    revision: { ...base, summary: 'Updated content, description, status, tags and entry type in one write' },
    selected: false,
  },
}

/** Cyrillic at rail width. */
export const Cyrillic: Story = {
  args: {
    revision: { ...base, author_kind: 'human', summary: 'Уточнены формулировки в двух абзацах' },
    selected: false,
  },
}
