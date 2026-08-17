import type { Meta, StoryObj } from '@storybook/vue3'
import EditableField from './EditableField.vue'

/**
 * A value that a pencil turns into a form.
 *
 * The triple — a header with a pencil, the value, an edit form — was written
 * out eight times across `DetailsPanel` and `TaskDetailModal`, which also
 * shared ten byte-identical rule bodies. Every copy carried the same twelve-line
 * pencil icon and the same `title="Edit"`, which is ambiguous when four of them
 * sit in one panel; the label is derived from the field's own name now.
 */
const meta: Meta<typeof EditableField> = {
  title: 'Forms/EditableField',
  component: EditableField,
  parameters: {
    docs: {
      description: {
        component:
          'The draft is local: the owner hears about it only on Save, so a '
          + 'cancelled edit costs nothing and a refused one is the owner’s to report.',
      },
    },
  },
}
export default meta

type Story = StoryObj<typeof EditableField>

/** The ordinary case: a single-line value a writer may change. */
export const Editable: Story = {
  args: {
    label: 'Goal',
    value: 'Decide whether pgvector is enough for retrieval over long documents.',
    editable: true,
    placeholder: 'What is this research trying to achieve?',
  },
}

/** A viewer sees the value and no way in — the pencil and the double-click both go. */
export const ReadOnly: Story = {
  args: {
    label: 'Goal',
    value: 'Decide whether pgvector is enough for retrieval over long documents.',
    editable: false,
  },
}

/**
 * Nothing set yet. The empty text is muted and says what to do, rather than
 * leaving a blank where a value should be.
 */
export const Empty: Story = {
  args: {
    label: 'Description',
    value: '',
    editable: true,
    emptyText: 'Click the pencil to add a description',
  },
}

/** A long value that wants room: `multiline` swaps the input for a textarea. */
export const Multiline: Story = {
  args: {
    label: 'AI Instruction',
    value:
      'Prefer primary sources. When a claim comes from a vendor benchmark, say so in the entry and record the version it was run against.',
    editable: true,
    multiline: true,
    rows: 6,
    placeholder: 'How should the agent work on this research?',
  },
}

/**
 * A value that is not a string. The slot exists because this component must
 * not decide what a value means — tags are a list, and the panel renders them
 * as tags while still editing them as a comma-separated line.
 */
export const CustomRendering: Story = {
  render: (args) => ({
    components: { EditableField },
    setup: () => ({ args }),
    template: `
      <EditableField v-bind="args">
        <span class="tag tag-hue-1">retrieval</span>
        <span class="tag tag-hue-3">pgvector</span>
        <span class="tag tag-hue-5">benchmarks</span>
      </EditableField>
    `,
  }),
  args: {
    label: 'Tags',
    value: 'retrieval, pgvector, benchmarks',
    editable: true,
    placeholder: 'tag1, tag2, tag3',
  },
}

/**
 * A very long unbroken value beside the pencil, which is where a header row
 * with a fixed-size button usually comes apart.
 */
export const LongValue: Story = {
  args: {
    label: 'Memory',
    value:
      'https://example.com/a-very-long-url-that-will-not-break-anywhere-because-it-contains-no-spaces-at-all-and-keeps-going',
    editable: true,
    multiline: true,
  },
}
