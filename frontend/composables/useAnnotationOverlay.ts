/**
 * Drawing marks over prose the page did not build.
 *
 * A block document is rendered through `v-html`, so there is no component to
 * hang an underline on: the only way to mark a sentence is to walk the text
 * nodes afterwards and wrap the range. That is the whole reason this exists as
 * a composable full of DOM work rather than as a template.
 *
 * Two rules run through everything here.
 *
 * **Matching is the server's rule, restated.** The server finds a quote by
 * lowercasing and collapsing whitespace (`normalizeQuote`), and this walks the
 * DOM building exactly that string with a map back to the nodes it came from.
 * If the two ever disagree, the queue says a mark is anchored and the page
 * cannot find the text to underline it — the symptom is silent, so the rule is
 * written twice on purpose and must be changed in both places.
 *
 * **The selection is read, not seized.** An earlier version painted its own
 * highlight over the range and called `removeAllRanges`, so the popover could
 * take focus without losing it. That traded a rare problem for a constant one:
 * selecting a sentence in order to COPY it — far and away the commoner gesture
 * — wiped the selection out from under the reader on every mouseup. So the
 * quote, its context and its rectangle are captured up front, the browser keeps
 * its own selection, and if focus later collapses it nothing is lost: the
 * popover already holds everything the mark needs, and echoes the quote back so
 * the reader can see what they are marking.
 */

export const MARK_CLASS = 'ann-mark'

/** How much either side of a selection is recorded to tell two identical
 *  sentences apart. Matches the server's cap; sending more is harmless, sending
 *  a different *end* of the prefix is not — the words nearest the quote are the
 *  ones that identify it. */
const CONTEXT = 120

export interface CapturedSelection {
  blockId: string
  quote: { exact: string; prefix: string; suffix: string }
  rect: DOMRect
}

interface CharPos {
  node: Text
  offset: number
}

/** Builds the server's comparison string for an element, with a way back to the
 *  nodes each character came from. */
function textMap(root: Element): { text: string; map: CharPos[] } {
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
    // Anything shown only to a screen reader is not prose a person can select,
    // and letting it into the haystack means matching against text nobody sees.
    acceptNode(node) {
      return (node.parentElement?.closest('.sr-only'))
        ? NodeFilter.FILTER_REJECT
        : NodeFilter.FILTER_ACCEPT
    },
  })
  let text = ''
  const map: CharPos[] = []
  let lastWasSpace = true
  let lastBlock: Element | null = null

  let node = walker.nextNode() as Text | null
  while (node) {
    // A block boundary is a space, because that is what the browser puts there.
    // `<li>one</li><li>two</li>` walks as "one"+"two" but selects as "one\ntwo",
    // so without this a selection spanning two list items or two table cells
    // produced a quote the document could never be searched for again — stored
    // silently, findable never.
    const block = blockAncestor(node)
    if (lastBlock && block !== lastBlock && !lastWasSpace) {
      text += ' '
      map.push({ node, offset: 0 })
      lastWasSpace = true
    }
    lastBlock = block

    // Marks already painted are part of the prose; the wrappers around them are
    // not. Skipping the text inside them would make a second paint unable to
    // find anything.
    const raw = node.data
    for (let i = 0; i < raw.length; i++) {
      const ch = raw[i]!
      if (/\s/.test(ch)) {
        if (lastWasSpace) continue
        text += ' '
        map.push({ node, offset: i })
        lastWasSpace = true
        continue
      }
      text += ch.toLowerCase()
      map.push({ node, offset: i })
      lastWasSpace = false
    }
    node = walker.nextNode() as Text | null
  }

  // A trailing space would make a quote ending at the block's end fail to match.
  while (text.endsWith(' ')) {
    text = text.slice(0, -1)
    map.pop()
  }
  return { text, map }
}

/** The nearest ancestor that lays its content out as its own line. */
function blockAncestor(node: Text): Element | null {
  return node.parentElement?.closest('li, td, th, p, div, h1, h2, h3, h4, h5, h6, blockquote, figcaption') ?? null
}

function normalize(s: string) {
  return s.toLowerCase().split(/\s+/).filter(Boolean).join(' ')
}

/** A range over the characters the map points at. */
function rangeFor(map: CharPos[], from: number, length: number): Range | null {
  const start = map[from]
  const end = map[from + length - 1]
  if (!start || !end) return null
  const range = document.createRange()
  range.setStart(start.node, start.offset)
  range.setEnd(end.node, end.offset + 1)
  return range
}

