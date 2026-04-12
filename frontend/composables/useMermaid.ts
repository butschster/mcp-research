import mermaid from 'mermaid'

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

export async function renderMermaidBlocks(container: HTMLElement) {
  ensureInit()
  const blocks = container.querySelectorAll<HTMLElement>('pre > code.language-mermaid')
  if (!blocks.length) return

  for (const code of blocks) {
    const pre = code.parentElement!
    const source = code.textContent || ''
    const id = `mermaid-${++counter}`

    try {
      const { svg } = await mermaid.render(id, source)
      const wrapper = document.createElement('div')
      wrapper.className = 'mermaid-diagram'
      wrapper.innerHTML = svg
      pre.replaceWith(wrapper)
    } catch {
      // leave the code block as-is on parse errors
      pre.classList.add('mermaid-error')
    }
  }
}
