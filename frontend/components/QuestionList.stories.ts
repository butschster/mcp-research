import type { Meta, StoryObj } from '@storybook/vue3'
import QuestionList from './QuestionList.vue'
import {
  mockQuestion,
  mockQuestionPending,
  mockQuestionDeferred,
  mockQuestionsGrouped,
} from '../__mocks__/question'
import { markupQuestionText } from '../__mocks__/markup'
import { withShare, withoutShare } from '../__mocks__/share'

/**
 * The questions of one session, grouped by status.
 *
 * A card is a link to the question page only when there is a question page to
 * reach. Without a session and a research it is not, and under a share link
 * there is none by design: the question, its answer and its children are all on
 * the session page already, and a route per question is one more payload to get
 * the redaction right on. An inert card keeps its shape and loses its
 * affordance — no pointer, no hover — rather than linking to `#`.
 */
const meta: Meta<typeof QuestionList> = {
  title: 'Session/QuestionList',
  component: QuestionList,
  tags: ['autodocs'],
  // Share state is module state; this gives the ordinary stories a known
  // starting point rather than whatever the last story left behind. The
  // trade-offs are in __mocks__/share.ts.
  decorators: [withoutShare()],
  argTypes: {
    researchSlug: { control: 'text' },
    sessionId: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof QuestionList>

export const MultipleGroups: Story = {
  args: {
    questions: mockQuestionsGrouped,
    researchSlug: 'R1',
    sessionId: 'SS1',
  },
}

export const AllAnswered: Story = {
  args: {
    questions: {
      answered: [
        mockQuestion,
        {
          ...mockQuestion,
          id: 'q_010',
          code: 'Q10',
          text: 'How does provide/inject compare to composable-based dependency injection?',
          answer: 'Provide/inject is best for deeply nested component trees where prop drilling becomes unwieldy.',
          area: 'architecture',
          priority: 'medium',
        },
      ],
      pending: [],
      in_progress: [],
      deferred: [],
      skipped: [],
    },
    researchSlug: 'R1',
    sessionId: 'SS1',
  },
}

export const WithChildQuestions: Story = {
  args: {
    questions: {
      answered: [mockQuestion],
      pending: [
        mockQuestionPending,
        {
          ...mockQuestionPending,
          id: 'q_020',
          code: 'Q20',
          text: 'What about reactivity across module boundaries?',
          parent_id: 'q_002',
          area: 'state',
          priority: 'low',
        },
      ],
      in_progress: [],
      deferred: [],
      skipped: [],
    },
    researchSlug: 'R1',
    sessionId: 'SS1',
  },
}

export const ManyAreas: Story = {
  args: {
    questions: {
      answered: [mockQuestion],
      pending: [
        mockQuestionPending,
        {
          ...mockQuestionPending,
          id: 'q_030',
          code: 'Q30',
          text: 'How do composables interact with SSR hydration?',
          area: 'ssr',
          priority: 'high',
        },
      ],
      in_progress: [
        {
          ...mockQuestionPending,
          id: 'q_040',
          code: 'Q40',
          text: 'What patterns exist for composable testing?',
          status: 'in_progress',
          area: 'testing',
          priority: 'medium',
        },
      ],
      deferred: [mockQuestionDeferred],
      skipped: [],
    },
    researchSlug: 'R1',
    sessionId: 'SS1',
  },
}

/**
 * A question whose text contains markup.
 *
 * `q.text` is the only field on this card that reaches `v-html`, through
 * `renderRefs`; the answer preview underneath is `{{ }}` and was never at risk.
 * The tags must read as text and `[[E3]]` must still be a link — an executed
 * payload prints `XSS EXECUTED` where the image tag is.
 *
 * A question is written by whoever is running the interview, so this is not a
 * hypothetical field to find markup in: `<Component>` and `a < b` appear in
 * real ones.
 */
export const MarkupInQuestionText: Story = {
  args: {
    questions: {
      answered: [mockQuestion],
      pending: [
        { ...mockQuestionPending, id: 'q_markup', code: 'Q7', text: markupQuestionText },
      ],
      in_progress: [],
      deferred: [],
      skipped: [],
    },
    researchSlug: 'R1',
    sessionId: 'SS1',
  },
}

export const Empty: Story = {
  args: {
    questions: {},
    researchSlug: 'R1',
    sessionId: 'SS1',
  },
}

/**
 * Inside a share link. The cards render exactly as they do above and go
 * nowhere: there is no per-question route on a shared view, and a link that
 * lands an anonymous visitor on a login wall is worse than no link.
 */
export const InsideAShare: Story = {
  decorators: [withShare()],
  args: {
    questions: mockQuestionsGrouped,
    researchSlug: 'R7',
    sessionId: 'SS1',
  },
}

/**
 * Rendered without a session — the graph sidebar does this. There is nothing to
 * link to, so the cards are inert here too. They used to point at `#`, which
 * scrolled the page to the top.
 */
export const WithoutSession: Story = {
  args: {
    questions: mockQuestionsGrouped,
  },
}
