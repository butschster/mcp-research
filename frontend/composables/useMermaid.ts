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

/** "Open in mermaid.live" — the source travels in the URL fragment, so the
 *  diagram is never uploaded anywhere. */
async function liveLink(source: string, mode: 'view' | 'edit'): Promise<HTMLAnchorElement> {
  const a = document.createElement('a')
  a.className = 'mermaid-live-link'
  a.href = await mermaidLiveUrl(source, mode)
  a.target = '_blank'
  a.rel = 'noopener noreferrer'
  a.textContent = mode === 'edit' ? 'Open in mermaid.live to debug' : 'Open in mermaid.live'
  return a
}

export async function renderMermaidBlocks(container: HTMLElement) {
  ensureInit()
  const blocks = container.querySelectorAll<HTMLElement>('pre > code.language-mermaid')
  if (!blocks.length) return

  for (const code of blocks) {
    const pre = code.parentElement!
    // Callers run this on mount *and* on a content watcher, which can fire twice
    // over the same DOM. A failed block stays in place, so without this it would
    // collect a second debug link.
    if (pre.classList.contains('mermaid-error')) continue

    const source = code.textContent || ''
    const id = `mermaid-${++counter}`

    try {
      const { svg } = await mermaid.render(id, source)
      const wrapper = document.createElement('div')
      wrapper.className = 'mermaid-diagram'
      wrapper.innerHTML = svg
      wrapper.appendChild(await liveLink(source, 'view'))
      pre.replaceWith(wrapper)
    } catch {
      // leave the code block as-is on parse errors, with a way to open it in the
      // editor that will say what is wrong with it
      pre.classList.add('mermaid-error')
      pre.after(await liveLink(source, 'edit'))
    }
  }
}
