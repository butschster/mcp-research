import type { Meta, StoryObj } from '@storybook/vue3'
import ModalHeader from './ModalHeader.vue'

/**
 * The bar across the top of a dialog.
 *
 * Five dialogs carried this by hand, each with its own copy of the same
 * sixteen-line close icon — and each close button announced as "button" until
 * all five were fixed one at a time, which is what a defect costs when the
 * markup exists five times.
 *
 * `ModalOverlay` owns the keyboard contract; this owns the visual one.
 */
const meta: Meta<typeof ModalHeader> = {
  title: 'Overlays/ModalHeader',
  component: ModalHeader,
  parameters: {
    docs: {
      description: {
        component:
          'A heading and a close button. Pass `title` for plain text, or use '
          + 'the default slot when the heading carries markup.',
      },
    },
  },
}
export default meta

type Story = StoryObj<typeof ModalHeader>

/** The ordinary case. */
export const Default: Story = {
  args: { title: 'New Task', titleId: 'demo-title' },
}

/**
 * A heading with markup — `StatusChangeModal` puts a status dot inside its
 * title, which is why the heading is a slot and not only a prop.
 */
export const WithMarkup: Story = {
  render: () => ({
    components: { ModalHeader },
    template: `
      <ModalHeader title-id="demo-markup">
        Move to
        <span class="badge badge-completed">Completed</span>
      </ModalHeader>
    `,
  }),
}

/**
 * A long title against a close button that must not move. The button is
 * `flex-shrink: 0` in the shared vocabulary; this is the story that would show
 * it if that ever stopped being true.
 */
export const LongTitle: Story = {
  args: {
    title: 'Evaluating pgvector against a dedicated vector database for retrieval over long research documents',
    titleId: 'demo-long',
  },
}

/** A title carrying a short code, as the task detail dialog does. */
export const WithCode: Story = {
  render: () => ({
    components: { ModalHeader },
    template: `
      <ModalHeader title-id="demo-code">
        <span class="short-code">T7</span>
        Migrate the embedding store
      </ModalHeader>
    `,
  }),
}
