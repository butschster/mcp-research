/**
 * Converts [[E3]], [[R2:E5]], [[R2]], [[RM1]], [[RM1:N3]] references in text to clickable HTML links.
 */
export function renderRefs(text: string, researchSlug: string): string {
  if (!text) return ''
  // Normalize literal \n and \t from MCP clients before rendering
  text = text.replace(/\\n/g, '\n').replace(/\\t/g, '\t')
  return text.replace(/\[\[([^\]]+)\]\]/g, (_, ref: string) => {
    const parts = ref.split(':')
    let href: string
    if (parts.length === 2 && parts[0].startsWith('RM')) {
      // [[RM1:N3]] — link to roadmap (node context)
      href = `/research/${researchSlug}/roadmap/${parts[0]}`
    } else if (parts.length === 2) {
      // [[R2:E5]] — cross-research entry
      href = `/research/${parts[0]}/entry/${parts[1]}`
    } else if (ref.startsWith('RM')) {
      // [[RM1]] — roadmap link
      href = `/research/${researchSlug}/roadmap/${ref}`
    } else if (ref.startsWith('R')) {
      // [[R2]] — research link
      href = `/research/${ref}`
    } else {
      // [[E3]] — entry in same research
      href = `/research/${researchSlug}/entry/${ref}`
    }
    return `<a href="${href}" class="crossref-link">${ref}</a>`
  })
}
