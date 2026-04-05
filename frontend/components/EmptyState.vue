<template>
  <div class="empty-state">
    <div v-if="icon" class="empty-icon">{{ icon }}</div>
    <p class="empty-title">{{ title }}</p>
    <p v-if="description" class="empty-desc card-meta">{{ description }}</p>
    <div v-if="command" class="empty-command">
      <code class="command-text">{{ command }}</code>
      <button class="copy-btn" :class="{ copied }" @click="copy">
        {{ copied ? '&#x2713; Copied' : 'Copy' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  icon?: string
  title: string
  description?: string
  command?: string
}>()

const copied = ref(false)
async function copy() {
  if (!props.command) return
  await navigator.clipboard.writeText(props.command)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}
</script>

<style scoped>
.empty-icon  { font-size: 2.5rem; margin-bottom: var(--space-5); line-height: 1; opacity: 0.7; }
.empty-title { font-size: var(--type-lg); font-weight: 600; margin-bottom: var(--space-2); color: var(--color-text); line-height: var(--line-tight); letter-spacing: -0.01em; }
.empty-desc  { font-size: var(--type-sm); max-width: 40ch; margin-left: auto; margin-right: auto; }
</style>
