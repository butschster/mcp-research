export function useApi<T>(path: string, options: Record<string, any> = {}) {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase || ''

  return useFetch<T>(path, {
    ...options,
    baseURL: baseURL || undefined,
    key: path,
  })
}
