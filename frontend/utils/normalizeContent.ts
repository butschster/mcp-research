/**
 * Replaces literal escape sequences (\n, \t) with actual characters.
 * MCP clients sometimes send these instead of real newlines/tabs.
 * Apply before markdown parsing to ensure correct rendering.
 */
export function normalizeContent(text: string): string {
  if (!text) return ''
  return text.replace(/\\n/g, '\n').replace(/\\t/g, '\t')
}
