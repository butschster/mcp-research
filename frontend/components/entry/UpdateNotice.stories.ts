import type { Meta, StoryObj } from '@storybook/vue3'
import UpdateNotice from './UpdateNotice.vue'

const meta: Meta<typeof UpdateNotice> = {
  title: 'Entry/UpdateNotice',
  component: UpdateNotice,
  tags: ['autodocs'],
  argTypes: { onReview: { action: 'review' } },
}
export default meta
type Story = StoryObj<typeof UpdateNotice>

export const NewDocument: Story = {
  args: { state: { kind: 'new', current_revision: 1, seen_revision: 0, unseen_revisions: 1 } },
}

export const ChangedDocument: Story = {
  args: { state: { kind: 'changed', current_revision: 7, seen_revision: 4, unseen_revisions: 3 } },
}
