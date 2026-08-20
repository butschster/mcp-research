import type { Annotation } from '../composables/useAnnotations'

/**
 * One factory, eight story files.
 *
 * An `Annotation` carries twenty-odd fields and every one of the annotation
 * components reads a different handful of them — the row wants the anchor and
 * the attempt count, the thread wants the rejections, the pass review wants the
 * resolution. Spelling the whole payload out per story meant a field added to
 * the type would be missing from seven files and present in the eighth, which
 * is drift that still compiles.
 *
 * `makeAnnotation` returns a plausible open `verify` mark; every story says only
 * what makes its case different.
 */
export function makeAnnotation(overrides: Partial<Annotation> = {}): Annotation {
  const code = overrides.code ?? 'A1'
  const base: Annotation = {
    id: `ann_${code.toLowerCase()}`,
    code,
    research_id: 'res_001',
    entry_id: 'ent_001',
    block_id: 'blk_7f31',
    quote: {
      exact: 'Composables replaced mixins outright by the end of 2023.',
      prefix: 'Across the ecosystem, ',
      suffix: ' Teams that stayed on the Options API…',
    },
    anchored_revision: 4,
    kind: 'verify',
    body: 'Which survey says "outright"? The Vue docs still document mixins.',
    author_kind: 'user',
    author_name: 'Pavel',
    user_id: 'usr_001',
    status: 'open',
    rejections: [],
    attempts: 0,
    created_at: '2025-03-18T09:12:00Z',
    updated_at: '2025-03-18T09:12:00Z',
    anchor: {
      state: 'anchored',
      strategy: 'block+quote',
      confidence: 1,
      block_id: 'blk_7f31',
      block_index: 3,
      block_type: 'markdown',
      text: 'Composables replaced mixins outright by the end of 2023.',
    },
    entry_code: 'E1',
    entry_title: 'Component Composition Patterns',
    entry_type: 'blocks',
  }
  return { ...base, ...overrides }
}

/** The ordinary case: open, anchored, waiting for the agent. */
export const mockAnnotation = makeAnnotation()

/** A `dig` — the agent is asked to write a child document, not to argue. */
export const mockAnnotationDig = makeAnnotation({
  code: 'A2',
  kind: 'dig',
  quote: { exact: 'Provide/inject is the escape hatch for deep trees.' },
  body: 'This deserves its own entry — when does the escape hatch stop being one?',
  anchor: { state: 'anchored', strategy: 'block+quote', confidence: 1, block_id: 'blk_9c02', block_index: 5, block_type: 'markdown' },
})

/** A `disagree` — both positions get recorded, the text does not get rewritten. */
export const mockAnnotationDisagree = makeAnnotation({
  code: 'A3',
  kind: 'disagree',
  quote: { exact: 'Keeping components under 200 lines is a hard rule.' },
  body: 'It is a heuristic, not a rule. A 300-line form with no branching reads fine.',
  anchor: { state: 'anchored', strategy: 'block+quote', confidence: 1, block_id: 'blk_a41f', block_index: 7, block_type: 'markdown' },
})

/** Answered by the agent and waiting for a person to accept or send it back. */
export const mockAnnotationAnswered = makeAnnotation({
  code: 'A4',
  status: 'answered',
  resolution:
    'Could not be confirmed. The 2023 State of Vue survey reports 61% composable ' +
    'adoption, not a replacement — the sentence now says "largely displaced". ' +
    'Source recorded in [[E3]].',
  resolved_revision: 6,
  answered_at: '2025-03-19T14:02:00Z',
  updated_at: '2025-03-19T14:02:00Z',
  session_id: 'sess_002',
  task_id: 'task_003',
})

/** The block is still there; the sentence under the mark was edited away. */
export const mockAnnotationDrifted = makeAnnotation({
  code: 'A5',
  kind: 'disagree',
  quote: { exact: 'Pinia is a drop-in replacement for Vuex in every case.' },
  body: 'Not for dynamic module registration — that has no equivalent.',
  anchored_revision: 2,
  anchor: { state: 'drifted', strategy: 'block', confidence: 0.72, block_id: 'blk_31bd', block_index: 2, block_type: 'markdown' },
})

/** The sentence turned up elsewhere in the document; confidence is printed. */
export const mockAnnotationMoved = makeAnnotation({
  code: 'A6',
  kind: 'dig',
  quote: { exact: 'Reactivity is built on Proxy, which cannot observe property addition on arrays.' },
  body: 'Half true — arrays are patched separately. Worth its own entry.',
  anchor: { state: 'moved', strategy: 'quote', confidence: 0.64, block_id: 'blk_5e88', block_index: 9, block_type: 'markdown' },
})

/** The marked text is gone from the document entirely. */
export const mockAnnotationOrphaned = makeAnnotation({
  code: 'A7',
  quote: { exact: 'Every composable should return a readonly ref.' },
  body: 'Where does this come from? It contradicts [[E2]].',
  anchored_revision: 3,
  anchor: { state: 'orphaned', strategy: 'none', confidence: 0, block_index: -1 },
})

