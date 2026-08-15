<template>
  <div class="modal-header">
    <h3 :id="titleId" class="modal-title">
      <slot>{{ title }}</slot>
    </h3>
    <button class="modal-close" type="button" aria-label="Close" @click="$emit('close')">
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <line x1="18" y1="6" x2="6" y2="18" />
        <line x1="6" y1="6" x2="18" y2="18" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
/**
 * The bar across the top of a dialog: its name, and the way out.
 *
 * Five components carried this — `CreateTaskModal`, `StatusChangeModal`,
 * `TaskDetailModal`, `DetailsPanel` and `RoadmapNodePopover` — each with its
 * own copy of the same sixteen-line close icon, and each having declared
 * `.modal-header` and `.modal-close` again in its own scoped block. All five
 * close buttons also announced as "button" until recently, which is the shape
 * of a defect that has to be fixed five times when the markup exists five
 * times.
 *
 * `ModalOverlay` owns the keyboard contract — the focus trap, Escape, focus
 * restore. This owns the visual one. The `titleId` is passed through rather
 * than generated here because it is the overlay's `aria-labelledby` that has to
 * point at it, and the dialog is the thing that knows its own name.
 */
defineProps<{
  /** The heading. Use the default slot instead when it needs markup. */
  title?: string
  /** The id `ModalOverlay`'s `labelledby` points at. */
  titleId?: string
}>()

defineEmits<{ close: [] }>()
</script>

<style scoped>
/* `.modal-header`, `.modal-title` and `.modal-close` stay global: the dialog
   vocabulary is shared, and the file that owns it records why — nine
   `.modal-title`s in two disagreeing designs was the state before it. */
</style>
