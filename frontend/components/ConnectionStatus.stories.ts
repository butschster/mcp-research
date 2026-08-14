import type { Meta, StoryObj } from '@storybook/vue3'
import ConnectionStatus from './ConnectionStatus.vue'

/**
 * The footer's answer to "is what I am looking at still true?".
 *
 * Five states, three appearances. Grace looks exactly like Live and Stale looks
 * exactly like Reconnecting — that is the design, not an oversight: a one-second
 * wifi handover must not repaint the chrome, and the escalation past twenty
 * seconds is carried by a toast rather than by the dot changing again.
 *
 * These stories drive the shipped component. It used to be un-story-able —
 * it opened its own WebSocket and its state was whatever the network did — so
 * this file held a hand-drawn copy of it that could drift from the real thing
 * without anything failing. The component now takes props and `app.vue` passes
 * them down from `useConnectionBanner()`, so what is below is the real markup,
 * the real classes and the real copy.
 */
const meta: Meta<typeof ConnectionStatus> = {
  title: 'Base/ConnectionStatus',
  component: ConnectionStatus,
  tags: ['autodocs'],
  argTypes: {
    state: {
      control: 'select',
      options: ['live', 'grace', 'reconnecting', 'stale', 'offline'],
    },
    reason: {
      control: 'inline-radio',
      options: [null, 'gave_up', 'auth'],
    },
    lastSyncedAt: { control: 'number' },
  },
}
export default meta
type Story = StoryObj<typeof ConnectionStatus>

// Fixed, so the tooltip reads the same on every run rather than drifting with
// the clock. Rendered in the reader's locale and zone, as the product does.
const SYNCED_AT = new Date('2026-08-14T14:02:00').getTime()

/** Healthy. The quietest thing in the product, and deliberately so. */
export const Live: Story = {
  args: { state: 'live', lastSyncedAt: SYNCED_AT },
}

/**
 * Under five seconds of gap. Indistinguishable from Live — a lid closing, a
 * wifi handover and a `make build` restart all land here and none of them is
 * worth a reader's attention.
 */
export const Grace: Story = {
  args: { state: 'grace', lastSyncedAt: SYNCED_AT },
}

/** Past the grace window: amber, pulsing, and now clickable to retry. */
export const Reconnecting: Story = {
  args: { state: 'reconnecting', lastSyncedAt: SYNCED_AT },
}

/**
 * First connect, before the socket has ever opened — there is no last sync to
 * report, so the tooltip drops that sentence instead of inventing a time.
 */
export const ReconnectingNeverSynced: Story = {
  args: { state: 'reconnecting', lastSyncedAt: null },
}

/**
 * Twenty seconds in. Identical to Reconnecting by design: the escalation is a
 * toast, because a dot that changes twice teaches nobody anything.
 */
export const Stale: Story = {
  args: { state: 'stale', lastSyncedAt: SYNCED_AT },
}

/** The client stopped retrying. Grey, still, and one click from another attempt. */
export const OfflineGaveUp: Story = {
  args: { state: 'offline', reason: 'gave_up', lastSyncedAt: SYNCED_AT },
}

/**
 * Signed out, or the key was revoked. The only offline state that is **not**
 * actionable: a credential that stopped being valid is not fixed by trying
 * again, and a control that does nothing when pressed is worse than one that is
 * not offered. Recovery is the toast's "Sign in again".
 */
export const OfflineAuth: Story = {
  args: { state: 'offline', reason: 'auth', lastSyncedAt: SYNCED_AT },
}

/**
 * At 768px and below, `main.css` hides `.ws-label` and the state would be
 * carried by the colour of a 6px dot and nothing else. The `sr-only` text and
 * the computed `aria-label` are what survive — inspect the DOM here, not the
 * pixels.
 */
export const MobileLabelHidden: Story = {
  args: { state: 'reconnecting', lastSyncedAt: SYNCED_AT },
  parameters: { viewport: { defaultViewport: 'mobile' } },
}

/** Every state side by side, which is where the three appearances show. */
export const AllStates: Story = {
  render: () => ({
    components: { ConnectionStatus },
    setup: () => ({ syncedAt: SYNCED_AT }),
    template: `
      <div style="display: flex; flex-direction: column; gap: var(--space-3); align-items: flex-start;">
        <ConnectionStatus state="live" :last-synced-at="syncedAt" />
        <ConnectionStatus state="grace" :last-synced-at="syncedAt" />
        <ConnectionStatus state="reconnecting" :last-synced-at="syncedAt" />
        <ConnectionStatus state="stale" :last-synced-at="syncedAt" />
        <ConnectionStatus state="offline" reason="gave_up" :last-synced-at="syncedAt" />
        <ConnectionStatus state="offline" reason="auth" :last-synced-at="syncedAt" />
      </div>
    `,
  }),
}

/**
 * In place, in the footer it actually lives in — the one context where its
 * quietness can be judged.
 */
export const InFooter: Story = {
  render: (args) => ({
    components: { ConnectionStatus },
    setup: () => ({ args }),
    template: `
      <footer style="border-top: 1px solid var(--color-border); padding: var(--space-4) var(--space-6);">
        <div style="display: flex; align-items: center; justify-content: space-between;">
          <span class="card-meta">Research</span>
          <div style="display: flex; align-items: center; gap: var(--space-4);">
            <ConnectionStatus v-bind="args" />
          </div>
        </div>
      </footer>
    `,
  }),
  args: { state: 'reconnecting', lastSyncedAt: SYNCED_AT },
}
