<template>
  <div ref="rootRef" class="action-menu">
    <button
      class="btn btn-icon"
      :title="title"
      aria-haspopup="menu"
      :aria-expanded="open ? 'true' : 'false'"
      @click.stop="open = !open"
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/></svg>
    </button>

    <Transition name="action-menu">
      <div
        v-if="open"
        class="action-menu-list"
        :class="`action-menu-list--${props.align}`"
        role="menu"
        @click="open = false"
      >
        <slot />
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    /** Tooltip on the trigger button. */
    title?: string
    /** Which edge the panel is anchored to. */
    align?: 'left' | 'right'
  }>(),
  { title: 'More actions', align: 'right' }
)

const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)

function onClickOutside(e: MouseEvent) {
  if (rootRef.value && !rootRef.value.contains(e.target as Node)) {
    open.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && open.value) {
    open.value = false
    // Return focus to the trigger so keyboard users are not dropped at the top.
    rootRef.value?.querySelector('button')?.focus()
  }
}

onMounted(() => {
  document.addEventListener('click', onClickOutside)
  document.addEventListener('keydown', onKeydown)
})
onUnmounted(() => {
  document.removeEventListener('click', onClickOutside)
  document.removeEventListener('keydown', onKeydown)
})

defineExpose({ close: () => (open.value = false) })
</script>

<style scoped>
.action-menu { position: relative; }

.action-menu-list {
  position: absolute;
  top: calc(100% + var(--space-1));
  min-width: 180px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-2);
  /* The header this sits in is raised to --z-in-page, which is below the sticky
     nav on purpose. --z-overlay keeps the panel above in-page content without
     escaping that header. */
  z-index: var(--z-overlay);
  padding: var(--space-1) 0;
}
.action-menu-list--right { right: 0; }
.action-menu-list--left { left: 0; }

/* Slotted content belongs to the parent's scope, so it needs :slotted(). */
.action-menu-list :slotted(.action-menu-item) {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  width: 100%;
  padding: 0.45rem 0.75rem;
  border: none;
  background: none;
  color: var(--color-text-muted);
  font-size: var(--type-sm);
  text-align: left;
  cursor: pointer;
  text-decoration: none;
  white-space: nowrap;
  transition: background var(--transition-fast), color var(--transition-fast);
}
.action-menu-list :slotted(.action-menu-item:hover) {
  background: var(--color-surface-hover);
  color: var(--color-text);
  text-decoration: none;
}
.action-menu-list :slotted(.action-menu-item:disabled) {
  opacity: 0.5;
  cursor: not-allowed;
}
.action-menu-list :slotted(.action-menu-item--danger) {
  color: var(--color-error);
}
.action-menu-list :slotted(.action-menu-divider) {
  height: 1px;
  margin: var(--space-1) 0;
  background: var(--color-border);
}

.action-menu-enter-active,
.action-menu-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.action-menu-enter-from,
.action-menu-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
