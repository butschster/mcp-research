import type { Meta, StoryObj } from '@storybook/vue3'
import { ref, h, defineComponent } from 'vue'
import SearchModal from './SearchModal.vue'

const meta: Meta<typeof SearchModal> = {
  title: 'Search/SearchModal',
  component: SearchModal,
  tags: ['autodocs'],
}
export default meta
type Story = StoryObj<typeof SearchModal>

/**
 * Shows the trigger button in its default state.
 * The modal opens on click or Cmd+K.
 */
export const TriggerButton: Story = {}

/**
 * Demonstrates the open modal with search results.
 * Since SearchModal manages its own open state and fetches data internally,
 * use the play function to open it.
 */
export const OpenState: Story = {
  play: async ({ canvasElement }) => {
    const trigger = canvasElement.querySelector('.search-trigger') as HTMLElement
    if (trigger) trigger.click()
  },
}
