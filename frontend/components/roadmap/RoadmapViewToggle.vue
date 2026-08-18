<template>
  <SegmentedToggle
    :model-value="modelValue"
    :options="options"
    label="Roadmap view"
    @update:model-value="v => emit('update:modelValue', v as RoadmapView)"
  />
</template>

<script setup lang="ts">
export type RoadmapView = 'graph' | 'stages' | 'timeline'

defineProps<{ modelValue: RoadmapView }>()
const emit = defineEmits<{ 'update:modelValue': [RoadmapView] }>()

const options: { value: RoadmapView; label: string }[] = [
  { value: 'graph', label: 'Graph' },
  { value: 'stages', label: 'Stages' },
  { value: 'timeline', label: 'Timeline' },
]
</script>
