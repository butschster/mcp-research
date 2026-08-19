<template>
  <ModalOverlay :visible="visible" size="lg" flush :labelledby="titleId" @close="emit('close')">
    <ModalHeader :title="`Import into ${sectionName}`" :title-id="titleId" @close="emit('close')" />

    <div v-if="preview" class="import-body">
      <p class="import-file">
        {{ preview.filename }} · {{ kb }} KB ·
        {{ preview.title_source === 'frontmatter' ? 'front matter read' : 'no front matter' }}
      </p>

      <p v-if="staleSpec" class="inline-error">
        The fields this section records changed while this was open, so what you
        are looking at was checked against the old ones.
        <button type="button" class="btn btn-sm" @click="emit('reread')">Read the file again</button>
      </p>

      <!-- The one editable thing, and the tallest, brightest object in the
           dialog for that reason. Everything else is read-only: the moment a
           second field is editable the metadata fields have to be too, and that
           is MetadataBlock plus the entry page's title bar rebuilt inside a
           dialog — one set of values with two editors. -->
      <div class="import-field">
        <label class="form-label" for="import-title">Title</label>
        <input id="import-title" v-model="titleDraft" type="text" class="text-input" />
        <p class="import-hint">{{ titleHint }}</p>
      </div>

      <dl class="import-read">
        <dt>Status</dt>
        <dd><StatusBadge :status="preview.status" /></dd>
        <dt v-if="preview.tags.length">Tags</dt>
        <dd v-if="preview.tags.length"><TagList :tags="preview.tags" /></dd>
        <dt v-if="preview.description">Description</dt>
        <dd v-if="preview.description">{{ preview.description }}</dd>
      </dl>

      <p v-for="(w, i) in preview.warnings || []" :key="i" class="import-warning">{{ w }}</p>

      <template v-if="groups.length">
        <h3 class="import-heading">What we did with the front matter</h3>
        <ul class="import-ledger">
          <ResearchImportNoteGroup
            v-for="g in groups"
            :key="g.label"
            :tone="g.tone"
            :label="g.label"
            :items="g.items"
            :default-open="g.tone === 'attention'"
          />
        </ul>
      </template>

      <h3 class="import-heading">The document</h3>
      <pre
        class="import-well"
        tabindex="0"
        role="region"
        :aria-label="`The document, first lines of ${preview.filename}`"
      ><code>{{ bodyHead }}</code></pre>
      <p class="import-hint">
        <template v-if="preview.body_lines > BODY_LINES">
          First {{ BODY_LINES }} of {{ preview.body_lines }} lines. All of it is saved.
        </template>
        <template v-else>{{ preview.body_lines }} lines.</template>
      </p>

      <p v-if="error" class="inline-error" role="alert">{{ error }}</p>
    </div>

    <div class="modal-footer import-footer">
      <button type="button" class="btn" :disabled="committing" @click="emit('close')">Cancel</button>
      <button type="button" class="btn btn-primary" :disabled="committing || !titleDraft.trim()" @click="submit">
        {{ committing ? 'Creating…' : 'Create document' }}
      </button>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
/**
 * What we made of the file, before anything is written.
 *
 * `ObsidianExportDialog` is the nearest shape in the catalog and it is an
 * options form. This is a parse report with one editable field in it, which
 * nothing here does.
 *
 * There is no `v-html` anywhere in this dialog, deliberately. Rendering the
 * body would mean `parseMarkdown` then `linkRefs`, on the one screen in the
 * product whose content came out of a file somebody emailed you — and it would
 * turn `[[E99]]` into a live link on the screen whose whole job is to say that
 * reference points at nothing.
 */
import type { ImportPreview } from '~/composables/useSectionImport'

const props = defineProps<{
  visible: boolean
  fileBytes: number
  sectionName: string
  preview: ImportPreview | null
  committing: boolean
  error: string
  staleSpec: boolean
}>()

const emit = defineEmits<{
  commit: [{ title: string }]
  reread: []
  close: []
}>()

/** The dialog is not a text editor. */
const BODY_LINES = 40

const titleId = useId()
const titleDraft = ref('')

watch(
  () => [props.visible, props.preview] as const,
  ([visible]) => {
    if (visible && props.preview) titleDraft.value = props.preview.title
  },
  { immediate: true },
)

const kb = computed(() => Math.max(1, Math.round(props.fileBytes / 1024)))

const titleHint = computed(() => {
  switch (props.preview?.title_source) {
    case 'frontmatter':
      return "From the file's front matter."
    case 'heading':
      return "From the document's first heading."
    default:
      // The one case somebody usually wants to change, so it says so plainly.
      return 'From the filename — the file gave no title and the document has no heading.'
  }
})

const bodyHead = computed(() => (props.preview?.body ?? '').split('\n').slice(0, BODY_LINES).join('\n'))

/**
 * The ledger, ordered attention → unknown → references → refused → ignored.
 *
 * `ignored` is last and quietest because every file that came out of this
 * product's own download carries `code`, `research`, `section`, `created` and
 * `updated`. Put that group first and people learn within two imports that the
 * ledger is noise, which is the one thing it cannot afford to become.
 */
