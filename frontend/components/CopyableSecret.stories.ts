import type { Meta, StoryObj } from '@storybook/vue3'
import { onUnmounted } from 'vue'
import CopyableSecret from './CopyableSecret.vue'
import ToastHost from './ToastHost.vue'
import { mockShareUrl } from '../__mocks__/share'

/**
 * A value the server will not show again, with a copy button.
 *
 * The clipboard fallback is the reason this is a component rather than three
 * lines of markup. `navigator.clipboard` is simply absent on plain HTTP, which
 * is a normal way to run this product on a LAN, and a dead Copy button on the
 * one screen that shows a secret once is the worst possible place to find that
 * out. The behaviour existed in two places already and would have drifted.
 *
 * The success toast comes from the real `useToasts()`, so these stories mount a
 * `ToastHost` beside the block — pressing Copy raises a real toast.
 */
const meta: Meta<typeof CopyableSecret> = {
  title: 'Base/CopyableSecret',
  component: CopyableSecret,
  tags: ['autodocs'],
  decorators: [
    () => ({
      components: { ToastHost },
      template: '<div style="max-width: 560px;"><story /><ToastHost /></div>',
    }),
  ],
  argTypes: {
    value: { control: 'text' },
    hint: { control: 'text' },
    toast: { control: 'text' },
    onCopied: { action: 'copied' },
  },
}
export default meta
type Story = StoryObj<typeof CopyableSecret>

/** A share URL: long, opaque and impossible to retype, which is the whole
 *  reason the copy affordance is the feature rather than a convenience. */
export const ShareLink: Story = {
  args: { value: mockShareUrl, toast: 'Share link copied' },
}

/** With a hint. The one that matters most is the password warning — a link and
 *  its password in the same message is one forward away from being no password
 *  at all. */
export const WithHint: Story = {
  args: {
    value: mockShareUrl,
    hint: 'Send the password separately — not in the same message.',
    toast: 'Share link copied',
  },
}

/**
 * Copying is not available: a plain-HTTP deployment on a LAN, where the
 * clipboard API is absent entirely.
 *
 * The block swaps the code for a readonly input, focuses and selects it, and
 * says why — rather than leaving a button that silently does nothing. The hint,
 * if there was one, is replaced: the instruction to select the text manually is
 * the more urgent of the two.
 *
 * The story removes `navigator.clipboard` for its own lifetime and presses Copy
 * for you; the property is restored when the story unmounts.
 */
export const ClipboardUnavailable: Story = {
  render: () => ({
    components: { CopyableSecret },
    setup() {
      // An own property shadows the one on Navigator.prototype; deleting it
      // uncovers the real getter again when the story goes away.
      Object.defineProperty(navigator, 'clipboard', { value: undefined, configurable: true })
      onUnmounted(() => {
        delete (navigator as { clipboard?: unknown }).clipboard
      })
      return { link: mockShareUrl }
    },
    template: `<CopyableSecret :value="link" hint="Send the password separately — not in the same message." />`,
  }),
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    const copy = canvasElement.querySelector('.copy-btn') as HTMLElement | null
    copy?.click()
  },
}

/** A value with nowhere to break: it wraps inside the block rather than pushing
 *  the Copy button off the end. Share tokens are hex, so this is the shape of
 *  every real one, only longer. */
export const VeryLongValue: Story = {
  args: {
    value: `https://research.intruforce.com/s/${'9f2c7b1e40a54d8fbb63d21c7e05c41a'.repeat(4)}`,
  },
}

/** An empty value — a create that answered without a URL. The block still
 *  renders rather than collapsing, because the caller decides whether to show it
 *  at all, and an empty frame is a visible fault where a missing one is not. */
export const EmptyValue: Story = {
  args: { value: '' },
}
