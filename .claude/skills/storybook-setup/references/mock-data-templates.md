# Mock Data Templates

Shared fixtures for Storybook stories. Place in `frontend/__mocks__/` and import from stories as needed.

## research.ts

```ts
export const mockResearch = {
  id: 'res_001',
  code: 'R1',
  name: 'Vue Component Architecture',
  goal: 'Investigate best practices for Vue 3 component design and composition patterns',
  description: 'A deep dive into Composition API patterns, component decomposition strategies, and reusable primitive extraction.',
  instruction: 'Focus on real-world scalability concerns. Compare Options API vs Composition API patterns. Document trade-offs.',
  status: 'active',
  tags: ['vue', 'architecture', 'frontend'],
  memory: [
    'Composition API preferred over Options API for complex components',
    'Keep components under 200 lines for readability',
    'Extract shared primitives when pattern appears 3+ times',
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
```

## section.ts

```ts
export const mockSection = {
  id: 'sec_001',
  name: 'fundamentals',
  display_name: 'Fundamentals',
  description: 'Core concepts and building blocks',
  status: 'active',
  entries_count: 5,
}

export const mockSectionCompleted = {
  id: 'sec_002',
  name: 'advanced_patterns',
  display_name: 'Advanced Patterns',
  description: 'Complex composition and architecture patterns',
  status: 'completed',
  entries_count: 8,
}

export const mockSectionDraft = {
  id: 'sec_003',
  name: 'testing',
  display_name: 'Testing Strategies',
  description: null,
  status: 'draft',
  entries_count: 0,
}

export const mockSections = [mockSection, mockSectionCompleted, mockSectionDraft]
```

## entry.ts

```ts
export const mockEntry = {
  id: 'ent_001',
  code: 'E1',
  title: 'Component Composition Patterns',
  description: 'Analysis of composable patterns in Vue 3',
  content: '# Component Composition\\n\\nVue 3 introduces the Composition API which allows better code organization...\\n\\n## Key Patterns\\n\\n- **Composables** - Reusable logic extraction\\n- **Provide/Inject** - Dependency injection\\n- **Render functions** - Programmatic templates\\n\\nSee also [[E2]] for related patterns and [[R2:E1]] for state management.',
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
  content: '# Reactive State\\n\\nVue 3 reactivity system is built on Proxy...',
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

export const mockEntriesBySection = {
  sec_001: mockEntries.slice(0, 3),
  sec_002: mockEntries.slice(3),
}
```

## session.ts

```ts
export const mockSession = {
  id: 'sess_001',
  code: 'SS1',
  title: 'Initial Research Session',
  focus: 'Explore [[E1]] composition patterns and [[E2]] reactivity',
  status: 'active',
  notes: 'Started with component architecture review. Key findings so far point to composables as the primary pattern.',
  created_at: '2025-03-10T10:00:00Z',
}

export const mockSessionCompleted = {
  ...mockSession,
  id: 'sess_002',
  code: 'SS2',
  title: 'Deep Dive: Composables',
  focus: 'Analyze real-world composable patterns',
  status: 'completed',
  notes: 'Reviewed 15+ composable implementations. Documented best practices.',
}

export const mockSessions = [mockSession, mockSessionCompleted]

export const mockSessionProgress = {
  total: 8,
  answered: 5,
  pending: 2,
  deferred: 1,
  skipped: 0,
}
```

## question.ts

```ts
export const mockQuestion = {
  id: 'q_001',
  code: 'Q1',
  text: 'What are the key differences between Options API and Composition API for large-scale apps?',
  answer: 'The Composition API offers better TypeScript support, more flexible code organization through composables, and avoids the "this" context issues of Options API. For large apps, the ability to extract and share logic via composables is the main advantage.',
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

export const mockQuestionsGrouped: Record<string, typeof mockQuestion[]> = {
  answered: [mockQuestion],
  pending: [mockQuestionPending],
  deferred: [mockQuestionDeferred],
  skipped: [],
  in_progress: [],
}
```

## task.ts

