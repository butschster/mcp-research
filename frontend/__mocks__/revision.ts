/**
 * Mock payloads for the entry revision history feature.
 *
 * Shapes follow the API exactly:
 *   - a revision is `domain.EntryRevision`   (GET /api/entries/{id}/revisions)
 *   - a diff is `service.EntryDiff`          (GET /api/entries/{id}/diff)
 *   - a session change is `service.SessionEntryChange`
 *                                            (GET /api/sessions/{id}/changes)
 *
 * The word-level breakdowns below are what `service.diffWords` + `keepSide`
 * actually emit: every token keeps the space that followed it, the remove side
 * carries `equal` + `remove` words and the add side `equal` + `add`, so joining
 * a line's words reproduces its text.
 */

export type DiffOp = 'equal' | 'add' | 'remove'

export interface DiffWord {
  op: DiffOp
  text: string
}

export interface DiffLine {
  op: DiffOp
  text: string
  /** Present only where a removed line was paired with a similar added one. */
  words?: DiffWord[]
}

export interface DiffResult {
  lines: DiffLine[]
  added: number
  removed: number
  unchanged: number
  unified: string
  truncated?: boolean
}

export interface Revision {
  id: string
  entry_id: string
  research_id: string
  revision: number
  title: string
  description: string
  content?: string
  entry_type: string
  status: string
  tags: string[]
  author_kind: 'agent' | 'human' | 'import' | 'restore'
  session_id?: string
  user_id?: string
  summary?: string
  created_at: string
  session_code?: string
  session_title?: string
}

export interface EntryDiff {
  entry_id: string
  entry_code?: string
  from: Revision | null
  to: Revision
  title?: DiffResult
  content: DiffResult
  summary: string
}

export interface SessionEntryChange {
  entry_id: string
  entry_code: string
  title: string
  section_id: string
  created: boolean
  revisions: Revision[]
  from_revision: number
  to_revision: number
  diff: DiffResult | null
  summary: string
}

const ENTRY_ID = 'ent_2f7c9a41'
const RESEARCH_ID = 'res_91ab33d0'

function revision(rev: Partial<Revision> & { revision: number }): Revision {
  return {
    id: `erv_${String(rev.revision).padStart(2, '0')}c4d1`,
    entry_id: ENTRY_ID,
    research_id: RESEARCH_ID,
    title: 'Streaming Ingestion Trade-offs',
    description: 'Throughput, latency and back-pressure behaviour of the two candidate pipelines.',
    entry_type: 'markdown',
    status: 'completed',
    tags: ['ingestion', 'latency', 'kafka'],
    author_kind: 'agent',
    created_at: '2025-03-18T09:12:00Z',
    ...rev,
  }
}

/** Newest first, the order the API returns and the panel renders. */
export const mockRevisions: Revision[] = [
  revision({
    revision: 5,
    author_kind: 'restore',
    created_at: '2025-03-18T09:12:00Z',
    summary: 'Restored revision 3',
    session_id: 'ses_4d1',
    session_code: 'SS3',
    session_title: 'Latency deep dive',
  }),
  revision({
    revision: 4,
    author_kind: 'human',
    created_at: '2025-03-17T16:40:00Z',
    summary: 'Updated content, tags',
    session_id: 'ses_4d1',
    session_code: 'SS3',
    session_title: 'Latency deep dive',
  }),
  revision({
    revision: 3,
    author_kind: 'agent',
    created_at: '2025-03-17T11:05:00Z',
    summary: 'Updated content',
    session_id: 'ses_4d1',
    session_code: 'SS3',
    session_title: 'Latency deep dive',
  }),
  revision({
    revision: 2,
    author_kind: 'agent',
    status: 'active',
    created_at: '2025-03-16T14:22:00Z',
    summary: 'Updated content, status',
    session_id: 'ses_2b8',
    session_code: 'SS2',
    session_title: 'Vendor benchmarks',
  }),
  revision({
    revision: 1,
    title: 'Ingestion notes',
    status: 'draft',
    tags: ['ingestion'],
    author_kind: 'agent',
    created_at: '2025-03-15T10:00:00Z',
    // The first revision of an entry carries no summary — nothing preceded it.
    summary: '',
    session_id: 'ses_1a0',
    session_code: 'SS1',
    session_title: 'Kick-off interview',
  }),
]

