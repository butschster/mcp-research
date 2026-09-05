import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import MemoryList from './MemoryList.vue'
import type { MemoryItem } from './MemoryList.vue'

const notes: MemoryItem[] = [
  { id: 'memory-agent', text: 'Check the source before citing it.', author: 'agent', created_at: '2026-09-01T12:00:00Z', session_code: 'SS1', version: 1 },
  { id: 'memory-legacy', text: 'A legacy note with no invented provenance.', author: 'unknown', created_at: null, version: 1 },
]
const noop = async () => {}
const meta: Meta<typeof MemoryList> = {
  title: 'Research/Settings/MemoryList',
  component: MemoryList,
  tags: ['autodocs'],
  args: { researchId: 'R1', canWrite: true, items: notes, onAdd: noop, onUpdate: noop, onDelete: noop, onReload: noop },
  decorators: [() => ({ template: '<div style="max-width: 860px"><story /></div>' })],
}
export default meta
type Story = StoryObj<typeof MemoryList>

// Separate saved and displayed copies model a stale client and a subsequent fetch.
function interactive(mode: 'editable' | 'conflict' | 'deleted' = 'editable'): Story['render'] {
  return args => ({
    components: { MemoryList },
    setup() {
      const items = ref(args.items.map(item => ({ ...item })))
      let saved = items.value.map(item => ({ ...item }))
      let nextId = 1
      if (mode === 'conflict') saved[0] = { ...saved[0]!, text: 'A colleague saved this newer source-checking guidance.', version: 2 }
      if (mode === 'deleted') saved = saved.filter(item => item.id !== 'memory-agent')
      async function reload() { items.value = saved.map(item => ({ ...item })) }
      async function add(text: string) {
        saved.push({ id: `added-${nextId++}`, text, author: 'user', created_at: '2026-09-05T12:00:00Z', version: 1 })
        await reload()
      }
      async function update(id: string, text: string, version: number) {
        const item = saved.find(item => item.id === id)
        if (!item || item.version !== version) throw Object.assign(new globalThis.Error('Memory item changed; reload before editing.'), { status: 409 })
        item.text = text
        item.version++
        await reload()
      }
      async function remove(ids: string[]) { saved = saved.filter(item => !ids.includes(item.id)); await reload() }
      return { args, items, add, update, remove, reload }
    },
    template: '<MemoryList v-bind="args" :items="items" :on-add="add" :on-update="update" :on-delete="remove" :on-reload="reload" />',
  })
}

export const Editable: Story = { render: interactive() }
export const ReadOnly: Story = { args: { canWrite: false } }
export const Conflict: Story = {
  render: interactive('conflict'),
  parameters: { docs: { description: { story: 'Edit and save to encounter a conflict, then load the latest saved text and save the retained draft.' } } },
}
export const ReloadDeleted: Story = {
  render: interactive('deleted'),
  parameters: { docs: { description: { story: 'Edit the first note, then Refresh. The server has deleted it; its draft remains available to copy.' } } },
}
export const Empty: Story = { args: { items: [] }, render: interactive() }
export const LegacyEmpty: Story = {
  args: { items: [{ id: 'legacy-empty', text: '', author: 'unknown', created_at: null, version: 1 }] },
  render: interactive(),
}
const longNotes: MemoryItem[] = Array.from({ length: 24 }, (_, index) => ({
  id: `multilingual-${index}`, author: index % 2 ? 'user' : 'agent', version: 1, created_at: '2026-09-01T12:00:00Z',
  text: `Source review ${index + 1}: preserve original language and verify [[E3]].\nქართული წყაროს შემოწმება. 日本語の引用を確認してください。 التحقق من المصدر الأصلي.\nhttps://example.org/research/${'long-source-identifier-'.repeat(16)}`,
}))
export const LongMultilingual: Story = { args: { items: longNotes }, render: interactive() }
export const Pending: Story = {
  args: { onReload: () => new Promise(() => {}) },
  play: async ({ canvasElement }) => { (canvasElement.querySelector('.toolbar button') as HTMLButtonElement).click() },
}
export const Error: Story = {
  args: { onAdd: async () => { throw new globalThis.Error('Could not save: connection interrupted. Your draft has been kept.') } },
  play: async ({ canvasElement }) => {
    const input = canvasElement.querySelector('textarea')!
    input.value = 'Keep this note while the connection recovers.'
    input.dispatchEvent(new Event('input', { bubbles: true }))
    await Promise.resolve()
    canvasElement.querySelector('form')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }))
  },
}
export const AllStates: Story = {
  render: args => ({
    components: { MemoryList },
    setup: () => ({ args, notes, longNotes }),
    template: `<div style="display: grid; gap: 48px">
      <div><h2>Editable</h2><MemoryList v-bind="args" /></div>
      <div><h2>Read only</h2><MemoryList v-bind="args" :can-write="false" /></div>
      <div><h2>Empty</h2><MemoryList v-bind="args" :items="[]" /></div>
      <div><h2>Long multilingual text</h2><MemoryList v-bind="args" :items="longNotes.slice(0, 1)" /></div>
    </div>`,
  }),
}
