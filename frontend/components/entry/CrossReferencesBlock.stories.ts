import type { Meta, StoryObj } from '@storybook/vue3'
import CrossReferencesBlock from './CrossReferencesBlock.vue'
import {
  mockOutgoingRefs,
  mockIncomingRefs,
  mockOutgoingRef,
  mockOutgoingRefCrossResearch,
  mockOutgoingRefUnresolved,
  mockIncomingRef,
} from '../../__mocks__/crossref'

const meta: Meta<typeof CrossReferencesBlock> = {
  title: 'Entry/CrossReferencesBlock',
  component: CrossReferencesBlock,
  tags: ['autodocs'],
  argTypes: {
    researchSlug: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof CrossReferencesBlock>

export const OutgoingOnly: Story = {
  args: {
    outgoing: mockOutgoingRefs,
    incoming: [],
    researchSlug: 'R1',
  },
}

export const IncomingOnly: Story = {
  args: {
    outgoing: [],
    incoming: [
      mockIncomingRef,
      { ...mockIncomingRef, source_id: 'ent_005', entry_code: 'E5', entry_title: 'Performance Optimization' },
    ],
    researchSlug: 'R1',
  },
}

export const BothDirections: Story = {
  args: {
    outgoing: [mockOutgoingRef, mockOutgoingRefCrossResearch],
    incoming: [mockIncomingRef],
    researchSlug: 'R1',
  },
}

export const WithUnresolved: Story = {
  args: {
    outgoing: [mockOutgoingRef, mockOutgoingRefUnresolved],
    incoming: [],
    researchSlug: 'R1',
  },
}

export const Empty: Story = {
  args: {
    outgoing: [],
    incoming: [],
    researchSlug: 'R1',
  },
}
