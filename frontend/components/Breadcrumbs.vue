<template>
  <div class="breadcrumbs" role="navigation" aria-label="Breadcrumb">
    <template v-for="(crumb, i) in crumbs" :key="i">
      <span v-if="i > 0" class="breadcrumb-sep">/</span>
      <NuxtLink v-if="crumb.to" :to="crumb.to" class="breadcrumb-link">{{ crumb.label }}</NuxtLink>
      <span v-else class="breadcrumb-current">{{ crumb.label }}</span>
    </template>
  </div>
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
  flex-wrap: wrap;
  gap: 0.375rem;
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  margin-bottom: var(--space-3);
}
.breadcrumb-sep {
  color: var(--color-text-muted);
  opacity: 0.3;
  user-select: none;
  font-size: var(--type-xs);
}
.breadcrumb-link {
  color: var(--color-text-muted);
  text-decoration: none;
  transition: color var(--transition-fast);
}
.breadcrumb-link:hover {
  color: var(--color-primary);
  text-decoration: none;
}
.breadcrumb-current {
  color: var(--color-text);
  font-weight: 500;
}
</style>
