import type { Meta, StoryObj } from '@storybook/vue3'
import ConfirmModal from './ConfirmModal.vue'

const meta: Meta<typeof ConfirmModal> = {
  title: 'Base/ConfirmModal',
  component: ConfirmModal,
  tags: ['autodocs'],
  argTypes: {
    visible: { control: 'boolean' },
    title: { control: 'text' },
    message: { control: 'text' },
    confirmLabel: { control: 'text' },
    variant: { control: 'select', options: ['danger', 'info'] },
    loading: { control: 'boolean' },
  },
}
export default meta
type Story = StoryObj<typeof ConfirmModal>

export const Danger: Story = {
  args: {
    visible: true,
    title: 'Delete Entry',
    message: 'Are you sure you want to delete "Architecture Overview"? This action cannot be undone.',
    confirmLabel: 'Delete',
    variant: 'danger',
    loading: false,
  },
}

export const Info: Story = {
  args: {
    visible: true,
    title: 'Archive Research',
    message: 'This will archive the research and hide it from the main list. You can restore it later.',
    confirmLabel: 'Archive',
    variant: 'info',
    loading: false,
  },
}

export const Loading: Story = {
  args: {
    visible: true,
    title: 'Delete Entry',
    message: 'Are you sure you want to delete this entry?',
    confirmLabel: 'Delete',
    variant: 'danger',
    loading: true,
  },
}

export const DefaultLabels: Story = {
  args: {
    visible: true,
    title: 'Confirm Action',
    message: 'Are you sure you want to proceed?',
  },
}
