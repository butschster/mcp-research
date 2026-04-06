export function useApi<T>(path: string, options: Record<string, any> = {}) {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase || ''
  const { token } = useAuth()

  return useFetch<T>(path, {
    ...options,
    baseURL: baseURL || undefined,
    key: path,
    onRequest({ options: reqOpts }) {
      if (token.value) {
        reqOpts.headers = reqOpts.headers instanceof Headers
          ? reqOpts.headers
          : new Headers(reqOpts.headers as Record<string, string> || {})
        ;(reqOpts.headers as Headers).set('Authorization', `Bearer ${token.value}`)
      }
    },
  })
}
