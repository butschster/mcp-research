/**
 * Links a mermaid source into the live editor at https://mermaid.live.
 *
 * The editor keeps its whole state in the URL fragment — nothing is uploaded,
 * the diagram never leaves the browser. The fragment is `<serde>:<payload>`,
 * where the payload encodes this JSON (the editor's own `State`, see
 * mermaid-live-editor src/lib/types.d.ts — note `mermaid` is a JSON *string*,
 * not an object):
 *
 *   { code, mermaid, updateDiagram, rough, panZoom }
 *
 * Two serdes are accepted. `pako:` is what the editor itself writes: zlib
 * deflate, then url-safe base64. `base64:` is the uncompressed form, still
 * supported on read. We emit `pako:` via the native CompressionStream — it
 * produces the zlib wrapper pako does — and fall back to `base64:` where that
 * API is missing, which only costs a longer URL.
 */

// Matches the editor's default config, so the diagram looks the way a fresh
// editor would draw it rather than inheriting our dark theme onto their page.
const MERMAID_CONFIG = JSON.stringify({ theme: 'default' }, null, 2)

function toBase64Url(bytes: Uint8Array): string {
  let binary = ''
  for (const b of bytes) binary += String.fromCharCode(b)
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

// Takes the text rather than the bytes so the encoder's own return type flows
// through: since TS 5.7 a bare Uint8Array is not a BufferSource, and annotating
// the parameter would pin this file to that lib version.
async function deflate(text: string): Promise<Uint8Array | null> {
  if (typeof CompressionStream === 'undefined') return null
  try {
    const cs = new CompressionStream('deflate')
    const writer = cs.writable.getWriter()
    void writer.write(new TextEncoder().encode(text))
    void writer.close()
    return new Uint8Array(await new Response(cs.readable).arrayBuffer())
  } catch {
    return null
  }
}

/**
 * @param code  mermaid source
 * @param mode  `view` renders it read-only, `edit` opens the editor — which is
 *              what you want for a diagram that failed to parse here.
 */
export async function mermaidLiveUrl(code: string, mode: 'view' | 'edit' = 'view'): Promise<string> {
  const state = JSON.stringify({
    code,
    mermaid: MERMAID_CONFIG,
    updateDiagram: true,
    rough: false,
    panZoom: true,
  })
  const packed = await deflate(state)
  const hash = packed
    ? `pako:${toBase64Url(packed)}`
    : `base64:${toBase64Url(new TextEncoder().encode(state))}`
  return `https://mermaid.live/${mode}#${hash}`
}
