<template>
  <NuxtLink :to="`/research/${researchSlug}/entry/${entry.code || entry.id}`" class="card entry-card">
    <div class="entry-card-header">
      <div class="entry-title-row">
        <ShortCode v-if="entry.code" :code="entry.code" />
        <h3 class="card-title">{{ entry.title }}</h3>
      </div>
      <StatusBadge :status="entry.status" />
    </div>
    <p v-if="entry.description" class="card-meta mt-2" v-html="renderRefs(entry.description, researchSlug)"></p>
    <TagList v-if="entry.tags?.length" :tags="entry.tags" class="mt-3" />
  </NuxtLink>
</template>

<script setup lang="ts">
defineProps<{
  entry: {
    id: string
    code?: string
    title: string
    status: string
    description?: string
    tags?: string[]
  }
  researchSlug: string
}>()
</script>

<style scoped>
.entry-card { display: block; text-decoration: none; color: inherit; }
.entry-card-header { display: flex; justify-content: space-between; align-items: flex-start; gap: var(--space-2); }
.entry-title-row { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }
</style>