const groups = computed(() => {
  const p = props.preview
  if (!p) return []
  const out: Array<{ tone: 'attention' | 'note'; label: string; items: any[] }> = []

  // Only what was offered and could not be used. A field the file never
  // answered is not a value this section "could not take", and saying so about
  // every document from every other tool — which is what a required field
  // guarantees — is how the warning tone stops meaning anything by the fifth
  // import.
  const attention = [
    ...(p.metadata_report.invalid_values || []).map((i) => ({
      kind: 'left out',
      ...i,
      reason: `${i.reason} The document imports without it.`,
    })),
    ...(p.fields || []).filter((f) => !f.applied).map((f) => ({ kind: 'replaced', key: f.key, value: f.value, reason: f.reason || '' })),
  ]
  if (attention.length) {
    out.push({ tone: 'attention', label: label(attention.length, 'value this section could not take', 'values this section could not take'), items: attention })
  }

  // Its own group, and a quiet one: this blocks nothing, and the entry page's
  // metadata block answers it in two clicks once the document exists.
  const missing = p.metadata_report.missing_required || []
  if (missing.length) {
    out.push({
      tone: 'note',
      label: label(missing.length, 'field this section requires is unanswered', 'fields this section requires are unanswered'),
      items: missing.map((key) => ({
        key,
        reason: 'The file does not answer it. This does not stop the import — you can fill it in on the document afterwards.',
      })),
    })
  }

  const unknown = p.metadata_report.unknown_keys || []
  if (unknown.length) {
    out.push({ tone: 'note', label: label(unknown.length, 'key this section does not declare', 'keys this section does not declare'), items: unknown })
  }

  const refs = p.unresolved_refs || []
  if (refs.length) {
    out.push({
      tone: 'note',
      label: label(refs.length, 'reference does not resolve here', 'references do not resolve here'),
      items: refs.map((r) => ({
        key: `[[${r.ref}]]`,
        reason: r.count > 1
          ? `Nothing in this research answers to it. It appears ${r.count} times, and is kept exactly as written.`
          : 'Nothing in this research answers to it. It is kept exactly as written.',
      })),
    })
  }

  if ((p.refused || []).length) {
    out.push({ tone: 'note', label: label(p.refused!.length, 'key refused', 'keys refused'), items: p.refused! })
  }
  // Applied-with-a-caveat notes sit with the ignored ones: nothing was lost,
  // but a guess was made and somebody may want to know which.
  const readNotUsed = [
    ...(p.ignored || []),
    ...(p.fields || []).filter((f) => f.applied).map((f) => ({ key: f.key, value: f.value, reason: f.reason || '' })),
  ]
  if (readNotUsed.length) {
    out.push({ tone: 'note', label: label(readNotUsed.length, 'key read, not used as given', 'keys read, not used as given'), items: readNotUsed })
  }
  return out
})

function label(n: number, one: string, many: string) {
  return `${n} ${n === 1 ? one : many}`
}

function submit() {
  const title = titleDraft.value.trim()
  if (!title) return
  emit('commit', { title })
}
</script>

<style scoped>
/* Pinned, because the card scrolls as one region: the arithmetic puts the
   footer 74px past the bottom on a 900px screen for a file with no notes at
   all, and 500px past it for one with a full ledger. A confirm dialog whose
   confirm is below the fold is not confirming anything. */
.import-footer {
  position: sticky;
  bottom: 0;
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
}

.import-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  padding: var(--space-5) var(--space-6);
}

.import-file {
  margin: 0;
  font-size: var(--type-xs);
  color: var(--color-text-faint);
}

.import-field { display: flex; flex-direction: column; gap: var(--space-2); }
.import-field .form-label { margin-bottom: 0; }
.import-hint { margin: 0; font-size: var(--type-xs); color: var(--color-text-faint); }

.import-read {
  display: grid;
  grid-template-columns: minmax(0, auto) minmax(0, 1fr);
  align-items: baseline;
  gap: var(--space-2) var(--space-4);
  margin: 0;
  font-size: var(--type-sm);
}
.import-read dt { margin: 0; color: var(--color-text-muted); }
.import-read dd { margin: 0; color: var(--color-text); min-width: 0; overflow-wrap: anywhere; }

.import-warning {
  margin: 0;
  padding: var(--space-2) var(--space-3);
  border-left: 2px solid var(--color-warning);
  background: var(--color-surface-hover);
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}

.import-heading {
  margin: 0;
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-border);
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-faint);
}

.import-ledger { margin: 0; padding: 0; }

.import-well {
  margin: 0;
  max-height: 22rem;
  overflow: auto;
  padding: var(--space-3);
  border-radius: var(--radius-sm);
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  line-height: 1.6;
  /* Sideways inside itself, never the page. */
  white-space: pre;
}

@media (max-width: 768px) {
  .import-body { padding: var(--space-4); gap: var(--space-4); }
  /* Label above value, the same rule MetadataBlock uses at this width. */
  .import-read { grid-template-columns: minmax(0, 1fr); gap: 0; }
  .import-read dt { margin-top: var(--space-2); font-size: var(--type-xs); }
}
</style>
