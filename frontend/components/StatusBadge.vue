<template>
  <span :class="['badge', `badge-${status}`]">{{ label }}</span>
</template>

<script setup lang="ts">
const props = defineProps<{ status: string }>()

const labelMap: Record<string, string> = {
  active:      '● Active',
  completed:   '✓ Completed',
  archived:    '▾ Archived',
  draft:       '○ Draft',
  pending:     '○ Pending',
  in_progress: '◐ In Progress',
  answered:    '✓ Answered',
  deferred:    '→ Deferred',
  skipped:     '⊘ Skipped',
  blocked:     '■ Blocked',
  failed:      '× Failed',
  high:        '↑ High',
  medium:      '• Medium',
  low:         '↓ Low',
}

const label = computed(() => labelMap[props.status] ?? props.status)
</script>
