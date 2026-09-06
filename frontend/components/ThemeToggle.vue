<script setup lang="ts">
import { computed } from 'vue'
import { useTheme } from '~/composables/useTheme'

/**
 * `sm` is for a strip that cannot afford a 36px control — the share banner is
 * a sticky 36px band, and on a phone `.btn-icon` grows to 44px, which would
 * make the band taller than the content it labels. The visual box drops to
 * `--control-h-sm`; the hit box stays 44px tall through a `::before` overlay,
 * so the target is not what got smaller.
 */
withDefaults(defineProps<{ size?: 'default' | 'sm' }>(), { size: 'default' })

const { theme, toggleTheme } = useTheme()
const label = computed(() => `Switch to ${theme.value === 'light' ? 'dark' : 'light'} theme`)
</script>

<template>
  <button
    type="button"
    class="btn-icon theme-toggle"
    :class="{ 'theme-toggle--sm': size === 'sm' }"
    :aria-label="label"
    :title="label"
    @click="toggleTheme"
  >
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
.theme-toggle--sm {
  position: relative;
  min-height: var(--control-h-sm);
  min-width: var(--control-h-sm);
}
/* 32px visual, 44px target: the overlay reaches 6px past every edge. */
.theme-toggle--sm::before {
  content: '';
  position: absolute;
  inset: -6px;
}
@media (max-width: 768px) {
  /* system.css grows every .btn-icon to --control-h-touch here; the small
     variant keeps its box and relies on the overlay for the touch target. */
  .theme-toggle.theme-toggle--sm {
    min-height: var(--control-h-sm);
    min-width: var(--control-h-sm);
  }
}
</style>
