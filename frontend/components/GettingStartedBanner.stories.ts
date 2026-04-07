import type { Meta, StoryObj } from '@storybook/vue3'
import GettingStartedBanner from './GettingStartedBanner.vue'

const meta: Meta<typeof GettingStartedBanner> = {
  title: 'Base/GettingStartedBanner',
  component: GettingStartedBanner,
  tags: ['autodocs'],
  argTypes: {
    hasResearches: { control: 'boolean' },
  },
  decorators: [
    (story) => ({
      components: { story },
      setup() {
        // Clear localStorage so the banner always shows in Storybook
        localStorage.removeItem('gs-dismissed')
      },
      template: '<story />',
    }),
  ],
}
export default meta
type Story = StoryObj<typeof GettingStartedBanner>

export const NoResearches: Story = {
  args: { hasResearches: false },
}

export const WithResearches: Story = {
  args: { hasResearches: true },
}
