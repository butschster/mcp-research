<template>
  <div class="mindmap-node entry-node" @click="navigate">
    <div class="entry-header">
      <div class="entry-title-row">
        <span v-if="data.entrySlug" class="mm-code">{{ data.entrySlug }}</span>
        <span class="entry-title">{{ truncate(data.title, 40) }}</span>
      </div>
      <StatusBadge :status="data.status" />
    </div>
    <p v-if="data.description" class="entry-desc">{{ truncate(data.description, 80) }}</p>
    <div v-if="data.tags?.length" class="entry-tags">
      <span v-for="tag in visibleTags" :key="tag" :class="['mm-tag', `tag-hue-${tagHue(tag)}`]">{{ tag }}</span>
      <span v-if="extraTagCount > 0" class="mm-tag mm-tag-more">+{{ extraTagCount }}</span>
    </div>
    <Handle type="target" :position="targetPosition" />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{
  data: {
    title: string
    description: string
    status: string
    tags: string[]
    createdAt: string
    researchSlug: string
    entrySlug: string
  }
  targetPosition?: Position
}>()

const visibleTags = computed(() => (props.data.tags ?? []).slice(0, 3))
const extraTagCount = computed(() => Math.max(0, (props.data.tags ?? []).length - 3))

function truncate(text: string, len: number): string {
  return text.length > len ? text.slice(0, len) + '...' : text
}

function tagHue(tag: string): number {
  return [...tag].reduce((acc, c) => acc + c.charCodeAt(0), 0) % 6
}

function navigate() {
  navigateTo(`/research/${props.data.researchSlug}/entry/${props.data.entrySlug}`)
}
</script>

<style scoped>
.entry-node {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--space-4) var(--space-5);
  min-width: 340px;
  max-width: 420px;
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
.entry-node:hover {
  border-color: var(--color-border-strong);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}
.entry-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
  margin-bottom: var(--space-1);
}
.entry-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}
.entry-title {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-text);
  line-height: 1.3;
}
.entry-desc {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  line-height: 1.35;
  margin-bottom: var(--space-2);
}
.entry-tags {
  display: flex;
  gap: 0.25rem;
  flex-wrap: wrap;
}
.mm-tag {
  font-size: 0.625rem;
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
  line-height: 1;
}
.mm-tag-more { opacity: 0.6; }
.mm-code {
  font-size: 0.625rem;
  font-weight: 700;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.1rem 0.3rem;
  border-radius: 3px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
</style>
