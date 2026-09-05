/**
 * Builds an interactive frame around a rendered mermaid diagram: pan, zoom,
 * fullscreen and a link to the live editor — what mermaid.live gives you, in
 * place.
 *
 * The DOM is assembled by hand rather than by a component because both callers
 * need it as a detached node: the markdown path swaps it in for a `pre` inside
 * v-html output, and the block renderer drops it into a container Vue is told
 * to leave alone. Nothing here is reactive; a viewer lives and dies with the
 * node it is mounted in.
 */
import mermaid from 'mermaid'
import { mermaidLiveUrl } from '~/composables/useMermaidLive'
import { readThemeColors } from '~/utils/theme'

let initializedTheme = ''

function ensureInit() {
  const theme = document.documentElement.dataset.theme || 'light'
  if (initializedTheme === theme) return
  initializedTheme = theme
  const colors = readThemeColors()
  mermaid.initialize({
    startOnLoad: false,
    theme: 'base',
    themeVariables: {
      darkMode: theme === 'dark',
      background: colors.surface,
      primaryColor: colors.raised,
      primaryTextColor: colors.text,
      primaryBorderColor: colors.primary,
      lineColor: colors.muted,
      secondaryColor: colors.recessed,
      tertiaryColor: colors.background,
      textColor: colors.text,
      nodeTextColor: colors.text,
      edgeLabelBackground: colors.surface,
      fontFamily: "'Outfit', sans-serif",
    },
    flowchart: { curve: 'basis' },
    sequence: { mirrorActors: false },
  })
}

let counter = 0

// One listener, weakly held source, no per-diagram document listeners left
// behind after navigation. Replace only the SVG, preserving pan and zoom.
const diagramSources = new WeakMap<HTMLElement, string>()
let themeRevision = 0
if (typeof window !== 'undefined') {
  window.addEventListener('dovod:theme-change', async () => {
    const revision = ++themeRevision
    for (const stage of document.querySelectorAll<HTMLElement>('.mermaid-stage')) {
      const source = diagramSources.get(stage)
      if (!source || !stage.isConnected) continue
      try {
        ensureInit()
        const { svg } = await mermaid.render(`mermaid-${++counter}`, source)
        if (revision !== themeRevision || !stage.isConnected) return
        const previous = stage.querySelector('svg')
        const width = previous?.style.width
        const height = previous?.style.height
        stage.innerHTML = svg
        const next = stage.querySelector('svg')
        if (next) {
          next.style.maxWidth = 'none'
          next.style.width = width || '100%'
          next.style.height = height || '100%'
          next.setAttribute('preserveAspectRatio', 'xMidYMid meet')
        }
      } catch { /* Keep the last valid diagram if a redraw fails. */ }
    }
  })
}

/** Bounds for zooming by hand. Fitting has its own, lower floor: a huge diagram
 *  must still be shown whole, however small that makes it — and a diagram fitted
 *  below MIN_SCALE lowers the floor to its own scale, or "zoom out" would zoom
 *  in. The frame's own height lives in CSS, since fullscreen has to override it. */
const MIN_SCALE = 0.2
const MAX_SCALE = 8
const FIT_MIN_SCALE = 0.05
/** Inline a small diagram stays its own size; fullscreen may enlarge it. */
const INLINE_MAX_FIT = 1
const FULLSCREEN_MAX_FIT = 3
const FIT_PADDING = 12

const ICON = {
  zoomIn:
    '<svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="11" cy="11" r="7"/><path d="M8 11h6M11 8v6M16.5 16.5 21 21"/></svg>',
  zoomOut:
    '<svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="11" cy="11" r="7"/><path d="M8 11h6M16.5 16.5 21 21"/></svg>',
  expand:
    '<svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 9V4h5M15 4h5v5M20 15v5h-5M9 20H4v-5"/></svg>',
  collapse:
    '<svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 4v5H4M20 9h-5V4M15 20v-5h5M4 15h5v5"/></svg>',
}

const clamp = (v: number, lo: number, hi: number) => Math.min(Math.max(v, lo), hi)

function button(label: string, html: string, onClick: () => void): HTMLButtonElement {
  const b = document.createElement('button')
  b.type = 'button'
  b.className = 'mermaid-btn'
  b.title = label
  b.setAttribute('aria-label', label)
  b.innerHTML = html
  b.addEventListener('click', onClick)
  return b
}

/** "Open in mermaid.live" — the source travels in the URL fragment, so the
 *  diagram is never uploaded anywhere. */
