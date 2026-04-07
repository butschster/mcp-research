export const mockQuestion = {
  id: 'q_001',
  code: 'Q1',
  text: 'What are the key differences between Options API and Composition API for large-scale apps?',
  answer: 'The Composition API offers better TypeScript support, more flexible code organization through composables, and avoids the "this" context issues of Options API.',
  status: 'answered',
  priority: 'high',
  area: 'architecture',
  session_id: 'sess_001',
  parent_id: null,
}

export const mockQuestionPending = {
  ...mockQuestion,
  id: 'q_002',
  code: 'Q2',
  text: 'How do composables handle shared state across components?',
  answer: null,
  status: 'pending',
  priority: 'medium',
  area: 'state',
}

export const mockQuestionDeferred = {
  ...mockQuestion,
  id: 'q_003',
  code: 'Q3',
  text: 'What are the performance implications of reactive() vs ref()?',
  answer: null,
  status: 'deferred',
  priority: 'low',
  area: 'performance',
}

export const mockQuestions = [mockQuestion, mockQuestionPending, mockQuestionDeferred]

export const mockQuestionsGrouped: Record<string, any[]> = {
  answered: [mockQuestion],
  pending: [mockQuestionPending],
  deferred: [mockQuestionDeferred],
  skipped: [],
  in_progress: [],
}
