/**
 * Time formatting shared by anything that shows when something was written.
 *
 * `relativeTime` already existed privately inside ResearchCard; it is lifted
 * here so the revision history and the card agree on what "2h ago" means.
 */
export function relativeTime(iso: string): string {
  if (!iso) return ''
  const then = new Date(iso).getTime()
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
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleString(undefined, opts)
}
