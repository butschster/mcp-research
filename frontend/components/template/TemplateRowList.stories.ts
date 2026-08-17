import type { Meta, StoryObj } from '@storybook/vue3'
import TemplateRowList from './TemplateRowList.vue'
import {
  mockGlobalTemplates,
  mockTeamTemplates,
  mockTemplateWithoutCriteria,
} from '../../__mocks__/template'

/**
 * The catalogue list. Same `.data-row` skeleton as the skills list and none of
 * its code: a template row carries a *pair* of criteria and no buttons, because
 * there is nothing to do to a methodology from a list.
 */
const meta: Meta<typeof TemplateRowList> = {
  title: 'Templates/TemplateRowList',
  component: TemplateRowList,
}
export default meta
type Story = StoryObj<typeof TemplateRowList>

export const Shipped: Story = {
  args: {
    templates: mockGlobalTemplates,
    heading: 'Ships with the app',
    note: 2,
    blurb: 'Written by us and updated with the binary.',
  },
}

/** A team's own, and its edited copy of a shipped one. */
export const TeamOwned: Story = {
  args: { templates: mockTeamTemplates, heading: 'Your teams', note: 2 },
}

/** No matching line: the row says so, because an agent will never choose it. */
export const WithoutCriteria: Story = { args: { templates: [mockTemplateWithoutCriteria] } }

export const Empty: Story = {
  args: {
    templates: [],
    heading: 'Ships with the app',
    emptyText: 'None loaded. That is a fault on the server rather than a setting.',
  },
}
