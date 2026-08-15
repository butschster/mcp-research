export function useApi<T>(path: string, options: Record<string, any> = {}) {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase || ''
  const { token } = useAuth()

  return useFetch<T>(path, {
    ...options,
    baseURL: baseURL || undefined,
    key: path,
    onRequest({ options: reqOpts }) {
      reqOpts.headers = reqOpts.headers instanceof Headers
        ? reqOpts.headers
        : new Headers(reqOpts.headers as Record<string, string> || {})
      if (token.value) {
        ;(reqOpts.headers as Headers).set('Authorization', `Bearer ${token.value}`)
      }
      // Stamps the write with this tab, so the event it produces can be
      // recognised here and not treated as somebody else's change.
      if (!import.meta.server) {
        ;(reqOpts.headers as Headers).set('X-Client-Id', realtimeClientId())
      }
    },
  })
}
