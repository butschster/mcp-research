<template>
  <ModalOverlay :visible="visible" size="sm" :labelledby="titleId" @close="$emit('cancel')">
    <div class="confirm-modal">
      <div class="confirm-icon" :class="variant">
        <svg v-if="variant === 'danger'" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
        <svg v-else width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/></svg>
      </div>
      <h3 :id="titleId" class="confirm-title">{{ title }}</h3>
      <p class="confirm-message">{{ message }}</p>
      <div class="confirm-actions">
        <button class="btn btn-sm" @click="$emit('cancel')">Cancel</button>
        <button
          :class="['btn btn-sm', variant === 'danger' ? 'btn-danger btn-danger--solid' : 'btn-primary']"
          :disabled="loading"
          @click="$emit('confirm')"
        >
          {{ loading ? 'Wait...' : confirmLabel }}
        </button>
      </div>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
defineProps<{
  visible: boolean
  title: string
  message: string
  confirmLabel?: string
  variant?: 'danger' | 'info'
  loading?: boolean
}>()

defineEmits<{
  confirm: []
  cancel: []
}>()

// Without this every confirmation — including the four destructive ones in
// team management — announces itself as a bare "dialog".
const titleId = `confirm-title-${useId()}`
</script>

<style scoped>
.confirm-modal { text-align: center; }
.confirm-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-bottom: var(--space-4);
}
.confirm-icon.danger {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
}
.confirm-icon.info {
  background: var(--color-primary-muted);
  color: var(--color-primary);
}
.confirm-title {
  font-size: var(--type-base);
  font-weight: var(--weight-semibold);
  margin: 0 0 var(--space-2);
}
.confirm-message {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  margin: 0 0 var(--space-6);
  line-height: 1.5;
}
.confirm-actions {
  display: flex;
  gap: var(--space-3);
  justify-content: center;
}
</style>
