import type { Meta, StoryObj } from '@storybook/vue3'
import { ref, defineComponent, h } from 'vue'

const ConnectionStatusMock = defineComponent({
  props: {
    state: { type: String as () => 'connected' | 'reconnecting' | 'disconnected', default: 'connected' },
  },
  setup(props) {
    const label = {
      connected: 'Live',
      reconnecting: 'Reconnecting\u2026',
      disconnected: 'Offline',
    }[props.state]

    const tooltip = {
      connected: 'Real-time updates active',
      reconnecting: 'Reconnecting to server\u2026',
      disconnected: 'Not connected',
    }[props.state]

    return () =>
      h('div', { class: 'connection-status', title: tooltip, style: 'display:flex;align-items:center;gap:0.375rem;font-size:var(--type-xs);font-weight:500;color:var(--color-text-muted);cursor:default;' }, [
        h('span', {
          class: `ws-dot ws-${props.state}`,
          style: `width:6px;height:6px;border-radius:50%;flex-shrink:0;${
            props.state === 'connected' ? 'background:var(--color-success);box-shadow:0 0 6px rgba(52,211,153,0.5);' :
            props.state === 'reconnecting' ? 'background:var(--color-warning);' :
            'background:var(--color-text-muted);opacity:0.5;'
          }`,
        }),
        h('span', { class: 'ws-label' }, label),
      ])
  },
})

const meta: Meta<typeof ConnectionStatusMock> = {
  title: 'Base/ConnectionStatus',
  component: ConnectionStatusMock,
  tags: ['autodocs'],
  argTypes: {
    state: {
      control: 'select',
      options: ['connected', 'reconnecting', 'disconnected'],
    },
  },
}
export default meta
type Story = StoryObj<typeof ConnectionStatusMock>

export const Connected: Story = { args: { state: 'connected' } }
export const Reconnecting: Story = { args: { state: 'reconnecting' } }
export const Disconnected: Story = { args: { state: 'disconnected' } }

export const AllStates: Story = {
  render: () => ({
    components: { ConnectionStatusMock },
    template: `
      <div style="display: flex; gap: 2rem; align-items: center;">
        <ConnectionStatusMock state="connected" />
        <ConnectionStatusMock state="reconnecting" />
        <ConnectionStatusMock state="disconnected" />
      </div>
    `,
  }),
}
