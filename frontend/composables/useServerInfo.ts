export interface ServerInfo {
  status: string
  version: string
  in_memory: boolean
  write_api: boolean
  auth_enabled: boolean
}

/**
 * What the server says about itself.
 *
 * Module state and one request for the life of the tab: `/api/health` is
 * already fetched by the in-memory warning banner, and the footer wanting the
 * version is not a reason to ask twice. It is deliberately not reactive to
 * anything — a running server's version does not change, and if it did, the
 * page it is printed on came from the old one.
 */
const info = ref<ServerInfo | null>(null)
let inFlight: Promise<void> | null = null

export function useServerInfo() {
  const base = useRuntimeConfig().public.apiBase || ''

  function load() {
    if (import.meta.server || info.value || inFlight) return inFlight ?? Promise.resolve()
    // No credential: /api/health is public, and the footer must render on a
    // page where nobody is signed in.
    inFlight = $fetch<ServerInfo>(`${base}/api/health`)
      .then((res) => { info.value = res })
      .catch(() => { /* the footer simply says nothing */ })
      .finally(() => { inFlight = null })
    return inFlight
  }

  if (!import.meta.server && !info.value) void load()

  return {
    info: readonly(info),
    /** e.g. `v1.4.0`, or `dev` for a local build. Empty until it arrives. */
    version: computed(() => {
      const v = info.value?.version
      if (!v) return ''
      return /^\d/.test(v) ? `v${v}` : v
    }),
    load,
  }
}
