/**
 * Converts [[E3]], [[R2:E5]], [[R2]] references in text to clickable HTML links.
 */
export function renderRefs(text: string, researchSlug: string): string {
  if (!text) return ''
  return text.replace(/\[\[([^\]]+)\]\]/g, (_, ref: string) => {
    const parts = ref.split(':')
    let href: string
    if (parts.length === 2) {
      href = `/research/${parts[0]}/entry/${parts[1]}`
    } else if (ref.startsWith('R')) {
      href = `/research/${ref}`
    } else {
      href = `/research/${researchSlug}/entry/${ref}`
    }
    return `<a href="${href}" class="crossref-link">${ref}</a>`
  })
}
