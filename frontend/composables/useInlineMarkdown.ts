/**
 * Inline rendering for block-document text fields.
 *
 * Block text carries a restricted markdown subset — **bold**, *italic*, `code`,
 * [label](url) — plus [[E3]] cross-references. It must never reach v-html raw:
 * escape first, then apply a fixed set of replacements over the escaped string,
 * so a payload containing markup renders as text instead of executing.
 */
import { renderRefs } from '~/composables/useCrossRefs'

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

/** http(s) and domain-relative only — the same rule the server normalizer applies. */
function safeHref(url: string): string | null {
  const decoded = url.replace(/&amp;/g, '&')
  if (/^https?:\/\//i.test(decoded)) return decoded
  if (decoded.startsWith('/') && !decoded.startsWith('//')) return decoded
  return null
}

/**
 * Renders one text field to HTML. Order matters: escaping happens before any
 * markup is introduced, links are validated, and bold runs before italic so `**`
 * is not eaten by the single-asterisk rule.
 */
export function renderInline(text: string | null | undefined, researchSlug = ''): string {
  if (!text) return ''
  let out = escapeHtml(text)

  // [label](url) — the label is already escaped; the href is validated.
  out = out.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (_m, label: string, url: string) => {
    const href = safeHref(url)
    if (!href) return label
    const external = /^https?:\/\//i.test(href)
    const attrs = external ? ' rel="nofollow noopener" target="_blank"' : ''
    return `<a href="${escapeHtml(href)}"${attrs}>${label}</a>`
  })

  out = out.replace(/`([^`]+)`/g, '<code>$1</code>')
  out = out.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  out = out.replace(/(^|[^*])\*([^*]+)\*(?!\*)/g, '$1<em>$2</em>')

  // Cross-references last, so [[E3]] inside a code span or a link label is left
  // alone by the earlier rules and still becomes a link here.
  return renderRefs(out, researchSlug)
}

/** Plain text of a field, for titles and attributes where markup has no place. */
export function stripInline(text: string | null | undefined): string {
  if (!text) return ''
  return text
    .replace(/\[([^\]]+)\]\([^)\s]+\)/g, '$1')
    .replace(/[`*]/g, '')
    .replace(/\[\[([^\]]+)\]\]/g, '$1')
}
