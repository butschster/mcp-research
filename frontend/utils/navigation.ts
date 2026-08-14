/**
 * A `?next=` destination is followed only when it is a path on this origin.
 *
 * Without the check, `?next=https://elsewhere.example` turns the sign-in page
 * into an open redirect — and a redirect that carries a freshly authenticated
 * reader somewhere else is exactly the useful kind for an attacker.
 */
export function safeNext(value: unknown): string | null {
  if (typeof value !== 'string' || value === '') return null
  // A leading `//` is a protocol-relative URL, not a local path.
  if (!value.startsWith('/') || value.startsWith('//')) return null
  return value
}
