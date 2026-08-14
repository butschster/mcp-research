import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import TransferModal from './TransferModal.vue'
import {
  mockTeam,
  mockTeamIntegrations,
  mockTeamPersonal,
  mockTeams,
} from '../../__mocks__/team'

/**
 * Moves a research between teams.
 *
 * It is the only operation that changes who can see a research, so it states the
 * consequence in both directions — who loses access, who gains it — before it
 * runs, and names the target as it is picked rather than after the fact.
 *
 * `teams` is the caller's owned teams, current one included; the dialog filters
 * that one out itself, so the page does not have to. When nothing is left it
 * explains and offers the way out, because a disabled menu item that never says
 * why is how a feature gets reported as broken.
 *
 * The short code does not move with it: `/research/R7` links have to survive a
 * transfer, and codes are global rather than per team.
 */
const meta: Meta<typeof TransferModal> = {
  title: 'Research/TransferModal',
  component: TransferModal,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
  argTypes: {
    visible: { control: 'boolean' },
    research: { control: 'object' },
    currentTeamId: { control: 'text' },
    currentTeamName: { control: 'text' },
    teams: { control: 'object' },
    busy: { control: 'boolean' },
    error: { control: 'text' },
    onTransfer: { action: 'transfer' },
    onClose: { action: 'close' },
  },
}
export default meta
type Story = StoryObj<typeof TransferModal>

const research = { code: 'R7', name: 'Ценообразование в SaaS' }

/** Out of the personal team and into a shared one — the common direction, and
 *  the moment a research stops being private. */
export const FromPersonalTeam: Story = {
  args: {
    visible: true,
    research,
    currentTeamId: mockTeamPersonal.id,
    currentTeamName: mockTeamPersonal.name,
    teams: [mockTeamPersonal, mockTeam, mockTeamIntegrations],
  },
}

/** Several owned teams to choose between. The consequence line follows the
 *  picker, so it always names the team currently selected. */
export const SeveralTargets: Story = {
  args: {
    visible: true,
    research,
    currentTeamId: mockTeam.id,
    currentTeamName: mockTeam.name,
    teams: mockTeams.filter((t) => t.role === 'owner'),
  },
}

/** Exactly one target: the select holds a single option, preselected, so the
 *  dialog is one button press. */
export const SingleTarget: Story = {
  args: {
    visible: true,
    research,
    currentTeamId: mockTeam.id,
    currentTeamName: mockTeam.name,
    teams: [mockTeam, mockTeamIntegrations],
  },
}

/** Nowhere to move it to — the only team you own is the one it is already in.
 *  The dialog explains and links to team creation instead of showing an empty
 *  picker and a dead Move button. */
export const NoOtherTeams: Story = {
  args: {
    visible: true,
    research,
    currentTeamId: mockTeam.id,
    currentTeamName: mockTeam.name,
    teams: [mockTeam],
  },
}

/** The same dead end reached from the personal team, which is where a first-time
 *  user meets it. */
export const NoOtherTeamsFromPersonal: Story = {
  args: {
    visible: true,
    research,
    currentTeamId: mockTeamPersonal.id,
    currentTeamName: mockTeamPersonal.name,
    teams: [mockTeamPersonal],
  },
}

/** Moving. Both buttons lock so the transfer cannot be fired twice at two
 *  different targets. */
export const Busy: Story = {
  args: {
    visible: true,
    research,
    currentTeamId: mockTeam.id,
    currentTeamName: mockTeam.name,
    teams: mockTeams.filter((t) => t.role === 'owner'),
    busy: true,
  },
}

/** The server refused — someone else moved or deleted the target team while the
 *  dialog was open. Said inline, with the picker still usable. */
export const TransferRefused: Story = {
  args: {
    visible: true,
    research,
    currentTeamId: mockTeam.id,
    currentTeamName: mockTeam.name,
    teams: mockTeams.filter((t) => t.role === 'owner'),
    error: 'That team no longer exists. Reload and try again.',
  },
}

/** A research with no code yet and a long English title, next to Cyrillic team
 *  names — the header line wraps rather than pushing the dialog wide. */
export const LongNames: Story = {
  args: {
    visible: true,
    research: { name: 'Competitive landscape for mid-market analytics suites, second pass' },
    currentTeamId: mockTeam.id,
    currentTeamName: 'Product Research',
    teams: [
      mockTeam,
      { ...mockTeamIntegrations, name: 'Отдел интеграций и партнёрских программ' },
    ],
  },
}

/** Open, pick, move: the target resets to the first option on every opening, so
 *  a dialog reopened after a cancel never carries a stale selection into the
 *  next transfer. */
export const Interactive: Story = {
  render: () => ({
    components: { TransferModal },
    setup() {
      const owned = mockTeams.filter((t) => t.role === 'owner')
      const visible = ref(false)
      const busy = ref(false)
      const currentTeamId = ref(mockTeamPersonal.id)
      const currentTeamName = ref(mockTeamPersonal.name)

      function transfer(teamId: string) {
        busy.value = true
        setTimeout(() => {
          const team = owned.find((t) => t.id === teamId)
          currentTeamId.value = teamId
          currentTeamName.value = team?.name ?? ''
          busy.value = false
          visible.value = false
        }, 500)
      }

      return { owned, visible, busy, currentTeamId, currentTeamName, research, transfer }
    },
    template: `
      <div style="padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-3); align-items: flex-start;">
        <button class="btn btn-sm" @click="visible = true">Move to another team</button>
        <p style="margin: 0; font-size: var(--type-xs); color: var(--color-text-muted);">
          <strong>{{ research.code }}</strong> is in <code>{{ currentTeamName }}</code>
        </p>
        <TransferModal
          :visible="visible"
          :research="research"
          :current-team-id="currentTeamId"
          :current-team-name="currentTeamName"
          :teams="owned"
          :busy="busy"
          @transfer="transfer"
          @close="visible = false"
        />
      </div>
    `,
  }),
}
