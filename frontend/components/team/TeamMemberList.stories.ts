import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import TeamMemberList from './TeamMemberList.vue'
import type { TeamMember, TeamRole } from '~/composables/useTeams'
import {
  mockMembers,
  mockMembersOneOwner,
  mockMembersSolo,
  myUserId,
} from '../../__mocks__/team'

/**
 * The people in a team.
 *
 * Name and email each get a column and both wrap rather than truncate: a
 * shortened email is not an email, and these are the two strings a reader came
 * to check. Your own row says "you"; everyone else's says when they joined.
 *
 * `canManage` is the whole difference between the two shapes of this list. An
 * owner gets a role select and a remove button per row; anyone else gets the
 * role as plain text and no controls at all — not disabled ones, because a
 * dimmed button still says "you could have done this".
 *
 * The last owner is refused in both places, on the select and on the ✕, with the
 * same reason. It is the server's rule, repeated here so the control is never
 * offered in a state the server would reject.
 */
const meta: Meta<typeof TeamMemberList> = {
  title: 'Team/TeamMemberList',
  component: TeamMemberList,
  tags: ['autodocs'],
  argTypes: {
    members: { control: 'object' },
    myUserId: { control: 'text' },
    canManage: { control: 'boolean' },
    busyUserId: { control: 'text' },
    onChangeRole: { action: 'changeRole' },
    onRemove: { action: 'remove' },
  },
}
export default meta
type Story = StoryObj<typeof TeamMemberList>

/** An owner looking at a healthy team: two owners, so every select and every ✕
 *  is live. */
export const AsOwner: Story = {
  args: { members: mockMembers, myUserId, canManage: true },
}

/** The same team seen by an editor or a viewer. Roles are readable, nothing is
 *  actionable, and the rows keep their alignment because the removed button
 *  leaves a zero-width placeholder rather than collapsing the grid. */
export const AsMember: Story = {
  args: { members: mockMembers, myUserId, canManage: false },
}

/** A team of one. You are the only member and therefore the last owner, so both
 *  of your own controls are refused — the page pairs this with a line offering
 *  an invite. */
export const SingleMember: Story = {
  args: { members: mockMembersSolo, myUserId, canManage: true },
}

/** Four members, one owner. That owner's select and ✕ are disabled and both
 *  carry the reason; the other three rows are untouched. Making someone else an
 *  owner first is what unlocks it. */
export const LastOwner: Story = {
  args: { members: mockMembersOneOwner, myUserId, canManage: true },
}

/** A role change in flight on one row. The row dims, its select locks and its ✕
 *  goes with it, so a second click cannot race the first. */
export const ChangeInFlight: Story = {
  args: {
    members: mockMembers,
    myUserId,
    canManage: true,
    busyUserId: 'usr_marat',
  },
}

/** The strings that break a table: a long Cyrillic name, and an address with no
 *  hyphen, dot or underscore to break at. Both wrap inside their column instead
 *  of widening the row. */
export const LongNamesAndEmails: Story = {
  args: {
    members: [
      {
        team_id: 'tm_product',
        user_id: myUserId,
        role: 'owner',
        email: 'pavel@intruforce.com',
        name: 'Pavel Buchnev',
        created_at: mockMembers[0]!.created_at,
      },
      {
        team_id: 'tm_product',
        user_id: 'usr_long',
        role: 'editor',
        email: 'ekaterinakonstantinovnakovalevskaya@issledovaniyarynkaipartnerstv.example',
        name: 'Екатерина Константиновна Ковалевская-Преображенская',
        created_at: mockMembers[2]!.created_at,
      },
    ] as TeamMember[],
    myUserId,
    canManage: true,
  },
}

/** A member registered without a display name: the row falls back to the
 *  address and the avatar to its first letter, rather than showing a blank line
 *  where a name should be. */
export const MemberWithoutName: Story = {
  args: {
    members: [mockMembers[0]!, mockMembers[1]!, mockMembers[4]!],
    myUserId,
    canManage: true,
  },
}

/** The list wired up: changing a role or pressing ✕ applies to local state, so
 *  the last-owner guard can be watched appearing and disappearing as the second
 *  owner is demoted or removed. */
export const Interactive: Story = {
  render: () => ({
    components: { TeamMemberList },
    setup() {
      const members = ref<TeamMember[]>(mockMembers.map((m) => ({ ...m })))
      const busyUserId = ref<string | null>(null)

      function changeRole(userId: string, role: TeamRole) {
        busyUserId.value = userId
        setTimeout(() => {
          const member = members.value.find((m) => m.user_id === userId)
          if (member) member.role = role
          busyUserId.value = null
        }, 400)
      }

      function remove(member: TeamMember) {
        members.value = members.value.filter((m) => m.user_id !== member.user_id)
      }

      return { members, busyUserId, changeRole, remove, myUserId }
    },
    template: `
      <TeamMemberList
        :members="members"
        :my-user-id="myUserId"
        :can-manage="true"
        :busy-user-id="busyUserId"
        @change-role="changeRole"
        @remove="remove"
      />
    `,
  }),
}