/** An entry written once by an import and never edited. */
export const mockRevisionsSingle: Revision[] = [
  revision({
    revision: 1,
    author_kind: 'import',
    created_at: '2025-02-02T08:30:00Z',
    summary: 'Imported research',
    session_id: undefined,
    session_code: undefined,
    session_title: undefined,
  }),
]

/** A long-lived entry: the list scrolls and the restore column repeats. */
export const mockRevisionsMany: Revision[] = Array.from({ length: 24 }, (_, i) => {
  const number = 24 - i
  // Every fifth write is a restore, and a restore's summary always names the
  // revision it brought back — the two fields cannot disagree in real data.
  const restored = number % 5 === 0
  const kinds: Revision['author_kind'][] = ['agent', 'agent', 'human', 'agent']
  return revision({
    revision: number,
    author_kind: restored ? 'restore' : kinds[number % kinds.length]!,
    created_at: new Date(Date.UTC(2025, 2, 1 + number, 9 + (number % 8), (number * 7) % 60)).toISOString(),
    summary: restored ? `Restored revision ${Math.max(1, number - 3)}` : 'Updated content',
    session_id: `ses_${number % 3}`,
    session_code: `SS${(number % 3) + 1}`,
    session_title: 'Latency deep dive',
  })
})

// --- Diffs -----------------------------------------------------------------

/**
 * One sentence rewritten, one sentence appended. The rewritten pair is similar
 * enough (token overlap ≥ 0.4) that the backend attaches a word-level
 * breakdown to both sides of it.
 */
export const mockDiffWordLevel: DiffResult = {
  lines: [
    { op: 'equal', text: '# Streaming Ingestion Trade-offs' },
    { op: 'equal', text: '' },
    {
      op: 'remove',
      text: 'Batch size 500 keeps p99 latency under 200ms in the [[E3]] benchmark.',
      words: [
        { op: 'equal', text: 'Batch ' },
        { op: 'equal', text: 'size ' },
        { op: 'remove', text: '500 ' },
        { op: 'equal', text: 'keeps ' },
        { op: 'equal', text: 'p99 ' },
        { op: 'equal', text: 'latency ' },
        { op: 'equal', text: 'under ' },
        { op: 'remove', text: '200ms ' },
        { op: 'equal', text: 'in ' },
        { op: 'equal', text: 'the ' },
        { op: 'equal', text: '[[E3]] ' },
        { op: 'equal', text: 'benchmark.' },
      ],
    },
    {
      op: 'add',
      text: 'Batch size 2000 keeps p99 latency under 350ms in the [[E3]] benchmark.',
      words: [
        { op: 'equal', text: 'Batch ' },
        { op: 'equal', text: 'size ' },
        { op: 'add', text: '2000 ' },
        { op: 'equal', text: 'keeps ' },
        { op: 'equal', text: 'p99 ' },
        { op: 'equal', text: 'latency ' },
        { op: 'equal', text: 'under ' },
        { op: 'add', text: '350ms ' },
        { op: 'equal', text: 'in ' },
        { op: 'equal', text: 'the ' },
        { op: 'equal', text: '[[E3]] ' },
        { op: 'equal', text: 'benchmark.' },
      ],
    },
    { op: 'add', text: 'Above 4000 the broker rejects produce requests — see [[R2:E5]].' },
    { op: 'equal', text: '' },
    { op: 'equal', text: '## Open questions' },
    { op: 'equal', text: '' },
    { op: 'equal', text: '- Does back-pressure interact with the retry budget?' },
  ],
  added: 2,
  removed: 1,
  unchanged: 6,
  unified: [
    '  # Streaming Ingestion Trade-offs',
    '  ',
    '- Batch size 500 keeps p99 latency under 200ms in the [[E3]] benchmark.',
    '+ Batch size 2000 keeps p99 latency under 350ms in the [[E3]] benchmark.',
    '+ Above 4000 the broker rejects produce requests — see [[R2:E5]].',
    '  ',
    '  ## Open questions',
    '  ',
    '',
  ].join('\n'),
}