/** Sent back twice, with the reasons the agent has to read before trying again. */
export const mockAnnotationRejected = makeAnnotation({
  code: 'A8',
  status: 'answered',
  attempts: 2,
  resolution: 'Rephrased the claim to "widely adopted" and linked the survey.',
  resolved_revision: 8,
  answered_at: '2025-03-20T11:40:00Z',
  rejections: [
    { reason: 'Rewording the sentence is not a source. Cite the survey or say it could not be confirmed.', revision: 5, at: '2025-03-19T16:20:00Z' },
    { reason: 'The link goes to the survey homepage, not the question it answers.', revision: 7, at: '2025-03-20T10:05:00Z' },
  ],
})

/** A mark on a markdown document — no blocks, so it is found by its text alone. */
export const mockAnnotationMarkdown = makeAnnotation({
  code: 'A9',
  entry_id: 'ent_003',
  entry_code: 'E3',
  entry_title: 'Template Syntax Deep Dive',
  entry_type: 'markdown',
  block_id: undefined,
  quote: { exact: 'v-if and v-for on the same element is merely discouraged.' },
  body: 'It is a lint error in the recommended config, not a preference.',
  anchor: { state: 'moved', strategy: 'quote', confidence: 0.81, block_index: 0 },
})

/** Closed after the answer was accepted. */
export const mockAnnotationClosed = makeAnnotation({
  code: 'A10',
  status: 'closed',
  resolution: 'Confirmed against the 2023 survey; the sentence now names the number.',
  resolved_revision: 6,
  answered_at: '2025-03-19T14:02:00Z',
  closed_at: '2025-03-19T15:30:00Z',
})

/** A second document, so a grouped list has more than one group to draw. */
export const mockAnnotationsSecondEntry: Annotation[] = [
  makeAnnotation({
    code: 'A11',
    entry_id: 'ent_002',
    entry_code: 'E2',
    entry_title: 'Reactive State Management',
    kind: 'dig',
    quote: { exact: 'shallowRef is an optimisation of last resort.' },
    body: 'When is it the first resort? Large frozen payloads, maybe.',
    anchor: { state: 'anchored', strategy: 'block+quote', confidence: 1, block_index: 1, block_type: 'markdown' },
  }),
  makeAnnotation({
    code: 'A12',
    entry_id: 'ent_002',
    entry_code: 'E2',
    entry_title: 'Reactive State Management',
    kind: 'disagree',
    status: 'answered',
    attempts: 1,
    quote: { exact: 'watchEffect is always preferable to watch.' },
    body: 'It hides its dependencies, which is the opposite of preferable in a big component.',
    resolution: 'Both positions recorded — see the trade-off table added to [[E2]].',
    resolved_revision: 4,
    answered_at: '2025-03-20T08:15:00Z',
    rejections: [{ reason: 'The first answer just deleted the sentence.', revision: 3, at: '2025-03-19T18:00:00Z' }],
    anchor: { state: 'drifted', strategy: 'block', confidence: 0.55, block_index: 4, block_type: 'markdown' },
  }),
]

/** A queue as the research page shows it: two documents, every anchor state. */
export const mockAnnotationQueue: Annotation[] = [
  mockAnnotation,
  mockAnnotationDig,
  mockAnnotationDrifted,
  mockAnnotationMoved,
  mockAnnotationOrphaned,
  ...mockAnnotationsSecondEntry,
]

/** What a finished pass leaves behind: everything answered, nothing accepted. */
export const mockAnnotationsAnswered: Annotation[] = [
  mockAnnotationAnswered,
  mockAnnotationRejected,
  makeAnnotation({
    code: 'A13',
    kind: 'dig',
    status: 'answered',
    quote: { exact: 'Provide/inject is the escape hatch for deep trees.' },
    body: 'Deserves its own entry.',
    resolution: 'Written up as [[E7]], linked from the paragraph above.',
    resolved_revision: 6,
    answered_at: '2025-03-19T14:04:00Z',
    anchor: { state: 'anchored', strategy: 'block+quote', confidence: 1, block_index: 5, block_type: 'markdown' },
  }),
  makeAnnotation({
    code: 'A14',
    entry_id: 'ent_002',
    entry_code: 'E2',
    entry_title: 'Reactive State Management',
    kind: 'disagree',
    status: 'answered',
    quote: { exact: 'reactive() and ref() perform identically.' },
    body: 'Not for large arrays — reactive proxies every index.',
    resolution: 'Both positions recorded with a benchmark; the claim is now qualified.',
    resolved_revision: 4,
    answered_at: '2025-03-20T08:20:00Z',
    anchor: { state: 'moved', strategy: 'quote', confidence: 0.7, block_index: 2 },
  }),
]

/** Gutter positions as `useAnnotationOverlay.positions()` measures them:
 *  pixels from the top of the entry card, sorted, one per painted mark. */
export function markPositions(items: Annotation[], tops: number[]) {
  return items.map((a, i) => ({ id: a.id, code: a.code, top: tops[i] ?? i * 80 }))
}