export async function mermaidLiveAnchor(source: string, mode: 'view' | 'edit'): Promise<HTMLAnchorElement> {
  const a = document.createElement('a')
  a.className = 'mermaid-live-link'
  a.href = await mermaidLiveUrl(source, mode)
  a.target = '_blank'
  a.rel = 'noopener noreferrer'
  a.textContent = mode === 'edit' ? 'Open in mermaid.live to debug' : 'mermaid.live'
  return a
}

// Safari still ships these prefixed.
type FsElement = HTMLElement & { webkitRequestFullscreen?: () => void }
type FsDocument = Document & { webkitFullscreenElement?: Element; webkitExitFullscreen?: () => void }

function fullscreenElement(): Element | null {
  const d = document as FsDocument
  return document.fullscreenElement ?? d.webkitFullscreenElement ?? null
}

/**
 * Renders `source` and returns the interactive viewer, or null if mermaid could
 * not parse it — the caller decides what to show instead.
 */
export async function createMermaidViewer(source: string): Promise<HTMLElement | null> {
  ensureInit()

  let svg: string
  try {
    ;({ svg } = await mermaid.render(`mermaid-${++counter}`, source))
  } catch {
    return null
  }

  const view = document.createElement('div')
  view.className = 'mermaid-diagram'

  const canvas = document.createElement('div')
  canvas.className = 'mermaid-canvas'
  canvas.tabIndex = 0
  canvas.setAttribute('role', 'img')
  canvas.setAttribute('aria-label', 'Mermaid diagram. Drag to pan, ctrl and scroll to zoom.')

  const stage = document.createElement('div')
  stage.className = 'mermaid-stage'
  diagramSources.set(stage, source)
  stage.innerHTML = svg
  canvas.appendChild(stage)
  view.appendChild(canvas)

  // The viewBox is the diagram's own coordinate system; mermaid's inline
  // max-width would otherwise fight the transform we scale with.
  const svgEl = stage.querySelector('svg')
  const [, , vw, vh] = (svgEl?.getAttribute('viewBox') || '').split(/[\s,]+/).map(Number)
  const width = vw && vw > 0 ? vw : 800
  const height = vh && vh > 0 ? vh : 600
  if (svgEl) {
    svgEl.style.maxWidth = 'none'
    svgEl.style.width = `${width}px`
    svgEl.style.height = `${height}px`
    svgEl.setAttribute('preserveAspectRatio', 'xMidYMid meet')
  }
  // Only the diagram's own height is inline; the cap and the fullscreen
  // override are CSS. An inline `height` would beat both — inline declarations
  // outrank every author rule — and fullscreen would keep the inline frame.
  view.style.setProperty('--mermaid-natural-h', `${Math.round(height)}px`)

  let scale = 1
  let x = 0
  let y = 0
  // Once the reader has moved the diagram, a container resize must not yank it
  // back to the default framing.
  let touched = false
  // The scale the diagram is framed at; the manual zoom floor never sits above
  // it, or a fitted-below-20% diagram could not be zoomed back out.
  let fitScale = 1

  const label = document.createElement('button')
  label.type = 'button'
  label.className = 'mermaid-btn mermaid-zoom-label'
  label.title = 'Reset zoom'
  label.setAttribute('aria-label', 'Reset zoom')

  const apply = () => {
    stage.style.transform = `translate(${x}px, ${y}px) scale(${scale})`
    label.textContent = `${Math.round(scale * 100)}%`
  }

  const isFullscreen = () => fullscreenElement() === view
  /** Real fullscreen and the overlay fallback fill the screen alike, and every
   *  behaviour that depends on "there is nothing else on screen" asks this. */
  const isEnlarged = () => isFullscreen() || view.classList.contains('is-overlay')

  const fit = () => {
    // Breathing room, so a diagram that exactly fills its frame does not sit
    // flush against the border — and clear of the toolbar in the corner.
    const cw = canvas.clientWidth - 2 * FIT_PADDING
    const ch = canvas.clientHeight - 2 * FIT_PADDING
    if (cw <= 0 || ch <= 0) return
    const max = isEnlarged() ? FULLSCREEN_MAX_FIT : INLINE_MAX_FIT
    scale = clamp(Math.min(cw / width, ch / height), FIT_MIN_SCALE, max)
    fitScale = scale
    x = (canvas.clientWidth - width * scale) / 2
    y = (canvas.clientHeight - height * scale) / 2
    apply()
  }

  const reset = () => {
    touched = false
    fit()
  }

  /** Zooms around a point in client coordinates, so what is under the cursor
   *  stays under the cursor. */
  const zoomAt = (clientX: number, clientY: number, factor: number) => {
    const rect = canvas.getBoundingClientRect()
    const px = clientX - rect.left
    const py = clientY - rect.top
    // Half the fitted scale, when that is smaller than the usual floor: a
    // diagram framed at 16% must still have somewhere to go when zoomed out,
    // or the button is dead the moment the reader arrives.
    const next = clamp(scale * factor, Math.min(MIN_SCALE, fitScale / 2), MAX_SCALE)
    if (next === scale) return
    x = px - ((px - x) / scale) * next
    y = py - ((py - y) / scale) * next
    scale = next
    touched = true
    apply()
  }

  const zoomCentre = (factor: number) => {
    const rect = canvas.getBoundingClientRect()
    zoomAt(rect.left + rect.width / 2, rect.top + rect.height / 2, factor)
  }

  label.addEventListener('click', reset)

  /** Fills the screen without the API — an iframe that was not granted
   *  fullscreen still owes the reader a way to enlarge a diagram.
   *
   *  The node moves to the body for the duration: `position: fixed` is contained
   *  by any transformed ancestor, and the entry card lifts on hover, so an
   *  overlay left in place would fill the card instead of the screen. */
  let placeholder: Comment | null = null

  const toggleOverlay = () => {
    if (placeholder) {
      view.classList.remove('is-overlay')
      // The placeholder marks the exact spot to return to — and if the page
      // rewrote that region while the overlay was up, it is gone and so is the
      // spot, so the viewer goes with it rather than reappearing out of place.
      if (placeholder.parentNode) placeholder.replaceWith(view)
      else view.remove()
      placeholder = null
    } else if (view.parentNode) {
      placeholder = document.createComment('mermaid overlay')
      view.replaceWith(placeholder)
      document.body.appendChild(view)
      view.classList.add('is-overlay')
    }
    sync()
  }

  const fsButton = button('Fullscreen', ICON.expand, () => {
    if (isFullscreen()) {
      const d = document as FsDocument
      ;(d.exitFullscreen ?? d.webkitExitFullscreen)?.call(document)
      return
    }
    if (view.classList.contains('is-overlay')) return toggleOverlay()

    const el = view as FsElement
    const request = el.requestFullscreen ?? el.webkitRequestFullscreen
    if (!request) return toggleOverlay()
    const started = request.call(el) as Promise<void> | undefined
    // A rejected promise here is a refusal, not a crash — take the overlay.
    started?.catch(toggleOverlay)
  })

  function sync() {
    const full = isEnlarged()
    view.classList.toggle('is-fullscreen', full)
    fsButton.innerHTML = full ? ICON.collapse : ICON.expand
    fsButton.title = full ? 'Exit fullscreen' : 'Fullscreen'
    fsButton.setAttribute('aria-label', fsButton.title)
    requestAnimationFrame(reset)
  }

  // fullscreenchange bubbles from the element, so no document-level listener is
  // left behind when the viewer is thrown away.
  const onFsChange = () => sync()
  view.addEventListener('fullscreenchange', onFsChange)
  view.addEventListener('webkitfullscreenchange', onFsChange)

  const bar = document.createElement('div')
  bar.className = 'mermaid-toolbar'
  bar.append(
    button('Zoom out', ICON.zoomOut, () => zoomCentre(1 / 1.25)),
    label,
    button('Zoom in', ICON.zoomIn, () => zoomCentre(1.25)),
    fsButton,
    await mermaidLiveAnchor(source, 'view')
  )
  view.appendChild(bar)

  canvas.addEventListener(
    'wheel',
    (e) => {
      // Inline, a bare wheel belongs to the page — hijacking it traps the reader
      // scrolling past a diagram. Fullscreen there is nothing else to scroll.
      if (!isEnlarged() && !e.ctrlKey && !e.metaKey) return
      e.preventDefault()
      // deltaY is in pixels, lines or pages depending on the device: Firefox
      // reports ±3 lines per notch where Chrome reports ±100 pixels, which
      // without this made a notch move the scale by half a percent.
      // 33px per line puts a Firefox notch (3 lines) on the same footing as a
      // Chrome one (100px).
      const step = e.deltaMode === WheelEvent.DOM_DELTA_LINE ? e.deltaY * 33
        : e.deltaMode === WheelEvent.DOM_DELTA_PAGE ? e.deltaY * 400
        : e.deltaY
      zoomAt(e.clientX, e.clientY, Math.exp(-clamp(step, -300, 300) * 0.002))
    },
    { passive: false }
  )

  canvas.addEventListener('pointerdown', (e) => {
    if (e.button !== 0) return
    // Inline, a finger belongs to the page: a 70vh-tall diagram would otherwise
    // swallow the swipe a reader uses to scroll past it. Enlarged, the diagram
    // is all there is, so it takes the touch (and CSS drops touch-action then).
    if (e.pointerType === 'touch' && !isEnlarged()) return

    // Stop the browser turning this drag into a text selection. CSS user-select
    // covers the diagram itself; this also stops a drag that leaves the canvas
    // from selecting its way across the rest of the page, and clears whatever
    // was selected before the reader grabbed the diagram.
    e.preventDefault()
    document.getSelection()?.removeAllRanges()
    // preventDefault also suppresses the focus the click would have given, and
    // the canvas is the element the arrow keys and +/- are bound to. Focus it by
    // hand; :focus-visible keeps the ring off for a mouse.
    canvas.focus()

    const startX = e.clientX - x
    const startY = e.clientY - y
    canvas.setPointerCapture(e.pointerId)
    canvas.classList.add('is-panning')

    const move = (ev: PointerEvent) => {
      x = ev.clientX - startX
      y = ev.clientY - startY
      touched = true
      apply()
    }
    const up = (ev: PointerEvent) => {
      // Unregister first: releasePointerCapture throws NotFoundError once the
      // pointer is gone, which is exactly the pointercancel case, and that used
      // to leave a live pointermove behind that dragged on plain hover.
      canvas.removeEventListener('pointermove', move)
      canvas.removeEventListener('pointerup', up)
      canvas.removeEventListener('pointercancel', up)
      canvas.classList.remove('is-panning')
      try {
        canvas.releasePointerCapture(ev.pointerId)
      } catch {
        /* the pointer was already released */
      }
    }
    canvas.addEventListener('pointermove', move)
    canvas.addEventListener('pointerup', up)
    canvas.addEventListener('pointercancel', up)
  })

  canvas.addEventListener('dblclick', reset)

  // Everything the mouse can do, the keyboard can do. Bound to the whole viewer,
  // not the canvas: after clicking Fullscreen the focus is on that button, and
  // Escape has to reach this from there.
  view.addEventListener('keydown', (e) => {
    const step = e.shiftKey ? 100 : 40
    const moves: Record<string, () => void> = {
      '+': () => zoomCentre(1.25),
      '=': () => zoomCentre(1.25),
      '-': () => zoomCentre(1 / 1.25),
      '0': reset,
      ArrowLeft: () => ((x += step), (touched = true), apply()),
      ArrowRight: () => ((x -= step), (touched = true), apply()),
      ArrowUp: () => ((y += step), (touched = true), apply()),
      ArrowDown: () => ((y -= step), (touched = true), apply()),
      // Real fullscreen leaves on Escape by itself; the overlay has to be told.
      Escape: () => {
        if (view.classList.contains('is-overlay')) toggleOverlay()
      },
    }
    const act = moves[e.key]
    if (!act) return
    e.preventDefault()
    act()
  })

  // The node is detached at this point, so the first fit has to wait for a size.
  if (typeof ResizeObserver !== 'undefined') {
    // A discarded viewer keeps its observer alive through the document's
    // observation list, and with it the whole detached SVG. Removal collapses
    // the box to zero, which is an observation — so it disconnects itself.
    const ro = new ResizeObserver(() => {
      if (!canvas.isConnected) return ro.disconnect()
      if (!touched) fit()
    })
    ro.observe(canvas)
  } else {
    requestAnimationFrame(fit)
  }

  apply()
  return view
}

/**
 * What to show when mermaid refuses the source: the source itself, plus a link
 * to the editor, which will say what is wrong with it.
 */
export async function createMermaidFallback(source: string): Promise<HTMLElement> {
  const wrap = document.createElement('div')
  wrap.className = 'mermaid-broken'
  const pre = document.createElement('pre')
  pre.className = 'mermaid-error'
  const code = document.createElement('code')
  code.textContent = source
  pre.appendChild(code)
  wrap.append(pre, await mermaidLiveAnchor(source, 'edit'))
  return wrap
}
