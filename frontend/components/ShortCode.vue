<template>
  <span class="short-code">{{ code }}</span>
</template>

<script setup lang="ts">
defineProps<{ code: string }>()
</script>

<style scoped>
.short-code {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
</style>
