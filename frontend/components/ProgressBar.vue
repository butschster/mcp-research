<template>
  <div class="progress-bar">
    <div class="progress-bar-fill" :style="{ width: pct + '%' }"></div>
  </div>
  <slot />
</template>

<script setup lang="ts">
const props = defineProps<{ value: number; total: number }>()
const pct = computed(() => props.total > 0 ? Math.round((props.value / props.total) * 100) : 0)
</script>
