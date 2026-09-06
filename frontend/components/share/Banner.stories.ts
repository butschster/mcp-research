import type { Meta, StoryObj } from '@storybook/vue3'
import ShareBanner from './Banner.vue'
import {
  includeDefault,
  includeEntriesOnly,
  includeEverything,
} from '../../__mocks__/share'

const DAY = 86_400_000
const inDays = (days: number) => new Date(Date.now() + days * DAY).toISOString()

/**
 * The permanent strip across a shared view.
 *
 * It exists so nobody mistakes this page for their own workspace — the page
 * below it is, deliberately, the same page the owner sees, and without the strip
 * a client would have no way of knowing they are looking at somebody's private
 * research through a keyhole.
 *
 * Two things are always on it: that this is read-only, and what the link
 * reaches. The rest — why the page cannot be edited, when the link lapses —
 * lives behind the disclosure, because it is read once and then never again.
 *
 * The byline is dropped entirely when the share has no owner name. A banner that
 * says "shared by" followed by a blank is worse than one that says nothing.
 *
 * **The right-hand cluster sits outside the disclosure button**, not inside it:
 * the activity blip and the theme toggle. A button inside a button is invalid
 * markup and unreachable from the keyboard, and the toggle has to be reachable —
 * the share shell is chromeless, so this strip is the only place on the whole
 * page a visitor can change the theme.
 *
 * The owner's name appears twice, and that is deliberate. Below 480px the
 * byline is dropped from the strip so it stays one line on a phone, and the
 * disclosure carries the name instead — one tap away rather than gone.
 */
const meta: Meta<typeof ShareBanner> = {
  title: 'Share/Banner',
  component: ShareBanner,
  tags: ['autodocs'],
  parameters: { layout: 'fullscreen' },
  argTypes: {
    ownerName: { control: 'text' },
    include: { control: 'object' },
    expiresAt: { control: 'text' },
    live: { control: 'boolean' },
  },
  args: {
    ownerName: 'Elena Marsh',
    include: includeDefault,
    expiresAt: inDays(23),
    live: false,
  },
}
export default meta
type Story = StoryObj<typeof ShareBanner>

/** The ordinary link: entries, roadmaps and downloads, lapsing in three weeks. */
export const Default: Story = {}

/** Everything the owner could include. Entries lead the list because they are
 *  what the reader came for; the rest read as additions to them. */
export const EverythingIncluded: Story = {
  args: { include: includeEverything, expiresAt: null },
}

/** The narrowest link there is. "Entries only" rather than a list of one — a
 *  single-item list invites the question of what the other items would be. */
export const EntriesOnly: Story = {
  args: { include: includeEntriesOnly },
}

/** No owner name: the byline is absent rather than blank, and the strip still
 *  says the two things it exists to say. */
export const WithoutOwnerName: Story = {
  args: { ownerName: '' },
}

/** Tomorrow. Said in days rather than as a date, because "tomorrow" is
 *  actionable and "12 Sep" has to be worked out. */
export const ExpiringTomorrow: Story = {
  args: { expiresAt: inDays(1) },
}

/** Beyond a month the phrasing switches to a date: "in 41 days" is a number
 *  nobody can place, and at that distance the exact day is the useful form. */
export const ExpiringInFortyDays: Story = {
  args: { expiresAt: inDays(41) },
}

/** No end date. The disclosure still names the thing that can happen — the
 *  person who shared it can turn it off — so "no end date" is not read as
 *  "permanent". */
export const NeverExpires: Story = {
  args: { expiresAt: null },
}

/** Something changed under the reader's hands in the last few seconds. The blip
 *  is the only acknowledgement; nothing scrolls and nothing takes focus,
 *  because a repaint is data changing, not navigation. */
export const Live: Story = {
  args: { live: true, include: includeEverything },
}

/** A 200-character owner name with no spaces — the strip wraps rather than
 *  pushing the indicator off the end, and the disclosure survives it too. */
export const VeryLongOwnerName: Story = {
  args: {
    ownerName:
      'Отдел-стратегических-исследований-и-конкурентного-анализа-департамента-развития-продукта',
    include: includeEverything,
    expiresAt: inDays(3),
  },
}

/**
 * The disclosure open, which is where the visitor is told the two things the
 * strip has no room for: why the page cannot be edited, and when the link
 * lapses. The owner's name is repeated here — see the note on the component
 * above — so a phone that has dropped the byline still has it.
 */
export const Expanded: Story = {
  parameters: { docs: { story: { autoplay: true } } },
  args: { include: includeEverything, expiresAt: inDays(3) },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await openDisclosure(canvasElement)
  },
}

/** ≤768px: the contents clause is dropped from the strip and the disclosure
 *  becomes the only place it is said. It pushes content down rather than
 *  overlaying it — an overlay over content the visitor has not seen yet is the
 *  wrong trade. Open the disclosure to see it.
 *
 *  The theme toggle stays. It is `size="sm"`, which is the reason that variant
 *  exists: at 36px — what `.btn-icon` grows to on a phone — it would set the
 *  height of the strip instead of sitting inside it. */
export const Mobile: Story = {
  parameters: { viewport: { defaultViewport: 'mobile' } },
  args: { include: includeEverything },
}

/**
 * 375px with the disclosure open: the byline is gone from the strip, which is
 * now one line, and the same name is in the panel below. Below 480px is the only
 * place these two facts are visible at once.
 */
export const MobileExpanded: Story = {
  parameters: { viewport: { defaultViewport: 'mobile' }, docs: { story: { autoplay: true } } },
  args: { include: includeEverything, ownerName: 'Марат Ибрагимов' },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await openDisclosure(canvasElement)
  },
}

/**
 * Clicks the strip open once it is mounted.
 *
 * There is no `@storybook/test` in this project, so the catalogue polls — the
 * same helper shape `ActionMenu.stories.ts` uses. The disclosure is found by
 * `aria-controls`, because the theme toggle beside it is also a button.
 */
async function openDisclosure(root: HTMLElement): Promise<void> {
  for (let i = 0; i < 50; i++) {
    const trigger = root.querySelector<HTMLElement>('button[aria-controls="share-banner-detail"]')
    if (trigger) {
      trigger.click()
      return
    }
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
  throw new Error('Share banner disclosure never appeared')
}

/** All of it side by side, which is how the wording differences show. */
export const AllStates: Story = {
  render: () => ({
    components: { ShareBanner },
    setup: () => ({
      rows: [
        { label: 'Everything, never expires', props: { ownerName: 'Elena Marsh', include: includeEverything, expiresAt: null } },
        { label: 'Entries only', props: { ownerName: 'Elena Marsh', include: includeEntriesOnly, expiresAt: inDays(23) } },
        { label: 'No owner name', props: { include: includeDefault, expiresAt: inDays(23) } },
        { label: 'Expires tomorrow', props: { ownerName: 'Марат Ибрагимов', include: includeDefault, expiresAt: inDays(1) } },
        { label: 'Expires in 40 days', props: { ownerName: 'Elena Marsh', include: includeDefault, expiresAt: inDays(40) } },
        { label: 'Live', props: { ownerName: 'Elena Marsh', include: includeEverything, expiresAt: null, live: true } },
      ],
    }),
    template: `
      <div style="display: flex; flex-direction: column; gap: var(--space-5);">
        <div v-for="row in rows" :key="row.label">
          <p style="margin: 0 0 var(--space-1); font-size: var(--type-xs); color: var(--color-text-muted);">{{ row.label }}</p>
          <ShareBanner v-bind="row.props" />
        </div>
      </div>
    `,
  }),
}
