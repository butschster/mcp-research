/**
 * Dropping one markdown file into one section, in two calls.
 *
 * The state machine lives here rather than in `EntriesView`, which already owns
 * the search, the tag filter, the card/table toggle and the grouping. A fifth
 * responsibility with four phases inside it is how that file stops being
 * reviewable.
 *
 * The two calls are the whole design: the preview writes nothing, and the human
 * sees what we made of the file before anything exists. The commit sends the
 * parsed fields rather than the file again — an edit made in the preview
 * survives, and the request has no field through which `session`, `author` or a
 * revision block could come back, so that refusal is enforced by the shape of
 * the payload rather than by remembering.
 */
import type { Ref } from 'vue'

export interface ImportNote {
  key: string
  value?: string
  reason: string
}

export interface FieldNote {
  key: string
  value?: string
  applied: boolean
  reason?: string
}

export interface UnresolvedRef {
  ref: string
  count: number
}

export interface ImportPreview {
  filename: string
  title: string
  title_source: 'frontmatter' | 'heading' | 'filename'
  description?: string
  status: string
  tags: string[]
  body: string
  body_lines: number
  metadata?: Record<string, unknown>
  metadata_report: {
    stored?: string[]
    unknown_keys?: ImportNote[]
    invalid_values?: ImportNote[]
    missing_required?: string[]
    spec_version: number
  }
  fields?: FieldNote[]
  ignored?: ImportNote[]
  refused?: ImportNote[]
  unresolved_refs?: UnresolvedRef[]
  warnings?: string[]
}

export interface ImportRefusal {
  message: string
  /** Whether picking a different file is the way forward, or retrying this one. */
  retryable: boolean
}

export type ImportPhase = 'idle' | 'reading' | 'previewing' | 'committing'

export function useSectionImport(sectionId: Ref<string>) {
  const { authFetch } = useAuth()
  const base = useRuntimeConfig().public.apiBase || ''

  const phase = ref<ImportPhase>('idle')
  const file = ref<File | null>(null)
  const preview = ref<ImportPreview | null>(null)
  const refusal = ref<ImportRefusal | null>(null)
  const commitError = ref('')
  // Set when a `section.updated` arrives while the dialog is open: the field
  // spec the preview was validated against is no longer the one the commit will
  // be checked by, and importing under a stale one produces a rejection the
  // person has no way to understand.
  const staleSpec = ref(false)

  /** The server's own message, never a sentence of ours. */
  function messageOf(e: any, fallback: string): string {
    return e?.data?.error || e?.response?._data?.error || fallback
  }

  async function read(f: File) {
    file.value = f
    preview.value = null
    refusal.value = null
    commitError.value = ''
    staleSpec.value = false
    phase.value = 'reading'

    const form = new FormData()
    form.append('file', f, f.name)
    try {
      const res = await authFetch<{ data: ImportPreview }>(
        `${base}/api/sections/${sectionId.value}/import/preview`,
        { method: 'POST', body: form },
      )
      preview.value = res.data
      phase.value = 'previewing'
    } catch (e: any) {
      // A refusal here is about the file, so a different file is the way
      // forward — except a 5xx, which is about us.
      const status = e?.statusCode ?? e?.response?.status ?? 0
      refusal.value = {
        message: messageOf(e, 'That file could not be read.'),
        retryable: status >= 500,
      }
      phase.value = 'idle'
      file.value = null
    }
  }

  /** Re-reads the same file, for when the section's fields changed underneath. */
  async function reread() {
    if (file.value) await read(file.value)
  }

  async function commit(overrides: { title: string }): Promise<{ id: string; code: string } | null> {
    if (!preview.value) return null
    commitError.value = ''
    phase.value = 'committing'
    try {
      const res = await authFetch<{ data: { id: string; code: string } }>(
        `${base}/api/sections/${sectionId.value}/import`,
        {
          method: 'POST',
          body: {
            title: overrides.title,
            description: preview.value.description ?? '',
            status: preview.value.status,
            tags: preview.value.tags,
            body: preview.value.body,
            metadata: preview.value.metadata ?? {},
          },
        },
      )
      reset()
      return res.data
    } catch (e: any) {
      // The dialog stays open. Everything the person did — the title they
      // fixed, the report they read — is still on screen, and throwing it away
      // to show a toast would make them do the whole thing again.
      commitError.value = messageOf(e, 'That document could not be created.')
      phase.value = 'previewing'
      return null
    }
  }

  function reset() {
    phase.value = 'idle'
    file.value = null
    preview.value = null
    refusal.value = null
    commitError.value = ''
    staleSpec.value = false
  }

  function dismissRefusal() {
    refusal.value = null
  }

  function markSpecStale() {
    if (phase.value === 'previewing') staleSpec.value = true
  }

  return {
    phase,
    file,
    preview,
    refusal,
    commitError,
    staleSpec,
    read,
    reread,
    commit,
    reset,
    dismissRefusal,
    markSpecStale,
  }
}
