<template>
  <section v-if="specs.length || orphaned.length" class="metadata-block" aria-label="Recorded by this section">
    <div class="metadata-head">
      <span class="metadata-title">Recorded by this section</span>
      <span v-if="versionLabel" class="metadata-version" :title="versionTitle">{{ versionLabel }}</span>
      <div v-if="editable" class="metadata-actions no-print">
        <template v-if="editing">
          <button class="btn-sm btn-primary" :disabled="saving" @click="save">
            {{ saving ? 'Saving...' : 'Save' }}
          </button>
          <button class="btn-sm" :disabled="saving" @click="cancel">Cancel</button>
        </template>
        <button v-else class="btn-sm" @click="startEditing">Edit values</button>
      </div>
    </div>

    <p v-if="error" class="inline-error" role="alert">{{ error }}</p>

    <dl class="metadata-list">
      <!-- The system status sits with the declared fields on purpose. A
           section-declared `stage` directly beneath it is what stops an author
           writing the review state into the prose, where it drifts away from
           the field nobody sees. -->
      <template v-if="specs.length">
        <dt class="metadata-label">Status</dt>
        <dd class="metadata-value">
          <StatusBadge :status="status" />
          <span class="metadata-flag">system</span>
        </dd>
      </template>

      <template v-for="f in specs" :key="f.key">
        <dt class="metadata-label">{{ fieldLabel(f) }}</dt>
        <dd class="metadata-value">
          <template v-if="editing">
            <select
              v-if="inputKind(f) === 'select'"
              v-model="draft[f.key]"
              class="select metadata-input"
            >
              <option value="">not filled</option>
              <!-- A stored value outside the vocabulary rendered as nothing
                   selected: the value somebody came to fix was not on screen,
                   and touching the control lost it for good. -->
              <option v-if="strayValue(f)" :value="strayValue(f)">{{ strayValue(f) }} (not an option)</option>
              <option v-for="o in f.options ?? []" :key="o" :value="o">{{ o }}</option>
            </select>

            <!-- A repeated enum used to fall through to a free-text box, which
                 throws away the only thing that makes a type worth declaring:
                 that it narrows the answer. -->
            <select
              v-else-if="f.type === 'enum' && f.repeated"
              v-model="multi[f.key]"
              class="select metadata-input"
              multiple
              :size="Math.min((f.options ?? []).length, 4)"
            >
              <option v-for="o in f.options ?? []" :key="o" :value="o">{{ o }}</option>
            </select>
            <input
              v-else
              v-model="draft[f.key]"
              class="text-input metadata-input"
              :type="editorInputType(f)"
              :placeholder="f.repeated ? 'comma, separated' : ''"
            />
            <button
              v-if="f.required"
              type="button"
              class="metadata-unknown"
              :class="{ 'is-on': draft[f.key] === UNKNOWN }"
              title="Record that nobody knows. This answers the field without inventing a value."
              @click="toggleUnknown(f.key)"
            >unknown</button>
            <span v-if="f.help" class="metadata-help">{{ f.help }}</span>
          </template>

          <template v-else>
            <!-- Shown while reading, not only while editing. The note says where
                 the value comes from, and the person deciding whether to fill it
                 in is the one reading the document. -->
            <span v-if="isUnknown(values[f.key])" class="metadata-unknown-value">Unknown</span>
            <span
              v-else-if="isFilled(values[f.key])"
              class="metadata-text"
              v-html="renderRefs(displayValue(f, values[f.key]), researchSlug)"
            ></span>
            <span v-else class="metadata-empty">Not filled</span>
            <span
              v-if="f.required && !isFilled(values[f.key])"
              class="metadata-flag metadata-flag--required"
            >Required</span>
            <span
              v-if="issueFor(f.key)"
              class="metadata-flag metadata-flag--invalid"
              :title="issueFor(f.key)"
            >Invalid</span>
            <span v-if="issueFor(f.key)" class="metadata-help">{{ issueFor(f.key) }}</span>
            <span v-if="f.help && !isFilled(values[f.key])" class="metadata-help">{{ f.help }}</span>
          </template>
        </dd>
      </template>
    </dl>

    <!-- Kept, not deleted: dropping a field decides what gets collected next,
         not what was already written down. -->
    <template v-if="orphaned.length">
      <p class="metadata-subhead">No longer collected</p>
      <dl class="metadata-list metadata-list--muted">
        <template v-for="key in orphaned" :key="key">
          <dt class="metadata-label">{{ key }}</dt>
          <dd class="metadata-value">
            <span class="metadata-text" v-html="renderRefs(formatValue(values[key]), researchSlug)"></span>
          </dd>
        </template>
      </dl>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { renderRefs } from '~/composables/useCrossRefs'
import {
  displayValue, fieldLabel, formatValue, inputKind, isFilled, isUnknown, orphanedKeys, parseInput,
  type FieldSpec, type MetadataStatus,
} from '~/composables/useFieldSpec'

const props = defineProps<{
  specs: FieldSpec[]
  values: Record<string, unknown>
  status: string
  researchSlug: string
  editable?: boolean
  entrySpecVersion?: number
  sectionSpecVersion?: number
  metadataStatus?: MetadataStatus | null
  /**
   * Persists the values and resolves when the server has agreed.
   *
   * A function prop rather than an emit, because this one has to be awaited:
   * Vue's `emit` returns undefined, so `await emit(...)` resolves immediately,
   * the busy state clears before the request lands and the catch below can
   * never run. The alternative house pattern — parent-owned busy and error —
   * would also have to tell the child when to close the editor, which is one
   * more thing to keep in step.
   */
  onSave?: (values: Record<string, unknown>) => Promise<void>
}>()

