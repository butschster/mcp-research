import type { Meta, StoryObj } from '@storybook/vue3'
import EmptyState from './EmptyState.vue'

const meta: Meta<typeof EmptyState> = {
  title: 'Base/EmptyState',
  component: EmptyState,
  tags: ['autodocs'],
  argTypes: {
    icon: { control: 'text' },
    title: { control: 'text' },
    description: { control: 'text' },
    command: { control: 'text' },
  },
}
export default meta
type Story = StoryObj<typeof EmptyState>

export const IconOnly: Story = {
  args: {
    icon: '\uD83D\uDD2C',
    title: 'No researches yet',
  },
}

export const WithDescription: Story = {
  args: {
    icon: '\uD83D\uDCCB',
    title: 'No entries found',
    description: 'Start a research session to generate entries automatically.',
  },
}

export const WithCommand: Story = {
  args: {
    icon: '\uD83D\uDE80',
    title: 'Get started with Research',
    description: 'Add this MCP server to Claude and run the initialization prompt.',
    command: 'Use the research/initialize prompt',
  },
}

export const MinimalTitle: Story = {
  args: {
    title: 'Nothing here',
  },
}

/**
 * A share link that does not work — revoked, expired, or mistyped.
 *
 * One screen for all three, because the server answers all three with the same
 * 404 and the UI must not speculate about which. There is deliberately no
 * button: the next action is "ask whoever sent it", which the copy names, and a
 * `Sign in` here would be exactly the wall this feature exists to avoid.
 */
export const ShareLinkDead: Story = {
  args: {
    icon: '🔗',
    title: "This link isn't available",
    description:
      'It may have been turned off, it may have expired, or the address may be incomplete — these links are long and easy to cut short. Ask the person who sent it for a new one.',
  },
}

/**
 * The server did not answer at all. A separate screen from the one above,
 * because "check how you copied the link" is the wrong instruction for a server
 * that is down — and here there *is* something to do, so the slot carries a
 * button.
 */
export const ShareLinkUnreachable: Story = {
  render: () => ({
    components: { EmptyState },
    template: `
      <EmptyState
        title="Couldn't open this link"
        description="The server didn't answer. The link is probably fine — try again in a moment."
      >
        <button class="btn btn-sm btn-primary">Try again</button>
      </EmptyState>
    `,
  }),
}

/**
 * A part of the research the share link does not include — reached only by
 * typing a URL, since the entry point is absent. The way back is into the
 * shared research, which is the one place this visitor can go.
 */
export const ShareLinkExcludedSection: Story = {
  render: () => ({
    components: { EmptyState },
    template: `
      <EmptyState
        title="Not part of this link"
        description="The person who shared this research didn't include this. Ask them if you need it."
      >
        <a class="btn btn-sm" href="/s/9f2c7b1e40a54d8fbb63d21c7e05c41a">Back to the research</a>
      </EmptyState>
    `,
  }),
}

/**
 * An entry deleted while it was being read: the fetch 404s after an
 * `entry.deleted` event arrives. The reassurance matters — the rest of the
 * research is still there, and a visitor with no account cannot tell that from
 * a bare error.
 */
export const ShareEntryRemoved: Story = {
  render: () => ({
    components: { EmptyState },
    template: `
      <EmptyState
        title="This entry was removed"
        description="It was deleted while you were reading. The rest of the research is still here."
      >
        <a class="btn btn-sm" href="/s/9f2c7b1e40a54d8fbb63d21c7e05c41a">Back to the research</a>
      </EmptyState>
    `,
  }),
}

/**
 * An empty shared research. "You can leave it open" is the honest next action
 * for a read-only visitor and it is true — the socket is live and the page
 * repaints itself. Without an owner name the sentence loses its subject and
 * becomes "Nothing has been added to this research yet."
 */
export const ShareResearchEmpty: Story = {
  args: {
    icon: '📄',
    title: 'Nothing here yet',
    description:
      "Elena Marsh hasn't added anything to this research yet. This page updates by itself when they do — you can leave it open.",
  },
}
