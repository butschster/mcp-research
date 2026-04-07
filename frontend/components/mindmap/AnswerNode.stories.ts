import type { Meta, StoryObj } from '@storybook/vue3'
import AnswerNode from './AnswerNode.vue'

const meta: Meta<typeof AnswerNode> = {
  title: 'Mindmap/AnswerNode',
  component: AnswerNode,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 400px;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof AnswerNode>

export const Default: Story = {
  args: {
    data: {
      answer: 'The refresh token rotation is handled by middleware that intercepts 401 responses and queues subsequent requests until a new access token is obtained.',
      questionCode: 'Q1',
      sessionId: 'sess-001',
      researchSlug: 'R1',
    },
  },
}

export const ShortAnswer: Story = {
  args: {
    data: {
      answer: 'Yes, WAL mode is enabled by default.',
      questionCode: 'Q3',
      sessionId: 'sess-001',
      researchSlug: 'R1',
    },
  },
}

export const LongAnswer: Story = {
  args: {
    data: {
      answer: 'The OAuth2 implementation follows RFC 6749 with PKCE extension (RFC 7636). The authorization server issues short-lived access tokens (15 min) and longer-lived refresh tokens (7 days). Token rotation occurs on each refresh, and the old refresh token is immediately invalidated to prevent replay attacks. The implementation also supports Dynamic Client Registration per RFC 7591.',
      questionCode: 'Q5',
      sessionId: 'sess-002',
      researchSlug: 'R2',
    },
  },
}

export const WithCrossRef: Story = {
  args: {
    data: {
      answer: 'See [[E3]] for the detailed implementation. The approach mirrors what was documented in [[R2:E5]] regarding PKCE validation.',
      questionCode: 'Q2',
      sessionId: 'sess-001',
      researchSlug: 'R1',
    },
  },
}

export const EmptyAnswer: Story = {
  args: {
    data: {
      answer: '',
      questionCode: 'Q7',
      sessionId: 'sess-003',
      researchSlug: 'R1',
    },
  },
}
