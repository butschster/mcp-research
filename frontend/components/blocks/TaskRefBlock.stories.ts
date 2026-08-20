import type { Meta, StoryObj } from '@storybook/vue3'
import TaskRefBlock from './TaskRefBlock.vue'
import { mockApi, neverResolves } from '../../__mocks__/api'
import { withoutShare } from '../../__mocks__/share'

const meta: Meta<typeof TaskRefBlock> = {
  title: 'Blocks/TaskRefBlock',
  component: TaskRefBlock,
  tags: ['autodocs'],
  // Share state is module state, so a story viewed after any `withShare()` one
  // inherits it: a task code chip would link into /s/{token}/ and [[E4]] would
  // render inert, both silently. This gives every story here a known start.
  decorators: [withoutShare()],
  argTypes: {
    blockId: { control: 'text' },
    refs: { control: 'object' },
    tasks: { control: 'object' },
    researchSlug: { control: 'text' },
    note: { control: 'text' },
    showProgress: { control: 'boolean' },
    readonly: { control: 'boolean' },
    status: { control: 'select', options: ['idle', 'loading', 'ready', 'error', 'excluded'] },
    onRetry: { action: 'retry' },
  },
  parameters: {
    docs: {
      description: {
        component:
          'Existing tasks projected into a document as a checklist. The block holds ' +
          'references only — a tick is a status change on the task itself, so the board, ' +
          'the MCP tools and this document cannot disagree. The page fetches the tasks and ' +
          'hands them down; the block never fetches. A reference that resolves to nothing ' +
          'is dropped rather than drawn as a ghost, and the progress denominator counts ' +
          'rendered rows so a list holding a deleted task can still reach the end.',
      },
    },
  },
}
export default meta
type Story = StoryObj<typeof TaskRefBlock>

const tasks = [
  { id: 'id1', code: 'T1', title: 'Close the perimeter', status: 'completed', priority: 'medium' },
  { id: 'id2', code: 'T4', title: 'Put the scanner inside the network', status: 'in_progress', priority: 'medium' },
  { id: 'id3', code: 'T7', title: 'Rotate the gateway certificates', status: 'blocked', priority: 'high' },
  { id: 'id4', code: 'T9', title: 'Write the runbook', status: 'pending', priority: 'low' },
  { id: 'id5', code: 'T11', title: 'Re-run the failed migration on a copy', status: 'failed', priority: 'high' },
  { id: 'id6', code: 'T12', title: 'Move the runbook into the wiki', status: 'deferred', priority: 'low' },
]

const refs = ['T1', 'T4', 'T7', 'T9', 'T11', 'T12']

export const Default: Story = {
  args: {
    blockId: 'tref0001',
    refs,
    note: 'Before the next call.',
    showProgress: true,
    tasks,
    status: 'ready',
    researchSlug: 'R1',
  },
}

/**
 * All six tickable statuses, and no note.
 *
 * `completed` needs no badge — the tick says it — and `pending` is the default,
 * so a badge on every row would be noise. The other four say what they are.
 * Every one of them stays tickable: a checkbox that looks live and refuses the
 * click is worse than one that does the thing.
 */
export const EveryStatus: Story = { args: { ...(Default.args as any), note: '' } }

/** No progress line: `show_progress: false` in the document. */
export const WithoutProgress: Story = {
  args: { ...(Default.args as any), showProgress: false },
}

/** Nothing done yet. The bar is `tone="done"` so a fresh list is not painted
 *  red for having made no progress against a plan it does not have. */
export const NothingDone: Story = {
  args: {
    ...(Default.args as any),
    tasks: tasks.map((t) => ({ ...t, status: 'pending' })),
  },
}

/** A reference the research no longer holds. The row is dropped, the count
 *  follows it, and a writer gets a footnote naming what did not resolve. */
export const PartiallyResolved: Story = {
  args: {
    ...(Default.args as any),
    refs: [...refs, 'T5', 'T6'],
  },
}

/** The same document to a reader: the rows and the count, and no footnote. A
 *  broken reference is the writer's problem to fix, and naming it to somebody
 *  who cannot fix it is noise about the document's plumbing. */
export const PartiallyResolvedReadOnly: Story = {
  args: { ...(PartiallyResolved.args as any), readonly: true },
}

/** Every reference gone. Said in a sentence rather than as an empty list, which
 *  would read as "no work" instead of "nothing left to show". */
export const AllReferencesGone: Story = {
  args: { ...(Default.args as any), refs: ['T5', 'T6'], tasks: [] },
}

/** The same emptiness to somebody who may not write: the sentence stays, the
 *  offer to open the task board does not. Sending a reader to a board they
 *  cannot act on is an invitation to a locked door. */
export const AllReferencesGoneReadOnly: Story = {
  args: { ...(AllReferencesGone.args as any), readonly: true },
}

/**
 * The page has not answered yet: one skeleton per reference, so the document
 * does not jump when the rows arrive. The note is drawn anyway — it is in the
 * document and it is already true.
 *
 * `idle` renders identically and has no story of its own: it is the same
 * branch, and it is what the page hands down for the moment between mount and
 * the request going out.
 */
export const Loading: Story = {
  args: { ...(Default.args as any), status: 'loading', tasks: [] },
}

/** The fetch failed. The codes still show — the reader learns what the block
 *  points at — and the retry re-asks the page. */
export const LoadFailed: Story = {
  args: { ...(Default.args as any), status: 'error', tasks: [] },
}

/** A share link created without tasks. The route 404s by design, so there is
 *  no retry to offer. */
export const TasksNotShared: Story = {
  args: { ...(Default.args as any), status: 'excluded', tasks: [] },
}

