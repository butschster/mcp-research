/**
 * Inline rendering for block-document text fields.
 *
 * Block text carries a restricted markdown subset — **bold**, *italic*, `code`,
 * [label](url) — plus [[E3]] cross-references. It must never reach v-html raw:
 * escape first, then apply a fixed set of replacements over the escaped string,
 * so a payload containing markup renders as text instead of executing.
 */
import { linkRefs } from '~/composables/useCrossRefs'
import { escapeHtml } from '~/utils/escapeHtml'
import { safeUrl } from '~/utils/safeUrl'

/**
 * The URL rule moved to `utils/safeUrl` when the markdown parser started
 * needing it too. Block text and a full markdown document should not disagree
 * about which links are allowed, and while the rule lived here they did: this
 * copy had no case for `mailto:` or a `#fragment`, and did not strip the
 * control characters that let `java\tscript:` past a check for `javascript:`.
 */
const safeHref = safeUrl

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
  // alone by the earlier rules and still becomes a link here. `linkRefs`, not
  // `renderRefs`: `out` was escaped on the first line and escaping it a second
  // time would print the entities rather than the characters.
  return linkRefs(out, researchSlug)
}

/** Plain text of a field, for titles and attributes where markup has no place. */
export function stripInline(text: string | null | undefined): string {
  if (!text) return ''
  return text
    .replace(/\[([^\]]+)\]\([^)\s]+\)/g, '$1')
    .replace(/[`*]/g, '')
    .replace(/\[\[([^\]]+)\]\]/g, '$1')
}
