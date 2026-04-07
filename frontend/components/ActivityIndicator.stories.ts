import type { Meta, StoryObj } from '@storybook/vue3'
import ActivityIndicator from './ActivityIndicator.vue'

const meta: Meta<typeof ActivityIndicator> = {
  title: 'Base/ActivityIndicator',
  component: ActivityIndicator,
  tags: ['autodocs'],
  argTypes: {
    active: { control: 'boolean' },
    label: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof ActivityIndicator>

export const Active: Story = {
  args: { active: true, label: 'Researching...' },
}

export const ActiveNoLabel: Story = {
  args: { active: true },
}

export const Inactive: Story = {
  args: { active: false, label: 'Researching...' },
}

export const CustomLabel: Story = {
  args: { active: true, label: 'Processing entries' },
}
