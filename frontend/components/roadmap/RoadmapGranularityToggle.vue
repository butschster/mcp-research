<template>
  <SegmentedToggle
    :model-value="modelValue"
    :options="options"
    label="Timeline zoom"
    @update:model-value="v => emit('update:modelValue', v as TimeUnit)"
  />
</template>

<script setup lang="ts">
import type { TimeUnit } from '~/utils/roadmap'

defineProps<{ modelValue: TimeUnit }>()
const emit = defineEmits<{ 'update:modelValue': [TimeUnit] }>()

const options: { value: TimeUnit; label: string }[] = [
  { value: 'month', label: 'Month' },
  { value: 'quarter', label: 'Quarter' },
  { value: 'year', label: 'Year' },
]
</script>