/**
 * Wraps a range in spans.
 *
 * `surroundContents` throws the moment a range crosses an element boundary —
 * which any sentence containing a link or bold text does — so the range is
 * split per text node and wrapped piece by piece.
 */
function wrap(range: Range, className: string, attrs: Record<string, string> = {}): HTMLElement[] {
  const pieces: Array<{ node: Text; start: number; end: number }> = []
  const walker = document.createTreeWalker(
    range.commonAncestorContainer.nodeType === Node.TEXT_NODE
      ? range.commonAncestorContainer.parentNode!
      : range.commonAncestorContainer,
    NodeFilter.SHOW_TEXT,
  )

  let node = walker.nextNode() as Text | null
  while (node) {
    if (range.intersectsNode(node)) {
      const start = node === range.startContainer ? range.startOffset : 0
      const end = node === range.endContainer ? range.endOffset : node.data.length
      if (end > start) pieces.push({ node, start, end })
    }
    node = walker.nextNode() as Text | null
  }

  const spans: HTMLElement[] = []
  // Backwards, so splitting a node does not invalidate the offsets recorded for
  // the pieces still to come.
  for (const piece of pieces.reverse()) {
    const target = piece.start > 0 ? piece.node.splitText(piece.start) : piece.node
    if (piece.end - piece.start < target.data.length) {
      target.splitText(piece.end - piece.start)
    }
    const span = document.createElement('span')
    span.className = className
    for (const [k, v] of Object.entries(attrs)) span.setAttribute(k, v)
    target.parentNode?.insertBefore(span, target)
    span.appendChild(target)
    spans.push(span)
  }
  return spans.reverse()
}

/** Unwraps every span of a class, putting its text back where it was. */
function unwrap(root: Element, className: string) {
  root.querySelectorAll<HTMLElement>(`span.${className}`).forEach((span) => {
    const parent = span.parentNode
    if (!parent) return
    while (span.firstChild) parent.insertBefore(span.firstChild, span)
    parent.removeChild(span)
    // Rejoins the text nodes the wrap split, so the next paint sees the prose as
    // one node again rather than as a growing pile of fragments.
    parent.normalize()
  })
}

