<template>
  <nav class="breadcrumbs" aria-label="Breadcrumb">
    <template v-for="(crumb, i) in crumbs" :key="i">
      <span v-if="i > 0" class="breadcrumb-sep">/</span>
      <NuxtLink v-if="crumb.to" :to="crumb.to" class="breadcrumb-link">{{ crumb.label }}</NuxtLink>
      <span v-else class="breadcrumb-current">{{ crumb.label }}</span>
    </template>
  </nav>
</template>

<script setup lang="ts">
defineProps<{
  crumbs: { label: string; to?: string }[]
}>()
</script>

<style scoped>
.breadcrumbs {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.8125rem;
  margin-bottom: 0.5rem;
}
.breadcrumb-sep {
  color: var(--color-text-muted);
  opacity: 0.5;
}
.breadcrumb-link {
  color: var(--color-text-muted);
  text-decoration: none;
}
.breadcrumb-link:hover {
  color: var(--color-primary);
  text-decoration: underline;
}
.breadcrumb-current {
  color: var(--color-text);
  font-weight: 500;
}
</style>
