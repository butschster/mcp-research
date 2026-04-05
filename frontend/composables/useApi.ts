export function useApi<T>(path: string, options: Record<string, any> = {}) {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase || ''
  const { token } = useAuth()

  const headers: Record<string, string> = { ...(options.headers || {}) }
  if (token.value) {
    headers['Authorization'] = `Bearer ${token.value}`
  }

  return useFetch<T>(path, {
    ...options,
    baseURL: baseURL || undefined,
    key: path,
    headers,
  })
}
