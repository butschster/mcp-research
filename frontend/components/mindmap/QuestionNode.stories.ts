import type { Meta, StoryObj } from '@storybook/vue3'
import QuestionNode from './QuestionNode.vue'
import { markupImg } from '../../__mocks__/markup'

const meta: Meta<typeof QuestionNode> = {
  title: 'Mindmap/QuestionNode',
  component: QuestionNode,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 440px;"><story /></div>',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof QuestionNode>

export const Pending: Story = {
  args: {
    data: {
      id: 'q-001',
      code: 'Q1',
      text: 'How does the token refresh flow handle concurrent requests?',
      status: 'pending',
      answer: '',
      sessionId: 'sess-001',
      sessionTitle: 'Initial Exploration',
      researchSlug: 'R1',
    },
  },
}

export const Answered: Story = {
  args: {
    data: {
      id: 'q-002',
      code: 'Q3',
      text: 'What happens when the refresh token expires mid-request?',
      status: 'answered',
      answer: 'The client receives a 401, clears local tokens, and redirects to the login page. A retry queue is not implemented.',
      sessionId: 'sess-001',
      sessionTitle: 'Security Deep Dive',
      researchSlug: 'R1',
    },
  },
}

export const Skipped: Story = {
  args: {
    data: {
      id: 'q-003',
      code: 'Q5',
      text: 'Is there a mechanism to revoke all sessions for a user?',
      status: 'skipped',
      answer: '',
      sessionId: 'sess-002',
      sessionTitle: 'Follow-up Session',
      researchSlug: 'R1',
    },
  },
}

export const LongQuestion: Story = {
  args: {
    data: {
      id: 'q-004',
      code: 'Q8',
      text: 'Given the current architecture where stdio, SSE, and Streamable HTTP all run in one process, what are the failure modes if the WebSocket hub blocks?',
      status: 'pending',
      answer: '',
      sessionId: 'sess-003',
      sessionTitle: 'Architecture Review',
      researchSlug: 'R2',
    },
  },
}

export const WithCrossRef: Story = {
  args: {
    data: {
      id: 'q-005',
      code: 'Q2',
      text: 'How does [[E3]] relate to the PKCE flow described in [[R2:E5]]?',
      status: 'pending',
      answer: '',
      sessionId: 'sess-001',
      sessionTitle: 'Cross-reference Analysis',
      researchSlug: 'R1',
    },
  },
}

/**
 * A question whose text contains markup.
 *
 * `data.text` reaches `v-html` through `renderRefs`, which escapes it: the tags
 * read as text and `[[E3]]` beside them is still a link. The answer line
 * underneath takes the other path — `parseMarkdownInline` then `linkRefs` — so
 * both rules of the split are visible on one node.
 *
 * Kept under 70 characters because the node truncates there; the sibling story
 * below shows what happens when it does not fit.
 */
export const MarkupInText: Story = {
  args: {
    data: {
      id: 'q-007',
      code: 'Q9',
      text: '<b>bold</b> and <script>alert(1)</script>? [[E3]]',
      status: 'answered',
      answer: 'It renders as text. The reference next to it still links.',
      sessionId: 'sess-001',
      sessionTitle: 'Rendering Rules',
      researchSlug: 'R1',
    },
  },
}

/**
 * The same payload, long enough that the 70-character cut lands **inside** the
 * image tag.
 *
 * `truncate()` runs before `renderRefs()`, so the half tag is escaped along with
 * everything else and shows up as the broken text it is. In the other order —
 * escape, then slice — the cut could fall inside an entity or between a tag's
 * name and its closing bracket, and the browser would be handed something it
 * has to guess about. Worth having a story for, because both orders look
 * equally reasonable in the source line.
 *
 * `[[E3]]` is past the cut here and is gone; the story above carries it.
 */
export const MarkupTruncatedMidTag: Story = {
  args: {
    data: {
      id: 'q-008',
      code: 'Q11',
      text: `Is the truncation safe here: ${markupImg} [[E3]]`,
      status: 'pending',
      answer: '',
      sessionId: 'sess-001',
      sessionTitle: 'Rendering Rules',
      researchSlug: 'R1',
    },
  },
}

export const NoCode: Story = {
  args: {
    data: {
      id: 'q-006',
      code: '',
      text: 'Can we replace the SQLite driver without CGo?',
      status: 'answered',
      answer: 'Yes, modernc.org/sqlite is a pure Go translation of SQLite and is already in use.',
      sessionId: 'sess-001',
      sessionTitle: 'Technical Feasibility',
      researchSlug: 'R3',
    },
  },
}
