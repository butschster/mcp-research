<template>
  <div class="empty-state">
    <div v-if="icon" class="empty-icon">{{ icon }}</div>
    <p class="empty-title">{{ title }}</p>
    <p v-if="description" class="empty-desc card-meta">{{ description }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  icon?: string
  title: string
  description?: string
}>()
</script>

<style scoped>
.empty-state { text-align: center; padding: 2.5rem 1rem; color: var(--color-text-muted); }
.empty-icon  { font-size: 2rem; margin-bottom: 0.75rem; }
.empty-title { font-size: 0.9375rem; font-weight: 500; margin-bottom: 0.25rem; color: var(--color-text); }
.empty-desc  { font-size: 0.8125rem; }
</style>
