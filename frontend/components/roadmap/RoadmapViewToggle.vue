<template>
  <div class="segmented rm-view-toggle" role="group" aria-label="Roadmap view">
    <button
      v-for="opt in options"
      :key="opt.value"
      type="button"
      :aria-pressed="modelValue === opt.value"
      @click="emit('update:modelValue', opt.value)"
    >
      {{ opt.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
export type RoadmapView = 'graph' | 'stages' | 'timeline'

defineProps<{ modelValue: RoadmapView }>()
const emit = defineEmits<{ 'update:modelValue': [RoadmapView] }>()

// Not a TabBar: these switch a rendered body, they do not own tab panels, so a
// pressed-button segmented control (aria-pressed, plain Tab to each) is the right
// metaphor rather than a roving-tabindex tablist.
const options: { value: RoadmapView; label: string }[] = [
  { value: 'graph', label: 'Graph' },
  { value: 'stages', label: 'Stages' },
  { value: 'timeline', label: 'Timeline' },
]
</script>

<style scoped>
.rm-view-toggle {
  flex-shrink: 0;
  /* Box to the toolbar's .btn-sm height (~26px) so it doesn't bulge above the
     button row it sits in: 22 + 1px padding ×2 + 1px border ×2 = 26. */
  padding: 1px;
}
.rm-view-toggle > button {
  height: 22px;
}
</style>
