/**
 * Mock payloads for read-only share links, and the two decorators that put a
 * story inside or outside one.
 *
 * Shapes follow the API exactly:
 *   - a row is `ShareRow`          (GET /api/researches/{id}/shares)
 *   - the visitor payload is `SharePayloadMeta` (GET /api/shared/{token})
 *
 * Expiry timestamps are relative to now, because the row prints "expires in N
 * days" and a fixed date would read "expired" a week after this file was
 * written. One row keeps SQLite's space-separated `created_at`, which is what
 * the server actually sends.
 */
import type { ShareRow } from '~/components/research/ShareRowList.vue'
import { useShare, type ShareInclude, type SharePayloadMeta } from '~/composables/useShare'

const DAY = 86_400_000
const inDays = (days: number) => new Date(Date.now() + days * DAY).toISOString()
const daysAgo = (days: number) => new Date(Date.now() - days * DAY).toISOString()

/** Long enough to wrap in the reveal block, as a real one does. */
export const mockShareToken = '9f2c7b1e40a54d8fbb63d21c7e05c41a'
export const mockShareUrl = `https://research.intruforce.com/s/${mockShareToken}`

export const includeEverything: ShareInclude = {
  sessions: true,
  tasks: true,
  roadmaps: true,
  export: true,
}

/** The default the create form proposes: entries, roadmaps and downloads. */
export const includeDefault: ShareInclude = {
  sessions: false,
  tasks: false,
  roadmaps: true,
  export: true,
}

/** The narrowest link there is. Entries are never optional. */
export const includeEntriesOnly: ShareInclude = {
  sessions: false,
  tasks: false,
  roadmaps: false,
  export: false,
}

export const mockShareMeta: SharePayloadMeta = {
  label: 'Client review, March',
  owner_name: 'Elena Marsh',
  research_id: 'res_007',
  research_code: 'R7',
  expires_at: inDays(23),
  include: includeDefault,
}

export const mockShareRows: ShareRow[] = [
  {
    id: 'shr_live',
    label: 'Client review, March',
    include: includeDefault,
    has_password: true,
    expires_at: null,
    revoked_at: null,
    last_seen_at: new Date(Date.now() - 2 * 3_600_000).toISOString(),
    view_count: 47,
    // SQLite's format, straight off the wire.
    created_at: daysAgo(12).replace('T', ' ').slice(0, 19),
    created_by_name: 'Elena Marsh',
  },
  {
    id: 'shr_never_opened',
    label: 'Board pack',
    include: includeEntriesOnly,
    has_password: false,
    expires_at: inDays(6),
    revoked_at: null,
    last_seen_at: null,
    view_count: 0,
    created_at: daysAgo(1),
    created_by_name: 'Екатерина Ковалёва',
  },
  {
    id: 'shr_tomorrow',
    label: 'Q1 handoff to Northlight',
    include: includeEverything,
    has_password: false,
    expires_at: inDays(1.2),
    revoked_at: null,
    last_seen_at: daysAgo(4),
    view_count: 1,
    created_at: daysAgo(29),
    created_by_name: 'Elena Marsh',
  },
  {
    id: 'shr_expired',
    label: 'Пилот с партнёром, февраль',
    include: includeDefault,
    has_password: false,
    expires_at: daysAgo(3),
    revoked_at: null,
    last_seen_at: daysAgo(5),
    view_count: 12,
    created_at: daysAgo(33),
    created_by_name: 'Марат Ибрагимов',
  },
  {
    id: 'shr_revoked',
    label: 'Sent to the wrong address',
    include: includeDefault,
    has_password: false,
    expires_at: null,
    revoked_at: daysAgo(2),
    last_seen_at: daysAgo(9),
    view_count: 3,
    created_at: daysAgo(18),
    created_by_name: 'Elena Marsh',
  },
]

/** A link this tab issued and is still holding, keyed by share id. */
export const mockRecoverableShareLinks: Record<string, string> = {
  shr_live: mockShareUrl,
}

/** Sixty rows: the dialog body scrolls and nothing paginates. */
export const mockManyShareRows: ShareRow[] = Array.from({ length: 60 }, (_, i) => ({
  ...mockShareRows[i % mockShareRows.length]!,
  id: `shr_bulk_${i}`,
  label: `Weekly digest ${i + 1}`,
  view_count: (i * 7) % 130,
  created_at: daysAgo(i / 2),
}))

/**
 * The one the label column exists for: 200 characters, which wrap rather than
 * truncate, because a truncated label is not a label.
 */
export const mockLongLabelShareRow: ShareRow = {
  ...mockShareRows[0]!,
  id: 'shr_long_label',
  label:
    'Pricing benchmark shared with the Northlight procurement group ahead of the March steering committee, superseding the February pack that went out before the seat-tier numbers were corrected',
  has_password: false,
}

/**
 * Puts a story inside a share link, and takes it out again.
 *
 * The share is module state — `renderRefs`, the path helpers and three
 * components read it as a plain function call, which is the whole reason it is
 * module state and not a prop. A story therefore sets it in `setup()`, before
 * anything below renders.
 *
 * **One shared state, so one story at a time.** On a story's own canvas that is
 * exactly right. On an autodocs page, which mounts every story of a component
 * side by side against the same module, whichever mounts last wins and the
 * others will have re-rendered against it — the href in a docs preview is the
 * last story's context, not its own. The canvas view is the authoritative one,
 * and there is no fix short of giving each story its own JS context.
 *
 * `withoutShare()` belongs on the meta of every component that reads share
 * state, so the ordinary stories start from a known state rather than from
 * whatever the previous story left behind.
 */
export function withShare(meta: SharePayloadMeta = mockShareMeta) {
  return (story: any) => ({
    components: { story },
    setup() {
      const { claim, setPayload } = useShare()
      claim(mockShareToken)
      setPayload({ share: meta, research: {}, sections: [] })
      return {}
    },
    template: '<story />',
  })
}

/** Puts a story back outside any share — the ordinary, signed-in view. */
export function withoutShare() {
  return (story: any) => ({
    components: { story },
    setup() {
      useShare().release()
      return {}
    },
    template: '<story />',
  })
}
