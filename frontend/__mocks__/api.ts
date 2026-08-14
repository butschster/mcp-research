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

/** Reads the `?to=` a diff request carries, so a story can answer per revision. */
export function revisionParam(url: string, name = 'to'): number | null {
  const match = new RegExp(`[?&]${name}=(\\d+)`).exec(url)
  return match ? Number(match[1]) : null
}
