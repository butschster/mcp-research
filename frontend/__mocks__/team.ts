/**
 * Mock payloads for teams, members and invitations.
 *
 * Shapes follow the API exactly:
 *   - a team is `Team`         (GET /api/teams)
 *   - a member is `TeamMember` (GET /api/teams/{id}/members)
 *   - an invite is `TeamInvite`(GET /api/teams/{id}/invites)
 *
 * Timestamps are relative to now, because the invite list prints "expires in N
 * days" and a fixed date would read "expired" a fortnight after this file was
 * written. One of them keeps SQLite's space-separated form, which is what the
 * server actually sends and what the two date parsers here have to survive.
 *
 * The product is used in Russian, so the fixtures are mixed on purpose: names
 * and team names that wrap are the normal case, not the edge case.
 */
import type { Team, TeamInvite, TeamMember } from '~/composables/useTeams'

const DAY = 86_400_000
const inDays = (days: number) => new Date(Date.now() + days * DAY).toISOString()
const daysAgo = (days: number) => new Date(Date.now() - days * DAY).toISOString()

/** The signed-in user in every story below. */
export const myUserId = 'usr_pavel'

export const mockTeam: Team = {
  id: 'tm_product',
  name: 'Product Research',
  personal: false,
  role: 'owner',
  member_count: 5,
  research_count: 18,
}

export const mockTeamEditor: Team = {
  id: 'tm_growth',
  name: 'Growth',
  personal: false,
  role: 'editor',
  member_count: 4,
  research_count: 7,
}

export const mockTeamViewer: Team = {
  id: 'tm_analytics',
  name: 'Аналитика рынка',
  personal: false,
  role: 'viewer',
  member_count: 11,
  research_count: 42,
}

export const mockTeamIntegrations: Team = {
  id: 'tm_integrations',
  name: 'Отдел интеграций',
  personal: false,
  role: 'owner',
  member_count: 3,
  research_count: 5,
}

export const mockTeamLegal: Team = {
  id: 'tm_legal',
  name: 'Legal & Compliance',
  personal: false,
  role: 'viewer',
  member_count: 2,
  research_count: 3,
}

/** Every account has one, and it is the row that says nothing new. */
export const mockTeamPersonal: Team = {
  id: 'tm_personal',
  name: 'Personal',
  personal: true,
  role: 'owner',
  member_count: 1,
  research_count: 12,
}

/** The order the API returns them in — personal first, which is what the list
 *  has to undo. */
export const mockTeams: Team[] = [
  mockTeamPersonal,
  mockTeam,
  mockTeamEditor,
  mockTeamViewer,
  mockTeamIntegrations,
  mockTeamLegal,
]

export const mockMembers: TeamMember[] = [
  {
    team_id: mockTeam.id,
    user_id: myUserId,
    role: 'owner',
    email: 'pavel@intruforce.com',
    name: 'Pavel Buchnev',
    created_at: daysAgo(214),
  },
  {
    team_id: mockTeam.id,
    user_id: 'usr_ekaterina',
    role: 'owner',
    email: 'ekaterina.kovaleva@intruforce.com',
    name: 'Екатерина Ковалёва',
    // SQLite's format, straight off the wire.
    created_at: daysAgo(96).replace('T', ' ').slice(0, 19),
  },
  {
    team_id: mockTeam.id,
    user_id: 'usr_marat',
    role: 'editor',
    email: 'marat.ibragimov@intruforce.com',
    name: 'Марат Ибрагимов',
    created_at: daysAgo(41),
  },
  {
    team_id: mockTeam.id,
    user_id: 'usr_anna',
    role: 'viewer',
    email: 'anna.sorensen@northlight.example',
    name: 'Anna Sørensen',
    created_at: daysAgo(12),
  },
  {
    // A service account registered without a display name: the row falls back
    // to the address, and the avatar to its first letter.
    team_id: mockTeam.id,
    user_id: 'usr_ci',
    role: 'viewer',
    email: 'ci-export@intruforce.com',
    name: '',
    created_at: daysAgo(3),
  },
]

/** Only you, and therefore the last owner. */
export const mockMembersSolo: TeamMember[] = [mockMembers[0]!]

/** One owner among four: the guard the server would apply is applied here too. */
export const mockMembersOneOwner: TeamMember[] = mockMembers.map((m, i) =>
  i === 1 ? { ...m, role: 'editor' as const } : m,
)

export const mockInvites: TeamInvite[] = [
  {
    id: 'inv_live',
    team_id: mockTeam.id,
    email: 'maria.orlova@intruforce.com',
    role: 'editor',
    inviter_name: 'Pavel Buchnev',
    inviter_email: 'pavel@intruforce.com',
    expires_at: inDays(9),
    created_at: daysAgo(5),
  },
  {
    id: 'inv_stale_link',
    team_id: mockTeam.id,
    email: 'dmitry@partner.example',
    role: 'viewer',
    inviter_name: 'Екатерина Ковалёва',
    inviter_email: 'ekaterina.kovaleva@intruforce.com',
    expires_at: inDays(13),
    created_at: daysAgo(1),
  },
  {
    id: 'inv_today',
    team_id: mockTeam.id,
    email: 'aleksey.zhukov@intruforce.com',
    role: 'viewer',
    inviter_name: 'Pavel Buchnev',
    expires_at: inDays(0.4),
    created_at: daysAgo(13),
  },
  {
    id: 'inv_expired',
    team_id: mockTeam.id,
    email: 'legal@old-partner.example',
    role: 'viewer',
    inviter_name: 'Pavel Buchnev',
    expires_at: daysAgo(3),
    created_at: daysAgo(17),
  },
]

/** A token still in memory from this tab's session, keyed by invite id. */
export const mockRecoverableLinks: Record<string, string> = {
  inv_live: 'https://research.intruforce.com/invite/8f2a1c9d4e6b7a3f5c0d8e1b2a4f6c9d',
}

export const mockInviteLink =
  'https://research.intruforce.com/invite/8f2a1c9d4e6b7a3f5c0d8e1b2a4f6c9d'
