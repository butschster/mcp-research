import type { Meta, StoryObj } from '@storybook/vue3'
import EntriesView from './EntriesView.vue'
import { mockEntry, mockEntryDraft } from '../../__mocks__/entry'
import { mockSections, mockSection, mockSectionCompleted } from '../../__mocks__/section'

const meta: Meta<typeof EntriesView> = {
  title: 'Research/EntriesView',
  component: EntriesView,
  tags: ['autodocs'],
  decorators: [
    () => ({
      template: '<div style="max-width: 800px"><story /></div>',
    }),
  ],
  argTypes: {
    mode: { control: 'select', options: ['all', 'section'] },
    loading: { control: 'boolean' },
  },
}
export default meta
type Story = StoryObj<typeof EntriesView>

const entriesWithSections = [
  { ...mockEntry, section_id: 'sec_001' },
  { ...mockEntryDraft, section_id: 'sec_001' },
  { ...mockEntry, id: 'ent_004', code: 'E4', title: 'Slots and Render Functions', tags: ['vue', 'slots'], status: 'active', section_id: 'sec_002' },
  { ...mockEntry, id: 'ent_005', code: 'E5', title: 'Performance Optimization', tags: ['vue', 'performance'], status: 'pending', section_id: 'sec_002' },
]

const tags = [
  { tag: 'vue', count: 4 },
  { tag: 'composables', count: 2 },
  { tag: 'slots', count: 1 },
  { tag: 'performance', count: 1 },
  { tag: 'reactivity', count: 1 },
]

export const AllEntriesGrouped: Story = {
  args: {
    entries: entriesWithSections,
    sections: [mockSection, mockSectionCompleted],
    researchSlug: 'R1',
    loading: false,
    mode: 'all',
    tags,
  },
}

export const Loading: Story = {
  args: {
    entries: [],
    sections: mockSections,
    researchSlug: 'R1',
    loading: true,
    mode: 'all',
    tags: [],
  },
}

export const SectionMode: Story = {
  args: {
    entries: [
      { ...mockEntry, section_id: 'sec_001' },
      { ...mockEntryDraft, section_id: 'sec_001' },
    ],
    sections: mockSections,
    researchSlug: 'R1',
    loading: false,
    mode: 'section',
    sectionInfo: mockSection,
    tags: [],
  },
}

export const TagFilterActive: Story = {
  args: {
    entries: entriesWithSections,
    sections: [mockSection, mockSectionCompleted],
    researchSlug: 'R1',
    loading: false,
    mode: 'all',
    tags,
  },
}

export const Empty: Story = {
  args: {
    entries: [],
    sections: mockSections,
    researchSlug: 'R1',
    loading: false,
    mode: 'all',
    tags: [],
  },
}
