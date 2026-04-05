<template>
  <NuxtLink :to="`/research/${research.id}`" class="card research-card">
    <div class="flex-between">
      <h3 class="card-title">{{ research.name }}</h3>
      <StatusBadge :status="research.status" />
    </div>

    <p v-if="research.goal" class="card-meta" style="margin-top:0.375rem;">
      {{ research.goal }}
    </p>

    <div class="card-footer">
      <div v-if="research.tags?.length" class="tags-row">
        <span
          v-for="tag in research.tags"
          :key="tag"
          :class="['tag', 'tag-clickable', `tag-hue-${tagHue(tag)}`]"
          @click.prevent.stop="emit('tagClick', tag)"
        >{{ tag }}</span>
      </div>
      <span v-if="research.updated_at" class="card-meta timestamp">
        {{ relativeTime(research.updated_at) }}
      </span>
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
    updated_at?: string
  }
}>()

const emit = defineEmits<{ tagClick: [tag: string] }>()

function tagHue(tag: string): number {
  return [...tag].reduce((acc, c) => acc + c.charCodeAt(0), 0) % 6
}

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  const days = Math.floor(hrs / 24)
  return `${days}d ago`
}
</script>

<style scoped>
.research-card { display: block; text-decoration: none; color: inherit; }
.flex-between  { display: flex; justify-content: space-between; align-items: flex-start; }
.card-footer   { display: flex; justify-content: space-between; align-items: center; margin-top: 0.75rem; gap: 0.5rem; }
.tags-row      { display: flex; gap: 0.375rem; flex-wrap: wrap; }
.timestamp     { font-size: 0.75rem; white-space: nowrap; flex-shrink: 0; }
.tag-clickable { cursor: pointer; transition: all 0.15s; }
.tag-clickable:hover { background: rgba(56,189,248,0.15); color: var(--color-primary); }
</style>
