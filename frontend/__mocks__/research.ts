export const mockResearch = {
  id: 'res_001',
  code: 'R1',
  name: 'Vue Component Architecture',
  goal: 'Investigate best practices for Vue 3 component design and composition patterns',
  description: 'A deep dive into Composition API patterns, component decomposition strategies, and reusable primitive extraction.',
  status: 'active',
  tags: ['vue', 'architecture', 'frontend'],
  memory: [
    { id: 'memory-1', text: 'Composition API preferred over Options API for complex components', author: 'agent', created_at: '2025-03-01T10:00:00Z', session_code: 'SS1', version: 1 },
    { id: 'memory-2', text: 'Keep components under 200 lines for readability', author: 'user', created_at: '2025-03-02T12:00:00Z', version: 1 },
    { id: 'memory-3', text: 'Extract shared primitives when pattern appears 3+ times', author: 'unknown', created_at: null, version: 1 },
  ],
  created_at: '2025-03-01T10:00:00Z',
  updated_at: '2025-03-20T14:30:00Z',
}

export const mockResearchCompleted = {
  ...mockResearch,
  id: 'res_002',
  code: 'R2',
  name: 'State Management Patterns',
  goal: 'Compare Pinia, Vuex, and composable-based state management',
  status: 'completed',
  tags: ['vue', 'state', 'pinia'],
}

export const mockResearchArchived = {
  ...mockResearch,
  id: 'res_003',
  code: 'R3',
  name: 'CSS Architecture Review',
  goal: 'Evaluate CSS custom properties vs Tailwind vs CSS-in-JS',
  status: 'archived',
  tags: ['css', 'design-system'],
  memory: [],
}

export const mockResearches = [mockResearch, mockResearchCompleted, mockResearchArchived]
