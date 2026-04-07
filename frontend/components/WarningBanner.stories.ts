import type { Meta, StoryObj } from '@storybook/vue3'
import { defineComponent, h } from 'vue'

const WarningBannerMock = defineComponent({
  props: {
    visible: { type: Boolean, default: true },
  },
  setup(props) {
    return () =>
      props.visible
        ? h('div', { class: 'warning-banner' }, [
            '\u26A0 ',
            h('strong', 'In-memory mode'),
            ' \u2014 data will be lost on restart. Run with ',
            h('code', '--db ./research.db'),
            ' to persist data.',
          ])
        : null
  },
})

const meta: Meta<typeof WarningBannerMock> = {
  title: 'Base/WarningBanner',
  component: WarningBannerMock,
  tags: ['autodocs'],
  argTypes: {
    visible: { control: 'boolean' },
  },
}
export default meta
type Story = StoryObj<typeof WarningBannerMock>

export const Visible: Story = { args: { visible: true } }
export const Hidden: Story = { args: { visible: false } }
