import type { Meta, StoryObj } from '@storybook/vue3'
import TemplateRowList from './TemplateRowList.vue'
import {
  mockGlobalTemplates,
  mockHouseTemplates,
  mockTeamTemplates,
  mockTemplateLongCriteria,
  mockTemplateWithoutCriteria,
} from '../../__mocks__/template'

/**
 * The catalogue list. Same `.data-row` skeleton as the skills list and none of
 * its code: a template row carries a *pair* of criteria and no buttons, because
 * there is nothing to do to a methodology from a list.
 *
 * Every fixture here is a *list* payload: no body, and no research count, since
 * `GET /api/templates` sends neither. The row's "started N researches" line is
 * therefore never reached on the page this component is used from.
 */
const meta: Meta<typeof TemplateRowList> = {
  title: 'Templates/TemplateRowList',
  component: TemplateRowList,
  /* /templates puts each group in a plain `.card` — not the `card--list` the
     skills page uses — and the card is what supplies the padding the rows sit
     in. Rendered bare, the catalog would show a spacing that exists nowhere. */
  decorators: [
    () => ({ template: '<div class="card" style="max-width: 760px"><story /></div>' }),
  ],
}
export default meta
type Story = StoryObj<typeof TemplateRowList>

export const Shipped: Story = {
  args: {
    templates: mockGlobalTemplates,
    heading: 'Ships with the app',
    note: 2,
    blurb: 'Written by us and updated with the binary. Editing one makes a copy for your team; the original stays as it is for everybody else.',
  },
}

/** Global, but written on this server. Nothing in the row but the badge says so
 *  — the group heading on /templates is the other half of the answer, and this
 *  group is only drawn there when it is non-empty. */
export const AddedOnThisServer: Story = {
  args: {
    templates: mockHouseTemplates,
    heading: 'Added on this server',
    note: 1,
    blurb: 'Written here by whoever runs this server, and visible to every team on it. Not part of the app, so an upgrade will not change them.',
  },
}

/** A team's own, and its edited copy of a shipped one. The page passes a
 *  `teamName` resolver so the badge's title can name the owner. */
export const TeamOwned: Story = {
  args: {
    templates: mockTeamTemplates,
    heading: 'Your teams',
    note: 2,
    blurb: 'Written by a team you belong to, and theirs alone. A copy of a shipped methodology replaces the original everywhere that team looks.',
    teamName: (id: string) => (id === 'team-1' ? 'Platform' : undefined),
  },
}

/** The longest a row can get: a name that has to wrap and two criteria at the
 *  length the service still accepts. */
export const LongText: Story = {
  args: {
    templates: [mockTemplateLongCriteria],
    heading: 'Your teams',
    note: 1,
    teamName: () => 'Platform',
  },
}

/** No matching line: the row says so, because an agent will never choose it. */
export const WithoutCriteria: Story = { args: { templates: [mockTemplateWithoutCriteria] } }

/** The shipped group is the one group /templates always draws, so it is the one
 *  that needs an empty line worth reading. Same text the page passes. */
export const Empty: Story = {
  args: {
    templates: [],
    heading: 'Ships with the app',
    note: 0,
    emptyText: 'None loaded. That is a fault on the server rather than a setting — the app ships with several.',
  },
}
