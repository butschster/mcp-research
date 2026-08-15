/**
 * Converts [[E3]], [[R2:E5]], [[R2]], [[RM1]], [[RM1:N3]] references in text to clickable HTML links.
 *
 * Under a share link the rule tightens: a reference renders as a link only when
 * its target is inside the shared research **and** the link includes that kind
 * of target. Everything else renders as bare text — no span, no class, no
 * title, no cursor change.
 *
 * That last part is deliberate and is the whole point. A distinct visual
 * treatment for a dead reference is itself a statement that something is there,
 * and `[[R2:E5]]` in a shared document must not tell a client that R2 exists.
 * The brackets are stripped, so `as covered in [[R2:E5]]` reads as `as covered
 * in R2:E5` — which is how a document code reads in prose, and the string was in
 * the author's own text already.
 */
export function renderRefs(text: string, researchSlug: string): string {
  if (!text) return ''
  // Normalize literal \n and \t from MCP clients before rendering
  text = text.replace(/\\n/g, '\n').replace(/\\t/g, '\t')

  const shared = shareActive()
  const include = shareInclude()
  const ownCode = shareResearchCode()

  return text.replace(/\[\[([^\]]+)\]\]/g, (_, ref: string) => {
    const parts = ref.split(':')
    const head = parts[0] ?? ''
    const tail = parts[1] ?? ''
    let href: string

    if (parts.length === 2 && head.startsWith('RM')) {
      // [[RM1:N3]] — link to roadmap (node context)
      if (shared && !include.roadmaps) return ref
      href = roadmapPath(researchSlug, head)
    } else if (parts.length === 2) {
      // [[R2:E5]] — cross-research entry. Inside a share it is a link only when
      // "R2" happens to be the shared research itself.
      if (shared && head !== ownCode) return ref
      href = shared ? entryPath(researchSlug, tail) : `/research/${head}/entry/${tail}`
    } else if (ref.startsWith('RM')) {
      // [[RM1]] — roadmap link
      if (shared && !include.roadmaps) return ref
      href = roadmapPath(researchSlug, ref)
    } else if (ref.startsWith('R')) {
      // [[R2]] — research link
      if (shared && ref !== ownCode) return ref
      href = shared ? researchPath(researchSlug) : `/research/${ref}`
    } else {
      // [[E3]] — entry in same research
      href = entryPath(researchSlug, ref)
    }

    return `<a href="${href}" class="crossref-link">${ref}</a>`
  })
}
