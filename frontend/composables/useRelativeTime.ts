/**
 * Time formatting shared by anything that shows when something was written.
 *
 * `relativeTime` already existed privately inside ResearchCard; it is lifted
 * here so the revision history and the card agree on what "2h ago" means.
 */
/**
 * Reads a timestamp the API might send in either shape.
 *
 * SQLite stores `2026-08-14 15:59:40` and the JSON encoder emits
 * `2026-08-14T15:59:40Z`; a bare `new Date()` reads the first as **local** time
 * and the second as UTC, so the same instant came out hours apart depending on
 * which route returned it. Everything that parses a timestamp goes through
 * here.
 */
export function parseTimestamp(iso: string): Date {
  if (!iso) return new Date(NaN)
  return new Date(iso.includes('T') ? iso : iso.replace(' ', 'T') + 'Z')
}

export function relativeTime(iso: string): string {
  if (!iso) return ''
  const then = parseTimestamp(iso).getTime()
  if (Number.isNaN(then)) return iso

  const diff = Date.now() - then
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  const days = Math.floor(hrs / 24)
  if (days < 30) return `${days}d ago`
  return absoluteTime(iso, { dateStyle: 'medium' })
}

/** The full timestamp, for a title attribute or a sentence that has room. */
export function absoluteTime(
  iso: string,
  opts: Intl.DateTimeFormatOptions = { dateStyle: 'medium', timeStyle: 'short' },
): string {
  if (!iso) return ''
  const d = parseTimestamp(iso)
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleString(undefined, opts)
}

/**
 * Wall-clock time from a millisecond timestamp, for "last update 14:02".
 *
 * Separate from `absoluteTime` because that one takes an ISO string through
 * `parseTimestamp`, and the realtime layer works in milliseconds throughout —
 * the mismatch is what produced two copies of this expression in the first
 * place, one in the tooltip and one in the toast, shown at the same moment.
 */
export function clockTime(ms: number | null | undefined): string {
  if (!ms) return ''
  return new Date(ms).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
