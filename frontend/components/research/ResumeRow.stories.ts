import type { Meta, StoryObj } from '@storybook/vue3'
import ResumeRow from './ResumeRow.vue'

/**
 * One line of the Continue block.
 *
 * Both variants — a proposed next action and an item inside a group — are the
 * same anatomy, so they are one component: who is expected to act, the short
 * code, what it is, and the badges that qualify it. The whole row is a single
 * link, because a row with a link inside it and clickable space around it gives
 * a keyboard user two targets for one destination.
 *
 * The `variant` prop is passed by the block and read by nothing: the template
 * never mentions it, and the two shapes differ only in which optional props
 * arrive. The stories keep passing it because that is what the caller does, but
 * there is no variant story here because there is no variant behaviour to show.
 */
const meta: Meta<typeof ResumeRow> = {
  title: 'Research/ResumeRow',
  component: ResumeRow,
  tags: ['autodocs'],
  decorators: [() => ({ template: '<div class="card card--list" style="max-width: 760px"><div class="data-rows"><story /></div></div>' })],
}
export default meta
type Story = StoryObj<typeof ResumeRow>

/** The agent's own work, with the evidence for the suggestion under it. */
export const AgentAction: Story = {
  args: {
    variant: 'action',
    code: 'T4',
    title: 'Compare provider pricing across the three vendors',
    href: '/research/R1/tasks?task=T4',
    actor: 'agent',
    reason: 'already in progress',
  },
}

/**
 * Waiting on a person. The pill reads "You", not "Human": the block is read by
 * the person it is talking about, and an agent cannot accept its own answer.
 */
export const WaitingOnYou: Story = {
  args: {
    variant: 'action',
    code: 'A5',
    title: '“nine stores in, four out”',
    href: '/research/R1/entry/E2',
    actor: 'human',
    reason: 'answered, waiting for you to accept or reject it',
  },
}

/** A group item: no actor pill, badges instead, and a trailing fact. */
export const GroupItem: Story = {
  args: {
    variant: 'item',
    code: 'T9',
    title: 'Wait on the vendor NDA before quoting',
    href: '/research/R1/tasks?task=T9',
    status: 'blocked',
    priority: 'medium',
    note: 'legal has not answered',
  },
}

/**
 * A changed document the reader has not seen, edited by a person.
 *
 * Two different facts sit in one row and must not be confused: the badge is
 * personal — this reader has not opened it — and the note is shared, and is the
 * one row an agent must not treat as its own stale draft.
 */
export const ChangedByAPerson: Story = {
  args: {
    variant: 'item',
    code: 'E1',
    title: 'Benchmarks on our own corpus',
    href: '/research/R1/entry/E1',
    note: 'edited by a person',
    meta: '2h ago',
    updateKind: 'changed',
    unseenRevisions: 2,
  },
}

/**
 * A title long enough to be a real document title in this product. It clamps at
 * two lines: the ledger below it must stay on screen, and the full text is on
 * the page the row links to.
 */
export const LongTitle: Story = {
  args: {
    variant: 'item',
    code: 'E42',
    title: 'Сравнение поставщиков векторного поиска по стоимости, задержке и требованиям к инфраструктуре, включая перенос индексов и стоимость повторной индексации всего корпуса документов поддержки',
    href: '/research/R1/entry/E42',
    status: 'completed',
    meta: '3d ago',
  },
}

/**
 * A mark that has been answered and is waiting for a person to accept it.
 *
 * The title is the quoted sentence, not the document — a mark lives on a
 * sentence, and the entry code trails at the end so the row says where without
 * spending the line on it.
 */
export const MarkAwaitingYou: Story = {
  args: {
    variant: 'item',
    code: 'A5',
    title: '“nine stores in, four out”',
    href: '/research/R1/entry/E2',
    status: 'answered',
    meta: 'E2',
  },
}

/**
 * Every shape the Continue block actually emits, in one place.
 *
 * They are one component because the anatomy never changes — actor, code,
 * title, qualifying badges — and only which optional props arrive does. Read
 * top to bottom: an action carries an actor pill and a `reason`; a task item
 * carries status and priority and falls back to its own `note`; a mark item
 * carries a quote and a trailing entry code; a changed document carries the
 * reader's personal badge and a relative time.
 */
export const AllRowShapes: Story = {
  render: () => ({
    components: { ResumeRow },
    setup: () => ({
      rows: [
        { variant: 'action', code: 'T4', title: 'Compare provider pricing across the three vendors', href: '/research/R1/tasks?task=T4', actor: 'agent', reason: 'already in progress' },
        { variant: 'action', code: 'A5', title: '“nine stores in, four out”', href: '/research/R1/entry/E2', actor: 'human', reason: 'answered, waiting for you to accept or reject it' },
        { variant: 'item', code: 'T9', title: 'Wait on the vendor NDA before quoting', href: '/research/R1/tasks?task=T9', status: 'blocked', priority: 'medium', note: 'legal has not answered' },
        { variant: 'item', code: 'T12', title: 'Benchmark rerun needs the new dataset', href: '/research/R1/tasks?task=T12', status: 'blocked', priority: 'low' },
        { variant: 'item', code: 'Q7', title: 'What does the enterprise tier actually cost?', href: '/research/R1/session/SS3/question/Q7', status: 'pending', priority: 'high' },
        { variant: 'item', code: 'A2', title: '“latency figures are quoted from 2023”', href: '/research/R1/entry/E1', status: 'open', meta: 'E1' },
        { variant: 'item', code: 'E1', title: 'Benchmarks on our own corpus', href: '/research/R1/entry/E1', note: 'edited by a person', meta: '2h ago', updateKind: 'changed', unseenRevisions: 2 },
      ],
    }),
    template: '<ResumeRow v-for="(row, i) in rows" :key="i" v-bind="row" />',
  }),
}
