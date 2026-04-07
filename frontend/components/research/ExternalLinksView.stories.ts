import type { Meta, StoryObj } from '@storybook/vue3'
import ExternalLinksView from './ExternalLinksView.vue'

const meta: Meta<typeof ExternalLinksView> = {
  title: 'Research/ExternalLinksView',
  component: ExternalLinksView,
  tags: ['autodocs'],
  argTypes: {
    loading: { control: 'boolean' },
  },
}
export default meta
type Story = StoryObj<typeof ExternalLinksView>

const mockGroups = [
  {
    domain: 'vuejs.org',
    links: [
      { url: 'https://vuejs.org/guide/extras/composition-api-faq.html', title: 'Composition API FAQ', entry_code: 'E1', entry_title: 'Component Composition Patterns' },
      { url: 'https://vuejs.org/guide/essentials/reactivity-fundamentals.html', title: 'Reactivity Fundamentals', entry_code: 'E2', entry_title: 'Reactive State Management' },
    ],
  },
  {
    domain: 'github.com',
    links: [
      { url: 'https://github.com/vuejs/core', title: 'Vue.js Core Repository', entry_code: 'E1', entry_title: 'Component Composition Patterns' },
    ],
  },
  {
    domain: 'developer.mozilla.org',
    links: [
      { url: 'https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Proxy', title: 'JavaScript Proxy', entry_code: 'E2', entry_title: 'Reactive State Management' },
      { url: 'https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Modules', title: 'JavaScript Modules', entry_code: null, entry_title: null },
    ],
  },
]

export const MultipleDomains: Story = {
  args: {
    groups: mockGroups,
    loading: false,
  },
}

export const Loading: Story = {
  args: {
    groups: [],
    loading: true,
  },
}

export const Empty: Story = {
  args: {
    groups: [],
    loading: false,
  },
}
