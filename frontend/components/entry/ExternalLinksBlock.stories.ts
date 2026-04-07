import type { Meta, StoryObj } from '@storybook/vue3'
import ExternalLinksBlock from './ExternalLinksBlock.vue'

const meta: Meta<typeof ExternalLinksBlock> = {
  title: 'Entry/ExternalLinksBlock',
  component: ExternalLinksBlock,
  tags: ['autodocs'],
}
export default meta
type Story = StoryObj<typeof ExternalLinksBlock>

export const FewLinks: Story = {
  args: {
    links: [
      { id: 'link_1', url: 'https://vuejs.org/guide/essentials/reactivity-fundamentals.html', title: 'Reactivity Fundamentals', domain: 'vuejs.org' },
      { id: 'link_2', url: 'https://github.com/vuejs/core', title: 'Vue.js Core Repository', domain: 'github.com' },
    ],
  },
}

export const ManyLinks: Story = {
  args: {
    links: [
      { id: 'link_1', url: 'https://vuejs.org/guide/extras/composition-api-faq.html', title: 'Composition API FAQ', domain: 'vuejs.org' },
      { id: 'link_2', url: 'https://vuejs.org/guide/essentials/reactivity-fundamentals.html', title: 'Reactivity Fundamentals', domain: 'vuejs.org' },
      { id: 'link_3', url: 'https://github.com/vuejs/core', title: 'Vue.js Core Repository', domain: 'github.com' },
      { id: 'link_4', url: 'https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Proxy', title: 'JavaScript Proxy', domain: 'developer.mozilla.org' },
      { id: 'link_5', url: 'https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Modules', title: 'JavaScript Modules', domain: 'developer.mozilla.org' },
      { id: 'link_6', url: 'https://pinia.vuejs.org/core-concepts/', title: 'Pinia Core Concepts', domain: 'pinia.vuejs.org' },
      { id: 'link_7', url: 'https://stackoverflow.com/questions/12345/vue-composables', title: null, domain: 'stackoverflow.com' },
    ],
  },
}

export const Empty: Story = {
  args: {
    links: [],
  },
}
