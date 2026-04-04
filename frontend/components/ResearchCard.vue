<template>
  <NuxtLink :to="`/research/${research.id}`" class="card" style="display: block; text-decoration: none; color: inherit;">
    <div style="display: flex; justify-content: space-between; align-items: start;">
      <h3 class="card-title">{{ research.name }}</h3>
      <StatusBadge :status="research.status" />
    </div>
    <p v-if="research.goal" class="card-meta" style="margin-top: 0.5rem;">{{ research.goal }}</p>
    <div v-if="research.tags?.length" style="margin-top: 0.75rem; display: flex; gap: 0.375rem; flex-wrap: wrap;">
      <span
        v-for="tag in research.tags"
        :key="tag"
        class="tag tag-clickable"
        @click.prevent.stop="emit('tagClick', tag)"
      >{{ tag }}</span>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
defineProps<{
  research: {
    id: string
    name: string
    goal: string
    status: string
    tags: string[]
  }
}>()

const emit = defineEmits<{
  tagClick: [tag: string]
}>()
</script>

<style scoped>
.tag-clickable {
  cursor: pointer;
  transition: all 0.15s;
}
.tag-clickable:hover {
  background: rgba(56, 189, 248, 0.15);
  color: var(--color-primary);
}
</style>
