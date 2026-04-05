<template>
  <span :class="['badge', `badge-${status}`]">
    <span class="badge-icon">{{ icon }}</span>
    {{ text }}
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{ status: string }>()

const STATUS_CONFIG: Record<string, { text: string; icon: string }> = {
  active:      { text: 'Active',      icon: '●' },
  completed:   { text: 'Completed',   icon: '✓' },
  archived:    { text: 'Archived',    icon: '○' },
  draft:       { text: 'Draft',       icon: '◌' },
  pending:     { text: 'Pending',     icon: '◷' },
  answered:    { text: 'Answered',    icon: '✓' },
  in_progress: { text: 'In Progress', icon: '▶' },
  deferred:    { text: 'Deferred',    icon: '→' },
  skipped:     { text: 'Skipped',     icon: '↷' },
  blocked:     { text: 'Blocked',     icon: '■' },
  failed:      { text: 'Failed',      icon: '×' },
  high:        { text: 'High',        icon: '↑' },
  medium:      { text: 'Medium',      icon: '•' },
  low:         { text: 'Low',         icon: '↓' },
}

const config = computed(() => STATUS_CONFIG[props.status] ?? { text: props.status, icon: '' })
const text = computed(() => config.value.text)
const icon = computed(() => config.value.icon)
</script>

<style scoped>
.badge-icon { margin-right: 0.125rem; }
</style>
