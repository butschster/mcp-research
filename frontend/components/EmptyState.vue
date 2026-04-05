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
.empty-state { text-align: center; padding: 2.5rem 1rem; color: var(--color-text-muted); }
.empty-icon  { font-size: 2rem; margin-bottom: 0.75rem; }
.empty-title { font-size: 0.9375rem; font-weight: 500; margin-bottom: 0.25rem; color: var(--color-text); }
.empty-desc  { font-size: 0.8125rem; }
</style>
