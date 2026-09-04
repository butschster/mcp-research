import type { Meta, StoryObj } from '@storybook/vue3'
import UpdateBadge from './UpdateBadge.vue'

const meta: Meta<typeof UpdateBadge> = {
  title: 'Entry/UpdateBadge',
  component: UpdateBadge,
  tags: ['autodocs'],
}
export default meta
type Story = StoryObj<typeof UpdateBadge>

export const NewDocument: Story = { args: { kind: 'new', unseenRevisions: 1 } }
export const OneChange: Story = { args: { kind: 'changed', unseenRevisions: 1 } }
export const SeveralChanges: Story = { args: { kind: 'changed', unseenRevisions: 4 } }
export const OmittedCount: Story = { args: { kind: 'changed' } }
