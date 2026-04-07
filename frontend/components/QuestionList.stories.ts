import type { Meta, StoryObj } from '@storybook/vue3'
import QuestionList from './QuestionList.vue'
import {
  mockQuestion,
  mockQuestionPending,
  mockQuestionDeferred,
  mockQuestionsGrouped,
} from '../__mocks__/question'

const meta: Meta<typeof QuestionList> = {
  title: 'Session/QuestionList',
  component: QuestionList,
  tags: ['autodocs'],
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

export const Empty: Story = {
  args: {
    questions: {},
    researchSlug: 'R1',
    sessionId: 'SS1',
  },
}
