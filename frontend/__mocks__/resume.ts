import type { EntryUpdate } from '../composables/useEntryUpdates'
import type { ResumeSummary, ResumeTaskItem } from '../composables/useResearchResume'

/** A group with everything filled in, so a story does not repeat the shape. */
function group<T>(items: T[], total = items.length, tool = 'task_list', href = '/research/R1/tasks') {
  return { items, returned: items.length, total, has_more: total > items.length, more: { tool, href } }
}

export const mockResumeSessions = [
  { id: 'sess_003', code: 'SS3', title: 'Provider pricing', focus: 'costs', status: 'active', updated_at: '2026-09-04T09:00:00Z' },
]

export const mockResumeRunning: ResumeTaskItem = { id: 'task_004', code: 'T4', title: 'Compare provider pricing across the three vendors', status: 'in_progress', priority: 'high', updated_at: '2026-09-04T09:10:00Z' }
export const mockResumeBlockedNDA: ResumeTaskItem = { id: 'task_009', code: 'T9', title: 'Wait on the vendor NDA before quoting', status: 'blocked', priority: 'medium', note: 'legal has not answered', updated_at: '2026-09-03T14:00:00Z' }
export const mockResumeBlockedRerun: ResumeTaskItem = { id: 'task_012', code: 'T12', title: 'Benchmark rerun needs the new dataset', status: 'blocked', priority: 'low', updated_at: '2026-09-03T11:00:00Z' }
export const mockResumeTodo: ResumeTaskItem = { id: 'task_015', code: 'T15', title: 'Draft the summary of the pricing findings', status: 'pending', priority: 'high', updated_at: '2026-09-02T08:00:00Z' }
export const mockResumeTasks = [mockResumeRunning, mockResumeBlockedNDA, mockResumeBlockedRerun, mockResumeTodo]

export const mockResumeQuestions = [
  { id: 'q_007', code: 'Q7', session_id: 'sess_003', session_code: 'SS3', text: 'What does the enterprise tier actually cost?', priority: 'high', status: 'pending' },
  { id: 'q_008', code: 'Q8', session_id: 'sess_003', session_code: 'SS3', text: 'Which vendor publishes latency percentiles?', priority: 'medium', status: 'in_progress' },
]

export const mockResumeMarks = [
  { id: 'ann_002', code: 'A2', entry_id: 'ent_001', entry_code: 'E1', entry_title: 'Benchmarks', kind: 'verify', status: 'open', quote: 'latency figures are quoted from 2023', updated_at: '2026-09-04T08:00:00Z' },
]

export const mockResumeAnswered = [
  { id: 'ann_005', code: 'A5', entry_id: 'ent_002', entry_code: 'E2', entry_title: 'Shortlist', kind: 'dig', status: 'answered', quote: 'nine stores in, four out', updated_at: '2026-09-04T07:00:00Z' },
]

export const mockResumeEntries = [
  { id: 'ent_001', code: 'E1', title: 'Benchmarks on our own corpus', section_id: 'sec_001', updated_at: '2026-09-04T09:20:00Z', author_kind: 'human' as const, revision: 4 },
  { id: 'ent_002', code: 'E2', title: 'The shortlist, and what fell off it', section_id: 'sec_001', updated_at: '2026-09-03T18:00:00Z', author_kind: 'agent' as const, revision: 2 },
]

export const mockResume: ResumeSummary = {
  schema_version: 1,
  generated_at: '2026-09-04T09:25:00Z',
  research: { id: 'res_001', code: 'R1', name: 'Vector store for the support assistant', status: 'active', role: 'owner', can_write: true },
  sessions: { items: mockResumeSessions, selected_id: 'sess_003', selection_required: false, active_count: 1 },
  work: {
    in_progress: group([mockResumeRunning], 1),
    blocked: group([mockResumeBlockedNDA, mockResumeBlockedRerun], 2),
    pending: group([mockResumeTodo], 5),
  },
  questions: {
    open: group(mockResumeQuestions, 4, 'session_get', '/research/R1/session/SS3'),
    deferred: group([], 0, 'session_get', '/research/R1/session/SS3'),
  },
  annotations: {
    to_work: group(mockResumeMarks, 3, 'annotation_list', '/research/R1/annotations'),
    awaiting_human: group(mockResumeAnswered, 1, 'annotation_list', '/research/R1/annotations'),
  },
  recent_entries: group(mockResumeEntries, 6, 'entry_list', '/research/R1?section=__all__'),
  next_actions: [
    {
      kind: 'continue_task',
      target: { type: 'task', id: 'task_004', code: 'T4', title: 'Compare provider pricing across the three vendors' },
      reason_code: 'task_in_progress', reason: 'already in progress', actor: 'agent',
      tool: 'task_update', href: '/research/R1/tasks',
    },
    {
      kind: 'answer_annotation',
      target: { type: 'annotation', id: 'ann_002', code: 'A2', title: 'Benchmarks', entry_code: 'E1' },
      reason_code: 'annotation_open', reason: 'an open verify mark on E1', actor: 'agent',
      tool: 'annotation_list', href: '/research/R1/annotations',
    },
    {
      kind: 'answer_question',
      target: { type: 'question', id: 'q_007', code: 'Q7', title: 'What does the enterprise tier actually cost?', session_code: 'SS3' },
      reason_code: 'question_open', reason: 'the next unanswered question in SS3', actor: 'agent',
      tool: 'question_update', href: '/research/R1/session/SS3',
    },
  ],
  truncated: false,
}

