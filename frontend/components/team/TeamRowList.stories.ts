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
 * A team is one line of information — a name, a headcount, what work is in it,
 * your role. Boxing each one produces a column of identical rectangles and a
 * scroll for what is really a short list.
 *
 * This is the list a reader meets first, and it was the one list on the teams
 * surface that did **not** use `.data-rows`/`.data-row`: it kept a private
 * subgrid through the split that named it as a consumer, which is how one page
 * ended up with 56px rows above 72px rows above 48px rows. It shares the frame
 * now, and the row is the same four zones as the member and invite lists — who,
 * what, standing, way in.
 *
 * The research count is new and is the point of the row: a team is a set of
 * people around some work, and the work was the one thing this list never said.
 *
 * The personal team still sorts last — it is the row that says nothing new —
 * but it is a link like the others now. It holds researches, and its page is
 * the only place that lists them.
 *
 * Two pages render it: `/teams` in full, and the Settings summary with `limit`.
 * One implementation on purpose — two would grow apart the first time either
 * page was touched.
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

/** One shared team you own — an Owner badge and a way in. */
export const SingleTeam: Story = {
  args: { teams: [mockTeam] },
}

/** Several, each carrying its role as a badge. It used to be an action word —
 *  "Manage" against "Open" — which named the reader's permission in the place a
 *  list normally names the row's own state. */
export const SeveralTeams: Story = {
  args: { teams: [mockTeam, mockTeamEditor, mockTeamViewer, mockTeamIntegrations] },
}

/** The personal team in the mix. It arrives first from the API and is sorted to
 *  the bottom here, and carries a "Personal" chip. It is a link: it holds
 *  twelve researches, and the team page is where they are listed. */
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

/** Russian department names at full length. The name wraps inside the identity
 *  block; the badge and the arrow keep their place at the end of the row. */
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

/** A single member reads "1 member", not "1 members". Both counts in the row
 *  are pluralised, so this is also where a team of one holding one research
 *  would read wrong twice. */
export const SingleMemberTeam: Story = {
  args: {
    teams: [{ ...mockTeamIntegrations, member_count: 1, research_count: 1 }, mockTeamPersonal],
  },
}

/** A team with nobody's work in it — the state a team is in for the whole
 *  window between being created and being useful, and the one the product
 *  previously said nothing about anywhere. Here it reads "0 researches"; on the
 *  team's own page it becomes a paragraph explaining that the owner's existing
 *  researches stayed in their personal team. */
export const EmptyTeam: Story = {
  args: {
    teams: [{ ...mockTeam, member_count: 1, research_count: 0 }, mockTeamPersonal],
  },
}

/** No teams at all. The component renders an empty rule set rather than a
 *  message: both callers decide what "nothing" means for them and show their
 *  own EmptyState instead. */
export const Empty: Story = {
  args: { teams: [] },
}
