/**
 * The export payload — `GET /api/researches/{id}/export`, and the redacted form
 * of it a share link serves.
 *
 * Shape follows the handler exactly: `research`, `sections[].entries[]`,
 * `sessions[].questions[]`, `tasks`, `roadmap_count`, `markdown`. An entry
 * carries `content_markdown` beside `content`, because a blocks entry ships its
 * document as JSON and the printable page must not have to parse it twice.
 *
 * Content uses `[[E3]]` cross-references, since that is what the author wrote
 * and rendering them is the one thing the document does to the text.
 */

export const mockExportData = {
  research: {
    id: 'res_007',
    code: 'R7',
    name: 'Pricing benchmark, Q3',
    goal: 'Understand how three competitors price seat-based tiers, and where our own tiering sits against them.',
    description:
      'Desk research plus four customer conversations. Numbers are list price, not negotiated, and are current as of the third week of the quarter.',
    status: 'active',
    tags: ['pricing', 'competitive', 'q3'],
    created_at: '2026-06-02T09:00:00Z',
    updated_at: '2026-07-28T16:40:00Z',
  },
  sections: [
    {
      id: 'sec_method',
      name: 'method',
      display_name: 'Method',
      description: 'How the numbers were gathered and what they exclude.',
      status: 'completed',
      entries: [
        {
          id: 'ent_701',
          code: 'E1',
          title: 'What counts as a seat',
          description: 'The definition three vendors disagree on.',
          status: 'completed',
          tags: ['definitions'],
          entry_type: 'markdown',
          content:
            'Two of the three vendors bill per *named* user; the third bills per concurrent session, which makes a straight per-seat comparison misleading below about twelve seats.\n\nThe correction we apply is in [[E3]], and the raw quotes are in [[E2]].',
        },
        {
          id: 'ent_702',
          code: 'E2',
          title: 'Sources and their dates',
          description: null,
          status: 'completed',
          tags: ['sources'],
          entry_type: 'markdown',
          content:
            '| Vendor | Page | Read on |\n| --- | --- | --- |\n| Kestrel | Public pricing | 14 Jul |\n| Northlight | Sales quote | 2 Jul |\n| Verge | Public pricing | 14 Jul |\n\nNorthlight does not publish; the quote was shared by a customer and may already be out of date.',
        },
      ],
    },
    {
      id: 'sec_findings',
      name: 'findings',
      display_name: 'Findings',
      description: 'What the numbers say once they are comparable.',
      status: 'active',
      entries: [
        {
          id: 'ent_703',
          code: 'E3',
          title: 'Seat pricing at Kestrel',
          description: 'Their tiers cluster around twelve seats.',
          status: 'answered',
          tags: ['kestrel', 'pricing'],
          entry_type: 'markdown',
          content:
            '## The shape of it\n\nKestrel prices in three bands, and the middle band is where almost every customer lands. The jump from band two to band three is 2.4x for 1.5x the seats, which is what pushes teams to sit on eleven seats rather than thirteen.\n\n```mermaid\nflowchart LR\n    A[1-5 seats] -->|$18/seat| B[6-12 seats]\n    B -->|$14/seat| C[13+ seats]\n    C -->|$34/seat| D[Enterprise]\n```\n\nThe definition problem this depends on is in [[E1]].',
        },
        {
          id: 'ent_704',
          code: 'E4',
          title: 'Where we sit',
          description: 'Two bands below Kestrel at the low end, level at the top.',
          status: 'draft',
          tags: ['us', 'pricing'],
          entry_type: 'markdown',
          content:
            'At five seats we are 22% cheaper. At forty we are within 3%, which is inside the noise of a negotiated deal.\n\n- Below ten seats, price is not why we lose\n- Between ten and twenty, it is the only thing mentioned\n- Above twenty, procurement asks about SSO before it asks about price',
        },
      ],
    },
    {
      id: 'sec_open',
      name: 'open_questions',
      display_name: 'Open questions',
      description: null,
      status: 'draft',
      entries: [],
    },
  ],
  sessions: [
    {
      id: 'sess_701',
      code: 'SS1',
      title: 'Competitor desk research',
      focus: 'Get comparable list prices for the three vendors',
      status: 'completed',
      notes: 'Numbers landed in [[E3]]; the definition wrangle became [[E1]].',
      created_at: '2026-06-04T10:00:00Z',
      questions: [
        {
          id: 'q_701',
          code: 'Q1',
          text: 'Do all three vendors price per named user?',
          answer:
            'No — Verge prices per concurrent session. See [[E1]] for the correction we apply before comparing.',
          status: 'answered',
          priority: 'high',
          area: 'definitions',
          parent_id: null,
        },
        {
          id: 'q_702',
          code: 'Q2',
          text: 'Is the Northlight quote list price or negotiated?',
          answer: null,
          status: 'pending',
          priority: 'medium',
          area: 'sources',
          parent_id: null,
        },
      ],
    },
  ],
  tasks: [
    {
      id: 'task_701',
      code: 'T1',
      title: 'Re-read Kestrel pricing after their September release',
      description: 'They pre-announced a tier change; the bands in [[E3]] may not survive it.',
      result: null,
      status: 'pending',
      priority: 'high',
      completed_at: null,
    },
    {
      id: 'task_702',
      code: 'T2',
      title: 'Confirm the Northlight number with a second customer',
      description: null,
      result: '**[Done]** Second customer quoted within 4% of the first. Recorded in [[E2]].',
      status: 'completed',
      priority: 'medium',
      completed_at: '2026-07-20T11:00:00Z',
    },
  ],
  roadmap_count: 2,
  markdown: '# Pricing benchmark, Q3\n\n(The markdown projection is what the .md download uses.)',
}

