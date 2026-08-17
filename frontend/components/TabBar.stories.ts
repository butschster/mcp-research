import type { Meta, StoryObj } from '@storybook/vue3'
import TabBar from './TabBar.vue'

const meta: Meta<typeof TabBar> = {
  title: 'Navigation/TabBar',
  component: TabBar,
  parameters: {
    docs: {
      description: {
        component:
          'The ARIA tab strip used by the settings and session pages. Selection follows focus, ' +
          'the arrows wrap, and the count badge is hidden from screen readers in favour of the ' +
          'sentence in `srCount` — "Skills 4/6" is read aloud as "Skills four six" otherwise.',
      },
    },
  },
}
export default meta
type Story = StoryObj<typeof TabBar>

export const Default: Story = {
  args: {
    label: 'Research settings',
    modelValue: 'overview',
    tabs: [
      { id: 'overview', label: 'Overview' },
      { id: 'skills', label: 'Skills', count: '2/6', srCount: '2 of 6 chosen' },
      { id: 'memory', label: 'Memory', count: 12, srCount: '12 notes' },
      { id: 'access', label: 'Access' },
    ],
  },
}

export const NoCounts: Story = {
  args: {
    label: 'Plain',
    modelValue: 'one',
    tabs: [
      { id: 'one', label: 'First' },
      { id: 'two', label: 'Second' },
    ],
  },
}

/** Long labels and full badges: the strip scrolls rather than letting the page do it. */
export const Overflowing: Story = {
  parameters: { viewport: { defaultViewport: 'mobile1' } },
  args: {
    label: 'Crowded',
    modelValue: 'a',
    tabs: [
      { id: 'a', label: 'Overview' },
      { id: 'b', label: 'Skills', count: '6/6', srCount: '6 of 6 chosen' },
      { id: 'c', label: 'Memory', count: 148, srCount: '148 notes' },
      { id: 'd', label: 'Access', count: 3, srCount: '3 live links' },
      { id: 'e', label: 'Something else entirely' },
    ],
  },
}
