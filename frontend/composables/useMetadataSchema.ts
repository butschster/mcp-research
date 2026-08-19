/**
 * The metadata rules, from the server that enforces them.
 *
 * `GET /api/metadata/schema` carries the field types, the reserved keys and
 * every cap, including the import limits. It exists so no client keeps its own
 * copy: a cap the client believes and the server enforces will disagree exactly
 * once, at the worst moment.
 *
 * One cache, because there were two parsers of one payload — the field-spec
 * editor read `types`/`reserved_keys`/`fields`/`required`, this read the import
 * caps, and disjoint field sets over one response is how a third parser gets
 * written. Fetched once per page load and shared: a research with nine sections
 * would otherwise ask nine times for one integer.
 */
export interface MetadataSchema {
  types?: Array<Record<string, any>>
  reserved_keys?: string[]
  caps?: Record<string, any>
}

const schema = ref<MetadataSchema | null>(null)
let inFlight: Promise<void> | null = null

/** This build's own values, so a failed fetch costs a stale hint rather than a
 *  broken control. The server refuses out-of-bounds input either way. */
const FALLBACK_MAX_BYTES = 1 << 20
const FALLBACK_EXTENSIONS = ['.md', '.markdown']

export function useMetadataSchema() {
  const { authFetch } = useAuth()
  const base = useRuntimeConfig().public.apiBase || ''

  function load() {
    if (inFlight || import.meta.server) return inFlight ?? Promise.resolve()
    inFlight = authFetch<{ data: MetadataSchema }>(`${base}/api/metadata/schema`)
      .then((res) => {
        schema.value = res?.data ?? null
      })
      .catch(() => {
        // Cleared so a later mount can try again; until then the fallbacks hold.
        inFlight = null
      })
    return inFlight
  }

  const maxBytes = computed<number>(() => {
    const n = schema.value?.caps?.import_max_bytes
    return typeof n === 'number' ? n : FALLBACK_MAX_BYTES
  })

  const extensions = computed<string[]>(() => {
    const list = schema.value?.caps?.import_extensions
    return Array.isArray(list) && list.length ? list : FALLBACK_EXTENSIONS
  })

  return { schema: readonly(schema), maxBytes, extensions, load }
}
