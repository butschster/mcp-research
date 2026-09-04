<template>
  <div v-if="tags?.length" class="tag-list">
    <!-- A clickable tag is the primary filter on two surfaces, and it was a
         `<span @click>`: no tab stop, no role, nothing announced, and no way to
         reach it at all without a mouse. When it does nothing it stays a span,
         because an inert button is its own kind of lie. -->
    <!-- `title` carries the full text, because the chip truncates: a tag is
         agent-written and `kubernetes-admission-controller-webhook-v2` is a
         real one. Before the cap a chip that long widened its card and, in a
         one-row toolbar, pushed every control after it onto the next line. -->
    <component
      :is="clickable ? 'button' : 'span'"
      v-for="tag in tags"
      :key="tag"
      :type="clickable ? 'button' : undefined"
      :aria-pressed="clickable ? String(activeTag === tag) : undefined"
      :title="tag"
      :class="['tag', `tag-hue-${tagHue(tag)}`, { 'tag-clickable': clickable, 'tag-active': activeTag === tag }]"
      @click="clickable ? $emit('tagClick', tag) : undefined"
    ><span class="tag-text">{{ tag }}</span><span v-if="clickable && counts?.[tag] && counts[tag] > 1" class="tag-count">{{ counts[tag] }}</span></component>
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
/* The ellipsis itself is `.tag-text`, global; the cap is the shared token. */
.tag-list > .tag { max-width: var(--tag-max); }
.tag-active { background: var(--color-primary-muted); color: var(--color-primary); }
.tag-clickable { cursor: pointer; transition: all var(--transition-fast); }
.tag-clickable:hover { background: var(--color-primary-muted); color: var(--color-primary); }
.tag-count { font-size: 0.75em; opacity: 0.7; margin-left: 0.15em; flex: none; }
</style>
