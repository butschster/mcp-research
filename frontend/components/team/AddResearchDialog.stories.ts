import type { Meta, StoryObj } from '@storybook/vue3'
import AddResearchDialog from './AddResearchDialog.vue'

/**
 * Pulls researches into a team.
 *
 * `research/TransferModal` asks the same question from the other end: standing
 * on one research, where should it go. This one stands on a team and asks what
 * belongs in it — the direction a brand-new team needs, and the one the product
 * had no control for at all. An owner who created a team and invited a
 * colleague had to visit each research in turn and find "Move to team…" in its
 * `⋯` menu.
 *
 * It sends one request for the whole selection, against a bulk endpoint added
 * with it. Twelve requests would be twelve ways to half-succeed and no single
 * place to say so.
 *
 * The selection is cleared on every opening: one that survives a cancel is a
 * selection nobody remembers making.
 */
const candidate = (id: string, code: string, name: string, extra: Record<string, unknown> = {}) => ({
  id,
  code,
  name,
  status: 'active',
  team_name: 'Personal',
  ...extra,
})

const candidates = [
  candidate('res_001', 'R1', 'Vue Component Architecture'),
  candidate('res_002', 'R2', 'State Management Patterns'),
  candidate('res_003', 'R3', 'CSS Architecture Review', { status: 'archived' }),
  candidate('res_004', 'R4', 'Исследование рынка систем управления знаниями', { team_name: 'Growth' }),
]

const meta: Meta<typeof AddResearchDialog> = {
  title: 'Team/AddResearchDialog',
  component: AddResearchDialog,
  tags: ['autodocs'],
  args: { visible: true, teamName: 'Product Research', candidates },
  argTypes: {
    visible: { control: 'boolean' },
    busy: { control: 'boolean' },
    error: { control: 'text' },
    candidates: { control: 'object' },
  },
}
export default meta
type Story = StoryObj<typeof AddResearchDialog>

/** The ordinary case. Each row names where the research is coming *from*, which
 *  is the fact that decides whether moving it takes it away from someone. */
export const Default: Story = {}

/** One candidate: the button reads "Move" rather than "Move 1 researches". */
export const SingleCandidate: Story = {
  args: { candidates: [candidates[0]!] },
}

/** Over eight, so the filter appears. Below that it is a control with nothing
 *  to do standing between the reader and the list. */
export const WithFilter: Story = {
  args: {
    candidates: Array.from({ length: 14 }, (_, i) =>
      candidate(`res_${i}`, `R${i + 1}`, `Research topic number ${i + 1}`),
    ),
  },
}

/** Nothing left to move. Said in a sentence rather than shown as an empty list
 *  under a disabled button that never explains itself — and the move button is
 *  removed entirely, since there is no version of it that could work. */
export const NothingToMove: Story = {
  args: { candidates: [] },
}

/** Mid-request. Cancel is disabled too: closing the dialog while the server is
 *  applying the move would leave the reader watching a page that has not
 *  changed yet, with nothing to explain why. */
export const Moving: Story = {
  args: { busy: true },
}

/** The server refused. Inline rather than a toast: the reader is looking at the
 *  dialog, their selection is still in it, and the retry is one click away. */
export const Refused: Story = {
  args: { error: 'Your role in this team does not allow this' },
}

/** An archived research among the candidates. Status is shown so that moving
 *  one is a decision, not a surprise — the team page lists every status, so it
 *  will appear there afterwards. */
export const WithArchived: Story = {
  args: { candidates: [candidates[2]!, candidates[0]!] },
}
