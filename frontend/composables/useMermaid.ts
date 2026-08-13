import { createMermaidViewer, mermaidLiveAnchor } from '~/composables/useMermaidViewer'

/**
 * Turns every ```mermaid fence inside rendered markdown into an interactive
 * diagram. The markup comes from v-html, so Vue does not track these nodes and
 * they are safe to replace.
 */
export async function renderMermaidBlocks(container: HTMLElement) {
  const blocks = container.querySelectorAll<HTMLElement>('pre > code.language-mermaid')
  if (!blocks.length) return

  for (const code of blocks) {
    const pre = code.parentElement!
    // Callers run this on mount *and* on a content watcher, which can fire twice
    // over the same DOM — and rendering is async, so two runs can be inside this
    // loop at once. A failed block stays in place, so the claim has to be staked
    // before the await, or it collects a second debug link per concurrent run.
    if (pre.classList.contains('mermaid-error') || pre.dataset.mermaidBusy) continue
    pre.dataset.mermaidBusy = '1'

    const source = code.textContent || ''
    const view = await createMermaidViewer(source)
    if (view) {
      pre.replaceWith(view)
    } else {
      // leave the source in place on parse errors, with a way to open it in the
      // editor that will say what is wrong with it
      pre.classList.add('mermaid-error')
      pre.after(await mermaidLiveAnchor(source, 'edit'))
    }
  }
}