export function useAnnotationOverlay() {
  /**
   * Takes the live selection, clamps it to a single block and paints our own
   * highlight over it.
   *
   * Clamped to one block because a mark addresses one block: a selection
   * crossing two is two marks, and silently keeping the first half is worse
   * than refusing. Returns null when there is nothing usable, which is also
   * what a collapsed selection looks like after a plain click.
   */
  function capture(root: Element): CapturedSelection | null {
    const sel = window.getSelection()
    if (!sel || sel.isCollapsed || sel.rangeCount === 0) return null

    const range = sel.getRangeAt(0)
    if (!root.contains(range.commonAncestorContainer)) return null

    // Read before anything touches the selection. Below, the highlight is
    // painted and `removeAllRanges` is called, and a Selection with no ranges
    // stringifies to "" — reading it afterwards produced an empty quote and
    // made every mark unsavable.
    const selected = sel.toString()
    const exact = normalize(selected)
    if (!exact) return null

    const startEl = range.startContainer.nodeType === Node.TEXT_NODE
      ? range.startContainer.parentElement
      : (range.startContainer as Element)
    const block = startEl?.closest('[data-block-id]') as HTMLElement | null

    // A markdown document has no blocks; the whole card is the haystack, and the
    // mark is stored with no block id at all.
    const scope: Element = block ?? root
    const blockId = block?.dataset.blockId ?? ''

    const { text, map } = textMap(scope)
    // Where the selection sits inside the block, so context can be read off the
    // same normalized string the server will search.
    const at = text.indexOf(exact)
    const prefix = at > 0 ? text.slice(Math.max(0, at - CONTEXT), at) : ''
    const suffix = at >= 0 ? text.slice(at + exact.length, at + exact.length + CONTEXT) : ''

    const rect = range.getBoundingClientRect()

    // The browser's selection is left exactly as the reader made it.
    return { blockId, quote: { exact: selected.trim().slice(0, 2000), prefix, suffix }, rect }
  }

  /**
   * Paints every mark that still has text to point at.
   *
   * Orphans are skipped here and shown in the list instead: there is nowhere in
   * the document to draw them, and inventing a place is exactly the silent
   * mis-anchoring the server refuses to do.
   */
  function paint(root: Element, annotations: Annotation[], mode: 'all' | 'open' | 'off') {
    clear(root)
    if (mode === 'off') return

    const visible = annotations.filter((a) => {
      if (mode === 'open' && a.status !== 'open' && a.status !== 'answered') return false
      return a.anchor && a.anchor.state !== 'orphaned'
    })

    // Grouped by block so each block's text is walked once, and so two marks on
    // the same repeated sentence take different occurrences — the same rule the
    // server follows when it resolves them.
    const byBlock = new Map<string, Annotation[]>()
    for (const a of visible) {
      const key = a.anchor?.block_id ?? ''
      byBlock.set(key, [...(byBlock.get(key) ?? []), a])
    }

    for (const [blockId, marks] of byBlock) {
      const scope = blockId
        ? (root.querySelector(`[data-block-id="${CSS.escape(blockId)}"]`) as Element | null)
        : root
      if (!scope) continue

      const taken = new Set<number>()
      for (const a of marks) {
        // Rebuilt per mark: wrapping the previous one split the text nodes the
        // map pointed at.
        const { text, map } = textMap(scope)
        const needle = normalize(a.quote.exact)
        if (!needle) continue

        // Consumption exists for a sentence that REPEATS inside one block, not
        // to refuse a second reader marking the same claim — "find me a source"
        // and "and I do not believe it" is an ordinary pair. So the last
        // occurrence is shared when nothing is free. This mirrors
        // `anchorIndex.take` in internal/service/annotation_anchor.go; the two
        // have to agree or the queue and the page disagree about the same mark.
        let at = text.indexOf(needle)
        let free = at
        while (free >= 0 && taken.has(free)) free = text.indexOf(needle, free + 1)
        if (free >= 0) {
          at = free
          taken.add(free)
        } else if (at < 0) {
          continue
        } else {
          // Share the last occurrence rather than dropping the mark.
          let last = at
          for (let next = text.indexOf(needle, last + 1); next >= 0; next = text.indexOf(needle, next + 1)) last = next
          at = last
        }

        const range = rangeFor(map, at, needle.length)
        if (!range) continue
        wrap(range, `${MARK_CLASS} ${MARK_CLASS}--${a.kind} ${MARK_CLASS}--${a.anchor?.state ?? 'anchored'}`, {
          'data-annotation-id': a.id,
          'data-annotation-code': a.code,
        })
      }
    }

    // What a screen reader announces at the mark.
    //
    // An `aria-label` rather than an appended `.sr-only` span: the span version
    // survived `clear()` — unwrap re-parents a mark's children and removes only
    // the wrapper — so from the second paint onward "(mark A1, verify, open,
    // anchored)" was part of the paragraph's text. Every later search ran
    // against a haystack polluted with its own labels, and marks downstream of
    // one silently stopped matching.
    for (const a of visible) {
      const first = root.querySelector(`span.${MARK_CLASS}[data-annotation-id="${CSS.escape(a.id)}"]`)
      if (!first) continue
      first.setAttribute('role', 'note')
      first.setAttribute('aria-label',
        `${a.quote.exact} — mark ${a.code}, ${a.kind}, ${a.status}, ${a.anchor?.state ?? 'anchored'}`)
    }
  }

  function clear(root: Element) {
    unwrap(root, MARK_CLASS)
  }

  /**
   * Where each mark sits vertically, for the gutter.
   *
   * Measured against the card rather than the viewport, so scrolling does not
   * move them and a scroll listener is not needed — a ResizeObserver on the
   * card is, because reflow does.
   */
  function positions(root: Element): Array<{ id: string; code: string; top: number }> {
    const cardTop = root.getBoundingClientRect().top
    const seen = new Set<string>()
    const out: Array<{ id: string; code: string; top: number }> = []

    root.querySelectorAll<HTMLElement>(`span.${MARK_CLASS}`).forEach((span) => {
      const id = span.dataset.annotationId
      if (!id || seen.has(id)) return
      seen.add(id)
      out.push({
        id,
        code: span.dataset.annotationCode ?? '',
        top: span.getBoundingClientRect().top - cardTop,
      })
    })
    return out.sort((a, b) => a.top - b.top)
  }

  return { capture, paint, clear, positions }
}
