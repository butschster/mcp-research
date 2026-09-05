import type { Meta, StoryObj } from '@storybook/vue3'
import { onUnmounted } from 'vue'
import ApiDocsHeader from './ApiDocsHeader.vue'

/**
 * The chrome bar of `/api-docs`. It is the only product surface on that page —
 * the rest is the rendered OpenAPI document — so it carries the way back to the
 * app, the address the reader is calling, and the raw file for whoever came for
 * the file rather than the page.
 *
 * It stays on screen in every state of the page, including while the document is
 * loading and when it failed, so the raw-file links work even when the viewer
 * does not.
 *
 * The copy button copies the **URL of the specification**, not its contents: the
 * reader's next move is `openapi-generator generate -i <url>`, and the YAML
 * button already opens the file for anyone who wanted the file.
 */
const meta: Meta<typeof ApiDocsHeader> = {
  title: 'Layout/ApiDocsHeader',
  component: ApiDocsHeader,
  tags: ['autodocs'],
  // Page chrome, like AppNav: it spans the window and its bottom border is the
  // seam against the document, which a centred column misrepresents.
  parameters: { layout: 'fullscreen' },
  argTypes: {
    baseUrl: { control: 'text' },
    specUrlYaml: { control: 'text' },
    specUrlJson: { control: 'text' },
    showApiKeysLink: { control: 'boolean' },
  },
  args: {
    baseUrl: 'https://research.example.com',
    specUrlYaml: 'https://research.example.com/api/openapi.yaml',
    specUrlJson: 'https://research.example.com/api/openapi.json',
    showApiKeysLink: true,
    signedIn: true,
  },
}
export default meta
type Story = StoryObj<typeof ApiDocsHeader>

export const Default: Story = {}

/**
 * Signed out, with accounts enabled. The link is the same destination through
 * the sign-in page, and says so — an unannounced bounce to `/login` from a
 * button labelled "API keys" reads as a failure. This label is also 46px wider,
 * which is what the narrow story is really testing.
 */
export const SignedOut: Story = { args: { signedIn: false } }

/**
 * With `auth_enabled: false` there are no accounts, so there are no API keys and
 * `/settings` is a dead end for this reader. The link is absent and nothing
 * takes its place.
 */
export const NoAccounts: Story = { args: { showApiKeysLink: false } }

// The clipboard is the one thing these stories cannot take as it comes. In a
// Storybook iframe `navigator.clipboard` may be present and still reject —
// permissions, or simply an unfocused frame — so a story that clicked Copy and
// hoped would document the success state or the refusal at random. Each of the
// two below installs the outcome it is about, for its own lifetime, the way
// `CopyableSecret.stories.ts` already does.
function stubClipboard(value: unknown) {
  // An own property shadows the one on Navigator.prototype; deleting it
  // uncovers the real getter again when the story goes away.
  Object.defineProperty(navigator, 'clipboard', { value, configurable: true })
  onUnmounted(() => {
    delete (navigator as { clipboard?: unknown }).clipboard
  })
}

function pressCopy({ canvasElement }: { canvasElement: HTMLElement }) {
  const copy = canvasElement.querySelector('.btn-icon') as HTMLElement | null
  copy?.click()
}

/**
 * The moment after Copy: the icon swaps to a check, the tooltip and the
 * accessible name both become "Copied", and the live region says so once.
 *
 * The story presses Copy for you against a clipboard that accepts. **It lasts
 * two seconds** — the component resets itself, and it is the same two seconds a
 * reader gets, so what reverts under you here is the real behaviour rather than
 * a story that expired. Press it again to see it.
 */
export const Copied: Story = {
  parameters: {
    // A play function does not run in docs by default, and without this the
    // docs page would show a Copied story that is not copied.
    docs: { story: { autoplay: true } },
  },
  render: (args) => ({
    components: { ApiDocsHeader },
    setup() {
      stubClipboard({ writeText: async () => {} })
      return { args }
    },
    template: '<ApiDocsHeader v-bind="args" />',
  }),
  play: pressCopy,
}

/**
 * The clipboard refuses — a plain-HTTP deployment on a LAN, where
 * `navigator.clipboard` is simply absent, which is a normal way to run this
 * product.
 *
 * A second row appears under the bar carrying the address itself, selectable,
 * with a Dismiss beside it — the icon does not change, and the `role="status"`
 * span announces the same thing. The row holds the URL rather than pointing at
 * the one in the bar, because they are different strings and the bar hides its
 * one below 769px. It does not expire: the text is the only fallback the reader
 * has, and taking it away after two seconds takes away the point.
 *
 * The story removes `navigator.clipboard` for its own lifetime and presses Copy
 * for you; the property is restored when the story unmounts.
 */
export const ClipboardRefused: Story = {
  parameters: { docs: { story: { autoplay: true } } },
  render: (args) => ({
    components: { ApiDocsHeader },
    setup() {
      stubClipboard(undefined)
      return { args }
    },
    template: '<ApiDocsHeader v-bind="args" />',
  }),
  play: pressCopy,
}

/** A reverse-proxied host with a long path is ordinary. The address truncates
 *  rather than pushing the controls off the bar; the full value is in `title`. */
export const LongBaseUrl: Story = {
  args: {
    baseUrl: 'https://research.internal.example-corporation.com/tooling/knowledge/mcp-research',
    specUrlYaml: 'https://research.internal.example-corporation.com/tooling/knowledge/mcp-research/api/openapi.yaml',
    specUrlJson: 'https://research.internal.example-corporation.com/tooling/knowledge/mcp-research/api/openapi.json',
  },
}

/** A local run, where the address is the one a person will actually paste. */
export const Localhost: Story = {
  args: {
    baseUrl: 'http://localhost:8088',
    specUrlYaml: 'http://localhost:8088/api/openapi.yaml',
    specUrlJson: 'http://localhost:8088/api/openapi.json',
  },
}

/**
 * Below 769px the bar wraps to two rows and drops the address — it is only
 * actionable beside the try-it panel, which is barely usable at this width. The
 * keys link stays, because on a phone it is the only route from this page to a
 * credential.
 *
 * 375px, the width the rest of this catalogue tests narrow layouts at.
 */
export const Narrow: Story = {
  parameters: { viewport: { defaultViewport: 'mobile' } },
}

/**
 * In place, above the document it fronts — the only view in which the size of
 * the title can be judged. Scalar renders the specification's own title as a
 * large heading immediately below, so "API reference" is deliberately small:
 * two headlines competing would push the document's first line under the fold.
 * The block below stands in for that heading.
 */
export const InPlace: Story = {
  render: (args) => ({
    components: { ApiDocsHeader },
    setup: () => ({ args }),
    template: `
      <div style="background: var(--color-bg);">
        <ApiDocsHeader v-bind="args" />
        <div style="padding: var(--space-6);">
          <h1 style="margin: 0 0 var(--space-2); font-size: var(--type-3xl); font-weight: var(--weight-bold);">MCP Research API</h1>
          <p class="card-meta" style="margin: 0;">OpenAPI 3.1 &middot; 126 operations</p>
        </div>
      </div>
    `,
  }),
}
