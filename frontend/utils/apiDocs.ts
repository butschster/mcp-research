// The API reference answers on several addresses, and the list is consumed in
// three places: the page's own `definePageMeta({ alias })`, `isChromeless` in
// app.vue, and the exemption in the auth middleware. Three copies means the
// fourth alias somebody adds is chromeless but not auth-exempt, and nobody
// notices until accounts are switched on.
//
// The aliases exist because the addresses a person guesses are the actual
// defect: `/docs` and `/swagger` already answer 200 with the SPA shell, so a
// developer who guesses one is shown the app with no explanation and concludes
// the documentation viewer is broken. Building the page and leaving the guesses
// missing would make that worse.
//
// `/scalar` is deliberately not among them: it names the renderer rather than
// the thing, and it is the one address that becomes a lie the day the renderer
// is swapped.
export const API_DOCS_PATH = '/api-docs'

export const API_DOCS_ALIASES = ['/docs', '/swagger', '/redoc', '/openapi'] as const

export const API_DOCS_PATHS = [API_DOCS_PATH, ...API_DOCS_ALIASES] as const

export function isApiDocsPath(path: string): boolean {
  const normalised = path.replace(/\/+$/, '') || '/'
  return (API_DOCS_PATHS as readonly string[]).includes(normalised)
}
