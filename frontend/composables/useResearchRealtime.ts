/**
 * Subscribe to the events of one research, whichever way the URL names it.
 *
 * Events carry the research's UUID. Every link in the product is built from the
 * short code — `/research/R7` — so six pages compared a code against a UUID,
 * matched nothing, and dropped every event they were sent. Real-time updates
 * were dead on all of them, silently, for anyone who arrived by clicking a
 * link, which is everyone.
 *
 * The comparison lives here so no page has to remember which identity its route
 * happened to use, and so the next page to be added inherits the fix.
 */

/**
 * Does this event concern the research the URL names?
 *
 * Exported because the revocation handler asks the same question, and a second
 * copy of this rule had already drifted from the first by the time it was
 * noticed. `name` is whatever the route carries — a short code from a link, a
 * UUID from a pasted address — and either identity is a match.
 */
export function matchesResearch(
  event: Pick<WsEvent, 'research_id' | 'research_code'>,
  name: string,
  resolvedId?: string | null,
) {
  if (!event.research_id && !event.research_code) return false
  return event.research_id === name
    || event.research_code === name
    || (!!resolvedId && event.research_id === resolvedId)
}

/** How long events are collected before the handler is called. */
const BURST_MS = 80

interface Options {
  /** Refetch after a dropped connection: events sent while it was down are gone. */
  onResync?: () => void
  /**
   * The research's UUID once it is loaded. Not required — the short code on the
   * event is enough — but it closes the window before the first fetch returns.
   */
  researchId?: MaybeRefOrGetter<string | undefined | null>
}

export function useResearchRealtime(
  slug: MaybeRefOrGetter<string>,
  handler: (event: WsEvent) => void,
  options: Options = {},
) {
  if (import.meta.server) return

  // One event per entity per burst. Adding twelve questions to a session emits
  // twelve events — correct, since each names its own question — and the page
  // only needs to hear "questions changed" once.
  let pending = new Map<string, WsEvent>()
  let timer: ReturnType<typeof setTimeout> | null = null

  function flush() {
    timer = null
    const batch = [...pending.values()]
    pending.clear()
    batch.forEach(handler)
  }

  function matches(event: WsEvent) {
    return matchesResearch(event, toValue(slug), toValue(options.researchId))
  }

  useRealtimeUpdates(
    (event) => {
      if (!matches(event)) return

      pending.set(event.entity, event)
      if (!timer) timer = setTimeout(flush, BURST_MS)
    },
    { onResync: options.onResync },
  )

  onUnmounted(() => {
    if (timer) clearTimeout(timer)
  })
}
