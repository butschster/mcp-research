import type { Meta, StoryObj } from '@storybook/vue3'
import TagList from './TagList.vue'

const meta: Meta<typeof TagList> = {
  title: 'Base/TagList',
  component: TagList,
  tags: ['autodocs'],
  argTypes: {
    tags: { control: 'object' },
    clickable: { control: 'boolean' },
    activeTag: { control: 'text' },
    counts: { control: 'object' },
  },
}
export default meta
type Story = StoryObj<typeof TagList>

export const Empty: Story = {
  args: { tags: [] },
}

export const FewTags: Story = {
  args: { tags: ['vue', 'composables', 'typescript'] },
}

export const ManyTags: Story = {
  args: {
    tags: ['vue', 'react', 'angular', 'svelte', 'typescript', 'javascript', 'css', 'architecture', 'performance', 'testing'],
  },
}

export const Clickable: Story = {
  args: {
    tags: ['vue', 'composables', 'typescript', 'architecture'],
    clickable: true,
  },
}

export const WithActiveTag: Story = {
  args: {
    tags: ['vue', 'composables', 'typescript', 'architecture'],
    clickable: true,
    activeTag: 'composables',
  },
}

export const WithCounts: Story = {
  args: {
    tags: ['vue', 'composables', 'typescript', 'architecture'],
    clickable: true,
    counts: { vue: 12, composables: 5, typescript: 8, architecture: 3 },
  },
}

export const WithActiveAndCounts: Story = {
  args: {
    tags: ['vue', 'composables', 'typescript', 'architecture'],
    clickable: true,
    activeTag: 'vue',
    counts: { vue: 12, composables: 5, typescript: 8, architecture: 3 },
  },
}
