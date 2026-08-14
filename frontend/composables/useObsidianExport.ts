export interface ObsidianExportOptions {
  sessions: boolean
  tasks: boolean
  roadmaps: boolean
  html: boolean
  revisions: boolean
}

export function defaultObsidianOptions(): ObsidianExportOptions {
  return { sessions: true, tasks: true, roadmaps: true, html: true, revisions: false }
}

/**
 * Only the options that differ from the server's defaults are sent, so an
 * ordinary download is a clean `?format=obsidian` — readable in a log, and
 * pasteable by anyone reading the docs.
 */
export function obsidianExportPath(researchId: string, opts: ObsidianExportOptions): string {
  const params = new URLSearchParams({ format: 'obsidian' })
  for (const key of ['sessions', 'tasks', 'roadmaps'] as const) {
    if (!opts[key]) params.set(key, 'false')
  }
  if (!opts.html) params.set('html', 'false')
  if (opts.revisions) params.set('revisions', 'true')
  return `/api/researches/${researchId}/export?${params.toString()}`
}

/** What a filesystem refuses in a name, mirroring the Go side's own list. */
const FORBIDDEN = /[/\\:*?"<>|]/g
const CONTROL = /[\u0000-\u001f]/g

/**
 * The name to save under when the response's own filename cannot be read —
 * which happens cross-origin unless the API exposes Content-Disposition. It
 * mirrors the Go side's sanitization so the two do not disagree about what a
 * research is called.
 */
export function vaultFallbackName(code: string, name: string): string {
  const clean = (s: string) =>
    s
      .replace(FORBIDDEN, '')
      .replace(CONTROL, ' ')
      .replace(/\s+/g, ' ')
      // A trailing dot or space is stripped by Windows anyway.
      .replace(/^[ .]+|[ .]+$/g, '')
  const parts = [clean(code), clean(name)].filter(Boolean)
  return `${parts.join(' — ') || 'research'}.zip`
}
