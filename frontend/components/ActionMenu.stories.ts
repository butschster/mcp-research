import type { Meta, StoryObj } from '@storybook/vue3'
import ActionMenu from './ActionMenu.vue'

/**
 * The `⋯` that holds a row's or a page's less-used actions.
 *
 * Every story here opens the menu before you see it. A closed menu is a 16px
 * button, and the button is not what is being judged — the panel is: its raised
 * surface, the icon column the items share, the header tile, and the rule
 * between the safe verbs and the destructive one.
 *
 * Items go in the default slot and carry `action-menu-item`; a
 * `action-menu-divider` separates groups, and `action-menu-header` is a
 * non-interactive fact at the top. All three are styled from here through
 * `:slotted()`, so a consumer only writes the class.
 *
 * `label` is the one exception to "the trigger is not what is being judged". It
 * turns the `⋯` into an ordinary `.btn` with a name and a chevron, for a menu
 * that has to be the thing the eye lands on rather than the thing found last —
 * the visitor's "Views" on a shared project, which is the only way to the
 * knowledge graph and the mind map from there. With a label the tooltip is
 * dropped: a button that says what it is needs no `title`, and a `title`
 * repeating the label is read out twice.
 *
 * It closes on an outside click, on Escape, and after any item is clicked — and
 * it exposes `close()` and `focusTrigger()` for what happens next. The second
 * is not decoration: an item unmounts the instant the menu closes, so anything
 * that saved the active element on open, a modal restoring focus above all, is
 * holding a detached node by the time it tries to go back, and the reader is
 * dropped at the top of the page. The entry page calls it after its delete
 * confirmation resolves either way. Neither is shown as a story here, because a
 * focus ring in a static frame proves nothing.
 */
const meta: Meta<typeof ActionMenu> = {
  title: 'Navigation/ActionMenu',
  component: ActionMenu,
  tags: ['autodocs'],
  argTypes: {
    title: { control: 'text' },
    label: { control: 'text' },
    align: { control: 'inline-radio', options: ['left', 'right'] },
    width: { control: 'inline-radio', options: ['default', 'wide'] },
  },
  parameters: {
    // Without this the docs page shows six closed triggers, because a play
    // function does not run in docs by default.
    docs: { story: { autoplay: true } },
  },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await openMenu(canvasElement)
  },
}
export default meta
type Story = StoryObj<typeof ActionMenu>

const downloadIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>`
const gearIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>`
const trashIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/></svg>`
const clockIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v5h5"/><path d="M3.05 13A9 9 0 1 0 6 5.3L3 8"/><path d="M12 7v5l4 2"/></svg>`
const graphIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.6" y1="10.7" x2="15.4" y2="6.3"/><line x1="8.6" y1="13.3" x2="15.4" y2="17.7"/></svg>`
const mapIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="9" width="6" height="6" rx="1"/><rect x="15" y="3" width="6" height="5" rx="1"/><rect x="15" y="16" width="6" height="5" rx="1"/><path d="M9 12h3v-6.5h3"/><path d="M12 12v6.5h3"/></svg>`

// The panel is absolutely positioned, so stories need room below the trigger.
const roomBelow = (story: string) => `<div style="min-height: 300px;">${story}</div>`

/** One verb. The panel still earns its keep by keeping it out of the header. */
export const SingleAction: Story = {
  args: { title: 'More actions', align: 'right', width: 'default' },
  render: (args) => ({
    components: { ActionMenu },
    setup: () => ({ args }),
    template: roomBelow(`
      <ActionMenu v-bind="args">
        <a href="#" class="action-menu-item">${downloadIcon} Export</a>
      </ActionMenu>
    `),
  }),
}

/**
 * Verbs only, which is what every consumer other than the document menu is:
 * the team member row, the invite row, the session and research headers.
 *
 * Worth keeping next to the header stories. The `.action-menu-header` rule was
 * added for one caller, and a slotted rule that disturbed the plain case would
 * disturb four screens at once.
 *
 * The rule above `Delete` is a `border-top` on the divider rather than a filled
 * block with air on both sides: a rule at the edge of a list is an edge, and
 * three items separated by two floating rules read as scaffolding.
 */
export const SeveralActions: Story = {
  args: { title: 'More actions', align: 'right', width: 'default' },
  render: (args) => ({
    components: { ActionMenu },
    setup: () => ({ args }),
    template: roomBelow(`
      <ActionMenu v-bind="args">
        <button class="action-menu-item">${gearIcon} Details</button>
        <a href="#" class="action-menu-item">${downloadIcon} Export</a>
        <button class="action-menu-item" disabled>${downloadIcon} Download JSON</button>
        <div class="action-menu-divider" role="separator"></div>
        <button class="action-menu-item action-menu-item--danger">${trashIcon} Delete</button>
      </ActionMenu>
    `),
  }),
}

export const AlignedLeft: Story = {
  args: { title: 'More actions', align: 'left', width: 'default' },
  render: (args) => ({
    components: { ActionMenu },
    setup: () => ({ args }),
    template: roomBelow(`
      <ActionMenu v-bind="args">
        <a href="#" class="action-menu-item">${downloadIcon} Export</a>
        <button class="action-menu-item">${gearIcon} Details</button>
      </ActionMenu>
    `),
  }),
}

/**
 * The two widths side by side, with the content that forced the choice.
 *
 * `default` is 180px and `wide` is 232px — 182px of text column, measured to
 * hold a Russian full name on one line. Both panels carry the same one-line
 * name, and the left one ellipsises it. Seen apart, either width looks
 * deliberate; the prop is only legible as a choice when both are on screen.
 */
export const BothWidths: Story = {
  parameters: { controls: { disable: true } },
  render: () => ({
    components: { ActionMenu },
    template: roomBelow(`
      <div style="display: flex; gap: 6rem;">
        <div>
          <p style="margin: 0 0 .5rem; font-size: .75rem; color: var(--color-text-faint);">width="default" — 180px</p>
          <ActionMenu title="More actions" align="left" width="default">
            <div class="action-menu-header">
              <p style="margin:0; font-size: var(--type-sm); overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">Александра Константинопольская</p>
            </div>
            <button class="action-menu-item">${clockIcon} Revision history</button>
            <button class="action-menu-item">${downloadIcon} Download .md</button>
          </ActionMenu>
        </div>
        <div>
          <p style="margin: 0 0 .5rem; font-size: .75rem; color: var(--color-text-faint);">width="wide" — 232px</p>
          <ActionMenu title="More actions" align="left" width="wide">
            <div class="action-menu-header">
              <p style="margin:0; font-size: var(--type-sm); overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">Александра Константинопольская</p>
            </div>
            <button class="action-menu-item">${clockIcon} Revision history</button>
            <button class="action-menu-item">${downloadIcon} Download .md</button>
          </ActionMenu>
        </div>
      </div>
    `),
  }),
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await openMenu(canvasElement, 0)
    await openMenu(canvasElement, 1)
  },
}

/**
 * A fact above the verbs — the raw `.action-menu-header` markup, without the
 * one component that currently uses it.
 *
 * Inset rather than full-bleed, because the panel's own top padding would leave
 * a 4px strip of raised colour above a full-bleed tile and read as a mistake.
 * Recessed rather than lightened, because hover in this product goes lighter
 * and a lighter tile reads as already-hovered. It has no hover and no pointer:
 * inertness is carried by the absence of feedback, which is what a person
 * actually tests it with.
 */
export const WithAHeader: Story = {
  args: { title: 'More actions', align: 'right', width: 'wide' },
  render: (args) => ({
    components: { ActionMenu },
    setup: () => ({ args }),
    template: roomBelow(`
      <ActionMenu v-bind="args">
        <div class="action-menu-header">
          <p style="margin:0; font-size: var(--type-sm); font-weight: var(--weight-medium);">R4 · Vector store evaluation</p>
          <p style="margin:.1rem 0 0; font-size: var(--type-xs); color: var(--color-text-faint);">12 documents · 3d ago</p>
        </div>
        <button class="action-menu-item">${gearIcon} Settings</button>
        <a href="#" class="action-menu-item">${downloadIcon} Export</a>
      </ActionMenu>
    `),
  }),
}

/** Where it actually sits: last in a header row, after the state and the verbs. */
export const InAPageHeader: Story = {
  args: { title: 'More actions', align: 'right', width: 'default' },
  render: (args) => ({
    components: { ActionMenu },
    setup: () => ({ args }),
    template: roomBelow(`
      <div class="page-header" style="display: flex; align-items: center; justify-content: space-between; gap: 1rem;">
        <h1 class="page-title">Session with a fairly long title that has to share the row</h1>
        <div style="display: flex; align-items: center; gap: 0.75rem; flex-shrink: 0;">
          <span class="status-badge">active</span>
          <ActionMenu v-bind="args">
            <a href="#" class="action-menu-item">${downloadIcon} Export</a>
          </ActionMenu>
        </div>
      </div>
    `),
  }),
}

/**
 * `label="Views"` — the visitor's menu on a shared project.
 *
 * A named `.btn` with a chevron instead of the `⋯`, because on a shared page
 * this is the only route to the graph and the map, and a reader who has never
 * seen this product before does not go looking inside an ellipsis for a
 * navigation they were not told about.
 *
 * The panel is unchanged. Only the trigger differs, which is the point: one
 * component, so the two menus cannot drift apart in behaviour.
 */
export const WithALabel: Story = {
  args: { label: 'Views', align: 'right', width: 'default' },
  render: (args) => ({
    components: { ActionMenu },
    setup: () => ({ args }),
    template: roomBelow(`
      <ActionMenu v-bind="args">
        <a href="#" class="action-menu-item">${graphIcon} Knowledge graph</a>
        <a href="#" class="action-menu-item">${mapIcon} Mind map</a>
      </ActionMenu>
    `),
  }),
}

/**
 * Both triggers in one row, which is the comparison the prop exists for.
 *
 * In a strip of icon buttons a labelled one is the only thing with a word on
 * it, and that is what makes it findable. Used twice in a row it would stop
 * being a landmark and become a second toolbar.
 */
export const LabelledAmongIcons: Story = {
  parameters: { controls: { disable: true } },
  render: () => ({
    components: { ActionMenu },
    template: roomBelow(`
      <div style="display: flex; align-items: center; gap: var(--space-2); justify-content: flex-end;">
        <button class="btn btn-icon" title="Print" aria-label="Print">${downloadIcon}</button>
        <ActionMenu label="Views" align="right" width="default">
          <a href="#" class="action-menu-item">${graphIcon} Knowledge graph</a>
          <a href="#" class="action-menu-item">${mapIcon} Mind map</a>
        </ActionMenu>
        <ActionMenu title="More actions" align="right" width="default">
          <a href="#" class="action-menu-item">${downloadIcon} Export</a>
        </ActionMenu>
      </div>
    `),
  }),
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    await openMenu(canvasElement, 0)
  },
}

/** A long label. The trigger grows to fit rather than truncating — a menu
 *  called "Vie…" is worse than a wider button — and the chevron stays glued to
 *  the end of the word. */
export const LongLabel: Story = {
  args: { label: 'Views and visualisations', align: 'right', width: 'wide' },
  render: (args) => ({
    components: { ActionMenu },
    setup: () => ({ args }),
    template: roomBelow(`
      <ActionMenu v-bind="args">
        <a href="#" class="action-menu-item">${graphIcon} Knowledge graph</a>
        <a href="#" class="action-menu-item">${mapIcon} Mind map</a>
      </ActionMenu>
    `),
  }),
}

/**
 * Opens the nth `⋯` once it is mounted.
 *
 * There is no `@storybook/test` in this project, so the catalogue polls — same
 * helper shape as `HistoryPanel.stories.ts`. The trigger carries no text, so it
 * is found by `aria-haspopup` rather than by label.
 */
async function openMenu(root: HTMLElement, index = 0): Promise<void> {
  for (let i = 0; i < 50; i++) {
    const triggers = Array.from(
      root.querySelectorAll<HTMLElement>('button[aria-haspopup="true"]')
    )
    const trigger = triggers[index]
    if (trigger) {
      trigger.click()
      return
    }
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
  throw new Error(`ActionMenu trigger #${index} never appeared`)
}