// Announced so the page can hold a remote refresh back while a draft is open —
// the same rule its content editor already follows.
const emit = defineEmits<{ (e: 'editing', open: boolean): void }>()

// The sentinel a person picks to say "we looked and cannot say". It starts with
// a space so nothing a person could type collides with it.
const UNKNOWN = ' unknown'

const editing = ref(false)
const saving = ref(false)
const error = ref('')
const draft = reactive<Record<string, string>>({})
const multi = reactive<Record<string, string[]>>({})

const values = computed(() => props.values ?? {})
const specs = computed(() => props.specs ?? [])
const orphaned = computed(() => orphanedKeys(specs.value, values.value))

const versionLabel = computed(() => {
  const own = props.entrySpecVersion ?? 0
  const now = props.sectionSpecVersion ?? 0
  if (!own) return ''
  return now > own ? `spec v${own} - section is on v${now}` : `spec v${own}`
})

const versionTitle = computed(
  () => `This document was last written against version ${props.entrySpecVersion ?? 0} of the section's fields.`,
)

function editorInputType(f: FieldSpec): string {
  const kind = inputKind(f)
  if (kind === 'date') return 'date'
  if (kind === 'number') return 'number'
  return 'text'
}

function issueFor(key: string): string {
  return props.metadataStatus?.issues?.find(i => i.key === key)?.reason ?? ''
}

function startEditing() {
  error.value = ''
  for (const f of specs.value) {
    const v = values.value[f.key]
    if (f.type === 'enum' && f.repeated) {
      multi[f.key] = Array.isArray(v) ? (v as string[]) : (v ? [String(v)] : [])
      continue
    }
    draft[f.key] = v === null ? UNKNOWN : formatValue(v)
  }
  editing.value = true
  emit('editing', true)
}

/**
 * A stored single-enum value the declaration does not list.
 *
 * It exists because the vocabulary can change under a document: the value was
 * valid when it was written, or an agent wrote something outside it and the
 * server kept it rather than discarding somebody's work.
 */
function strayValue(f: FieldSpec): string {
  if (f.type !== 'enum' || f.repeated) return ''
  const current = draft[f.key]
  if (!current || current === UNKNOWN) return ''
  return (f.options ?? []).includes(current) ? '' : current
}

function cancel() {
  editing.value = false
  emit('editing', false)
  error.value = ''
}

function toggleUnknown(key: string) {
  draft[key] = draft[key] === UNKNOWN ? '' : UNKNOWN
}

async function save() {
  saving.value = true
  error.value = ''
  try {
    const out: Record<string, unknown> = {}
    for (const f of specs.value) {
      if (f.type === 'enum' && f.repeated) {
        out[f.key] = multi[f.key] ?? []
        continue
      }
      const raw = draft[f.key] ?? ''
      if (raw === UNKNOWN) {
        out[f.key] = null
        continue
      }
      // An empty box clears the field, and is sent rather than omitted: the map
      // replaces what is stored, so an omitted key would read as "leave it".
      out[f.key] = parseInput(f, raw)
    }
    // Orphaned values ride along untouched. The editor cannot write them, and
    // dropping them here would delete somebody's work as a side effect of
    // saving something else.
    for (const key of orphaned.value) out[key] = values.value[key]
    await props.onSave?.(out)
    editing.value = false
    emit('editing', false)
  } catch (e: any) {
    error.value = e?.data?.error || e?.message || 'Could not save'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.metadata-block {
  margin-bottom: var(--space-5);
  padding-bottom: var(--space-5);
  border-bottom: 1px solid var(--color-border);
}

.metadata-head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
  font-size: var(--type-xs);
  color: var(--color-text-faint);
}

.metadata-title { font-weight: var(--weight-semibold); }
.metadata-version { font-variant-numeric: tabular-nums; }
.metadata-actions { margin-left: auto; display: flex; gap: var(--space-2); }


.metadata-list {
  display: grid;
  grid-template-columns: minmax(8rem, var(--metadata-col-max)) 1fr;
  column-gap: var(--space-5);
  row-gap: var(--space-2);
  margin: 0;
}

.metadata-list--muted { margin-top: var(--space-2); opacity: 0.72; }

.metadata-subhead {
  margin: var(--space-4) 0 0;
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-text-faint);
}

.metadata-label {
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-text-muted);
  overflow-wrap: anywhere;
}

.metadata-value {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
  margin: 0;
  font-size: var(--type-sm);
  color: var(--color-text);
  overflow-wrap: anywhere;
}

.metadata-text :deep(a) { color: var(--color-primary); }
.metadata-empty { color: var(--color-text-faint); font-style: italic; }
.metadata-unknown-value { color: var(--color-text-muted); }

.metadata-flag {
  font-size: var(--type-xs);
  color: var(--color-text-faint);
}
.metadata-flag--required { color: var(--color-warning); }
.metadata-flag--invalid { color: var(--color-danger); }

.metadata-help {
  flex-basis: 100%;
  font-size: var(--type-xs);
  color: var(--color-text-faint);
}

/* Height is stated rather than left to emerge from padding, so a select, a
   text input and the unknown toggle standing in one row come out level. */
.metadata-input {
  height: var(--control-h);
  min-width: 0;
  flex: 1 1 12rem;
}

.metadata-unknown {
  height: var(--control-h);
  padding: 0 var(--space-2);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  cursor: pointer;
}
.metadata-unknown.is-on {
  color: var(--color-text);
  border-color: var(--color-primary);
}

@media (max-width: 768px) {
  .metadata-list { grid-template-columns: 1fr; row-gap: var(--space-1); }
  .metadata-label { margin-top: var(--space-2); }
}
</style>
