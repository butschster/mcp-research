<template>
  <section class="memory-list" aria-label="Project memory" :aria-busy="busy">
    <p v-if="error" class="inline-error" role="alert">{{ error }}</p>
    <div class="toolbar">
      <p class="lead">Context your AI assistant can use when working on this project. Use Skills to guide how it works.</p>
      <button type="button" class="btn btn-secondary" :disabled="busy" @click="reload">Refresh</button>
    </div>
    <p v-if="!canWrite" class="lead">Your role allows reading these notes, but not changing them.</p>
    <form v-if="canWrite" class="card note-form" @submit.prevent="add">
      <label :for="newMemoryId">New memory note</label>
      <textarea :id="newMemoryId" ref="newInput" class="form-textarea" v-model="newText" rows="3" :disabled="busy" required />
      <button class="btn btn-primary" :disabled="busy || !newText.trim()">Add note</button>
    </form>
    <div v-if="selected.length && canWrite" class="toolbar">
      <span>{{ selected.length }} selected</span>
      <button type="button" class="btn btn-secondary" :disabled="busy" @click="deleting = [...selected]">Delete selected</button>
    </div>
    <ConfirmModal
      :visible="!!deleting.length"
      title="Delete memory notes"
      :message="`Delete ${deleting.length} selected ${deleting.length === 1 ? 'note' : 'notes'}? This cannot be undone. New notes are not affected.`"
      confirm-label="Confirm deletion"
      variant="danger"
      :loading="busy"
      @confirm="remove"
      @cancel="cancelDeletion"
    />
    <p v-if="!items.length" class="card lead">No memory notes yet.</p>
    <article v-for="item in displayedItems" :key="item.id" class="card memory-item">
      <div class="toolbar">
        <label v-if="canWrite" class="select-note">
          <input v-model="selected" type="checkbox" :value="item.id" :disabled="busy" />
          <span>Select note <span class="note-id">{{ item.id.slice(0, 8) }}</span></span>
        </label>
        <span class="provenance">
          {{ item.author === 'user' ? 'User' : item.author === 'agent' ? 'AI assistant' : 'Unknown author' }}
          · {{ item.created_at ? new Date(item.created_at).toLocaleString() : 'Date not recorded' }}
          <NuxtLink v-if="item.session_code" :to="`/research/${researchId}/session/${item.session_code}`"> · {{ item.session_code }}</NuxtLink>
        </span>
      </div>
      <form v-if="editing?.id === item.id" class="note-form" @submit.prevent="save">
        <label :for="`memory-${item.id}`">Note text</label>
        <textarea :id="`memory-${item.id}`" :ref="setEditInput" class="form-textarea" v-model="editText" rows="5" :disabled="busy" required />
        <div v-if="latestText !== null" class="saved-version">
          <h4>Current saved text</h4>
          <p class="note-text">{{ latestText }}</p>
          <p class="lead">Review the saved text and your draft before saving.</p>
        </div>
        <p v-if="editedNoteMissing" class="inline-error" role="alert">This note was deleted. Your draft is still available to copy into a new note.</p>
        <div class="actions">
          <button class="btn btn-primary" :disabled="busy || !editText.trim() || needsLatest || editedNoteMissing">Save</button>
          <button type="button" class="btn btn-secondary" :disabled="busy" @click="closeEditor">Cancel</button>
        </div>
        <div v-if="needsLatest && !editedNoteMissing" class="note-form">
          <p class="lead">This note changed elsewhere. Load its latest version to compare with your draft before saving.</p>
          <button type="button" class="btn" :disabled="busy" @click="loadLatest">Load latest version</button>
        </div>
      </form>
      <template v-else>
        <p class="note-text">{{ item.text }}</p>
        <div v-if="canWrite" class="actions">
          <button type="button" class="btn btn-secondary" :ref="element => setEditButton(item.id, element)" :disabled="busy" @click="edit(item)">Edit</button>
          <button type="button" class="btn btn-secondary" :disabled="busy" @click="deleting = [item.id]">Delete</button>
        </div>
      </template>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, useId } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import ConfirmModal from '../../ConfirmModal.vue'

export interface MemoryItem {
  id: string
  text: string
  author: 'agent' | 'user' | 'unknown'
  created_at: string | null
  session_id?: string
  session_code?: string
  version: number
}

