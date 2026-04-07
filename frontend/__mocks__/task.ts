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
  result: '**[Done]** Found 5 components over 250 lines. Created extraction plan for tasks.vue and index.vue.',
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
  result: '**[Rejected]** Custom CSS with design tokens provides better consistency.',
  status: 'failed',
  priority: 'low',
}

export const mockTasks = [mockTask, mockTaskHigh, mockTaskCompleted, mockTaskFailed]
