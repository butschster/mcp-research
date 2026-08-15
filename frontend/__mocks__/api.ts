/**
 * A routing table behind the Storybook stub of `useAuth().authFetch`.
 *
 * Components that fetch their own data (HistoryPanel, ChangesList) would
 * otherwise have to be re-implemented as markup stubs in their story files —
 * and a stub drifts the moment the component changes, which is worse than no
 * story at all. With this, the catalog renders the real component and only the
 * network is fake.
 *
 * Usage from a story:
 *
 *   setup() {
 *     mockApi({
 *       '/revisions': { data: { revisions, current: 5 } },
 *       '/diff': (url) => ({ data: diffFor(url) }),
 *     })
 *   }
 *
 * Keys are matched as substrings of the request URL, longest key first, so
 * '/api/sessions' beats '/api'. Values are either the payload or a function of
 * the URL. Anything unmatched resolves to `{}`, which is what the plain stub
 * did.
 */

export type MockRoute = (url: string, options?: unknown) => unknown
export type MockRoutes = Record<string, unknown>

let routes: MockRoutes = {}

/** Installs the routes for the story about to render. Call it in `setup()`. */
export function mockApi(next: MockRoutes): void {
  routes = next
}

export function resetMockApi(): void {
  routes = {}
}

/** A request that never settles — for documenting a loading state. */
export function neverResolves(): Promise<never> {
  return new Promise<never>(() => {})
}

/** The function the `useAuth` stub delegates to. */
export function runMockFetch(url: string, options?: unknown): Promise<any> {
  const key = Object.keys(routes)
    .filter((candidate) => url.includes(candidate))
    .sort((a, b) => b.length - a.length)[0]

  if (key === undefined) {
    return Promise.resolve({})
  }

  const value = routes[key]
  return Promise.resolve(typeof value === 'function' ? (value as MockRoute)(url, options) : value)
}

/**
 * The same routing table, for `useApi()` rather than `authFetch()`.
 *
 * Kept separate because the two have opposite defaults and both are load-
 * bearing. An unrouted `authFetch` resolves to `{}` — a component that awaited
 * it carries on. An unrouted `useApi` must stay `null`, which is what its stub
 * has always returned and what every story written against it assumes; a story
 * that has not asked for data must not start receiving another story's.
 *
 * `SearchModal` is the case that needs this: it takes no props, fetches its own
 * two endpoints, and the only way to put a result on screen — let alone a
 * result whose name contains markup — is to answer them.
 */
let dataRoutes: MockRoutes = {}

/** Installs the `useApi` payloads for the story about to render. Call it in `setup()`. */
export function mockApiData(next: MockRoutes): void {
  dataRoutes = next
}

/** Called from the global decorator, so routes never leak into the next story. */
export function resetMockApiData(): void {
  dataRoutes = {}
}

/** What the `useApi` stub returns for a URL: a payload, or null when unrouted. */
export function resolveMockApiData(url: string): unknown {
  const key = Object.keys(dataRoutes)
    .filter((candidate) => url.includes(candidate))
    .sort((a, b) => b.length - a.length)[0]

  if (key === undefined) return null

  const value = dataRoutes[key]
  return typeof value === 'function' ? (value as MockRoute)(url) : value
}

/** Reads the `?to=` a diff request carries, so a story can answer per revision. */
export function revisionParam(url: string, name = 'to'): number | null {
  const match = new RegExp(`[?&]${name}=(\\d+)`).exec(url)
  return match ? Number(match[1]) : null
}
