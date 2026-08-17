/**
 * Decides whether a URL out of authored content may be used as an `href`.
 *
 * Everything rendered in this product is written by an agent or pasted by a
 * person, and it is read by other people — teammates, and strangers holding a
 * share link. `javascript:` and `data:` in a link are the two that turn reading
 * into executing, and ordinary markdown link syntax reaches both.
 *
 * The rule is an allow-list on purpose. A deny-list has to anticipate
 * `JaVaScRiPt:`, `java\tscript:`, a leading NUL, and whatever the next parser
 * quirk turns out to be; an allow-list only has to recognise the four shapes a
 * research document legitimately contains.
 */

/** Control characters and whitespace a browser strips before resolving a URL. */
const IGNORE = /[\u0000-\u0020\u007f]/g

export function safeUrl(url: string): string | null {
  if (!url) return null
  // Entity-decoded because one caller hands us a string that has already been
  // escaped, where `&` arrives as `&amp;`.
  const decoded = url.replace(/&amp;/g, '&')
  // Tested against a copy with the characters a browser would drop removed, so
  // `java\tscript:` cannot pass a check that `javascript:` fails.
  const probe = decoded.replace(IGNORE, '')

  if (/^https?:\/\//i.test(probe)) return decoded
  if (/^mailto:/i.test(probe)) return decoded
  // Same-origin, but not protocol-relative — `//evil.example` is another host.
  if (probe.startsWith('/') && !probe.startsWith('//')) return decoded
  // A fragment stays on the page it is already on.
  if (probe.startsWith('#')) return decoded
  return null
}
