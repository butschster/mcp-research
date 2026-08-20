/**
 * A control a reader flips, written to the server behind them.
 *
 * Two of these existed before this file: the checklist branch in
 * `BlockRenderer.vue` and the task rows in `TaskRefBlock.vue`. Same four phases
 * in the same order, the comments copied word for word — and already disagreeing
 * about the two things that matter:
 *
 *   - a `403` meant "your session expired" in one and "you can't change tasks in
 *     this research" in the other, so a reader ticking a checklist and a task
 *     list in ONE document was told two different stories about one refusal;
 *   - fresh data from the server cleared the optimistic map unconditionally in
 *     the checklist, so an agent rewriting the entry mid-tick snapped the box
 *     back before the response landed. The task rows guarded against it. That
 *     guard is the default here.
 *
 * The shape is deliberately not a store: state is per component instance,
 * because two blocks in one document are two independent sets of controls.
 */

/** How the four phases leave their marks, keyed by whatever the caller calls a row. */
export interface OptimisticWrite<T> {
  /** Values applied locally and not yet agreed by the server. */
  optimistic: Ref<Record<string, T>>
  /** Rows with a request in the air. */
  pending: Ref<Set<string>>
  /** One message per row. Not a flag: the server says what went wrong, and
   *  "reload the page" is the wrong answer to most of it. */
  failed: Ref<Record<string, string>>
  /**
   * Apply `value` to `key` at once, then `send()` it.
   *
   * A key already in flight is ignored rather than queued — a double click on a
   * checkbox is one intention, not two.
   */
  write(key: string, value: T, send: () => Promise<unknown>, overrides?: Record<number, string>): Promise<void>
  /**
   * Fresh data from the server is the truth again.
   *
   * Rows still in flight keep their optimistic value: they are the only thing
   * holding the control where the reader put it, and dropping them is what made
   * a box snap back mid-request.
   */
  reset(): void
}

export function useOptimisticWrite<T>(): OptimisticWrite<T> {
  const optimistic = ref<Record<string, T>>({}) as Ref<Record<string, T>>
  const pending = ref<Set<string>>(new Set())
  const failed = ref<Record<string, string>>({})

  async function write(key: string, value: T, send: () => Promise<unknown>, overrides?: Record<number, string>) {
    if (pending.value.has(key)) return

    optimistic.value = { ...optimistic.value, [key]: value }
    pending.value = new Set(pending.value).add(key)
    const cleared = { ...failed.value }
    delete cleared[key]
    failed.value = cleared

    try {
      await send()
    } catch (e: any) {
      // Revert visibly. A silent revert is worse than an error: the reader
      // re-ticks and assumes it stuck.
      const back = { ...optimistic.value }
      delete back[key]
      optimistic.value = back
      failed.value = { ...failed.value, [key]: saveError(e, overrides) }
    } finally {
      const p = new Set(pending.value)
      p.delete(key)
      pending.value = p
    }
  }

  function reset() {
    const keep: Record<string, T> = {}
    for (const key of pending.value) {
      const v = optimistic.value[key]
      if (v !== undefined) keep[key] = v
    }
    optimistic.value = keep
    failed.value = {}
  }

  return { optimistic, pending, failed, write, reset }
}

/**
 * What a failed write says, from whatever `authFetch` threw.
 *
 * A plain function rather than a `use*`: there is no reactive state in it, and
 * this directory already ships pure helpers this way (`renderRefs`, `tagHue`,
 * `taskPath`).
 *
 * `overrides` is for a status whose meaning is specific to one caller — a `409`
 * on a document patch, a `404` on a task. Everything else has to read the same
 * everywhere, which is the whole reason this is one function.
 */
export function saveError(e: any, overrides?: Record<number, string>): string {
  const status = e?.status ?? e?.response?.status
  const detail = e?.data?.error ?? e?.response?._data?.error
  if (status && overrides?.[status]) return overrides[status]!
  if (status === 401) return 'Not saved — your session expired. Sign in again.'
  if (status === 403) return 'Not saved — you are not allowed to change this.'
  if (!status) return 'Not saved — the server could not be reached.'
  return detail ? `Not saved — ${detail}` : 'Not saved.'
}
