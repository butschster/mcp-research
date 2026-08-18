/** Templates as the API returns them, for stories and for anything that needs a
 *  catalogue without a server behind it.
 *
 *  `tier` and `source` are two different facts and the fixtures keep both: the
 *  tier says who owns it, the source says whether an upgrade rewrites it. The
 *  global tier holds both kinds — what ships in the binary, and what the
 *  operator of an instance added through the API.
 *
 *  The other thing these fixtures encode is that a list row and a single read
 *  are not the same object. `GET /api/templates` never sends the body, the
 *  research count or the resolved skills; only `GET /api/templates/{id}` does.
 *  So the list fixtures below carry none of the three, and a story that shows a
 *  row "started 9 researches" is showing a payload the server does not send. */

/** One entry of `skills_resolved`. `missing` is the one worth rendering: a team
 *  skill can be deleted after a template named it. */
export interface MockTemplateSkill {
  slug: string
  name?: string
  description?: string
  ambient?: boolean
  missing?: boolean
}

export interface MockTemplate {
  id: string
  slug: string
  name: string
  tier: 'global' | 'team'
  source: 'builtin' | 'user'
  /* Never omitted by the API. They arrive as '' when unset, which is why the
     components branch on falsiness rather than on the key being absent. */
  description: string
  when_to_use: string
  when_not_to_use: string
  skills: string[]
  /** Counted in SQL from the body, so a list can show a size without the text. */
  body_words: number
  version: number
  created_at: string
  updated_at: string
  /** Team tier only — a global template belongs to nobody. */
  team_id?: string
  /** Who wrote it. Absent on what ships in the binary. */
  user_id?: string
  /** The global slug this copies. A fork keeps its parent's slug, so the two
   *  are equal on a real fork. */
  forked_from?: string

  /* Single read only (`GET /api/templates/{id}`). Never in a list. */
  body?: string
  research_count?: number
  skills_resolved?: MockTemplateSkill[]
}

const at = (day: number) => `2025-08-${String(day).padStart(2, '0')}T09:14:00Z`

/** What ships. Two of the ten — enough to see a list, not a catalogue dump.
 *  No `user_id`: the loader wrote these, not a person. */
export const mockGlobalTemplates: MockTemplate[] = [
  {
    id: 'g1', slug: 'technology-comparison', name: 'Technology comparison',
    tier: 'global', source: 'builtin', version: 3, body_words: 812,
    created_at: at(1), updated_at: at(9),
    description: 'Choose between named alternatives and be able to defend the choice afterwards.',
    when_to_use:
      'Use when a decision is being made between named tools, libraries or platforms and the reasoning will have to survive review.',
    when_not_to_use:
      'Not when the choice is already made and what is wanted is justification, and not for a market scan — that is a landscape.',
    skills: ['evidence-grading'],
  },
  {
    id: 'g2', slug: 'user-interview-study', name: 'User interview study',
    tier: 'global', source: 'builtin', version: 2, body_words: 954,
    created_at: at(1), updated_at: at(4),
    description: 'Findings traceable to what real people actually said.',
    when_to_use:
      'Use when the question is what people do and why, and the evidence has to come from talking to them rather than from analytics.',
    when_not_to_use:
      'Not for measuring how many — that is a survey. Not when nobody can be recruited in the time available.',
    skills: ['structured-interviewing', 'evidence-grading'],
  },
]

/** Added on this server by whoever runs it: global, but not ours. Written
 *  through the API by an operator, so it has a `user_id` and version 1. */
export const mockHouseTemplates: MockTemplate[] = [
  {
    id: 'h1', slug: 'house-diligence', name: 'House diligence',
    tier: 'global', source: 'user', user_id: 'usr_ops', version: 1, body_words: 430,
    created_at: at(6), updated_at: at(6),
    description: 'The way this company runs a diligence, before anything else is decided.',
    when_to_use:
      'Use when a deal, a supplier or a partnership is being assessed here, and the answer has to go to the committee.',
    when_not_to_use: 'Not for a technical evaluation — use the comparison methodology for that.',
    skills: ['evidence-grading'],
  },
]

/** A team's own, and its edited copy of a shipped one. A fork keeps the parent's
 *  slug, which is why `slug` and `forked_from` match on the second. */
