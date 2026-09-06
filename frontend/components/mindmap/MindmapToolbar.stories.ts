import type { Meta, StoryObj } from '@storybook/vue3'
import MindmapToolbar from './MindmapToolbar.vue'

/** The mind map's controls, shared by the owner's page and the public share
 *  page. The share page offers fewer groups — a chip for a part the link
 *  excludes would advertise it. */
const meta: Meta<typeof MindmapToolbar> = {
  title: 'Mindmap/MindmapToolbar',
  component: MindmapToolbar,
}
export default meta
type Story = StoryObj<typeof MindmapToolbar>

const allGroups = [
  { key: 'entries', label: 'Documents' },
  { key: 'questions', label: 'Sessions' },
  { key: 'tasks', label: 'Tasks' },
]

export const Owner: Story = {
  args: {
    groups: allGroups,
    visibleGroups: new Set(['entries', 'questions', 'tasks']),
    showCrossrefs: false,
    layoutDirection: 'LR',
  },
}

export const SharedDocumentsOnly: Story = {
  args: {
    groups: [allGroups[0]!],
    visibleGroups: new Set(['entries']),
    showCrossrefs: true,
    layoutDirection: 'TB',
  },
}

export const Interactive: Story = {
  render: (args) => ({
    components: { MindmapToolbar },
    setup: () => {
      const visible = ref(new Set(['entries', 'tasks']))
      const crossrefs = ref(false)
      const direction = ref<'LR' | 'TB'>('LR')
      const log = ref<string[]>([])
      function toggle(key: string) {
        const next = new Set(visible.value)
        next.has(key) ? next.delete(key) : next.add(key)
        visible.value = next
      }
      return { args, visible, crossrefs, direction, log, toggle, allGroups }
    },
    template: `
      <MindmapToolbar
        :groups="allGroups"
        :visible-groups="visible"
        :show-crossrefs="crossrefs"
        :layout-direction="direction"
        @toggle-group="toggle"
        @toggle-crossrefs="crossrefs = !crossrefs"
        @set-direction="d => direction = d"
        @expand-all="log.push('expand all')"
        @collapse-all="log.push('collapse')"
        @fit="log.push('fit')"
      />
      <p style="margin-top: 12px; font-size: 12px; color: var(--color-text-muted)">{{ log.join(' · ') || 'Buttons report here.' }}</p>
    `,
  }),
}
