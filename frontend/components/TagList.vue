<template>
  <div v-if="tags?.length" class="tag-list">
    <span
      v-for="tag in tags"
      :key="tag"
      :class="['tag', `tag-hue-${tagHue(tag)}`, { 'tag-clickable': clickable, 'tag-active': activeTag === tag }]"
      @click="clickable ? $emit('tagClick', tag) : undefined"
    >{{ tag }}<span v-if="clickable && counts?.[tag] && counts[tag] > 1" class="tag-count">{{ counts[tag] }}</span></span>
  </div>
</template>

<script setup lang="ts">
import { tagHue } from '~/composables/useTagHue'

defineProps<{
  tags: string[]
  clickable?: boolean
  activeTag?: string
  counts?: Record<string, number>
}>()

defineEmits<{ tagClick: [tag: string] }>()
</script>

<style scoped>
.tag-list { display: flex; gap: var(--space-2); flex-wrap: wrap; }
.tag-active { background: var(--color-primary-muted); color: var(--color-primary); }
.tag-clickable { cursor: pointer; transition: all var(--transition-fast); }
.tag-clickable:hover { background: var(--color-primary-muted); color: var(--color-primary); }
.tag-count { font-size: 0.75em; opacity: 0.7; margin-left: 0.15em; }
</style>
