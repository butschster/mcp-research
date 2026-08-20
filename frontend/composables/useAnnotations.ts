/**
 * The marks a person left on places in the documents.
 *
 * Three surfaces read this one resource — the document page, the research
 * queue, and the counter in the research header — and they need different
 * amounts of it. The counter in particular must not pull the whole list to
 * print one number, which is what a single "load everything" composable would
 * have made it do.
 *
 * Nothing here caches across researches. A mark's anchor is computed by the
 * server against the document as it stands at the moment of the read, so a
 * cached list is a set of claims about a document an agent may have rewritten
 * since. Re-reading is the point.
 */

export type AnnotationKind = 'verify' | 'dig' | 'disagree'
export type AnnotationStatus = 'open' | 'answered' | 'closed' | 'dismissed'
export type AnchorState = 'anchored' | 'drifted' | 'moved' | 'orphaned'

export interface AnnotationAnchor {
  state: AnchorState
  strategy: string
  confidence: number
  block_id?: string
  block_index: number
  block_type?: string
  text?: string
}

export interface Annotation {
  id: string
  code: string
  research_id: string
  entry_id: string
  block_id?: string
  quote: { exact: string; prefix?: string; suffix?: string }
  anchored_revision: number
  kind: AnnotationKind
  body?: string
  author_kind: string
  author_name?: string
  user_id?: string
  status: AnnotationStatus
  resolution?: string
  resolved_revision?: number
  session_id?: string
  task_id?: string
  rejections?: Array<{ reason: string; revision?: number; at: string }>
  attempts: number
  created_at: string
  updated_at: string
  answered_at?: string
  closed_at?: string
  anchor?: AnnotationAnchor
  entry_code?: string
  entry_title?: string
  entry_type?: string
}

export interface AnnotationMeta {
  counts?: Record<string, number>
  by_anchor?: Record<string, number>
  by_entry?: Record<string, number>
}

export interface QueueFilters {
  status?: AnnotationStatus | ''
  kind?: AnnotationKind | ''
  anchor?: AnchorState | ''
  entry_id?: string
}

/** The glyph and the word that travel with every mark, so neither colour nor
 *  position is ever the only thing saying what it is. */
export const KIND_META: Record<AnnotationKind, { glyph: string; label: string; hint: string }> = {
  verify: { glyph: '?', label: 'Verify', hint: 'Find a source, or say plainly it could not be confirmed' },
  dig: { glyph: '↓', label: 'Dig', hint: 'Write a child document and link it from here' },
  disagree: { glyph: '≠', label: 'Disagree', hint: 'Record both positions — do not rewrite the text' },
}

/** Anchor state carries shape, not colour: solid, wavy, dashed, gone. */
export const ANCHOR_META: Record<AnchorState, { label: string; underline: string; hint: string }> = {
  anchored: { label: 'Anchored', underline: 'solid', hint: 'The marked text is where it was' },
  drifted: { label: 'Drifted', underline: 'wavy', hint: 'The block is here, but the sentence under this mark changed' },
  moved: { label: 'Moved', underline: 'dashed', hint: 'The sentence turned up elsewhere in the document' },
  orphaned: { label: 'Orphaned', underline: 'none', hint: 'The marked text is gone from the document' },
}

export function useAnnotations() {
  const { authFetch } = useAuth()
  const base = useRuntimeConfig().public.apiBase || ''

  function query(filters: QueueFilters = {}) {
    const params = new URLSearchParams()
    for (const [key, value] of Object.entries(filters)) {
      if (value) params.set(key, String(value))
    }
    const s = params.toString()
    return s ? `?${s}` : ''
  }

  /** Everything marked in one document, placed against the document as it is now. */
  async function listForEntry(entryId: string) {
    const res = await authFetch<{ data: Annotation[] | null }>(`${base}/api/entries/${entryId}/annotations`)
    return res?.data ?? []
  }

  /** The queue. */
  async function listForResearch(researchId: string, filters: QueueFilters = {}) {
    const res = await authFetch<{ data: Annotation[] | null; meta?: AnnotationMeta; counts?: Record<string, number> }>(
      `${base}/api/researches/${researchId}/annotations${query(filters)}`,
    )
    return {
      data: res?.data ?? [],
      meta: res?.meta ?? { counts: res?.counts ?? {} },
    }
  }

  /**
   * The counter in the research header.
   *
   * Asks for one row and reads the counts off the envelope. The list is capped
   * by the server anyway, but a page that only wants a number should not make
   * the server resolve anchors for fifteen documents to produce it.
   */
  async function counts(researchId: string) {
    const res = await listForResearch(researchId, { status: 'open', entry_id: '' })
    return res.meta.counts ?? {}
  }

  /** Creating waits for the server: the code and the anchor are its to decide,
   *  and drawing a mark only to take it back is worse than a moment's wait. */
  async function create(entryId: string, input: {
    block_id?: string
    quote: { exact: string; prefix?: string; suffix?: string }
    kind: AnnotationKind
    body?: string
  }) {
    const res = await authFetch<{ data: Annotation }>(`${base}/api/entries/${entryId}/annotations`, {
      method: 'POST',
      body: input,
    })
    return res.data
  }

  async function patch(id: string, input: Record<string, any>) {
    const res = await authFetch<{ data: Annotation }>(`${base}/api/annotations/${id}`, {
      method: 'PUT',
      body: input,
    })
    return res.data
  }

  /** One decision over a set — the accept-all at the end of reviewing a pass.
   *  The per-row result matters: "twelve of fourteen" is its own outcome. */
  async function bulk(researchId: string, ids: string[], status: AnnotationStatus, reason = '') {
    return await authFetch<{ data: Array<{ id: string; code?: string; ok: boolean; error?: string }>; applied: number; total: number }>(
      `${base}/api/researches/${researchId}/annotations/bulk`,
      { method: 'POST', body: { ids, status, reason } },
    )
  }

  async function remove(id: string) {
    await authFetch(`${base}/api/annotations/${id}`, { method: 'DELETE' })
  }

  return { listForEntry, listForResearch, counts, create, patch, bulk, remove }
}