const vendorNotes = [
  '## Candidates',
  '',
  'Both pipelines were run against the same 90-minute replay of production traffic.',
  '',
  '### Vendor A',
  '',
  '- Managed connectors, no operator to run',
  '- Exactly-once only within a single topic',
  '- Pricing is seat-based above 50 users',
  '',
  '### Vendor B',
  '',
  '- Self-hosted, one operator per cluster',
  '- Exactly-once across topics via transactions',
  '- Pricing is throughput-based, see [[E7]]',
  '',
  '## Measurements',
  '',
  'Numbers below are the median of five runs; the raw series live in [[E9]].',
  '',
]

/**
 * A long document with two edits far apart: unchanged runs collapse into two
 * separate "N unchanged lines" gaps.
 */
export const mockDiffWithGaps: DiffResult = {
  lines: [
    { op: 'add', text: '# Ingestion vendor comparison' },
    { op: 'add', text: '' },
    ...vendorNotes.map((text) => ({ op: 'equal' as const, text })),
    {
      op: 'remove',
      text: 'Vendor A sustained 40k events/s before back-pressure kicked in.',
      words: [
        { op: 'equal', text: 'Vendor ' },
        { op: 'equal', text: 'A ' },
        { op: 'equal', text: 'sustained ' },
        { op: 'remove', text: '40k ' },
        { op: 'equal', text: 'events/s ' },
        { op: 'equal', text: 'before ' },
        { op: 'equal', text: 'back-pressure ' },
        { op: 'equal', text: 'kicked ' },
        { op: 'equal', text: 'in.' },
      ],
    },
    {
      op: 'add',
      text: 'Vendor A sustained 62k events/s before back-pressure kicked in.',
      words: [
        { op: 'equal', text: 'Vendor ' },
        { op: 'equal', text: 'A ' },
        { op: 'equal', text: 'sustained ' },
        { op: 'add', text: '62k ' },
        { op: 'equal', text: 'events/s ' },
        { op: 'equal', text: 'before ' },
        { op: 'equal', text: 'back-pressure ' },
        { op: 'equal', text: 'kicked ' },
        { op: 'equal', text: 'in.' },
      ],
    },
    { op: 'equal', text: 'Vendor B sustained 58k events/s with a flatter tail.' },
    { op: 'equal', text: '' },
    { op: 'equal', text: '## Recommendation' },
    { op: 'equal', text: '' },
    { op: 'equal', text: 'Neither vendor is disqualified on throughput alone.' },
    { op: 'equal', text: '' },
    { op: 'equal', text: 'The decision hangs on the operating cost of the self-hosted option.' },
    { op: 'equal', text: '' },
    { op: 'equal', text: 'A follow-up session should price that out.' },
    { op: 'equal', text: '' },
    { op: 'remove', text: 'Decision deferred to the next review.' },
    { op: 'add', text: 'Decision: pilot Vendor B for one quarter, tracked in [[T12]].' },
  ],
  added: 4,
  removed: 2,
  unchanged: 29,
  unified: '',
}

/** Revision 1 of an entry: everything is an addition, nothing to collapse. */
export const mockDiffFirstVersion: DiffResult = {
  lines: [
    { op: 'add', text: '# Ingestion notes' },
    { op: 'add', text: '' },
    { op: 'add', text: 'First pass from the kick-off interview. Sources are in [[E1]].' },
    { op: 'add', text: '' },
    { op: 'add', text: '- Two candidate pipelines, both Kafka-backed' },
    { op: 'add', text: '- Latency budget is 500ms end to end' },
    { op: 'add', text: '- Nobody has measured the retry path yet' },
  ],
  added: 7,
  removed: 0,
  unchanged: 0,
  unified: '',
}

