import type { Meta, StoryObj } from '@storybook/vue3'
import AccessRevokedNotice from './AccessRevokedNotice.vue'

const meta: Meta<typeof AccessRevokedNotice> = {
  title: 'Base/AccessRevokedNotice',
  component: AccessRevokedNotice,
  tags: ['autodocs'],
  argTypes: {
    revocation: { control: 'object' },
  },
  parameters: {
    docs: {
      description: {
        component:
          'Replaces the body of a page whose subject the reader has just lost access to, ' +
          'while they are looking at it. Raised by a directed `access.revoked` event. ' +
          'It never redirects — being moved off the page without warning is worse than ' +
          'the stale page it replaces, and destructive if a draft is open. ' +
          'The title prefers the short code and falls back to the name, so the two ' +
          'scopes read differently: a research is “R7”, a team has no code and is named.',
      },
    },
  },
}
export default meta
type Story = StoryObj<typeof AccessRevokedNotice>

/** Removed from a team, so everything that team owned went with it. */
export const RemovedFromTeam: Story = {
  args: {
    revocation: {
      scope: 'team',
      id: 'team-1',
      name: 'Northwind Ops',
      reason: 'removed_from_team',
    },
  },
}

/** The research was moved to a team the reader is not in. */
export const ResearchTransferred: Story = {
  args: {
    revocation: {
      scope: 'research',
      id: '9f3c1e5a-0000-4000-8000-000000000000',
      code: 'R7',
      name: 'Payment rails',
      reason: 'research_transferred',
    },
  },
}

/**
 * A reason this build has no copy for, so it falls through to the bare
 * sentence. `team_deleted` is the concrete case: the spec lists it among the
 * emitted reasons, the server does not send it yet, and when it does this is
 * what the reader will get — a true statement that fails to say the team was
 * deleted. An empty or unrecognised reason lands in the same branch.
 */
export const UnhandledReason: Story = {
  args: {
    revocation: {
      scope: 'team',
      id: 'team-1',
      name: 'Northwind Ops',
      reason: 'team_deleted',
    },
  },
}

/**
 * The event arrived without a name — `useAccessRevoked` substitutes
 * “this workspace” rather than rendering empty quotes. The title still carries
 * the code, which is the part worth pasting into a request to be re-invited.
 */
export const NameUnavailable: Story = {
  args: {
    revocation: {
      scope: 'research',
      id: '9f3c1e5a-0000-4000-8000-000000000000',
      code: 'R7',
      name: 'this workspace',
      reason: 'research_transferred',
    },
  },
}

/**
 * Research names are user input and this one is long. The title stays short
 * because it prefers the code; the name lands in the description, which
 * `EmptyState` caps at 40ch — so the paragraph grows downwards and the layout
 * holds.
 */
export const LongResearchName: Story = {
  args: {
    revocation: {
      scope: 'research',
      id: '9f3c1e5a-0000-4000-8000-000000000000',
      code: 'R7',
      name: 'Payment rails — cross-border settlement, FX exposure and the reconciliation window we keep arguing about',
      reason: 'research_transferred',
    },
  },
}

/**
 * The same length in the worse place. A team has no short code, so its name is
 * the title — and `.empty-title` has no `overflow-wrap`, unlike the 40ch
 * description below it. This is the story to open when checking that a long
 * name cannot push the notice sideways.
 */
export const LongTeamName: Story = {
  args: {
    revocation: {
      scope: 'team',
      id: 'team-1',
      name: 'Northwind Operations, Settlement & Treasury Platform Working Group',
      reason: 'removed_from_team',
    },
  },
}

/**
 * A single unbroken token — a pasted URL or an identifier used as a name. There
 * is nothing to wrap at, which is the case that breaks a container rather than
 * merely filling it.
 */
export const UnbreakableName: Story = {
  args: {
    revocation: {
      scope: 'team',
      id: 'team-1',
      name: 'northwind-operations-settlement-and-treasury-platform-working-group-2026',
      reason: 'removed_from_team',
    },
  },
}

/**
 * Narrow viewport: `.empty-state` drops to smaller padding and the action
 * button takes the full width. Nothing else in this region has columns to
 * collapse.
 */
export const Mobile: Story = {
  args: {
    revocation: {
      scope: 'research',
      id: '9f3c1e5a-0000-4000-8000-000000000000',
      code: 'R7',
      name: 'Payment rails',
      reason: 'research_transferred',
    },
  },
  parameters: { viewport: { defaultViewport: 'mobile' } },
}