```ts
export const mockTask = {
  id: 'task_001',
  code: 'T1',
  title: 'Review composable patterns in production codebases',
  description: 'Analyze useAuth, useApi, and useRealtimeUpdates composables for common patterns and potential improvements.',
  result: null,
  status: 'pending',
  priority: 'medium',
  research_id: 'res_001',
  created_at: '2025-03-10T10:00:00Z',
  completed_at: null,
}

export const mockTaskHigh = {
  ...mockTask,
  id: 'task_002',
  code: 'T2',
  title: 'Document component decomposition strategy',
  description: 'Create guidelines for when to extract components vs keep inline.',
  status: 'in_progress',
  priority: 'high',
}

export const mockTaskCompleted = {
  ...mockTask,
  id: 'task_003',
  code: 'T3',
  title: 'Audit existing component sizes',
  description: 'Check all components for line count and identify candidates for decomposition.',
  result: '**[Done]** Found 5 components over 250 lines. Created extraction plan for tasks.vue (860 lines) and index.vue (918 lines).',
  status: 'completed',
  priority: 'medium',
  completed_at: '2025-03-18T16:00:00Z',
}

export const mockTaskFailed = {
  ...mockTask,
  id: 'task_004',
  code: 'T4',
  title: 'Migrate to Tailwind CSS',
  description: 'Evaluate switching from custom CSS to Tailwind.',
  result: '**[Rejected]** Custom CSS with design tokens provides better consistency for this project size.',
  status: 'failed',
  priority: 'low',
}

export const mockTasks = [mockTask, mockTaskHigh, mockTaskCompleted, mockTaskFailed]
```

## crossref.ts

```ts
export const mockOutgoingRef = {
  target_entry_id: 'ent_002',
  target_ref: 'E2',
  entry_code: 'E2',
  entry_title: 'Reactive State Management',
  research_code: 'R1',
  research_name: 'Vue Component Architecture',
  resolved: true,
}

export const mockOutgoingRefCrossResearch = {
  target_entry_id: 'ent_r2_001',
  target_ref: 'R2:E1',
  entry_code: 'E1',
  entry_title: 'Pinia Store Patterns',
  research_code: 'R2',
  research_name: 'State Management Patterns',
  resolved: true,
}

export const mockOutgoingRefUnresolved = {
  target_entry_id: null,
  target_ref: 'E99',
  entry_code: null,
  entry_title: null,
  research_code: null,
  research_name: null,
  resolved: false,
}

export const mockIncomingRef = {
  source_id: 'ent_004',
  source_type: 'entry',
  entry_code: 'E4',
  entry_title: 'Slots and Render Functions',
  research_code: 'R1',
  research_name: 'Vue Component Architecture',
}

export const mockOutgoingRefs = [mockOutgoingRef, mockOutgoingRefCrossResearch, mockOutgoingRefUnresolved]
export const mockIncomingRefs = [mockIncomingRef]
```

## external-link.ts

```ts
export const mockExternalLink = {
  id: 'link_001',
  url: 'https://vuejs.org/guide/reusability/composables.html',
  title: 'Composables - Vue.js Guide',
  domain: 'vuejs.org',
  entry_code: 'E1',
  entry_title: 'Component Composition Patterns',
}

export const mockExternalLinks = [
  mockExternalLink,
  { ...mockExternalLink, id: 'link_002', url: 'https://vuejs.org/api/composition-api-setup.html', title: 'Composition API: setup()', domain: 'vuejs.org' },
  { ...mockExternalLink, id: 'link_003', url: 'https://github.com/vueuse/vueuse', title: 'VueUse - Collection of Vue Composition Utilities', domain: 'github.com', entry_code: 'E2', entry_title: 'Reactive State Management' },
  { ...mockExternalLink, id: 'link_004', url: 'https://pinia.vuejs.org/core-concepts/', title: 'Core Concepts - Pinia', domain: 'pinia.vuejs.org', entry_code: 'E4', entry_title: 'Slots and Render Functions' },
]

export const mockExternalLinksGrouped = [
  { domain: 'vuejs.org', links: mockExternalLinks.slice(0, 2) },
  { domain: 'github.com', links: [mockExternalLinks[2]] },
  { domain: 'pinia.vuejs.org', links: [mockExternalLinks[3]] },
]
```
