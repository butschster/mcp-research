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
