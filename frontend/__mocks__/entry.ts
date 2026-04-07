export const mockEntry = {
  id: 'ent_001',
  code: 'E1',
  title: 'Component Composition Patterns',
  description: 'Analysis of composable patterns in Vue 3',
  content: '# Component Composition\n\nVue 3 introduces the Composition API which allows better code organization...\n\n## Key Patterns\n\n- **Composables** - Reusable logic extraction\n- **Provide/Inject** - Dependency injection\n- **Render functions** - Programmatic templates\n\nSee also [[E2]] for related patterns and [[R2:E1]] for state management.',
  status: 'completed',
  section_id: 'sec_001',
  session_id: 'sess_001',
  tags: ['vue', 'composables'],
  created_at: '2025-03-15T10:00:00Z',
  updated_at: '2025-03-15T14:30:00Z',
}

export const mockEntryDraft = {
  ...mockEntry,
  id: 'ent_002',
  code: 'E2',
  title: 'Reactive State Management',
  description: 'How ref, reactive, and computed work together',
  status: 'draft',
  tags: ['vue', 'reactivity'],
}

export const mockEntryNoTags = {
  ...mockEntry,
  id: 'ent_003',
  code: 'E3',
  title: 'Template Syntax Deep Dive',
  description: null,
  tags: [],
  session_id: null,
}

export const mockEntries = [
  mockEntry,
  mockEntryDraft,
  mockEntryNoTags,
  { ...mockEntry, id: 'ent_004', code: 'E4', title: 'Slots and Render Functions', tags: ['vue', 'slots'], status: 'active' },
  { ...mockEntry, id: 'ent_005', code: 'E5', title: 'Performance Optimization', tags: ['vue', 'performance'], status: 'pending' },
]
