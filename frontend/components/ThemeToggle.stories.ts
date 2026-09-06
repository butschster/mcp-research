import type { Meta, StoryObj } from '@storybook/vue3'
import ThemeToggle from './ThemeToggle.vue'

/**
 * The light/dark switch. One button, whose icon and label both name the theme it
 * would move to rather than the one you are in — "Switch to dark theme" is
 * actionable; "Dark theme" is a riddle about whether it is a state or a verb.
 *
 * **These stories are live.** It calls the real `useTheme`, so clicking one
 * flips the whole Storybook preview and writes the choice to `localStorage`
 * under the same key the product uses. Use the toolbar's paintbrush to put it
 * back; the decorator re-applies the global on the next render.
 *
 * `size="sm"` exists for one place: a strip that cannot afford a 36px control.
 * The share banner is a sticky 36px band, and on a phone `.btn-icon` grows to
 * 44px, which would make the band taller than the content it labels. The visual
 * box drops to `--control-h-sm`; the **hit box stays 44px** through a `::before`
 * overlay that reaches 6px above and below, so what got smaller is the drawing
 * and not the target. Both sizes are below.
 */
const meta: Meta<typeof ThemeToggle> = {
  title: 'Base/ThemeToggle',
  component: ThemeToggle,
  tags: ['autodocs'],
  argTypes: {
    size: { control: 'inline-radio', options: ['default', 'sm'] },
  },
}
export default meta
type Story = StoryObj<typeof ThemeToggle>

/** 36px, as it sits in the auth shell and the API docs header. */
export const Default: Story = {
  args: { size: 'default' },
}

/** 32px, as it sits in the share banner and on the standalone share screens. */
export const Small: Story = {
  args: { size: 'sm' },
}

/**
 * The two side by side, with their hit boxes drawn.
 *
 * The dashed outline is added by this story, not by the component — it marks
 * where a tap still lands. The small one's box is wider and taller than its
 * border, which is the whole trick and is invisible in a static screenshot of
 * either variant alone.
 */
export const BothSizes: Story = {
  parameters: { controls: { disable: true } },
  render: () => ({
    components: { ThemeToggle },
    template: `
      <div style="display: flex; align-items: center; gap: 4rem;">
        <div style="text-align: center;">
          <div style="outline: 1px dashed rgba(var(--color-primary-rgb), .5); display: inline-flex; padding: 0;">
            <ThemeToggle />
          </div>
          <p style="margin: .5rem 0 0; font-size: var(--type-xs); color: var(--color-text-faint);">default — 36px</p>
        </div>
        <div style="text-align: center;">
          <div style="outline: 1px dashed rgba(var(--color-primary-rgb), .5); display: inline-flex; padding: 6px 0;">
            <ThemeToggle size="sm" />
          </div>
          <p style="margin: .5rem 0 0; font-size: var(--type-xs); color: var(--color-text-faint);">sm — 32px box, 44px target</p>
        </div>
      </div>
    `,
  }),
}

/**
 * Where the small one actually lives: the right-hand end of a 36px band, beside
 * the activity indicator.
 *
 * This is the mock of the share banner's tools cluster, not the banner itself —
 * see `Share/Banner` for the real thing. It is here because the size only reads
 * as correct against the band it has to fit inside: at 36px the toggle would set
 * the height of the strip instead of sitting in it.
 */
export const InAStrip: Story = {
  parameters: { controls: { disable: true } },
  render: () => ({
    components: { ThemeToggle },
    template: `
      <div style="display: flex; align-items: center; justify-content: space-between; gap: var(--space-3); min-height: 36px; padding: 0 var(--space-4); background: var(--color-surface); border-bottom: 1px solid var(--color-border);">
        <span style="font-size: var(--type-xs); color: var(--color-text-muted);">
          <strong style="color: var(--color-text);">Read-only shared view</strong> — shared by Elena Marsh
        </span>
        <ThemeToggle size="sm" />
      </div>
    `,
  }),
}

/**
 * ≤768px, where `system.css` grows every `.btn-icon` to the touch height. The
 * small variant opts out of that and keeps its 32px box; the overlay is what
 * makes the tap land. Compare the two here — on a phone the default one is
 * visibly taller.
 */
export const Mobile: Story = {
  parameters: { viewport: { defaultViewport: 'mobile' }, controls: { disable: true } },
  render: () => ({
    components: { ThemeToggle },
    template: `
      <div style="display: flex; align-items: center; gap: var(--space-4);">
        <ThemeToggle />
        <ThemeToggle size="sm" />
      </div>
    `,
  }),
}
