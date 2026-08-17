import type { Meta, StoryObj } from '@storybook/vue3'
import TaskDetailModal from './TaskDetailModal.vue'
import { markupTaskTitle } from '../../__mocks__/markup'
import { mockTask, mockTaskHigh, mockTaskCompleted } from '../../__mocks__/task'

const meta: Meta<typeof TaskDetailModal> = {
  title: 'Tasks/TaskDetailModal',
  component: TaskDetailModal,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof TaskDetailModal>

export const FullTask: Story = {
  args: {
    task: {
      ...mockTaskCompleted,
      description: 'Check all components for line count and identify candidates for decomposition.\n\n- Components over 200 lines\n- Deeply nested templates\n- Mixed concerns',
    },
    researchSlug: 'R1',
  },
}

export const EmptyDescription: Story = {
  args: {
    task: {
      ...mockTask,
      description: null,
      result: null,
    },
    researchSlug: 'R1',
  },
}

export const HighPriorityTask: Story = {
  args: {
    task: mockTaskHigh,
    researchSlug: 'R1',
  },
}

export const Closed: Story = {
  args: {
    task: null,
    researchSlug: 'R1',
  },
}

/**
 * The one component that calls both `renderRefs` and `linkRefs`, in one view.
 *
 * **Title** — a raw field, rendered with `renderRefs`, which escapes. The tags
 * in it are text: `<b>bold</b>` shows its angle brackets, the script tag is
 * spelled out, and `[[E3]]` is still a link. If escaping regresses, the words
 * `XSS EXECUTED` appear in the title where the image tag was.
 *
 * **Description and Result** — markdown, run through `parseMarkdown` first and
 * then `linkRefs`, which does *not* escape: its input is HTML that was just
 * produced on purpose, and escaping it would print the tags instead of applying
 * them. The heading is a heading, the list is a list, `[[E3]]` is a link.
 *
 * The rule the two halves state together: **`renderRefs` for a field,
 * `linkRefs` for something already rendered.** Reaching for the second one with
 * a raw field is the bug this split exists to make hard, and it is a bug with no
 * visible symptom — the text simply appears, and appears correct.
 *
 * Line 2 of the description is the join between the two rules. It is markdown
 * containing a literal `<b>` tag, and the tag shows as text because
 * `parseMarkdown` escapes author HTML rather than passing it through the way
 * plain `marked` does. That is what lets `linkRefs` be safe on this path without
 * being a sanitiser itself: by the time it runs, nothing hostile is left.
 */
export const MarkupInTitleMarkdownInBody: Story = {
  args: {
    task: {
      ...mockTaskCompleted,
      code: 'T7',
      title: markupTaskTitle,
      description:
        '## What to check\n\n' +
        'Every field that reaches `v-html`, in order:\n\n' +
        '1. `task.title` — a raw field, so **escape it**\n' +
        '2. `task.description` — markdown, so render it; a literal <b>tag</b> in it is still text\n' +
        '3. `task.result` — same\n\n' +
        'Background in [[E3]].',
      result:
        '**[Done]** All thirteen call sites now go through `renderRefs`, and the ' +
        'four that hold rendered markdown call `linkRefs` instead.\n\n' +
        '- `EntryCard`, `EntriesView` — descriptions\n' +
        '- `QuestionList`, `QuestionNode` — question text\n' +
        '- `KanbanCard`, this modal — task titles\n\n' +
        'Cross-referenced from [[E3]].',
    },
    researchSlug: 'R1',
  },
}
