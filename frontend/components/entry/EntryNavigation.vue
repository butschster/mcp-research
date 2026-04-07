<template>
  <div v-if="prev || next" class="entry-nav no-print">
    <NuxtLink v-if="prev" :to="`/research/${researchSlug}/entry/${prev.code || prev.id}`" class="btn btn-sm entry-nav-btn">
      &larr; {{ prev.title }}
    </NuxtLink>
    <span v-else class="entry-nav-placeholder"></span>
    <NuxtLink v-if="next" :to="`/research/${researchSlug}/entry/${next.code || next.id}`" class="btn btn-sm entry-nav-btn entry-nav-next">
      {{ next.title }} &rarr;
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  prev?: { code?: string; id: string; title: string }
  next?: { code?: string; id: string; title: string }
  researchSlug: string
}>()
</script>

<style scoped>
.entry-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--space-8);
  padding-top: var(--space-6);
  border-top: 1px solid var(--color-border);
  gap: var(--space-4);
}
.entry-nav-btn {
  max-width: 45%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.entry-nav-next { margin-left: auto; text-align: right; }
.entry-nav-placeholder { flex: 1; }
</style>
