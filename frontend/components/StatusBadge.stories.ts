import type { Meta, StoryObj } from '@storybook/vue3'
import StatusBadge from './StatusBadge.vue'

const meta: Meta<typeof StatusBadge> = {
  title: 'Base/StatusBadge',
  component: StatusBadge,
  tags: ['autodocs'],
  argTypes: {
    status: {
      control: 'select',
      options: [
        'active', 'completed', 'archived', 'draft', 'pending',
        'answered', 'in_progress', 'deferred', 'skipped', 'blocked',
        'failed', 'high', 'medium', 'low',
      ],
    },
  },
}
export default meta
type Story = StoryObj<typeof StatusBadge>

export const Active: Story = { args: { status: 'active' } }
export const Completed: Story = { args: { status: 'completed' } }
export const Archived: Story = { args: { status: 'archived' } }
export const Draft: Story = { args: { status: 'draft' } }
export const Pending: Story = { args: { status: 'pending' } }
export const Answered: Story = { args: { status: 'answered' } }
export const InProgress: Story = { args: { status: 'in_progress' } }
export const Deferred: Story = { args: { status: 'deferred' } }
export const Skipped: Story = { args: { status: 'skipped' } }
export const Blocked: Story = { args: { status: 'blocked' } }
export const Failed: Story = { args: { status: 'failed' } }
export const High: Story = { args: { status: 'high' } }
export const Medium: Story = { args: { status: 'medium' } }
export const Low: Story = { args: { status: 'low' } }

export const AllStatuses: Story = {
  render: () => ({
    components: { StatusBadge },
    template: `
      <div style="display: flex; flex-wrap: wrap; gap: 0.75rem; align-items: center;">
        <StatusBadge v-for="s in statuses" :key="s" :status="s" />
      </div>
    `,
    setup() {
      const statuses = [
        'active', 'completed', 'archived', 'draft', 'pending',
        'answered', 'in_progress', 'deferred', 'skipped', 'blocked',
        'failed', 'high', 'medium', 'low',
      ]
      return { statuses }
    },
  }),
}
