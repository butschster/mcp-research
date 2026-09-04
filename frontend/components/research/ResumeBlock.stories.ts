import type { Meta, StoryObj } from '@storybook/vue3'
import ResumeBlock from './ResumeBlock.vue'
import {
  mockResume,
  mockResumeAmbiguous,
  mockResumeEmpty,
  mockResumeManySessions,
  mockResumeNoActiveSession,
  mockResumeUpdatesByEntry,
} from '../../__mocks__/resume'
import { withoutShare } from '../../__mocks__/share'

/**
 * "Continue" — the block a person reads when they come back to a research after
 * days away.
 *
 * It answers one question: what should happen next, without reading the
 * research. Three action rows are the focal point; the ledger under them is the
 * complete picture of what remains, one group open at a time. Everything else
 * is one click away and costs no height until asked for.
 *
 * Nothing here writes. Opening the summary marks no document as read, moves no
 * status and starts no session — the personal new/changed queue stays a
 * separate request for exactly that reason.
 */
const meta: Meta<typeof ResumeBlock> = {
  title: 'Research/ResumeBlock',
  component: ResumeBlock,
  tags: ['autodocs'],
  decorators: [
    // Every href in the block is built by `useResearchPaths`, which reads the
    // share module state. Without this a story clicked after a share story
    // would quietly render `/s/{token}/…` links.
    withoutShare(),
    (story) => ({
      components: { story },
      setup: () => {
        // The fold is remembered in one localStorage key for the whole product,
        // so a story that toggles it decides how every later story renders.
        // Clearing it here is what keeps the catalogue independent of the order
        // it was clicked through in.
        localStorage.removeItem('research_resume_open')
      },
      template: '<story />',
    }),
    () => ({ template: '<div style="max-width: 1000px"><story /></div>' }),
  ],
  args: { researchSlug: 'R1', canWrite: true },
}
export default meta
type Story = StoryObj<typeof ResumeBlock>

/** The ordinary case: one session, work in three states, marks both ways. */
export const Default: Story = {
  args: { summary: mockResume },
}

/**
 * The Blocked group opened. Exactly one may be open at a time — eight groups of
 * five rows is forty rows above the documents, which turns the research page
 * into a dashboard on a 13" laptop.
 *
 * The two rows differ in the way that matters for this group: T9 carries the
 * task's recorded note and T12 carries none. "Blocked" on its own does not say
 * whether a person is needed, so a row without a reason is the honest shape of
 * the common case, not a gap in the mock.
 */
export const GroupOpen: Story = {
  args: { summary: mockResume },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickText(canvasElement, 'Blocked')
  },
}

/**
 * The Changed group, with this reader's own unseen markers on it.
 *
 * Two facts sit in one row and are not the same fact. The `EntryUpdateBadge` is
 * **personal** — fed from the page's existing updates queue, so two people see
 * different badges on identical rows — while "edited by a person" is shared,
 * and is the row an agent must not treat as its own stale draft. This is the
 * only place `updatesByEntry` is read, and reading it still marks nothing seen.
 */
export const ChangedGroupOpen: Story = {
  args: { summary: mockResume, updatesByEntry: mockResumeUpdatesByEntry },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickText(canvasElement, 'Changed')
  },
}

/**
 * Two interviews open at once. **This is the state the server refuses to guess
 * at:** the repository's "find the active session" is a `LIMIT 1` with no
 * `ORDER BY`, so choosing silently would be a coin toss presented as a fact.
 * The picker offers both, the note says so, and the Questions counters are
 * absent — not disabled — until a session is chosen, because a counter that
 * opens an empty panel teaches distrust of the whole block.
 */
export const SelectionRequired: Story = {
  args: { summary: mockResumeAmbiguous },
}

/**
 * Undecided, and still reading a group.
 *
 * Tasks, marks and documents belong to the research, not to an interview, so
 * they stay countable and openable while the question of which session is being
 * continued is still open. Only Questions and Deferred wait for an answer — the
 * ledger here is the proof that the block degrades by one section rather than
 * going blank.
 */
export const SelectionRequiredGroupOpen: Story = {
  args: { summary: mockResumeAmbiguous },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickText(canvasElement, 'Blocked')
  },
}

/**
 * Months of interviews. The head stops being a chip and becomes a picker, and
 * the `{code} — {title}` labels start doing real work: SS5 and SS6 are both
 * paused and only the title tells them apart.
 */
export const ManySessions: Story = {
  args: { summary: mockResumeManySessions },
}

/**
 * The last session is closed and nothing replaced it.
 *
 * One session is a link, not a control, and it carries its real status — the
 * head has to say the interview is over rather than imply one is running. The
 * work below it is still the research's and still outstanding.
 */
export const LastSessionClosed: Story = {
  args: { summary: mockResumeNoActiveSession },
}

/** Nothing outstanding. Not a blank card: it says so, and offers the way back. */
export const NothingWaiting: Story = {
  args: { summary: mockResumeEmpty },
}

/**
 * A viewer sees the same shared picture — the work is the team's, not the
 * reader's — and the copy stops promising them an agent they cannot run.
 */
export const ViewerRole: Story = {
  args: { summary: mockResumeEmpty, canWrite: false },
}

/**
 * An archived research with work still on it. The heading changes to "Left
 * unfinished", the lead says the list is history rather than a queue, and the
 * actor pills are suppressed — nobody is expected to act on it.
 *
 * It also opens **collapsed**, overriding the remembered preference for this
 * render only: history should not take the top of the page from the documents.
 * The play function opens it, because the content is what this story is for.
 */
export const Archived: Story = {
  args: {
    summary: { ...mockResume, research: { ...mockResume.research, status: 'archived' } },
    archived: true,
  },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await clickText(canvasElement, 'Left unfinished')
  },
}

/** First paint. Skeletons are row-height, so nothing jumps when data lands. */
export const Loading: Story = {
  args: { summary: null, loading: true },
}

/**
 * A refetch in flight with data already on screen — a realtime burst, a session
 * change, or the ⟳ button.
 *
 * The old picture stays, unmoved and unskeletoned: swapping rows for placeholders
 * would throw away a correct answer to avoid admitting it is a few seconds old.
 * The section is `aria-busy` and ⟳ is disabled, so the change is announced
 * without being drawn.
 */
export const Refreshing: Story = {
  args: { summary: mockResume, refreshing: true },
}

/**
 * The first load failed. It offers a retry and says plainly that nothing was
 * changed — a summary that could not load is not a research that lost its work.
 */
export const FirstLoadError: Story = {
  args: { summary: null, error: 'The server could not be reached.' },
}

/**
 * A refresh failed with data already on screen.
 *
 * The rule that matters: **keep the old picture and date it.** Blanking the
 * block or showing zeroes would tell a reader the queue is empty, which is the
 * one wrong answer a summary of outstanding work can give.
 */
export const RefreshFailed: Story = {
  args: { summary: mockResume, error: 'The server could not be reached.' },
}

/** The size cap dropped detail. Totals and the links out survive it. */
export const Truncated: Story = {
  args: { summary: { ...mockResume, truncated: true } },
}

/** ≤768px: the picker takes the row, chips grow to the touch height. */
export const Phone: Story = {
  args: { summary: mockResume },
  parameters: { viewport: { defaultViewport: 'mobile1' } },
}

/** Polls for the control, the way the rest of this catalogue does. */
async function clickText(root: HTMLElement, label: string): Promise<void> {
  for (let i = 0; i < 50; i++) {
    const el = Array.from(root.querySelectorAll('button')).find(b => b.textContent?.trim().startsWith(label))
    if (el) {
      el.click()
      return
    }
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
}