/** Nothing outstanding: every total zero and no action proposed. */
export const mockResumeEmpty: ResumeSummary = {
  ...mockResume,
  sessions: { items: mockResumeSessions, selected_id: 'sess_003', selection_required: false, active_count: 1 },
  work: { in_progress: group([], 0), blocked: group([], 0), pending: group([], 0) },
  questions: { open: group([], 0), deferred: group([], 0) },
  annotations: { to_work: group([], 0), awaiting_human: group([], 0) },
  recent_entries: group([], 0),
  next_actions: [],
}

/** Two interviews open at once — the case the server refuses to guess. */
export const mockResumeAmbiguous: ResumeSummary = {
  ...mockResume,
  sessions: {
    items: [
      ...mockResumeSessions,
      { id: 'sess_004', code: 'SS4', title: 'Migration planning', focus: 'rollout', status: 'active', updated_at: '2026-09-04T08:00:00Z' },
    ],
    selected_id: undefined,
    selection_required: true,
    active_count: 2,
  },
  questions: { open: group([], 0, 'session_get', '/research/R1/sessions'), deferred: group([], 0, 'session_get', '/research/R1/sessions') },
  next_actions: [
    {
      kind: 'choose_session',
      target: { type: 'research', id: 'res_001', code: 'R1', title: 'Vector store for the support assistant' },
      reason_code: 'session_selection_required',
      reason: '2 sessions are active, say which one you are continuing',
      actor: 'human', tool: 'research_resume', href: '/research/R1/sessions',
    },
    ...mockResume.next_actions.slice(0, 2),
  ],
}

/**
 * The reader's own unseen-document markers, keyed by entry id the way
 * `indexEntryUpdates` produces them.
 *
 * This is the one personal thing in the block: two people looking at the same
 * research see the same Changed rows and different badges on them. It is fed
 * from the page's existing queue request, so reading the summary never marks
 * anything seen.
 */
export const mockResumeUpdatesByEntry: Record<string, EntryUpdate> = {
  ent_001: {
    entry_id: 'ent_001', entry_code: 'E1', research_id: 'res_001', section_id: 'sec_001',
    title: 'Benchmarks on our own corpus', entry_type: 'finding', status: 'active',
    current_revision: 4, seen_revision: 2, unseen_revisions: 2, kind: 'changed',
    updated_at: '2026-09-04T09:20:00Z',
  },
  ent_002: {
    entry_id: 'ent_002', entry_code: 'E2', research_id: 'res_001', section_id: 'sec_001',
    title: 'The shortlist, and what fell off it', entry_type: 'finding', status: 'active',
    current_revision: 1, seen_revision: 0, unseen_revisions: 1, kind: 'new',
    updated_at: '2026-09-03T18:00:00Z',
  },
}

/**
 * A research that has been interviewed for months.
 *
 * Six sessions is not a stress test — it is what a long-running research looks
 * like — and it is the point at which the picker stops being a two-item toggle
 * and starts needing its `{code} — {title}` labels to be readable.
 */
export const mockResumeManySessions: ResumeSummary = {
  ...mockResume,
  sessions: {
    items: [
      { id: 'sess_001', code: 'SS1', title: 'Scoping the problem', focus: 'scope', status: 'completed', updated_at: '2026-06-02T10:00:00Z' },
      { id: 'sess_002', code: 'SS2', title: 'Vendor landscape', focus: 'vendors', status: 'completed', updated_at: '2026-07-11T10:00:00Z' },
      ...mockResumeSessions,
      { id: 'sess_005', code: 'SS5', title: 'Migration planning', focus: 'rollout', status: 'completed', updated_at: '2026-08-20T10:00:00Z' },
      { id: 'sess_006', code: 'SS6', title: 'Self-hosting the index, and what it costs to run', focus: 'infrastructure', status: 'completed', updated_at: '2026-08-28T10:00:00Z' },
      { id: 'sess_007', code: 'SS7', title: 'Security review with the platform team', focus: 'security', status: 'archived', updated_at: '2026-09-01T10:00:00Z' },
    ],
    selected_id: 'sess_003',
    selection_required: false,
    active_count: 1,
  },
}

/**
 * The last session closed and nothing replaced it.
 *
 * One session means a link rather than a picker, and its real status is the
 * point: the head has to say the interview is over instead of implying one is
 * running.
 */
export const mockResumeNoActiveSession: ResumeSummary = {
  ...mockResume,
  sessions: {
    items: [{ id: 'sess_003', code: 'SS3', title: 'Provider pricing', focus: 'costs', status: 'completed', updated_at: '2026-09-04T09:00:00Z' }],
    selected_id: 'sess_003',
    selection_required: false,
    active_count: 0,
  },
}
