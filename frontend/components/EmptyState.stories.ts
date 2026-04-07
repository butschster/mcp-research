import type { Meta, StoryObj } from '@storybook/vue3'
import EmptyState from './EmptyState.vue'

const meta: Meta<typeof EmptyState> = {
  title: 'Base/EmptyState',
  component: EmptyState,
  tags: ['autodocs'],
  argTypes: {
    icon: { control: 'text' },
    title: { control: 'text' },
    description: { control: 'text' },
    command: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof EmptyState>

export const IconOnly: Story = {
  args: {
    icon: '\uD83D\uDD2C',
    title: 'No researches yet',
  },
}

export const WithDescription: Story = {
  args: {
    icon: '\uD83D\uDCCB',
    title: 'No entries found',
    description: 'Start a research session to generate entries automatically.',
  },
}

export const WithCommand: Story = {
  args: {
    icon: '\uD83D\uDE80',
    title: 'Get started with Research',
    description: 'Add this MCP server to Claude and run the initialization prompt.',
    command: 'Use the research/initialize prompt',
  },
}

export const MinimalTitle: Story = {
  args: {
    title: 'Nothing here',
  },
}
