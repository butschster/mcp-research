<script setup lang="ts">
import { computed } from 'vue'
import { useTheme } from '~/composables/useTheme'

const { theme, toggleTheme } = useTheme()
const label = computed(() => `Switch to ${theme.value === 'light' ? 'dark' : 'light'} theme`)
</script>

<template>
  <button type="button" class="btn-icon theme-toggle" :aria-label="label" :title="label" @click="toggleTheme">
    <svg v-if="theme === 'light'" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M20.9 13.1A9 9 0 0 1 10.9 3.1 9 9 0 1 0 20.9 13.1Z" />
    </svg>
    <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" aria-hidden="true">
      <circle cx="12" cy="12" r="4" /><path d="M12 2v2m0 16v2M2 12h2m16 0h2M4.9 4.9l1.4 1.4m11.4 11.4 1.4 1.4M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
    </svg>
  </button>
</template>

<style scoped>
.theme-toggle { flex-shrink: 0; }
</style>
