/**
 * Shortens a label to fit somewhere it would otherwise overflow.
 *
 * Eleven components had written this out, in two versions: some guarded the
 * empty case and some did not, so a node whose title had not loaded yet threw
 * in half of them and rendered in the other half. Which half you got depended
 * on which file you were looking at, which is the whole argument for there
 * being one.
 *
 * The ellipsis is the character, not three dots. It is one glyph wide, it is
 * what a screen reader announces as an ellipsis, and it cannot be mistaken for
 * the author having typed three full stops.
 */
export function truncate(text: string | null | undefined, max: number): string {
  if (!text) return ''
  if (text.length <= max) return text
  // Cutting mid-word is worse than cutting a word early, when there is a space
  // close enough to the limit to be the obvious break.
  const cut = text.slice(0, max)
  const space = cut.lastIndexOf(' ')
  return (space > max * 0.6 ? cut.slice(0, space) : cut).trimEnd() + '…'
}
