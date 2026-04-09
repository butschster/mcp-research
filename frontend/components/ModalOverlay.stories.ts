import type { Meta, StoryObj } from '@storybook/vue3'
import ModalOverlay from './ModalOverlay.vue'

const meta: Meta<typeof ModalOverlay> = {
  title: 'Base/ModalOverlay',
  component: ModalOverlay,
  tags: ['autodocs'],
  argTypes: {
    visible: { control: 'boolean' },
    size: { control: 'select', options: ['sm', 'md', 'lg'] },
    flush: { control: 'boolean' },
  },
}
export default meta
type Story = StoryObj<typeof ModalOverlay>

export const Default: Story = {
  args: { visible: true },
  render: (args) => ({
    components: { ModalOverlay },
    setup() { return { args } },
    template: `
      <ModalOverlay v-bind="args">
        <h3 style="margin: 0 0 0.5rem; font-weight: 600;">Modal Title</h3>
        <p style="color: var(--color-text-muted); font-size: var(--type-sm); margin: 0;">This is a default-sized modal with some content inside.</p>
      </ModalOverlay>
    `,
  }),
}

export const Small: Story = {
  args: { visible: true, size: 'sm' },
  render: (args) => ({
    components: { ModalOverlay },
    setup() { return { args } },
    template: `
      <ModalOverlay v-bind="args">
        <h3 style="margin: 0 0 0.5rem; font-weight: 600;">Confirm</h3>
        <p style="color: var(--color-text-muted); font-size: var(--type-sm); margin: 0 0 1rem;">Are you sure you want to delete this?</p>
        <div style="display: flex; gap: 0.5rem; justify-content: flex-end;">
          <button class="btn btn-sm">Cancel</button>
          <button class="btn btn-sm" style="background: var(--color-error); color: white;">Delete</button>
        </div>
      </ModalOverlay>
    `,
  }),
}

export const Large: Story = {
  args: { visible: true, size: 'lg' },
  render: (args) => ({
    components: { ModalOverlay },
    setup() { return { args } },
    template: `
      <ModalOverlay v-bind="args">
        <h3 style="margin: 0 0 0.5rem; font-weight: 600;">Research Details</h3>
        <p style="color: var(--color-text-muted); font-size: var(--type-sm); margin: 0 0 1rem;">
          A large modal for displaying detailed content, forms, or data tables.
        </p>
        <div style="background: var(--color-surface-hover); border-radius: var(--radius-sm); padding: 1rem; font-size: var(--type-sm); color: var(--color-text-muted);">
          <p style="margin: 0 0 0.5rem;">Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>
          <p style="margin: 0;">Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.</p>
        </div>
      </ModalOverlay>
    `,
  }),
}

export const Flush: Story = {
  args: { visible: true, size: 'lg', flush: true },
  render: (args) => ({
    components: { ModalOverlay },
    setup() { return { args } },
    template: `
      <ModalOverlay v-bind="args">
        <div style="padding: 0.75rem 1.5rem; border-bottom: 1px solid var(--color-border); display: flex; justify-content: space-between; align-items: center;">
          <span style="font-size: var(--type-sm); font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em; color: var(--color-text-muted);">Modal Title</span>
          <span style="color: var(--color-text-muted); cursor: pointer;">&#x2715;</span>
        </div>
        <div style="padding: 1.25rem 1.5rem;">
          <p style="color: var(--color-text-muted); font-size: var(--type-sm); margin: 0;">Flush mode removes card padding so the header border extends edge-to-edge.</p>
        </div>
      </ModalOverlay>
    `,
  }),
}

export const Hidden: Story = {
  args: { visible: false },
  render: (args) => ({
    components: { ModalOverlay },
    setup() { return { args } },
    template: `
      <div>
        <p style="color: var(--color-text-muted); font-size: var(--type-sm);">Modal is hidden (visible=false). Nothing renders.</p>
        <ModalOverlay v-bind="args">
          <p>You should not see this.</p>
        </ModalOverlay>
      </div>
    `,
  }),
}