/**
 * Two revisions compared that turn out to be identical (`?from=3&to=3`).
 * `lines` is non-empty and every op is `equal`, so nothing collapses — the
 * component deliberately shows the document rather than one giant gap.
 */
export const mockDiffNoChanges: DiffResult = {
  lines: [
    { op: 'equal', text: '# Streaming Ingestion Trade-offs' },
    { op: 'equal', text: '' },
    { op: 'equal', text: 'Batch size 2000 keeps p99 latency under 350ms in the [[E3]] benchmark.' },
    { op: 'equal', text: '' },
    { op: 'equal', text: '## Open questions' },
    { op: 'equal', text: '' },
    { op: 'equal', text: '- Does back-pressure interact with the retry budget?' },
  ],
  added: 0,
  removed: 0,
  unchanged: 7,
  unified: '',
}

/** Both sides empty — an entry whose body was blank on either revision. */
export const mockDiffEmpty: DiffResult = {
  lines: [],
  added: 0,
  removed: 0,
  unchanged: 0,
  unified: '',
}

/**
 * Over `maxDiffLines`: the backend gives up on alignment and reports the whole
 * document removed and re-added, with `truncated` set.
 */
export const mockDiffTruncated: DiffResult = {
  lines: [
    { op: 'remove', text: '# Event schema registry (v3)' },
    { op: 'remove', text: '' },
    { op: 'remove', text: '| field | type | since |' },
    { op: 'remove', text: '| --- | --- | --- |' },
    { op: 'remove', text: '| order_id | uuid | v1 |' },
    { op: 'add', text: '# Event schema registry (v4)' },
    { op: 'add', text: '' },
    { op: 'add', text: '| field | type | since |' },
    { op: 'add', text: '| --- | --- | --- |' },
    { op: 'add', text: '| order_id | uuid | v1 |' },
    { op: 'add', text: '| tenant_id | uuid | v4 |' },
  ],
  added: 4206,
  removed: 4102,
  unchanged: 0,
  unified: '',
  truncated: true,
}

/** Long unbroken tokens: a URL and a wide table row have to wrap, not clip. */
export const mockDiffLongLines: DiffResult = {
  lines: [
    { op: 'equal', text: '## Sources' },
    { op: 'equal', text: '' },
    {
      op: 'remove',
      text: '- Benchmark harness: https://internal.example.com/observability/dashboards/ingestion-latency?from=now-90d&to=now&panelId=42&refresh=30s&var-cluster=eu-central-1',
    },
    {
      op: 'add',
      text: '- Benchmark harness: https://internal.example.com/observability/dashboards/ingestion-latency-v2?from=now-180d&to=now&panelId=17&refresh=10s&var-cluster=eu-central-1&var-topic=orders.v4',
    },
    { op: 'add', text: '| partition | leader | p50 | p95 | p99 | throughput | retries | rebalance_seconds | notes |' },
  ],
  added: 2,
  removed: 1,
  unchanged: 2,
  unified: '',
}

// --- Entry diffs (what the history panel fetches) ---------------------------

function entryDiff(from: Revision | null, to: Revision, content: DiffResult): EntryDiff {
  return {
    entry_id: ENTRY_ID,
    entry_code: 'E4',
    from,
    to,
    content,
    summary: `+${content.added} −${content.removed}`,
  }
}

/** The default view of the panel: newest revision against the one before it. */
export const mockEntryDiff = entryDiff(mockRevisions[1]!, mockRevisions[0]!, mockDiffWordLevel)

/** No `from`: the panel headers this as "first version". */
export const mockEntryDiffFirstVersion: EntryDiff = {
  ...entryDiff(null, mockRevisions[4]!, mockDiffFirstVersion),
  // A retitled first revision has no title diff; a later rename does.
  title: undefined,
}

export const mockEntryDiffLong = entryDiff(mockRevisions[2]!, mockRevisions[1]!, mockDiffWithGaps)

