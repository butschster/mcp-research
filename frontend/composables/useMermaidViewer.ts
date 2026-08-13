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

let initialized = false

function ensureInit() {
  if (initialized) return
  initialized = true
  mermaid.initialize({
    startOnLoad: false,
    theme: 'dark',
    themeVariables: {
      darkMode: true,
      background: '#1a1a2e',
      primaryColor: '#f0b849',
      primaryTextColor: '#e0e0e0',
      primaryBorderColor: '#f0b849',
      lineColor: '#555',
      secondaryColor: '#2a2a3e',
      tertiaryColor: '#1e1e30',
      fontFamily: "'Inter', sans-serif",
    },
    flowchart: { curve: 'basis' },
    sequence: { mirrorActors: false },
  })
}

let counter = 0

/** Bounds for zooming by hand. Fitting has its own, lower floor: a huge diagram
 *  must still be shown whole, however small that makes it. */
const MIN_SCALE = 0.2
const MAX_SCALE = 8
const FIT_MIN_SCALE = 0.05
/** Inline, a diagram is a figure in an article — it gets a viewport of its own
 *  height, capped, and scales to fit. Fullscreen is where a big graph is read. */
const INLINE_MAX_HEIGHT = '70vh'
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
  // A short diagram gets exactly its own height; a tall one gets most of the
  // window and is scaled down to fit, which beats a slice of a wall of boxes.
  view.style.setProperty('--mermaid-h', `min(${Math.round(height)}px, ${INLINE_MAX_HEIGHT})`)

  let scale = 1
  let x = 0
  let y = 0
  // Once the reader has moved the diagram, a container resize must not yank it
  // back to the default framing.
  let touched = false

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

  const fit = () => {
    // Breathing room, so a diagram that exactly fills its frame does not sit
    // flush against the border — and clear of the toolbar in the corner.
    const cw = canvas.clientWidth - 2 * FIT_PADDING
    const ch = canvas.clientHeight - 2 * FIT_PADDING
    if (cw <= 0 || ch <= 0) return
    const max = isFullscreen() ? FULLSCREEN_MAX_FIT : INLINE_MAX_FIT
    scale = clamp(Math.min(cw / width, ch / height), FIT_MIN_SCALE, max)
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
    const next = clamp(scale * factor, MIN_SCALE, MAX_SCALE)
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
   *  fullscreen still owes the reader a way to enlarge a diagram. */
  const toggleOverlay = () => {
    view.classList.toggle('is-overlay')
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
    const full = isFullscreen() || view.classList.contains('is-overlay')
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
      if (!isFullscreen() && !view.classList.contains('is-overlay') && !e.ctrlKey && !e.metaKey) return
      e.preventDefault()
      zoomAt(e.clientX, e.clientY, Math.exp(-e.deltaY * 0.002))
    },
    { passive: false }
  )

  canvas.addEventListener('pointerdown', (e) => {
    if (e.button !== 0) return
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
      canvas.releasePointerCapture(ev.pointerId)
      canvas.classList.remove('is-panning')
      canvas.removeEventListener('pointermove', move)
      canvas.removeEventListener('pointerup', up)
      canvas.removeEventListener('pointercancel', up)
    }
    canvas.addEventListener('pointermove', move)
    canvas.addEventListener('pointerup', up)
    canvas.addEventListener('pointercancel', up)
  })

  canvas.addEventListener('dblclick', reset)

  // Everything the mouse can do, the keyboard can do — the canvas is focusable.
  canvas.addEventListener('keydown', (e) => {
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
    new ResizeObserver(() => {
      if (!touched) fit()
    }).observe(canvas)
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
