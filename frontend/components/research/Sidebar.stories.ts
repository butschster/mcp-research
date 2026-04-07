import type { Meta, StoryObj } from '@storybook/vue3'
import Sidebar from './Sidebar.vue'
import { mockSections } from '../../__mocks__/section'

const meta: Meta<typeof Sidebar> = {
  title: 'Research/Sidebar',
  component: Sidebar,
  tags: ['autodocs'],
  argTypes: {
    activeSection: { control: 'text' },
    totalEntryCount: { control: 'number' },
    linksTotal: { control: 'number' },
  },
}
export default meta
type Story = StoryObj<typeof Sidebar>

export const AllEntriesActive: Story = {
  args: {
    sections: mockSections,
    activeSection: '__all__',
    totalEntryCount: 13,
    linksTotal: 7,
  },
}

export const SectionActive: Story = {
  args: {
    sections: mockSections,
    activeSection: 'sec_001',
    totalEntryCount: 13,
    linksTotal: 7,
  },
}

export const LinksActive: Story = {
  args: {
    sections: mockSections,
    activeSection: '__links__',
    totalEntryCount: 13,
    linksTotal: 7,
  },
}

export const NoLinks: Story = {
  args: {
    sections: mockSections,
    activeSection: '__all__',
    totalEntryCount: 13,
    linksTotal: 0,
  },
}
