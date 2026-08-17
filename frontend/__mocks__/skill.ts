/** Skills as the API returns them, for stories and for anything that needs a
 *  list without a server behind it. */

export interface MockSkill {
  id: string
  slug: string
  name: string
  tier: 'builtin' | 'team' | 'private'
  description: string
  body?: string
  ambient?: boolean
  forked_from?: string
  needs_trigger?: boolean
  via_template?: boolean
  attached?: boolean
  attached_at?: string
  body_tokens: number
}

const daysAgo = (n: number) => new Date(Date.now() - n * 86_400_000).toISOString()

/** The four product skills: always on, never counted, never detachable. */
export const mockAmbientSkills: MockSkill[] = [
  {
    id: 'a1', slug: 'managing-a-research', name: 'Managing a research', tier: 'builtin',
    ambient: true, body_tokens: 809,
    description: 'Use when deciding what to create next in a research — a section, an entry, a session, a task or a roadmap.',
  },
  {
    id: 'a2', slug: 'writing-entries', name: 'Writing entries', tier: 'builtin',
    ambient: true, body_tokens: 652,
    description: 'Use when about to create or rewrite an entry, or when choosing between markdown and a blocks document.',
  },
  {
    id: 'a3', slug: 'writing-artifacts', name: 'Writing an artifact', tier: 'builtin',
    ambient: true, body_tokens: 588,
    description: 'Use when about to write an html block — a chart, a dashboard, an interactive layout.',
  },
  {
    id: 'a4', slug: 'building-roadmaps', name: 'Building roadmaps', tier: 'builtin',
    ambient: true, body_tokens: 588,
    description: 'Use when the finding is a path or a dependency network rather than a list.',
  },
]

/** What somebody picked. These are what the six-skill budget counts. */
export const mockChosenSkills: MockSkill[] = [
  {
    id: 'c1', slug: 'house-rules', name: 'How we run this one', tier: 'private',
    body_tokens: 240, attached_at: daysAgo(2),
    description: 'Use when writing anything in this particular research.',
  },
  {
    id: 'c2', slug: 'evidence-grading', name: 'Grading evidence, our way', tier: 'team',
    forked_from: 'evidence-grading', body_tokens: 700, attached_at: daysAgo(9),
    description: "Use when recording a claim that rests on a source, in this team's house style.",
  },
  {
    id: 'c3', slug: 'structured-interviewing', name: 'Structured interviewing', tier: 'builtin',
    body_tokens: 702, attached_at: daysAgo(9), via_template: true,
    description: 'Use when the next step is asking a person what they know rather than looking something up.',
  },
]

/** What the migration from `instruction` produces: a machine-written trigger. */
export const mockMigratedSkill: MockSkill = {
  id: 'm1', slug: 'research-r7-rules', name: 'R7 — working rules', tier: 'private',
  needs_trigger: true, body_tokens: 780, attached_at: daysAgo(1),
  description: 'Imported from the research instruction on 12 August.',
}

export const mockLibrarySkills: MockSkill[] = [
  {
    id: 'l1', slug: 'evidence-grading', name: 'Grading evidence', tier: 'builtin',
    body_tokens: 643, attached: false,
    description: 'Use when recording a claim that rests on a source, or when two sources disagree.',
  },
  {
    id: 'l2', slug: 'pricing-questions', name: 'Pricing questions', tier: 'team',
    body_tokens: 310, attached: false,
    description: 'Use when asking a customer about what they pay.',
  },
]

export const mockSkillWithBody: MockSkill = {
  ...mockLibrarySkills[0]!,
  body: [
    '# Grading evidence',
    '',
    'The job is not to decide what is true. It is to make visible what each claim',
    'rests on, so a reader can disagree with the evidence rather than with your tone.',
    '',
    '## Rank what you find',
    '',
    '1. **Measured** — we ran it, we counted it, we saw the invoice.',
    '2. **Primary record** — a filing, an audited number, a contract.',
    '3. **Direct testimony** — a practitioner saying what they did.',
    '',
    '**Never average two numbers into a third that nobody published.**',
  ].join('\n'),
}

/** A skill saved without a trigger line: in the index, and unopenable. */
export const mockSkillWithoutTrigger: MockSkill = {
  id: 'x1', slug: 'silent', name: 'Silent skill', tier: 'private',
  description: '', body_tokens: 120, attached_at: daysAgo(4),
}
