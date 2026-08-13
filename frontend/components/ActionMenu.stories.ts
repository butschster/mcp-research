import type { Meta, StoryObj } from '@storybook/vue3'
import ActionMenu from './ActionMenu.vue'

const meta: Meta<typeof ActionMenu> = {
  title: 'Navigation/ActionMenu',
  component: ActionMenu,
  tags: ['autodocs'],
  argTypes: {
    title: { control: 'text' },
    align: { control: 'inline-radio', options: ['left', 'right'] },
  },
  parameters: {
    docs: {
      description: {
        component:
          'Three-dot menu for page-level actions. Items go in the default slot and ' +
          'carry the `action-menu-item` class; use `action-menu-divider` to group them. ' +
          'Closes on outside click, on Escape (returning focus to the trigger), and ' +
          'after any item is clicked.',
      },
    },
  },
}
export default meta
type Story = StoryObj<typeof ActionMenu>

const downloadIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>`
const gearIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>`
const trashIcon = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/></svg>`

// The menu is absolutely positioned, so stories need room below the trigger.
const roomBelow = (story: string) => `<div style="min-height: 260px;">${story}</div>`

export const SingleAction: Story = {
  args: { title: 'More actions', align: 'right' },
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

export const SeveralActions: Story = {
  args: { title: 'More actions', align: 'right' },
  render: (args) => ({
    components: { ActionMenu },
    setup: () => ({ args }),
    template: roomBelow(`
      <ActionMenu v-bind="args">
        <button class="action-menu-item">${gearIcon} Details</button>
        <a href="#" class="action-menu-item">${downloadIcon} Export</a>
        <button class="action-menu-item" disabled>${downloadIcon} Download JSON</button>
        <div class="action-menu-divider"></div>
        <button class="action-menu-item action-menu-item--danger">${trashIcon} Delete</button>
      </ActionMenu>
    `),
  }),
}

export const AlignedLeft: Story = {
  args: { title: 'More actions', align: 'left' },
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

export const InAPageHeader: Story = {
  args: { title: 'More actions', align: 'right' },
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