/**
 * What a share export looks like when the link does not include sessions or
 * tasks: the arrays are absent, not empty, and `instruction` and `memory` never
 * appear on the research at all.
 */
export const mockExportDataShared = {
  research: mockExportData.research,
  sections: mockExportData.sections,
  roadmap_count: 0,
  markdown: mockExportData.markdown,
}

/** A research with sections and nothing in them. */
export const mockExportDataEmpty = {
  research: {
    id: 'res_new',
    code: 'R9',
    name: 'Onboarding drop-off',
    goal: '',
    description: '',
    status: 'active',
    tags: [],
  },
  sections: [
    { id: 'sec_a', name: 'method', display_name: 'Method', description: null, status: 'draft', entries: [] },
    { id: 'sec_b', name: 'findings', display_name: 'Findings', description: null, status: 'draft', entries: [] },
  ],
  sessions: [],
  tasks: [],
  roadmap_count: 0,
  markdown: '# Onboarding drop-off\n',
}

/** One entry, stored as blocks. The document renders the blocks rather than
 *  their markdown projection, because this page is what becomes the PDF. */
export const mockExportDataBlocks = {
  ...mockExportData,
  sections: [
    {
      id: 'sec_diagram',
      name: 'architecture',
      display_name: 'Architecture',
      description: 'One entry, written in blocks rather than markdown.',
      status: 'active',
      entries: [
        {
          id: 'ent_blocks',
          code: 'E7',
          title: 'How a quote reaches the comparison table',
          description: 'The path a number takes from a vendor page to [[E3]].',
          status: 'active',
          tags: ['method'],
          entry_type: 'blocks',
          content: JSON.stringify({
            blocks: [
              {
                type: 'paragraph',
                data: { text: 'Every number in [[E3]] arrives the same way, and the normalisation step is where the seat definition from [[E1]] is applied.' },
              },
              {
                type: 'mermaid',
                data: {
                  code: 'flowchart TD\n    A[Vendor page] --> B[Quote captured]\n    B --> C{Per named user?}\n    C -->|yes| D[Use as-is]\n    C -->|no| E[Normalise to named seats]\n    D --> F[Comparison table]\n    E --> F',
                  caption: 'The correction is applied once, at capture — see [[E1]].',
                },
              },
            ],
          }),
          content_markdown:
            'Every number in [[E3]] arrives the same way.\n\n```mermaid\nflowchart TD\n    A[Vendor page] --> F[Comparison table]\n```',
        },
      ],
    },
  ],
  sessions: [],
  tasks: [],
}
