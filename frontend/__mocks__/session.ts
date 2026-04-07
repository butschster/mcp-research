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
