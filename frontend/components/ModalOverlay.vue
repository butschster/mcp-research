<template>
  <Teleport to="body">
    <div v-if="visible" class="modal-overlay" @click.self="$emit('close')">
      <div :class="['modal-card', sizeClass]">
        <slot />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
const props = defineProps<{
  visible: boolean
  size?: 'sm' | 'md' | 'lg'
}>()

defineEmits<{ close: [] }>()

const sizeClass = computed(() => {
  if (props.size === 'lg') return 'modal-lg'
  if (props.size === 'sm') return 'modal-sm'
  return ''
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-overlay);
}
.modal-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  width: 100%;
  max-width: 460px;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.3);
}
.modal-lg {
  max-width: 720px;
  max-height: 85vh;
  overflow-y: auto;
}
.modal-sm { max-width: 360px; }
</style>
