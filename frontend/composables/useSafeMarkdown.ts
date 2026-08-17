/**
 * The one markdown parser this product renders with.
 *
 * `marked` passes raw HTML in a document straight through — that is what
 * markdown is specified to do — and every rendered field here lands in
 * `v-html`. So a `<script>` smuggled into an entry body ran in the reader's
 * origin, and on a share link it ran in a stranger's. `[click](javascript:…)`
 * did the same with ordinary link syntax and no HTML at all.
 *
 * What settles it is that the product already made the opposite decision
 * elsewhere and this path contradicted it: an entry whose author *declares* it
 * is HTML becomes a blocks document with an `html` block, which
 * `ArtifactFrame` renders inside a sandboxed iframe with no same-origin
 * access. HTML the author asked for was isolated; HTML the author hid was
 * executed. Escaping here makes the two agree, and leaves the iframe as the one
 * sanctioned way to ship markup.
 *
 * It is a private `Marked` instance rather than the global one because the
 * global was configured from eight different files, and a parser whose safety
 * depends on every future caller remembering to configure it is not safe. The
 * two share pages had in fact never called `setOptions` at all, so a shared
 * entry rendered with different options than the owner's copy of it.
 */
import { Marked } from 'marked'
import { escapeHtml } from '~/utils/escapeHtml'
import { safeUrl } from '~/utils/safeUrl'

const parser = new Marked({ gfm: true, breaks: true })

parser.use({
  renderer: {
    // Block and inline HTML both arrive here. Showing the markup as text is
    // deliberate: an author who typed `<div>` should see that they typed it,
    // rather than watch it vanish.
    html({ text }: { text: string }) {
      return escapeHtml(text)
    },
    link({ href, title, tokens }: any) {
      const text = this.parser.parseInline(tokens)
      const safe = safeUrl(href)
      // A refused link keeps its label and loses its destination; the reader
      // sees the words the author wrote, not a dead control.
      if (!safe) return text
      const external = /^https?:\/\//i.test(safe)
      const attrs = external ? ' rel="nofollow noopener" target="_blank"' : ''
      const t = title ? ` title="${escapeHtml(title)}"` : ''
      return `<a href="${escapeHtml(safe)}"${t}${attrs}>${text}</a>`
    },
    image({ href, title, text }: any) {
      const safe = safeUrl(href)
      if (!safe) return escapeHtml(text || '')
      const t = title ? ` title="${escapeHtml(title)}"` : ''
      return `<img src="${escapeHtml(safe)}" alt="${escapeHtml(text || '')}"${t}>`
    },
  },
})

/** A markdown document — an entry body, session notes, a task result. */
export function parseMarkdown(text: string | null | undefined): string {
  if (!text) return ''
  return parser.parse(text) as string
}

/** One line of markdown, with no wrapping paragraph. */
export function parseMarkdownInline(text: string | null | undefined): string {
  if (!text) return ''
  return parser.parseInline(text) as string
}
