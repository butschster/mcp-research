import type { Meta, StoryObj } from '@storybook/vue3'
import ToastHost from './ToastHost.vue'
import { useToasts } from '~/composables/useToasts'

/**
 * The app's notifications, in one corner.
 *
 * The host is mounted once in `app.vue`; a page raises a message with
 * `useToasts()` and never renders one itself. A success dismisses itself, a
 * failure waits — an error usually carries the only way to retry, and taking
 * that away on a timer loses the reader.
 *
 * These stories drive the real composable, so pressing the buttons is the
 * component working, not a mock of it.
 */
const meta: Meta<typeof ToastHost> = {
  title: 'ToastHost',
  component: ToastHost,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
}
export default meta
type Story = StoryObj<typeof ToastHost>

function stage(setup: (t: ReturnType<typeof useToasts>) => void, buttons: string) {
  return () => ({
    components: { ToastHost },
    setup() {
      const toasts = useToasts()
      toasts.clear()
      setup(toasts)
      return { toasts }
    },
    template: `
      <div style="min-height: 320px; padding: var(--space-6); display: flex; gap: var(--space-2); flex-wrap: wrap; align-items: flex-start;">
        ${buttons}
        <ToastHost />
      </div>
    `,
  })
}

/** A download finished. Dismisses itself after five seconds — unless the
 *  pointer is over it, which holds the countdown. */
export const Success: Story = {
  render: stage(
    (t) => t.success('R1 — Competitive Landscape.zip', 'Downloaded'),
    `<button class="btn btn-sm" @click="toasts.success('R1 — Competitive Landscape.zip', 'Downloaded')">Raise a success</button>`,
  ),
}

/** A failure with the one act that answers it. It stays until dismissed. */
export const ErrorWithAction: Story = {
  render: stage(
    (t) =>
      t.error('the server could not build the archive. Nothing was downloaded.', 'Export failed', {
        label: 'Try again',
        onClick: () => {},
      }),
    `<button class="btn btn-sm" @click="toasts.error('the server could not build the archive.', 'Export failed', { label: 'Try again', onClick: () => {} })">Raise a failure</button>`,
  ),
}

/** Progress that is not an outcome — the slow-download hint. */
export const Info: Story = {
  render: stage(
    (t) => t.push({ message: 'Still working — a large research takes a moment.' }),
    `<button class="btn btn-sm" @click="toasts.push({ message: 'Still working — a large research takes a moment.' })">Raise a hint</button>`,
  ),
}

/** All three at once, which is where the stacking and the variant edges show. */
export const Stacked: Story = {
  render: stage(
    (t) => {
      t.success('R1 — Competitive Landscape.zip', 'Downloaded')
      t.push({ message: 'Still working — a large research takes a moment.' })
      t.error('your session expired. Sign in again to download.', 'Export failed', {
        label: 'Sign in',
        onClick: () => {},
      })
    },
    `<button class="btn btn-sm" @click="toasts.clear()">Clear</button>`,
  ),
}

/** Five raised, four shown: the oldest goes, because a stack taller than the
 *  corner is a wall rather than a notification. */
export const OverflowsToFour: Story = {
  render: stage(
    (t) => {
      for (let i = 1; i <= 5; i++) t.error(`Failure number ${i}.`, 'Export failed')
    },
    `<button class="btn btn-sm" @click="toasts.clear()">Clear</button>`,
  ),
}

/** A long Cyrillic filename with nothing to break at. */
export const LongMessage: Story = {
  render: stage(
    (t) =>
      t.success(
        'R2 — Исследование рынка цены и упаковка с очень длинным именем без пробелов_и_подчёркиваний.zip',
        'Downloaded',
      ),
    `<button class="btn btn-sm" @click="toasts.clear()">Clear</button>`,
  ),
}
