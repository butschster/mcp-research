<template>
  <div class="field" @dblclick="editable && start()">
    <div class="field-header">
      <label class="field-label" :for="editing ? inputId : undefined">{{ label }}</label>
      <button
        v-if="editable && !editing"
        type="button"
        class="field-edit-btn"
        :aria-label="`Edit ${label.toLowerCase()}`"
        :title="`Edit ${label.toLowerCase()}`"
        @click="start"
      >
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
        </svg>
      </button>
    </div>

    <div v-if="!editing" class="field-value" :class="{ 'field-empty': isEmpty }">
      <!-- A field whose value is markdown renders it; the slot is how the
           owner says so, since this component must not decide what a string
           means. -->
      <slot :value="value">{{ isEmpty ? emptyText : value }}</slot>
    </div>

    <div v-else class="field-edit">
      <textarea
        v-if="multiline"
        :id="inputId"
        ref="field"
        v-model="draft"
        class="field-textarea"
        :rows="rows"
        :placeholder="placeholder"
        @keydown.escape="cancel"
      ></textarea>
      <input
        v-else
        :id="inputId"
        ref="field"
        v-model="draft"
        class="field-input"
        :placeholder="placeholder"
        @keydown.enter="commit"
        @keydown.escape="cancel"
      />
      <div class="field-edit-actions">
        <button type="button" class="btn btn-sm btn-primary" @click="commit">Save</button>
        <button type="button" class="btn btn-sm" @click="cancel">Cancel</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * A value that a pencil turns into a form.
 *
 * The triple — header with a pencil, a value, an edit form — was written out
 * eight times across `DetailsPanel` and `TaskDetailModal`, which also shared
 * ten byte-identical rule bodies between them. Each copy carried the same
 * twelve-line pencil icon and the same `title="Edit"`, which is ambiguous when
 * four of them sit in one panel; the label is built from the field's own name
 * now, so a screen reader hears "Edit goal" rather than the fourth "Edit".
 *
 * The draft is local. The owner is told only when Save is pressed, so a
 * cancelled edit costs nothing and a refused one is the owner's to report —
 * this component does not know whether the write succeeded.
 */
const props = withDefaults(defineProps<{
  label: string
  value?: string | null
  /** False for a viewer: the pencil and the double-click both go. */
  editable?: boolean
  multiline?: boolean
  rows?: number
  placeholder?: string
  /** What stands in for an absent value, in the muted style. */
  emptyText?: string
}>(), {
  value: '',
  editable: false,
  multiline: false,
  rows: 3,
  placeholder: '',
  emptyText: '—',
})

const emit = defineEmits<{ save: [value: string] }>()

const editing = ref(false)
const draft = ref('')
const field = ref<HTMLInputElement | HTMLTextAreaElement | null>(null)
const inputId = useId()

const isEmpty = computed(() => !props.value)

function start() {
  draft.value = props.value ?? ''
  editing.value = true
  nextTick(() => field.value?.focus())
}

function cancel() {
  editing.value = false
  draft.value = ''
}

function commit() {
  emit('save', draft.value)
  editing.value = false
}

/* If the value changes underneath — a realtime write, or the owner reloading
   after a save — an open editor is stale. Closing it is the honest response;
   keeping it would let Save overwrite somebody else's change with text typed
   against the version before it. */
watch(() => props.value, () => { if (editing.value) cancel() })
</script>

<style scoped>
/* None of this was global, which I learned by deleting it: an unstyled
   textarea falls back to the browser's white box, in a dark theme, on the
   panel that holds a research's instruction. */
/* The horizontal inset comes from the frame when there is one — the same
   --row-inset a rule list hands its rows — so a field in a list card lines up
   with every other line of text in it. The fallback is what a field wears
   anywhere else, so nothing outside such a card moves. */
.field { padding: var(--space-3) var(--row-inset, var(--space-4)); }
.field-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-1);
}
.field-label {
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-text-muted);
  letter-spacing: 0.02em;
}
.field-value { font-size: var(--type-sm); color: var(--color-text); line-height: 1.6; }
.field-empty { color: var(--color-text-faint); font-style: italic; }
.field-input,
.field-textarea {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-size: var(--type-sm);
  font-family: inherit;
  line-height: 1.5;
}
.field-textarea { resize: vertical; min-height: 60px; }
.field-input:focus,
.field-textarea:focus { outline: 2px solid var(--color-primary); outline-offset: -1px; }
.field-edit { display: flex; flex-direction: column; gap: var(--space-2); }
.field-edit-actions { display: flex; gap: var(--space-2); }
.field-edit-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  padding: 0;
  border: none;
  border-radius: var(--radius-xs);
  background: none;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: color var(--transition-fast), background var(--transition-fast);
}
.field-edit-btn:hover { color: var(--color-text); background: var(--color-surface-hover); }
</style>
