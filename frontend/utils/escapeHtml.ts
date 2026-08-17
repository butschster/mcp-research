/**
 * Makes a string safe to place inside HTML, in text or in an attribute value.
 *
 * It lives in `utils/` rather than beside one of its two callers because both
 * of them are the *start* of a rendering pipeline — `renderInline` for block
 * text, `renderRefs` for a plain field — and a shared helper owned by one of
 * them would have made the other import across a circular edge.
 *
 * Escaping `'` matters as much as `"`: an href interpolated into a
 * single-quoted attribute is just as open as a double-quoted one, and nothing
 * here can know which the caller wrote.
 */
export function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}