/** An export preview, or a reader who may look and not write: state is shown,
 *  nothing is offered. */
export const ReadOnly: Story = {
  args: { ...(Default.args as any), readonly: true },
}

/** A long title next to badges — the wrap the row is built to take without a
 *  breakpoint. */
export const LongTitles: Story = {
  args: {
    ...(Default.args as any),
    refs: ['T1', 'T4'],
    tasks: [
      {
        id: 'id1',
        code: 'T1',
        title:
          'Проверить, что ротация ключей шлюза выполняется по расписанию и что дежурный ' +
          'знает, где лежит инструкция',
        status: 'blocked',
        priority: 'high',
      },
      { id: 'id2', code: 'T4', title: 'Short one', status: 'completed', priority: 'low' },
    ],
  },
}


/**
 * Fifty rows — the cap the normalizer enforces.
 *
 * There is no fold, deliberately. `task_ref` is a working list, and folding it
 * would hide the thing the block exists to show; the reader who wants a shorter
 * list writes a shorter one.
 */
export const AtTheRowCap: Story = {
  args: {
    ...(Default.args as any),
    note: 'Everything still open on the perimeter work.',
    refs: Array.from({ length: 50 }, (_, i) => `T${i + 1}`),
    tasks: Array.from({ length: 50 }, (_, i) => ({
      id: `cap${i}`,
      code: `T${i + 1}`,
      title: `Rotate the certificate on gateway ${i + 1}`,
      status: i < 12 ? 'completed' : i % 7 === 0 ? 'in_progress' : 'pending',
      priority: i % 11 === 0 ? 'high' : 'medium',
    })),
  },
}

/** Waits for the block's rows before a play function clicks one. */
async function waitFor(root: HTMLElement, selector: string): Promise<HTMLElement | null> {
  for (let i = 0; i < 50; i++) {
    const found = root.querySelector(selector) as HTMLElement | null
    if (found) return found
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
  return null
}

/** A rejection the way `authFetch` delivers one, with the status the copy reads. */
function rejectWith(status?: number, error?: string): Promise<never> {
  return Promise.reject(Object.assign(new Error(error || 'request failed'), {
    status,
    data: error ? { error } : undefined,
  }))
}

/**
 * A tick posted and not yet answered.
 *
 * Reachable only by clicking, because the in-flight set is internal state and
 * no prop exposes it — so this story routes `PUT /api/tasks/{id}` to a request
 * that never settles and clicks the first open box.
 *
 * What to look at: the box is checked and dimmed, and it is **not disabled**.
 * Disabling a focused control blurs it, which throws a keyboard reader back to
 * the top of the document on every tick. The dimming is on the box alone — the
 * checklist dims the whole row, which puts the title under AA for as long as
 * the request runs.
 */
export const TickInFlight: Story = {
  args: { ...(Default.args as any) },
  render: (args: any) => ({
    components: { TaskRefBlock },
    setup() {
      mockApi({ '/api/tasks/': () => neverResolves() })
      return { args }
    },
    template: `<TaskRefBlock v-bind="args" />`,
  }),
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await waitFor(canvasElement, '.b-task-row')
    const boxes = canvasElement.querySelectorAll<HTMLInputElement>('.b-task-check input')
    boxes[1]?.click()
  },
}

/**
 * Three ticks the server refused, each with its own sentence.
 *
 * Also internal state, and reached the same way: the routes below reject with
 * the three shapes the block distinguishes — a 403, a 404, and a request that
 * never reached anything. Every box snaps back visibly, because a silent revert
 * is worse than an error: the reader re-ticks and assumes it stuck.
 *
 * The line lands under the row that failed rather than under the block, so two
 * `task_ref` blocks in one document cannot blame each other.
 */
export const TickRejected: Story = {
  args: { ...(Default.args as any) },
  render: (args: any) => ({
    components: { TaskRefBlock },
    setup() {
      mockApi({
        // A viewer in somebody else's research.
        '/api/tasks/id2': () => rejectWith(403),
        // Deleted between the document loading and the box being ticked.
        '/api/tasks/id3': () => rejectWith(404),
        // Offline, or the binary is not running.
        '/api/tasks/id4': () => rejectWith(),
      })
      return { args }
    },
    template: `<TaskRefBlock v-bind="args" />`,
  }),
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await waitFor(canvasElement, '.b-task-row')
    const boxes = canvasElement.querySelectorAll<HTMLInputElement>('.b-task-check input')
    for (const index of [1, 2, 3]) {
      boxes[index]?.click()
      await new Promise((resolve) => setTimeout(resolve, 30))
    }
  },
}

/**
 * Ticked by accident, then unticked — on the row that was `failed`.
 *
 * The second click sends back the status the row held before the first one, so
 * the `× Failed` badge returns. A flat `pending` would have erased the failure
 * with nothing said, which is the one genuinely lossy case in a two-state box.
 * The memory is this component's, for as long as the page lives; a row it does
 * not remember falls back to `pending`.
 *
 * Internal state again, so the story clicks rather than declares. The write
 * itself succeeds here — an unrouted `authFetch` resolves — which is what makes
 * the restored badge the point rather than a revert.
 */
export const UntickRestoresTheStatus: Story = {
  args: { ...(Default.args as any) },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await waitFor(canvasElement, '.b-task-row')
    const boxes = canvasElement.querySelectorAll<HTMLInputElement>('.b-task-check input')
    // T11, the failed one.
    boxes[4]?.click()
    await new Promise((resolve) => setTimeout(resolve, 60))
    boxes[4]?.click()
  },
}
