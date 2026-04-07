import type { Meta, StoryObj } from '@storybook/vue3'
import Breadcrumbs from './Breadcrumbs.vue'

const meta: Meta<typeof Breadcrumbs> = {
  title: 'Base/Breadcrumbs',
  component: Breadcrumbs,
  tags: ['autodocs'],
}
export default meta
type Story = StoryObj<typeof Breadcrumbs>

export const TwoCrumbs: Story = {
  args: {
    crumbs: [
      { label: 'Researches', to: '/' },
      { label: 'Vue Component Architecture' },
    ],
  },
}

export const FourCrumbs: Story = {
  args: {
    crumbs: [
      { label: 'Researches', to: '/' },
      { label: 'Vue Component Architecture', to: '/research/R1' },
      { label: 'Core Concepts', to: '/research/R1/section/sec_001' },
      { label: 'Composition Patterns' },
    ],
  },
}

export const SingleCrumb: Story = {
  args: {
    crumbs: [{ label: 'Researches' }],
  },
}