const props = defineProps<{
  items: MemoryItem[]
  researchId: string
  canWrite: boolean
  onAdd: (text: string) => Promise<unknown>
  onUpdate: (id: string, text: string, version: number) => Promise<unknown>
  onDelete: (ids: string[]) => Promise<unknown>
  onReload: () => Promise<unknown>
}>()
const busy = ref(false)
const error = ref('')
const newText = ref('')
const selected = ref<string[]>([])
const deleting = ref<string[]>([])
const editing = ref<MemoryItem | null>(null)
const editText = ref('')
const newMemoryId = `new-memory-${useId()}`
const newInput = ref<HTMLTextAreaElement | null>(null)
let editInput: HTMLTextAreaElement | null = null
const editButtons = new Map<string, HTMLElement>()
function setEditButton(id: string, element: Element | ComponentPublicInstance | null) {
  if (element) editButtons.set(id, element as HTMLElement)
  else editButtons.delete(id)
}
const conflict = ref(false)
const latestText = ref<string | null>(null)
const currentItem = computed(() => props.items.find(item => item.id === editing.value?.id))
const editedNoteMissing = computed(() => !!editing.value && !currentItem.value)
const needsLatest = computed(() => conflict.value || (!!editing.value && !!currentItem.value && currentItem.value.version !== editing.value.version))
// Keep the editor mounted even if a refresh reveals the note was deleted.
const displayedItems = computed(() => editedNoteMissing.value ? [...props.items, editing.value!] : props.items)
function setEditInput(element: Element | ComponentPublicInstance | null) { editInput = element as HTMLTextAreaElement | null }
let restoreEditorId: string | undefined
async function restoreEditorFocus() {
  if (!restoreEditorId) return
  const id = restoreEditorId
  restoreEditorId = undefined
  await nextTick()
  if (editButtons.has(id)) editButtons.get(id)!.focus()
  else newInput.value?.focus()
}
async function closeEditor() {
  restoreEditorId = editing.value?.id
  editing.value = null
  conflict.value = false
  latestText.value = null
  if (!busy.value) await restoreEditorFocus()
}
function cancelDeletion() { if (!busy.value) deleting.value = [] }

async function run(action: () => Promise<void>) {
  if (busy.value) return
  busy.value = true
  error.value = ''
  try { await action() }
  catch (e: any) { error.value = e?.data?.error || e?.message || 'Could not save. Your draft has been kept.' }
  finally { busy.value = false; await restoreEditorFocus() }
}
async function edit(item: MemoryItem) {
  editing.value = { ...item }
  editText.value = item.text
  conflict.value = false
  latestText.value = null
  await nextTick()
  editInput?.focus()
}
async function loadLatest() {
  if (busy.value) return
  await run(async () => {
    await props.onReload()
    // The parent updates the list; wait until its new value reaches our props.
    await nextTick()
    if (editing.value && currentItem.value) {
      latestText.value = currentItem.value.text
      editing.value = { ...currentItem.value }
      conflict.value = false
    }
  })
  // run has re-enabled the editor after the recovery control disappears.
  await nextTick()
  editInput?.focus()
}
function add() {
  if (!props.canWrite) return
  return run(async () => { await props.onAdd(newText.value); newText.value = '' })
}
function save() {
  if (!props.canWrite || !editing.value || needsLatest.value || editedNoteMissing.value) return
  const item = { ...editing.value }
  return run(async () => {
    try { await props.onUpdate(item.id, editText.value, item.version) }
    catch (e: any) {
      if (e?.status === 409 || e?.statusCode === 409 || e?.response?.status === 409 || /changed|conflict/i.test(e?.data?.error || e?.message || '')) conflict.value = true
      throw e
    }
    await closeEditor()
  })
}
function remove() {
  if (!props.canWrite) return
  return run(async () => {
    const ids = [...deleting.value]
    try { await props.onDelete(ids) }
    catch (e) { deleting.value = []; throw e }
    selected.value = selected.value.filter(id => !ids.includes(id))
    deleting.value = []
    if (editing.value && ids.includes(editing.value.id)) await closeEditor()
  })
}
function reload() { return run(async () => { await props.onReload() }) }
</script>

<style scoped>
.memory-list, .note-form { display: grid; gap: var(--space-4); }
.toolbar, .actions, .select-note { display: flex; align-items: center; gap: var(--space-3); flex-wrap: wrap; }
.toolbar { justify-content: space-between; }
.toolbar .lead { flex: 1; margin: 0; }
.note-form label { font-weight: var(--weight-semibold); }
.note-form textarea { box-sizing: border-box; resize: vertical; min-height: 96px; }
.note-form > button { justify-self: start; }
.note-text { white-space: pre-wrap; overflow-wrap: anywhere; line-height: 1.65; }
.provenance, .note-id { font-size: var(--type-xs); color: var(--color-text-muted); }
.lead { margin: 0; font-size: var(--type-sm); color: var(--color-text-muted); max-width: var(--measure-prose); }
.saved-version h4 { margin: 0; font-size: var(--type-sm); }
.memory-item { min-width: 0; }
@media (max-width: 640px) { .toolbar { align-items: flex-start; } .provenance { width: 100%; } }
</style>
