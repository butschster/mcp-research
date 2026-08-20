import type { Meta, StoryObj } from '@storybook/vue3'
import ProgressBar from './ProgressBar.vue'

const meta: Meta<typeof ProgressBar> = {
  title: 'Base/ProgressBar',
  component: ProgressBar,
  tags: ['autodocs'],
  argTypes: {
    value: { control: { type: 'range', min: 0, max: 100 } },
    total: { control: { type: 'number', min: 1 } },
    showLabel: { control: 'boolean' },
    tone: { control: 'inline-radio', options: ['auto', 'done'] },
  },
  parameters: {
    docs: {
      description: {
        component:
          'A thin completion bar. `tone` decides what the percentage *means*: `auto` ' +
          'reads it as progress against a plan and paints anything under 30% red, ' +
          'which is right for a session that has stalled. `done` pins the fill to the ' +
          'success colour, for a bar that counts completions with nothing scheduled ' +
          'behind them — a task list nobody has started has not failed at anything.',
      },
    },
  },
}
export default meta
type Story = StoryObj<typeof ProgressBar>

export const Empty: Story = { args: { value: 0, total: 100, showLabel: true } }
export const Quarter: Story = { args: { value: 25, total: 100, showLabel: true } }
export const Half: Story = { args: { value: 50, total: 100, showLabel: true } }
export const ThreeQuarters: Story = { args: { value: 75, total: 100, showLabel: true } }
export const Complete: Story = { args: { value: 100, total: 100, showLabel: true } }

export const WithoutLabel: Story = { args: { value: 60, total: 100, showLabel: false } }

/** Nothing to divide by: the bar reads 0% rather than NaN. Reached by a section
 *  with no entries yet, which is every section on the day it is created. */
export const NoTotal: Story = { args: { value: 0, total: 0, showLabel: true } }

/**
 * `tone="done"`, the tone a `task_ref` block asks for.
 *
 * Five tasks written and none closed yet. Under `auto` this same bar is red —
 * the state the tone exists to stop, because a freshly written list has made no
 * progress against a plan it does not have.
 */
export const DoneToneNothingClosed: Story = {
  args: { value: 0, total: 5, tone: 'done', showLabel: true },
}

/** Part-way through the same list. Still green: the count is the fact, the
 *  colour is not a judgement about it. */
export const DoneTonePartway: Story = {
  args: { value: 1, total: 5, tone: 'done', showLabel: true },
}

/** Both tones at every value. The two agree only at 100%, which is the whole
 *  reason a caller has to choose. */
export const AllTones: Story = {
  render: () => ({
    components: { ProgressBar },
    template: `
      <div style="display: grid; grid-template-columns: 3rem 1fr 1fr; gap: 0.75rem 1.5rem; align-items: center; max-width: 640px;">
        <div></div>
        <div style="font-size: 0.75rem; color: var(--color-text-muted);">tone="auto"</div>
        <div style="font-size: 0.75rem; color: var(--color-text-muted);">tone="done"</div>
        <template v-for="v in values" :key="v">
          <div style="font-size: 0.75rem; color: var(--color-text-muted); font-variant-numeric: tabular-nums;">{{ v }}%</div>
          <ProgressBar :value="v" :total="100" tone="auto" showLabel />
          <ProgressBar :value="v" :total="100" tone="done" showLabel />
        </template>
      </div>
    `,
    setup() {
      return { values: [0, 10, 25, 50, 75, 100] }
    },
  }),
}

/** Every value under the default tone: red under 30, amber to 70, blue to 100,
 *  green at the end. */
export const AllStates: Story = {
  render: () => ({
    components: { ProgressBar },
    template: `
      <div style="display: flex; flex-direction: column; gap: 1.5rem; max-width: 400px;">
        <div v-for="v in values" :key="v">
          <div style="font-size: 0.75rem; color: var(--color-text-muted); margin-bottom: 0.25rem;">{{ v }}%</div>
          <ProgressBar :value="v" :total="100" showLabel />
        </div>
      </div>
    `,
    setup() {
      return { values: [0, 10, 25, 50, 75, 100] }
    },
  }),
}
