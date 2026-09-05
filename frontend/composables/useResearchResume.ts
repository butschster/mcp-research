/**
 * The continuation summary for one research.
 *
 * It owns three things the component must not: when to refetch, which session
 * the reader last chose, and the guarantee that an old response can never
 * overwrite a newer one.
 *
 * Nothing here writes. The only request it makes is a GET, and in particular it
 * never touches the personal new/changed queue — opening a summary is not
 * reading the documents it names.
 */
import { relativeTime } from '~/composables/useRelativeTime'

export type ResumeActor = 'agent' | 'human'

export interface ResumeGroup<T> {
  items: T[]
  returned: number
  /** Null when the count could not be established — never rendered as zero. */
  total: number | null
  has_more: boolean
  more?: { tool?: string; href?: string }
}

export interface ResumeSessionItem {
  id: string
  code: string
  title: string
  focus?: string
  status: string
  updated_at: string
}

export interface ResumeTaskItem {
  id: string
  code: string
  title: string
  status: string
  priority: string
  /** The task's recorded result — where a blocked task says why, when it does. */
  note?: string
  updated_at: string
}

export interface ResumeQuestionItem {
  id: string
  code: string
  session_id: string
  session_code?: string
  text: string
  area?: string
  priority: string
  status: string
}

export interface ResumeAnnotationItem {
  id: string
  code: string
  entry_id: string
  entry_code?: string
  entry_title?: string
  kind: string
  status: string
  quote?: string
  updated_at: string
}

export interface ResumeEntryItem {
  id: string
  code: string
  title: string
  section_id: string
  updated_at: string
  /** Who wrote the newest revision. A human edit is a correction, not stale work. */
  author_kind?: 'agent' | 'human' | 'import' | 'restore'
  revision?: number
}

export interface ResumeAction {
  kind: string
  target: {
    type: string
    id?: string
    code?: string
    title?: string
    session_code?: string
    entry_code?: string
  }
  reason_code: string
  reason: string
  actor: ResumeActor
  tool?: string
  href?: string
}

export interface ResumeSummary {
  schema_version: number
  generated_at: string
  research: { id: string; code: string; name: string; status: string; role?: string; can_write: boolean }
  sessions: {
    items: ResumeSessionItem[]
    selected_id?: string
    selection_required: boolean
    active_count: number
  }
  work: {
    in_progress: ResumeGroup<ResumeTaskItem>
    blocked: ResumeGroup<ResumeTaskItem>
    pending: ResumeGroup<ResumeTaskItem>
  }
  questions: {
    open: ResumeGroup<ResumeQuestionItem>
    deferred: ResumeGroup<ResumeQuestionItem>
  }
  annotations: {
    to_work: ResumeGroup<ResumeAnnotationItem>
    awaiting_human: ResumeGroup<ResumeAnnotationItem>
  }
  recent_entries: ResumeGroup<ResumeEntryItem>
  next_actions: ResumeAction[]
  truncated: boolean
  note?: string
}

/** How long a burst of realtime events is collected before one refetch. */
const REFRESH_DEBOUNCE_MS = 400

/** Remembering the session is per research; remembering the fold is not. */
function sessionKey(researchCode: string) {
  return `research_resume_session_${researchCode}`
}

export function useResearchResume(researchIdOrCode: string, researchCode = researchIdOrCode) {
  const { authFetch } = useAuth()
  const base = useRuntimeConfig().public.apiBase || ''

  const summary = ref<ResumeSummary | null>(null)
  const loading = ref(true)
  const refreshing = ref(false)
  /** Set only when the summary could not be loaded at all. */
  const error = ref<string | null>(null)
  /**
   * True when the endpoint answered 403/404 — an older binary, a research that
   * went away, or a share context. The page then renders exactly as it did
   * before this feature existed rather than painting an error over it.
   */
  const unavailable = ref(false)

  const sessionId = ref('')
  /** Set once a 404 has been retried without the remembered session. */
  let retriedWithoutSession = false

  // A response older than the newest request is dropped. Without this a slow
  // first reply lands after a fast second one and resurrects counts that have
  // already changed — the same guard the entries search carries.
  let seq = 0
  let inFlight = false
  let dirty = false
  let timer: ReturnType<typeof setTimeout> | null = null

  async function load(initial = false) {
    const mine = ++seq
    if (initial) loading.value = true
    else refreshing.value = true
    inFlight = true

    const params = new URLSearchParams({ limit: '5' })
    if (sessionId.value) params.set('session_id', sessionId.value)

    try {
      const res = await authFetch<{ data: ResumeSummary }>(
        `${base}/api/researches/${encodeURIComponent(researchIdOrCode)}/resume?${params}`,
      )
      if (mine !== seq) return
      summary.value = res?.data ?? null
      error.value = null
      unavailable.value = false

      // The server decides which session the summary is about when the reader
      // has not; adopting its answer keeps the picker and the payload agreeing.
      const selected = summary.value?.sessions?.selected_id
      if (selected && selected !== sessionId.value) sessionId.value = selected
    } catch (e: any) {
      if (mine !== seq) return
      const status = e?.statusCode ?? e?.response?.status
      if (status === 404 && sessionId.value && !retriedWithoutSession) {
        // A remembered session that has since been deleted is a 404 from the
        // session lookup, not from the route. Treating it as "this server has
        // no resume endpoint" would hide the block until the page is reloaded.
        retriedWithoutSession = true
        sessionId.value = ''
        forgetSession()
        inFlight = false
        void load(initial)
        return
      }
      if (status === 403 || status === 404) {
        unavailable.value = true
        summary.value = null
        error.value = null
      } else {
        error.value = e?.data?.error || 'The server could not be reached.'
      }
    } finally {
      if (mine === seq) {
        loading.value = false
        refreshing.value = false
        inFlight = false
        // A refresh asked for while one was running fires exactly once more,
        // rather than queueing one request per event in the burst.
        if (dirty) {
          dirty = false
          void load()
        }
      }
    }
  }

  /** Coalesces a burst of realtime events into one refetch. */
  function scheduleRefresh() {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      timer = null
      if (inFlight) {
        dirty = true
        return
      }
      void load()
    }, REFRESH_DEBOUNCE_MS)
  }

  function forgetSession() {
    if (!import.meta.client) return
    try {
      localStorage.removeItem(sessionKey(researchCode))
    } catch {
      // Nothing to clean up if storage is refused.
    }
  }

  async function selectSession(id: string) {
    sessionId.value = id
    if (import.meta.client) {
      try {
        localStorage.setItem(sessionKey(researchCode), id)
      } catch {
        // A browser that refuses storage still gets a working picker.
      }
    }
    await load()
  }

  /**
   * The reader's last choice, read before the first request rather than after
   * it. Adopting it afterwards meant the first load answered about the server's
   * default session, then a second request replaced it — two round trips and a
   * picker that visibly changed value on arrival.
   */
  function restoreSession() {
    if (!import.meta.client || !researchCode) return
    try {
      sessionId.value = localStorage.getItem(sessionKey(researchCode)) || ''
    } catch {
      // A browser that refuses storage simply gets the server's choice.
    }
  }

  const generatedLabel = computed(() =>
    summary.value?.generated_at ? relativeTime(summary.value.generated_at) : '',
  )

  onUnmounted(() => {
    if (timer) clearTimeout(timer)
  })

  return {
    summary,
    loading,
    refreshing,
    error,
    unavailable,
    sessionId,
    generatedLabel,
    load,
    refresh: () => load(),
    scheduleRefresh,
    selectSession,
    restoreSession,
  }
}
