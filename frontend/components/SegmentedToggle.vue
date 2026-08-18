<template>
  <div class="segmented seg-toggle" role="group" :aria-label="label">
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
/**
 * A pressed-button segmented control over the `.segmented` primitive — for
 * switching a modelValue among a few simple options (a view, a zoom unit). It is
 * deliberately NOT a TabBar: TabBar owns tab panels with a roving tabindex; this
 * flips a value with `aria-pressed` and plain Tab. The roadmap view and zoom
 * toggles are thin typed wrappers around it, so the box (and its 26px toolbar
 * boxing) lives in exactly one place.
 */
defineProps<{
  modelValue: string
  options: { value: string; label: string }[]
  label: string
}>()
const emit = defineEmits<{ 'update:modelValue': [string] }>()
</script>

<style scoped>
.seg-toggle {
  flex-shrink: 0;
  /* Box to the toolbar's .btn-sm height (~26px) so it doesn't bulge above the
     button row it sits in: 22 + 1px padding ×2 + 1px border ×2 = 26. */
  padding: 1px;
}
.seg-toggle > button {
  height: 22px;
}
</style>
