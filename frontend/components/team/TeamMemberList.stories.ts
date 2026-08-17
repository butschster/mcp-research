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
 * Name and email were a column each, which made this row 72px against the
 * invite list's 48px under two identical headings — one page, two lists about
 * the same thing, half again the height. They are one identity block now: the
 * name a reader scans for, and under it the two facts that qualify it. That is
 * 56px, the rhythm every rule list on the page shares. Neither string
 * truncates — a shortened email is not an email.
 *
 * `canManage` is the whole difference between the two shapes of this list. An
 * owner gets a role select and a `⋯` per row; anyone else gets the role as
 * plain text and no controls at all — not disabled ones, because a dimmed
 * button still says "you could have done this".
 *
 * Removal moved into the `⋯`. It costs a click on a large team and wins anyway:
 * it used to be a 28px unlabelled ✕, which made the loudest control in the row
 * also the only irreversible one, and it announced itself as "button".
 *
 * The last owner is refused in both places, on the select and in the menu, with
 * the same reason. It is the server's rule, repeated here so the control is
 * never offered in a state the server would reject.
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

/** An owner looking at a healthy team: two owners, so every select and every
 *  menu is live. */
export const AsOwner: Story = {
  args: { members: mockMembers, myUserId, canManage: true },
}

/** The same team seen by an editor or a viewer. Roles are readable, nothing is
 *  actionable, and the rows keep their alignment because the absent menu leaves
 *  a placeholder its exact width — otherwise every role would sit at a
 *  different distance from the edge depending on who is reading. */
export const AsMember: Story = {
  args: { members: mockMembers, myUserId, canManage: false },
}

/** A team of one. You are the only member and therefore the last owner, so both
 *  of your own controls are refused — the page pairs this with a line offering
 *  an invite. */
export const SingleMember: Story = {
  args: { members: mockMembersSolo, myUserId, canManage: true },
}

/** Four members, one owner. That owner's select is disabled and so is Remove
 *  inside their menu, both carrying the reason; the other three rows are
 *  untouched. Making someone else an owner first is what unlocks it. */
export const LastOwner: Story = {
  args: { members: mockMembersOneOwner, myUserId, canManage: true },
}

/** A role change in flight on one row. The row goes faint and `inert`, which
 *  takes it out of the document for the pointer *and* for Tab — it used to be
 *  `opacity` plus `pointer-events: none`, so a keyboard user could still focus
 *  a control inside it and press it into nothing. */
export const ChangeInFlight: Story = {
  args: {
    members: mockMembers,
    myUserId,
    canManage: true,
    busyUserId: 'usr_marat',
  },
}

/** The strings that break a table: a long Cyrillic name, and an address with no
 *  hyphen, dot or underscore to break at. Both wrap inside the identity block
 *  instead of widening the row or pushing the menu off the end. */
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

/** The list wired up: changing a role or removing from the menu applies to
 *  local state, so the last-owner guard can be watched appearing and
 *  disappearing as the second owner is demoted or removed. */
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