/** No preceding revision to compare against, and a body that is one paragraph. */
export const mockEntryDiffSingle: EntryDiff = {
  entry_id: ENTRY_ID,
  entry_code: 'E4',
  from: null,
  to: mockRevisionsSingle[0]!,
  content: mockDiffFirstVersion,
  summary: '+7 −0',
}

/** A rename carries a `title` diff alongside the content one. */
export const mockEntryDiffRenamed: EntryDiff = {
  ...entryDiff(mockRevisions[4]!, mockRevisions[3]!, mockDiffWordLevel),
  title: {
    lines: [
      { op: 'remove', text: 'Ingestion notes' },
      { op: 'add', text: 'Streaming Ingestion Trade-offs' },
    ],
    added: 1,
    removed: 1,
    unchanged: 0,
    unified: '',
  },
}

/** Keyed by the `?to=` the panel asks for, so every row in the list resolves. */
export const mockEntryDiffByRevision: Record<number, EntryDiff> = {
  1: mockEntryDiffFirstVersion,
  2: mockEntryDiffRenamed,
  3: entryDiff(mockRevisions[3]!, mockRevisions[2]!, mockDiffLongLines),
  4: mockEntryDiffLong,
  // r5 restored r3, so it reads as an ordinary content change against r4.
  5: mockEntryDiff,
}

// --- Session changes --------------------------------------------------------

/** An entry the session edited but did not create. */
export const mockChangeModified: SessionEntryChange = {
  entry_id: ENTRY_ID,
  entry_code: 'E4',
  title: 'Streaming Ingestion Trade-offs',
  section_id: 'sec_findings',
  created: false,
  revisions: [mockRevisions[2]!, mockRevisions[1]!, mockRevisions[0]!],
  from_revision: 3,
  to_revision: 5,
  diff: mockDiffWordLevel,
  summary: '+2 −1',
}

/** An entry that first appeared in this session: from_revision is 1. */
export const mockChangeCreated: SessionEntryChange = {
  entry_id: 'ent_7bb0e412',
  entry_code: 'E7',
  title: 'Ingestion vendor comparison',
  section_id: 'sec_findings',
  created: true,
  revisions: [
    revision({ revision: 1, entry_id: 'ent_7bb0e412', title: 'Ingestion vendor comparison', summary: '' }),
    revision({ revision: 2, entry_id: 'ent_7bb0e412', title: 'Ingestion vendor comparison', summary: 'Updated content' }),
  ],
  from_revision: 1,
  to_revision: 2,
  diff: mockDiffWithGaps,
  summary: '+4 −2',
}


/** A created entry whose body is one paragraph: a small, ordinary diff. */
export const mockChangeCreatedSmall: SessionEntryChange = {
  entry_id: 'ent_31d0a7c2',
  entry_code: 'E9',
  title: 'Retry budget baseline',
  section_id: 'sec_open',
  created: true,
  revisions: [
    revision({ revision: 1, entry_id: 'ent_31d0a7c2', title: 'Retry budget baseline', summary: '' }),
  ],
  from_revision: 1,
  to_revision: 1,
  diff: mockDiffFirstVersion,
  summary: '+7 −0',
}

export const mockSessionChanges: SessionEntryChange[] = [
  mockChangeModified,
  mockChangeCreated,
  mockChangeCreatedSmall,
]

/** A long session: the summary line and the card list both have to hold up. */
export const mockSessionChangesMany: SessionEntryChange[] = Array.from({ length: 12 }, (_, i) => {
  const created = i % 3 === 0
  const base = created ? mockChangeCreatedSmall : mockChangeModified
  return {
    ...base,
    entry_id: `ent_bulk_${i}`,
    entry_code: `E${10 + i}`,
    title: [
      'Broker configuration baseline',
      'Consumer lag alerting',
      'Schema evolution rules',
      'Cost model for self-hosting',
      'Replay tooling gaps',
      'Retention policy per topic',
    ][i % 6]!,
    created,
    from_revision: created ? 1 : 2,
    to_revision: created ? 1 : 4,
  }
})
