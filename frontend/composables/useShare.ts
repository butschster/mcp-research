export interface ShareInclude {
  sessions: boolean
  tasks: boolean
  roadmaps: boolean
  export: boolean
}

export interface SharePayloadMeta {
  label?: string
  owner_name?: string
  research_id: string
  research_code?: string
  expires_at?: string | null
  include: ShareInclude
}

/**
 * The share link the reader arrived on, if any.
 *
 * Module state, for the same reason `useResearchRole` is: the components that
 * have to know — a cross-reference deep inside rendered markdown, a path helper
 * called from a template — are the ones hardest to reach with provide/inject,
 * and there is exactly one shared research on screen at a time.
 *
 * The default is "no share", so every existing page behaves as it always did
 * without knowing this file exists.
 */
const token = ref<string>('')
const meta = ref<SharePayloadMeta | null>(null)
/** What the password exchange handed back. Empty for an unprotected link. */
const unlock = ref<string>('')

/**
 * The research and its sections, fetched once by the shell.
 *
 * They live here rather than being provided down the tree because the shell is
 * a route parent: a child re-created by navigation would otherwise refetch the
 * payload on every click, and each of those fetches counts as a visit.
 */
const research = ref<any | null>(null)
const sections = ref<any[]>([])

const NO_INCLUDE: ShareInclude = { sessions: false, tasks: false, roadmaps: false, export: false }

export function shareActive() {
  return !!token.value
}

export function shareToken() {
  return token.value
}

export function shareInclude(): ShareInclude {
  return meta.value?.include ?? NO_INCLUDE
}

/** The research code, which is the only identifier that survives redaction. */
export function shareResearchCode() {
  return meta.value?.research_code || ''
}

/**
 * Whether a link still works: not revoked, not past its date.
 *
 * One definition, because three surfaces have to agree about it — the row, the
 * "all links revoked" line under the list, and the count on the research
 * page's Share button. An owner told a link is live in one place and dead in
 * another has no way to decide which to believe.
 */
export function isShareLive(share: { revoked_at?: string | null; expires_at?: string | null }) {
  if (share.revoked_at) return false
  return !share.expires_at || parseTimestamp(share.expires_at).getTime() > Date.now()
}

/**
 * What a link includes, in words.
 *
 * Also one definition, and for a sharper reason: the owner reads this on the
 * row and the visitor reads it in the banner, and the two were describing the
 * same four flags with different nouns.
 */
export function shareContents(include: ShareInclude, style: 'prose' | 'list' = 'list') {
  const parts = ['Documents']
  if (include?.roadmaps) parts.push('roadmaps')
  if (include?.sessions) parts.push(style === 'prose' ? 'interview sessions' : 'sessions')
  if (include?.tasks) parts.push('tasks')
  if (include?.export) parts.push('downloads')
  if (style !== 'prose') return parts.join(', ')
  if (parts.length === 1) return 'Documents only'
  return parts.slice(0, -1).join(', ') + ' and ' + parts[parts.length - 1]
}

/** When a link lapses, said the way a person would say it. */
export function shareExpiryPhrase(expiresAt?: string | null) {
  if (!expiresAt) return ''
  const when = parseTimestamp(expiresAt)
  if (Number.isNaN(when.getTime())) return 'at some point'
  const days = Math.round((when.getTime() - Date.now()) / 86_400_000)
  if (days <= 0) return 'today'
  if (days === 1) return 'tomorrow'
  if (days < 30) return `in ${days} days`
  return `on ${when.toLocaleDateString()}`
}

function unlockKey(t: string) {
  return `share_unlock_${t}`
}

export function useShare() {
  const config = useRuntimeConfig()
  const base = config.public.apiBase || ''

  /**
   * Adopts a token before the payload is known.
   *
   * Called first thing in the shell's setup, so the read-only lock and the path
   * helpers are already correct while the fetch is still in flight — a page that
   * became read-only a moment after rendering would flash edit controls at
   * somebody who has no account.
   */
  function claim(t: string) {
    token.value = t
    if (!import.meta.server) unlock.value = sessionStorage.getItem(unlockKey(t)) || ''
  }

  function setPayload(payload: { share: SharePayloadMeta; research: any; sections: any[] } | null) {
    meta.value = payload?.share ?? null
    research.value = payload?.research ?? null
    sections.value = payload?.sections ?? []
  }

  function setUnlock(value: string) {
    unlock.value = value
    if (import.meta.server) return
    // sessionStorage, not localStorage: a password typed on one tab is not a
    // standing grant on every tab of the browser, and it does not outlive the
    // window it was typed in.
    if (value) sessionStorage.setItem(unlockKey(token.value), value)
    else sessionStorage.removeItem(unlockKey(token.value))
  }

  /**
   * Forgets the share. Only a caller genuinely leaving the shared view calls
   * this; it is not an unmount hook, for the reason `useResearchRole.clear`
   * documents.
   */
  function release() {
    token.value = ''
    meta.value = null
    unlock.value = ''
    research.value = null
    sections.value = []
  }

  /**
   * The visitor's fetch.
   *
   * Deliberately not `useApi()`: that attaches an `Authorization` header
   * whenever a session exists, and an owner opening their own share link must be
   * served as a visitor — otherwise the one person who most needs to check what
   * the link exposes is the one person who cannot see it.
   */
  function shareFetch<T>(path: string, options: Record<string, any> = {}) {
    const headers: Record<string, string> = { ...(options.headers || {}) }
    if (unlock.value) headers['X-Share-Unlock'] = unlock.value
    return $fetch<T>(`${base}/api/shared/${token.value}${path}`, { ...options, headers })
  }

  return {
    token: readonly(token),
    meta: readonly(meta),
    unlock: readonly(unlock),
    active: computed(() => !!token.value),
    include: computed<ShareInclude>(() => meta.value?.include ?? NO_INCLUDE),
    ownerName: computed(() => meta.value?.owner_name || ''),
    researchId: computed(() => meta.value?.research_id || ''),
    researchCode: computed(() => meta.value?.research_code || ''),
    /** How every page under the shell names the research in a URL. */
    slug: computed(() => meta.value?.research_code || meta.value?.research_id || ''),
    expiresAt: computed(() => meta.value?.expires_at || null),
    label: computed(() => meta.value?.label || ''),
    // Not readonly(): these are handed straight to components whose props are
    // ordinary arrays, and a DeepReadonly wrapper makes every one of those a
    // type error for no safety the shell does not already provide.
    research: computed(() => research.value),
    sections: computed<any[]>(() => sections.value),
    claim,
    setPayload,
    setUnlock,
    release,
    shareFetch,
  }
}