export const mockTeamTemplates: MockTemplate[] = [
  {
    id: 't1', slug: 'our-playbook', name: 'Our discovery playbook',
    tier: 'team', source: 'user', team_id: 'team-1', user_id: 'usr_ada',
    version: 5, body_words: 601, created_at: at(2), updated_at: at(14),
    description: 'How this team opens a discovery, including the two questions we always forget.',
    when_to_use:
      'Use when starting discovery on a new customer problem in this team, before any solution has been sketched.',
    when_not_to_use: 'Not for a delivery post-mortem.',
    skills: ['structured-interviewing'],
  },
  {
    id: 't2', slug: 'literature-review', name: 'Literature review (ours)',
    tier: 'team', source: 'user', team_id: 'team-1', user_id: 'usr_ada',
    forked_from: 'literature-review', version: 2, body_words: 770,
    created_at: at(3), updated_at: at(11),
    description: 'The shipped review methodology with our exclusion rules bolted on.',
    when_to_use:
      'Use when the answer is expected to exist already in published work and the job is to find, exclude and summarise it.',
    when_not_to_use: 'Not when the field has no published work worth the search.',
    skills: ['evidence-grading'],
  },
]

/** No matching line: an agent will never choose it, and the row has to say so.
 *  A write cannot produce this any more — `when_to_use` is required — but a row
 *  that predates the check, or a built-in file whose frontmatter omits the key,
 *  still arrives looking like this. */
export const mockTemplateWithoutCriteria: MockTemplate = {
  id: 'x1', slug: 'silent', name: 'Untriggered methodology',
  tier: 'team', source: 'user', team_id: 'team-1', user_id: 'usr_ada',
  version: 1, body_words: 88, created_at: at(7), updated_at: at(7),
  description: '', when_to_use: '', when_not_to_use: '', skills: [],
}

/** Both criteria within a few runes of the 240 the service allows, with an
 *  unbroken product name in one. This is the width the row and the detail page
 *  are actually designed for. */
export const mockTemplateLongCriteria: MockTemplate = {
  id: 'x2', slug: 'platform-migration-readiness',
  name: 'Platform migration readiness assessment for long-lived internal services',
  tier: 'team', source: 'user', team_id: 'team-2', user_id: 'usr_kim',
  version: 1, body_words: 1840, created_at: at(5), updated_at: at(15),
  description: 'Whether a service can be moved at all, before anyone argues about where to.',
  when_to_use:
    'Use when a long-lived internal service is being considered for a move onto another platform — kubernetes-on-bare-metal-with-cilium-and-longhorn, a managed runtime, or somebody else entirely — and the decision has to survive the team',
  when_not_to_use:
    'Not when the move is already committed and dated, and not when the service is small enough that two people could rewrite it in a fortnight — the assessment then costs more than the migration it is meant to de-risk in the end',
  skills: ['evidence-grading', 'structured-interviewing'],
}

/** One template as a single read: body, research count and resolved skills, none
 *  of which a list carries. */
export const mockTemplateDetail: MockTemplate = {
  ...mockGlobalTemplates[0]!,
  research_count: 4,
  skills_resolved: [
    {
      slug: 'evidence-grading', name: 'Grading evidence',
      description: 'Use when recording a claim that rests on a source, or when two sources disagree.',
    },
  ],
  body: [
    '# Technology comparison',
    '',
    'This research ends in a choice somebody has to live with.',
    '',
    '## Before you propose anything',
    '',
    '1. **What breaks if this decision is wrong, and when would you find out?**',
    '2. **What are you not allowed to change?** Licence terms, an existing contract,',
    '   the language the team already writes.',
    '',
    '## Working rules',
    '',
    '**Criteria are fixed and weighted before any candidate is named.** Otherwise',
    'they get retrofitted around whichever option somebody already likes.',
    '',
    '**"Keep what we have" is candidate zero**, scored like the others.',
  ].join('\n'),
}

/** A single read whose skill list has rotted: the team skill it names was
 *  deleted, so the server reports it missing rather than dropping it. */
export const mockTemplateDetailMissingSkill: MockTemplate = {
  ...mockTemplateDetail,
  id: 'g1-missing',
  skills: ['evidence-grading', 'house-rules'],
  skills_resolved: [
    ...(mockTemplateDetail.skills_resolved ?? []),
    { slug: 'house-rules', missing: true },
  ],
}
