import type { Meta, StoryObj } from '@storybook/vue3'
import TeamRowList from './TeamRowList.vue'
import {
  mockTeam,
  mockTeamEditor,
  mockTeamIntegrations,
  mockTeamLegal,
  mockTeamPersonal,
  mockTeamViewer,
  mockTeams,
} from '../../__mocks__/team'

/**
 * The list of teams, as rules rather than cards.
 *
 * A team is one line of information — a name, a headcount, your role. Boxing
 * each one produces a column of identical rectangles and a scroll for what is
 * really a short list.
 *
 * Two pages render it: `/teams` in full, and the Settings summary with
 * `limit`. One implementation on purpose — two would grow apart the first time
 * either page was touched.
 *
 * The personal team always sorts last and is the one row that is not a link:
 * there is nothing to manage and nobody to invite, so it renders as static text
 * with the action column held open so the rows above stay aligned.
 */
const meta: Meta<typeof TeamRowList> = {
  title: 'Team/TeamRowList',
  component: TeamRowList,
  tags: ['autodocs'],
  argTypes: {
    teams: { control: 'object' },
    limit: { control: { type: 'number', min: 0, max: 10 } },
  },
}
export default meta
type Story = StoryObj<typeof TeamRowList>

/** One shared team you own — the row reads "Manage". */
export const SingleTeam: Story = {
  args: { teams: [mockTeam] },
}

/** Several, with the action text following your role: "Manage" where you own
 *  the team, "Open" where you only belong to it. */
export const SeveralTeams: Story = {
  args: { teams: [mockTeam, mockTeamEditor, mockTeamViewer, mockTeamIntegrations] },
}

/** The personal team in the mix. It arrives first from the API and is sorted to
 *  the bottom here, carries a "Personal" chip and is not a link. */
export const WithPersonalTeam: Story = {
  args: { teams: mockTeams },
}

/** Only the personal team: the whole list is the one row that says nothing new.
 *  `/teams` never reaches this — it shows an EmptyState inviting you to create a
 *  team instead — but Settings does, which is why the row still has to look
 *  finished on its own. */
export const OnlyPersonal: Story = {
  args: { teams: [mockTeamPersonal] },
}

/** The Settings summary: three rows and no more, with the page's own "All
 *  teams" link carrying the rest. The limit applies after the sort, so the
 *  personal team is the first thing dropped rather than the row that survives. */
export const LimitedToThree: Story = {
  args: { teams: mockTeams, limit: 3 },
}

/** Russian department names at full length, next to a five-digit-free headcount
 *  — the name wraps and the meta columns keep their place. */
export const CyrillicNames: Story = {
  args: {
    teams: [
      { ...mockTeamViewer, name: 'Аналитика рынка и конкурентная разведка' },
      { ...mockTeamIntegrations, name: 'Отдел интеграций и партнёрских программ' },
      { ...mockTeamLegal, name: 'Юридическая поддержка' },
      mockTeamPersonal,
    ],
  },
}

/** A single member reads "1 member", not "1 members" — the one plural in the
 *  row, and the one a headcount of one exposes. */
export const SingleMemberTeam: Story = {
  args: {
    teams: [{ ...mockTeamIntegrations, member_count: 1 }, mockTeamPersonal],
  },
}

/** No teams at all. The component renders an empty rule set rather than a
 *  message: both callers decide what "nothing" means for them and show their
 *  own EmptyState instead. */
export const Empty: Story = {
  args: { teams: [] },
}
