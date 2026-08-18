export interface DownloadError {
  status: number
  message: string
}

/**
 * Downloads a file the API only hands out to an authenticated request.
 *
 * A plain `<a href>` is the obvious thing and the wrong thing: the token lives
 * in localStorage and travels as a header, so a link navigation arrives
 * unauthenticated, gets a 401, and the browser then renders the JSON error body
 * as a page — destroying the SPA state to show the user `{"error":"..."}`.
 *
 * So: fetch the bytes with the header attached, then hand the blob to a
 * synthetic link. `$fetch.raw` rather than `authFetch`, because the filename the
 * server chose lives in a response header and `authFetch` returns only the body.
 */
export function useDownload() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase || ''
  const { token } = useAuth()

  const pending = ref(false)
  const slow = ref(false)
  const filename = ref<string | null>(null)
  const error = ref<DownloadError | null>(null)

  let controller: AbortController | null = null
  // Set when the caller aborts, so leaving the page is not reported as a
  // network failure.
  let cancelled = false
  let slowTimer: ReturnType<typeof setTimeout> | null = null
  let timeoutTimer: ReturnType<typeof setTimeout> | null = null

  // A large research takes a moment; the label has to say so before the user
  // decides nothing happened.
  const SLOW_AFTER_MS = 12_000
  const TIMEOUT_MS = 60_000

  function clearTimers() {
    if (slowTimer) clearTimeout(slowTimer)
    if (timeoutTimer) clearTimeout(timeoutTimer)
    slowTimer = null
    timeoutTimer = null
  }

  function reset() {
    error.value = null
    filename.value = null
    slow.value = false
  }

  function abort() {
    cancelled = true
    controller?.abort()
    controller = null
    clearTimers()
    pending.value = false
  }

  /**
   * @param opts.headers Extra request headers — the share unlock, for a
   *   password-protected link.
   * @param opts.anonymous Sends no `Authorization` even when a session exists.
   *   A share download must go out as a visitor's: an owner opening their own
   *   link would otherwise be served the unredacted vault, and the one person
   *   who most needs to check what the link exposes is the one who could not.
   */
  async function start(
    path: string,
    fallbackName: string,
    opts: { headers?: Record<string, string>; anonymous?: boolean } = {},
  ): Promise<boolean> {
    if (pending.value) return false
    reset()
    pending.value = true
    cancelled = false
    controller = new AbortController()
    slowTimer = setTimeout(() => { slow.value = true }, SLOW_AFTER_MS)
    let timedOut = false
    timeoutTimer = setTimeout(() => {
      timedOut = true
      controller?.abort()
    }, TIMEOUT_MS)

    try {
      const res = await $fetch.raw(path, {
        baseURL: baseURL || undefined,
        responseType: 'blob',
        signal: controller.signal,
        headers: {
          ...(token.value && !opts.anonymous ? { Authorization: `Bearer ${token.value}` } : {}),
          ...(opts.headers || {}),
        },
      })

      const name = filenameFromDisposition(res.headers.get('content-disposition')) || fallbackName
      const blob = res._data as Blob
      saveBlob(blob, name)
      filename.value = name
      return true
    } catch (e: any) {
      // The raw body is Go-side text ("research not found: sql: no rows") that
      // means nothing to a reader, so the status picks the message and the body
      // only reaches the console.
      if (cancelled) return false
      if (e?.response?._data) console.warn('download failed', e.response._data)
      const status = timedOut ? TIMED_OUT : (e?.response?.status ?? e?.statusCode ?? 0)
      error.value = { status, message: describe(status) }
      return false
    } finally {
      clearTimers()
      slow.value = false
      pending.value = false
      controller = null
    }
  }

  onBeforeUnmount(abort)

  return { pending, slow, filename, error, start, reset, abort }
}

/*
 * The messages name no noun.
 *
 * They were written for the vault archive and said "the archive", "export fewer
 * parts", "this research" — which became wrong the day a second caller
 * downloaded one document. Each caller supplies the noun in its own toast
 * title, which is where it belongs.
 */

/** A timeout of our own making, kept apart from the statuses a server sends. */
export const TIMED_OUT = -1

function describe(status: number): string {
  switch (status) {
  case TIMED_OUT:
    // The UI said "still working" moments ago; blaming the connection now
    // would contradict it.
    return 'the server took too long. Nothing was downloaded — try again.'
  case 401:
  case 403:
    return 'your session expired. Sign in again to download.'
  case 404:
    return 'this is no longer available. Reload the page to check.'
  case 0:
    return 'no answer from the server. Check your connection and try again.'
  default:
    return 'the server could not produce the file. Nothing was downloaded.'
  }
}

/**
 * Reads the name the server chose. `filename*` wins because it is the one that
 * survives a Cyrillic research name; the quoted `filename` is the ASCII
 * fallback beside it.
 */
export function filenameFromDisposition(header: string | null): string | null {
  if (!header) return null
  const star = header.match(/filename\*=UTF-8''([^;]+)/i)
  if (star?.[1]) {
    try {
      return decodeURIComponent(star[1].trim())
    } catch {
      // A malformed header is not worth failing a finished download over.
    }
  }
  const plain = header.match(/filename="([^"]*)"/i) || header.match(/filename=([^;]+)/i)
  return plain?.[1]?.trim() || null
}

/**
 * Hands a blob to the browser as a download.
 *
 * The anchor has to be in the document and the object URL has to outlive the
 * click: Firefox and Safari cancel a save from a detached anchor, and revoking
 * synchronously pulls the bytes out from under a download that has not started
 * reading them yet.
 */
export function saveBlob(blob: Blob, name: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = name
  a.style.display = 'none'
  document.body.appendChild(a)
  a.click()
  setTimeout(() => {
    a.remove()
    URL.revokeObjectURL(url)
  }, 0)
}
